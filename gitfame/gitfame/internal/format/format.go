package format

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"gitlab.com/slon/shad-go/gitfame/internal/statistics"
)

type StatsFormatter func(io.Writer, []*statistics.Stat) error

func FormatTabular(outputStream io.Writer, aggregatedStats []*statistics.Stat) error {
	writer := tabwriter.NewWriter(outputStream, 0, 0, 1, ' ', 0)
	header := "Name\tLines\tCommits\tFiles\n"
	if _, err := fmt.Fprint(writer, header); err != nil {
		return fmt.Errorf("failed to write tabular header: %w", err)
	}

	for _, stat := range aggregatedStats {
		line := fmt.Sprintf("%s\t%d\t%d\t%d\n", stat.Name, stat.Lines, stat.Commits, stat.Files)
		if _, err := fmt.Fprint(writer, line); err != nil {
			return fmt.Errorf("failed to write tabular data for %q: %w", stat.Name, err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush tabular writer: %w", err)
	}
	return nil
}

func FormatCSV(targetWriter io.Writer, statsData []*statistics.Stat) error {
	csvWriter := csv.NewWriter(targetWriter)
	headerRow := []string{"Name", "Lines", "Commits", "Files"}
	if err := csvWriter.Write(headerRow); err != nil {
		return fmt.Errorf("csv write header failed: %w", err)
	}

	for _, statEntry := range statsData {
		dataRow := []string{statEntry.Name, fmt.Sprint(statEntry.Lines), fmt.Sprint(statEntry.Commits), fmt.Sprint(statEntry.Files)}
		if err := csvWriter.Write(dataRow); err != nil {
			return fmt.Errorf("csv write row for %q failed: %w", statEntry.Name, err)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("csv flush failed: %w", err)
	}
	return nil
}

func FormatJSON(resultWriter io.Writer, statsList []*statistics.Stat) error {
	encoder := json.NewEncoder(resultWriter)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(statsList); err != nil {
		return fmt.Errorf("json encoding failed: %w", err)
	}
	return nil
}

func FormatJSONLines(output io.Writer, individualStats []*statistics.Stat) error {
	jsonEncoder := json.NewEncoder(output)
	for _, singleStat := range individualStats {
		if err := jsonEncoder.Encode(singleStat); err != nil {
			return fmt.Errorf("json line encoding for %q failed: %w", singleStat.Name, err)
		}
	}
	return nil
}
