package statistics

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"gitlab.com/slon/shad-go/gitfame/internal/git"
	"gitlab.com/slon/shad-go/gitfame/internal/reader"
)

type Stat struct {
	Name      string `json:"name"`
	Lines     int    `json:"lines"`
	Files     int    `json:"files"`
	Commits   int    `json:"commits"`
	commitSet map[string]struct{}
}

type commitData struct {
	Hash      string
	Author    string
	Committer string
}

func FromRepo(
	repoRoot,
	revision string,
	targetPaths []string,
	trackCommitter bool,
	concurrencyLevel int,
	progressCallback func(),
) (map[string]*Stat, error) {
	filePathChan := make(chan string, concurrencyLevel)

	type processingResult struct {
		authorStats map[string]*Stat
		err         error
	}
	resultChan := make(chan processingResult, concurrencyLevel)

	var workerWaitGroup sync.WaitGroup
	workerWaitGroup.Add(concurrencyLevel)

	for range concurrencyLevel {
		go func() {
			defer workerWaitGroup.Done()
			for filePath := range filePathChan {
				fileStats, err := collectFileStats(repoRoot, filePath, revision, trackCommitter)
				resultChan <- processingResult{fileStats, err}
			}
		}()
	}

	go func() {
		for _, filePath := range targetPaths {
			filePathChan <- filePath
		}
		close(filePathChan)
		workerWaitGroup.Wait()
		close(resultChan)
	}()

	aggregatedStats := make(map[string]*Stat)
	for result := range resultChan {
		if result.err != nil {
			fmt.Fprintf(os.Stderr, "Error processing file stats: %v\n", result.err)
			continue
		}

		for author, fileStat := range result.authorStats {
			stat, exists := aggregatedStats[author]
			if !exists {
				stat = &Stat{
					Name:      author,
					commitSet: make(map[string]struct{}),
				}
				aggregatedStats[author] = stat
			}
			stat.Files += fileStat.Files
			stat.Lines += fileStat.Lines
			for commitHash := range fileStat.commitSet {
				stat.commitSet[commitHash] = struct{}{}
			}
		}
		progressCallback()
	}

	for _, stat := range aggregatedStats {
		stat.Commits = len(stat.commitSet)
	}

	return aggregatedStats, nil
}

func collectFileStats(repoRoot, filePath, revision string, trackCommitter bool) (map[string]*Stat, error) {
	blameOutput, err := git.Blame(repoRoot, filePath, revision)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve blame for %q: %w", filePath, err)
	}

	authorStats := make(map[string]*Stat)
	commitCache := make(map[string]*commitData)

	reader := reader.NewWrapper(bytes.NewReader(blameOutput))
	isFirstLine := true

	for {
		line, err := reader.ReadLine()
		if errors.Is(err, io.EOF) {
			if isFirstLine {
				return handleEmptyFile(repoRoot, filePath, revision, trackCommitter)
			}
			break
		}
		isFirstLine = false
		if err != nil {
			return nil, err
		}

		if line[0] != '\t' && strings.Count(line, " ") == 3 {
			commitHash := line[:40]
			linesModified, err := strconv.Atoi(line[strings.LastIndex(line, " ")+1:])
			if err != nil {
				return nil, fmt.Errorf("invalid line modification count in blame output: %w", err)
			}

			commit, ok := commitCache[commitHash]
			if !ok {
				commit, err = parseCommitMetadata(reader)
				if err != nil {
					return nil, fmt.Errorf("commit parsing failed for hash %q: %w", commitHash, err)
				}
				commit.Hash = commitHash
				commitCache[commitHash] = commit
			}

			authorName := commit.Author
			if trackCommitter {
				authorName = commit.Committer
			}

			authorStat, exists := authorStats[authorName]
			if !exists {
				authorStat = &Stat{
					Files:     1,
					commitSet: make(map[string]struct{}),
				}
				authorStats[authorName] = authorStat
			}
			authorStat.Lines += linesModified
			authorStat.commitSet[commit.Hash] = struct{}{}
		}
	}

	return authorStats, nil
}

func handleEmptyFile(repoRoot, filePath, revision string, trackCommitter bool) (map[string]*Stat, error) {
	logOutput, err := git.LogLast(repoRoot, filePath, revision, "format:%H%n%an%n%cn%n")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve log for empty file %q: %w", filePath, err)
	}

	commitInfo := strings.Split(string(logOutput), "\n")
	authorName := commitInfo[1]
	if trackCommitter {
		authorName = commitInfo[2]
	}

	emptyFileStat := &Stat{
		Files:     1,
		Commits:   1,
		commitSet: map[string]struct{}{commitInfo[0]: {}},
	}
	return map[string]*Stat{authorName: emptyFileStat}, nil
}

func parseCommitMetadata(reader reader.LineReader) (*commitData, error) {
	commit := &commitData{}
	for {
		line, err := reader.ReadLine()
		if errors.Is(err, io.EOF) {
			return commit, nil
		}
		if err != nil {
			return nil, fmt.Errorf("reading commit metadata failed: %w", err)
		}

		switch {
		case strings.HasPrefix(line, "author "):
			commit.Author = strings.TrimPrefix(line, "author ")
		case strings.HasPrefix(line, "committer "):
			commit.Committer = strings.TrimPrefix(line, "committer ")
		case strings.HasPrefix(line, "filename"):
			return commit, nil
		}
	}
}
