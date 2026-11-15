package analyzer

import (
	"os"
	"testing"

	"loganalyzer/v2/internal/parser"
)

const sampleLog = `2026-10-20 14:32:10 INFO [auth] User login successful
2026-10-20 14:32:11 ERROR [db] Connection timeout
2026-10-20 14:32:12 WARN [api] rate limited
2026-10-20 14:32:13 DEBUG [auth] token refreshed
2026-10-20 14:32:14 FATAL [db] Database unavailable
2026-10-20 14:32:15 INFO [api] request handled
this is a malformed line
2026-10-20 14:33:00 ERROR [db] Connection timeout
`

func writeTempLog(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "sample-*.log")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestAnalyzeFiles_EmptyFile(t *testing.T) {
	p := writeTempLog(t, "")
	stats := AnalyzeFiles([]string{p}, nil)
	if stats.TotalLines != 0 || stats.ParsedLines != 0 || stats.MalformedLines != 0 {
		t.Errorf("empty file stats = %+v, want zeros", stats)
	}
	if len(stats.LevelCounts) != 0 {
		t.Errorf("LevelCounts not empty: %v", stats.LevelCounts)
	}
}

func TestAnalyzeFiles_Counts(t *testing.T) {
	p := writeTempLog(t, sampleLog)
	stats := AnalyzeFiles([]string{p}, nil)

	if stats.TotalLines != 8 {
		t.Errorf("TotalLines = %d, want 8", stats.TotalLines)
	}
	if stats.MalformedLines != 1 {
		t.Errorf("MalformedLines = %d, want 1", stats.MalformedLines)
	}
	if stats.ParsedLines != 7 {
		t.Errorf("ParsedLines = %d, want 7", stats.ParsedLines)
	}
	if stats.FilteredLines != 0 {
		t.Errorf("FilteredLines = %d, want 0", stats.FilteredLines)
	}
	if stats.LevelCounts["INFO"] != 2 {
		t.Errorf("INFO count = %d, want 2", stats.LevelCounts["INFO"])
	}
	if stats.LevelCounts["ERROR"] != 2 {
		t.Errorf("ERROR count = %d, want 2", stats.LevelCounts["ERROR"])
	}
	if stats.ErrorCounts["Connection timeout"] != 2 {
		t.Errorf("Connection timeout count = %d, want 2", stats.ErrorCounts["Connection timeout"])
	}
	if stats.ComponentCounts["db"] != 3 {
		t.Errorf("db component count = %d, want 3", stats.ComponentCounts["db"])
	}
	if stats.TimeBuckets["2026-10-20 14:32"] != 6 {
		t.Errorf("bucket 14:32 = %d, want 6", stats.TimeBuckets["2026-10-20 14:32"])
	}
	if stats.TimeBuckets["2026-10-20 14:33"] != 1 {
		t.Errorf("bucket 14:33 = %d, want 1", stats.TimeBuckets["2026-10-20 14:33"])
	}
}

func TestAnalyzeFiles_MissingFileContinues(t *testing.T) {
	p := writeTempLog(t, sampleLog)
	stats := AnalyzeFiles([]string{"/no/such/file.log", p}, nil)
	if stats.TotalLines != 8 {
		t.Errorf("TotalLines = %d, want 8 (second file still processed)", stats.TotalLines)
	}
}

func TestAnalyzeFiles_FilteredLines(t *testing.T) {
	p := writeTempLog(t, sampleLog)
	// keep only ERROR lines -> 2 kept, 5 filtered (7 valid - 2)
	filters := []Filter{func(e parser.LogEntry) bool {
		return e.Level == parser.LevelError
	}}
	stats := AnalyzeFiles([]string{p}, filters)
	if stats.ParsedLines != 7 {
		t.Errorf("ParsedLines = %d, want 7 (valid lines)", stats.ParsedLines)
	}
	if stats.FilteredLines != 5 {
		t.Errorf("FilteredLines = %d, want 5", stats.FilteredLines)
	}
	if stats.LevelCounts["ERROR"] != 2 {
		t.Errorf("ERROR count = %d, want 2", stats.LevelCounts["ERROR"])
	}
	if stats.LevelCounts["INFO"] != 0 {
		t.Errorf("INFO count = %d, want 0 (filtered out of level counts)", stats.LevelCounts["INFO"])
	}
}

func TestNewStats_MapsNotNil(t *testing.T) {
	s := NewStats()
	if s.LevelCounts == nil || s.ComponentCounts == nil || s.ErrorCounts == nil || s.TimeBuckets == nil {
		t.Error("NewStats returned nil maps")
	}
}
