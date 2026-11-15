package analyzer

import (
	"bufio"
	"context"
	"log"
	"loganalyzer/v2/internal/config"
	"loganalyzer/v2/internal/parser"
	"loganalyzer/v2/internal/reader"
	"os"
	sync "sync"
	"time"
)

func AnalyzeFiles(paths []string, filters []Filter) AnalysisStats {
	analysisStats := NewStats()
	startTime := time.Now()

	if filters == nil || len(filters) == 0 {
		filters = []Filter{KeepAll}
	}

	for _, path := range paths {
		// Try to open the file
		file, fileErr := os.Open(path)
		if fileErr != nil {
			log.Printf("failed to read file %s: %s\n", path, fileErr)
			continue
		}

		// Create scanner to read the file line by line
		scanner := bufio.NewScanner(file)

		// Increase the max size of the scanner buffer to read very long log lines
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1 MB buffer

		// Iterate on each line of the file
		for scanner.Scan() {
			line := scanner.Text()
			analysisStats.TotalLines++

			// Malformed and filtered lines are skipped from time buckets
			logEntry, parseErr := parser.ParseLine(line)
			if parseErr != nil {
				analysisStats.MalformedLines++

				// shortcut: per-line diagnostics dropped for throughput; reintroduce behind --verbose if needed
				//log.Printf("failed to parse line %d of file %s\n", analysisStats.TotalLines, parseErr)
			} else if applyFilters(logEntry, filters) {
				analysisStats.ParsedLines++
				analysisStats.LevelCounts[logEntry.Level.String()]++
				analysisStats.ComponentCounts[logEntry.Component]++

				if (logEntry.Level == parser.LevelError) || (logEntry.Level == parser.LevelFatal) {
					analysisStats.ErrorCounts[logEntry.Message]++
				}

				timestamp := logEntry.Timestamp.Truncate(time.Minute).Format("2006-01-02 15:04")
				analysisStats.TimeBuckets[timestamp]++
			} else {
				// Parsed correctly but excluded by filter
				analysisStats.ParsedLines++
				analysisStats.FilteredLines++
			}
		}

		// Check errors occurred during the scan
		if scanErr := scanner.Err(); scanErr != nil {
			log.Printf("error during reading file[%s]: %s\n", path, scanErr)
		}

		_ = file.Close()
	}

	processingTime := time.Since(startTime)
	analysisStats.ProcessingTime = processingTime

	return analysisStats
}

func AnalyzeConcurrently(ctx context.Context, flagCtx *config.Context, paths []string, filters []Filter) AnalysisStats {
	if filters == nil || len(filters) == 0 {
		filters = []Filter{KeepAll}
	}

	// Read all files in a separate goroutine
	jobs := make(chan parser.RawLine, flagCtx.Buffer)
	go reader.ReadFiles(ctx, paths, jobs)

	// Run the workers
	results := make(chan parser.ParseResult, flagCtx.Buffer)
	var wg sync.WaitGroup
	wg.Add(flagCtx.Workers)
	for i := 0; i < flagCtx.Workers; i++ {
		go func() {
			defer wg.Done()
			Worker(ctx, jobs, results, filters)
		}()
	}

	// Close results channel when all workers exit
	go func() {
		wg.Wait()
		close(results)
	}()

	return AggregateResults(ctx, results)
}
