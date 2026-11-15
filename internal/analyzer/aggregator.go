package analyzer

import (
	"context"
	"loganalyzer/v2/internal/parser"
	"time"
)

func AggregateResults(ctx context.Context, results chan parser.ParseResult) AnalysisStats {
	stats := NewStats()
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			stats.ProcessingTime = time.Since(startTime)
			return stats
		case pr, ok := <-results:
			if !ok {
				stats.ProcessingTime = time.Since(startTime)
				return stats // channel closed by waiter goroutine
			}
			stats.TotalLines++
			entry := pr.Entry

			if !pr.Valid {
				stats.MalformedLines++

				// shortcut: per-line diagnostics dropped for throughput; reintroduce behind --verbose if needed
				// log.Printf("failed to parse line %d of file %s\n", stats.TotalLines, pr.RawLine.File)
			} else if !pr.Filtered {
				stats.ParsedLines++
				stats.LevelCounts[entry.Level.String()]++
				stats.ComponentCounts[entry.Component]++

				if (entry.Level == parser.LevelError) || (entry.Level == parser.LevelFatal) {
					stats.ErrorCounts[entry.Message]++
				}

				timestamp := entry.Timestamp.Truncate(time.Minute).Format("2006-01-02 15:04")
				stats.TimeBuckets[timestamp]++
			} else {
				// Parsed correctly but excluded by filter
				stats.ParsedLines++
				stats.FilteredLines++
			}
		}
	}
}
