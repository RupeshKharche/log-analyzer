package analyzer

import (
	"context"
	"loganalyzer/v2/internal/parser"
)

func applyFilters(e parser.LogEntry, filters []Filter) bool {
	for _, filter := range filters {
		if !filter(e) {
			return false
		}
	}
	return true
}

func Worker(ctx context.Context, jobs <-chan parser.RawLine, results chan<- parser.ParseResult, filters []Filter) {
	for {
		select {
		case <-ctx.Done():
			return

		case rl, ok := <-jobs:
			if !ok { // channel closed
				return
			}
			pr := parser.ParseResult{RawLine: rl}
			entry, err := parser.ParseLine(rl.Text)
			if err != nil {
				pr.Valid = false
			} else {
				pr.Entry = entry
				pr.Valid = true
				if !applyFilters(entry, filters) {
					pr.Filtered = true
				}
			}

			select {
			case results <- pr: // Push the parsed line to the results channel
			case <-ctx.Done():
				return
			}
		}
	}
}
