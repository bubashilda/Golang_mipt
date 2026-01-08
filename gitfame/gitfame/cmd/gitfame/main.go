//go:build !solution

package main

import (
	"cmp"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"slices"
	"strings"

	"gitlab.com/slon/shad-go/gitfame/internal/filter"
	"gitlab.com/slon/shad-go/gitfame/internal/format"
	"gitlab.com/slon/shad-go/gitfame/internal/git"
	"gitlab.com/slon/shad-go/gitfame/internal/progress"
	"gitlab.com/slon/shad-go/gitfame/internal/statistics"

	flag "github.com/spf13/pflag"
)

var (
	sortingOptions = []string{"lines", "commits", "files"}
	outputFormats  = []string{"tabular", "csv", "json", "json-lines"}
)

var statsFormatters = map[string]format.StatsFormatter{
	"tabular":    format.FormatTabular,
	"csv":        format.FormatCSV,
	"json":       format.FormatJSON,
	"json-lines": format.FormatJSONLines,
}

var (
	repoPath         = flag.String("repository", ".", "path to the Git repository directory to analyze (default: current directory)")
	targetRevision   = flag.String("revision", "HEAD", "Git revision to analyze, such as a commit hash, tag, or branch name (default: HEAD)")
	useCommitterDate = flag.Bool("use-committer", false, "attribute statistics to the committer instead of the author of each commit")
	sortBy           = flag.String("order-by", "lines", fmt.Sprintf("sort results by one of: %s (default: lines)", strings.Join(sortingOptions, ", ")))
	outputFormat     = flag.String("format", "tabular", fmt.Sprintf("output format, one of: %s (default: tabular)", strings.Join(outputFormats, ", ")))
	fileExtensions   = flag.StringSlice("extensions", nil, "filter files by their extensions, e.g., '.go,.py' (comma-separated list)")
	programmingLangs = flag.StringSlice("languages", nil, "filter files by programming languages, e.g., 'go,python' (comma-separated list)")
	excludePatterns  = flag.StringSlice("exclude", nil, "exclude files matching glob patterns, e.g., '*/test/*,*.log' (comma-separated list)")
	restrictPatterns = flag.StringSlice("restrict-to", nil, "restrict analysis to files matching glob patterns, e.g., '*/src/*,*.go' (comma-separated list)")
	concurrencyLevel = flag.Int("workers", runtime.NumCPU(), "number of concurrent workers to process files (default: number of CPU cores)")
	showProgress     = flag.Bool("progress", false, "display a progress bar to track analysis progress")
	cpuProfilePath   = flag.String("cpuprofile", "", "save CPU profiling data to the specified file for performance debugging (hidden)")
)

func main() {
	flag.CommandLine.SortFlags = false
	_ = flag.CommandLine.MarkHidden("cpuprofile")
	flag.Parse()

	if *cpuProfilePath != "" {
		cpuProfileFile, err := os.Create(*cpuProfilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create CPU profiling file: %v\n", err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
			fmt.Fprintf(os.Stderr, "Could not start CPU profiling: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
		fmt.Fprintf(os.Stderr, "CPU profiling enabled, writing to: %s\n", *cpuProfilePath)
	}

	if !slices.Contains(outputFormats, *outputFormat) {
		fmt.Fprintf(os.Stderr, "Error: invalid output format '%s'. Supported formats are: %s.\n",
			*outputFormat, strings.Join(outputFormats, ", "))
		os.Exit(1)
	}
	if !slices.Contains(sortingOptions, *sortBy) {
		fmt.Fprintf(os.Stderr, "Error: invalid sorting option '%s'. Allowed values: %s.\n",
			*sortBy, strings.Join(sortingOptions, ", "))
		os.Exit(1)
	}

	fileFilter, err := filter.New(*fileExtensions, *programmingLangs, *excludePatterns, *restrictPatterns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize file filter: %v\nDetails: unable to compile glob patterns.\n", err)
		os.Exit(1)
	}

	allFiles, err := git.ListFiles(*repoPath, *targetRevision)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Git repository scan failed: could not list files at revision '%s' in '%s': %v\n",
			*targetRevision, *repoPath, err)
		os.Exit(1)
	}

	filteredFiles := fileFilter.Filter(allFiles)
	fmt.Fprintf(os.Stderr, "Processing %d files (out of %d total) after applying filters...\n",
		len(filteredFiles), len(allFiles))

	var (
		incrementProgress func()
		stopProgressBar   func()
	)
	progress := progress.New(len(filteredFiles))
	if *showProgress {
		stopProgressBar = progress.Run()
		incrementProgress = func() { progress.Update(1) }
	} else {
		incrementProgress = func() {}
		stopProgressBar = func() {}
	}

	repoStatistics, err := statistics.FromRepo(
		*repoPath,
		*targetRevision,
		filteredFiles,
		*useCommitterDate,
		*concurrencyLevel,
		incrementProgress,
	)
	stopProgressBar()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Statistics collection failed: %v\nCheck repository path, revision and permissions.\n", err)
		os.Exit(1)
	}

	statsSlice := make([]*statistics.Stat, 0, len(repoStatistics))
	for _, stat := range repoStatistics {
		statsSlice = append(statsSlice, stat)
	}

	switch *sortBy {
	case "lines":
		slices.SortFunc(statsSlice, func(a, b *statistics.Stat) int {
			return cmp.Or(
				-cmp.Compare(a.Lines, b.Lines),
				-cmp.Compare(a.Files, b.Files),
				-cmp.Compare(a.Commits, b.Commits),
				cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)),
			)
		})
	case "commits":
		slices.SortFunc(statsSlice, func(a, b *statistics.Stat) int {
			return cmp.Or(
				-cmp.Compare(a.Commits, b.Commits),
				-cmp.Compare(a.Lines, b.Lines),
				-cmp.Compare(a.Files, b.Files),
				cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)),
			)
		})
	case "files":
		slices.SortFunc(statsSlice, func(a, b *statistics.Stat) int {
			return cmp.Or(
				-cmp.Compare(a.Files, b.Files),
				-cmp.Compare(a.Lines, b.Lines),
				-cmp.Compare(a.Commits, b.Commits),
				cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)),
			)
		})
	}

	if err := statsFormatters[*outputFormat](os.Stdout, statsSlice); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format output as '%s': %v\nCheck if the output format is supported.\n",
			*outputFormat, err)
		os.Exit(1)
	}
}
