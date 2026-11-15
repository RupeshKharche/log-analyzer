package reporter

import (
	"bytes"
	"strings"
	"testing"

	"loganalyzer/v2/internal/analyzer"
)

func TestWriteText_ContainsSummary(t *testing.T) {
	stats := analyzer.NewStats()
	stats.TotalLines = 10
	stats.ParsedLines = 8
	stats.MalformedLines = 2
	stats.FilteredLines = 1
	stats.LevelCounts["INFO"] = 5
	stats.LevelCounts["ERROR"] = 3
	stats.ErrorCounts["Connection timeout"] = 2
	stats.ProcessingTime = 0 // avoid div-by-zero in throughput (guarded)

	var buf bytes.Buffer
	WriteText(&buf, []string{"a.log", "b.log"}, stats, 4)
	out := buf.String()

	// tabwriter expands tabs to spaces, so match labels loosely.
	checks := []string{
		"Log Analysis Report",
		"Files analyzed:",
		"Total lines:",
		"Parsed lines:",
		"Malformed lines:",
		"Filtered lines:",
		"INFO:",
		"ERROR:",
		"Connection timeout",
		"Workers:",
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("output missing %q\n--- output ---\n%s", c, out)
		}
	}
	// verify numeric values survived tabwriter expansion
	for _, want := range []string{"2", "10", "8", "1", "5", "3", "4"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing value %q\n--- output ---\n%s", want, out)
		}
	}
}

func TestWriteText_NoTopErrors(t *testing.T) {
	stats := analyzer.NewStats()
	stats.TotalLines = 1
	stats.ParsedLines = 1
	stats.LevelCounts["INFO"] = 1

	var buf bytes.Buffer
	WriteText(&buf, []string{"x.log"}, stats, 1)
	out := buf.String()
	if !strings.Contains(out, "Top Errors:") {
		t.Errorf("missing Top Errors header\n%s", out)
	}
}
