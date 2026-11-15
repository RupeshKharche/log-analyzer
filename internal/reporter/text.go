package reporter

import (
	"fmt"
	"io"
	"loganalyzer/v2/internal/analyzer"
	"loganalyzer/v2/internal/utils"
	"text/tabwriter"
)

func WriteText(w io.Writer, files []string, stats analyzer.AnalysisStats, workers int) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, "Log Analysis Report")
	fmt.Fprintln(tw, "===================")
	fmt.Fprintln(tw)
	fmt.Fprintf(tw, "Files analyzed:\t%d\n", len(files))
	fmt.Fprintf(tw, "Total lines:\t%d\n", stats.TotalLines)
	fmt.Fprintf(tw, "Parsed lines:\t%d\n", stats.ParsedLines)
	fmt.Fprintf(tw, "Malformed lines:\t%d\n", stats.MalformedLines)
	fmt.Fprintf(tw, "Filtered lines:\t%d\n", stats.FilteredLines)
	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "Log Levels:")
	for _, lvl := range []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"} {
		fmt.Fprintf(tw, "  %s:\t%d\n", lvl, stats.LevelCounts[lvl])
	}
	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "Top Errors:")
	top10Errors := utils.GetTopNEntries(stats.ErrorCounts, 10)
	for i, entry := range top10Errors {
		fmt.Fprintf(tw, "  %2d. %s\t%d\n", i+1, entry.Key, entry.Value)
	}
	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "Time Buckets:")
	top10Buckets := utils.GetTopNEntries(stats.TimeBuckets, 10)
	for i, entry := range top10Buckets {
		fmt.Fprintf(tw, "  %2d. %s\t%d\n", i+1, entry.Key, entry.Value)
	}
	fmt.Fprintln(tw)
	fmt.Fprintf(tw, "Duration:\t%s\n", stats.ProcessingTime)
	secs := stats.ProcessingTime.Seconds()
	if secs > 0 {
		fmt.Fprintf(tw, "Throughput:\t%.0f lines/sec\n", float64(stats.TotalLines)/secs)
	}
	fmt.Fprintf(tw, "Workers:\t%d\n", workers)
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("flush report: %w", err)
	}
	return nil
}
