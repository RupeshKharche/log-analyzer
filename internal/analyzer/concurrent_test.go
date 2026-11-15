package analyzer

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"loganalyzer/v2/internal/config"
	"loganalyzer/v2/internal/parser"
)

const bigSample = `2026-10-20 14:32:10 INFO [auth] User login successful
2026-10-20 14:32:11 ERROR [db] Connection timeout
2026-10-20 14:32:12 WARN [api] rate limited
2026-10-20 14:32:13 DEBUG [auth] token refreshed
2026-10-20 14:32:14 FATAL [db] Database unavailable
2026-10-20 14:32:15 INFO [api] request handled
this is a malformed line
2026-10-20 14:33:00 ERROR [db] Connection timeout
`

func writeTempLog2(t *testing.T, content string) string {
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

func makeFlagCtx(workers, buffer int) *config.Context {
	return config.NewContext(workers, "ALL", "ALL", "", "", "", "text", "stdout", buffer)
}

// statsEqual compares two AnalysisStats ignoring ProcessingTime.
func statsEqual(a, b *AnalysisStats) bool {
	if a.TotalLines != b.TotalLines || a.ParsedLines != b.ParsedLines ||
		a.MalformedLines != b.MalformedLines || a.FilteredLines != b.FilteredLines {
		return false
	}
	if len(a.LevelCounts) != len(b.LevelCounts) {
		return false
	}
	for k, v := range a.LevelCounts {
		if b.LevelCounts[k] != v {
			return false
		}
	}
	if len(a.ComponentCounts) != len(b.ComponentCounts) {
		return false
	}
	for k, v := range a.ComponentCounts {
		if b.ComponentCounts[k] != v {
			return false
		}
	}
	if len(a.ErrorCounts) != len(b.ErrorCounts) {
		return false
	}
	for k, v := range a.ErrorCounts {
		if b.ErrorCounts[k] != v {
			return false
		}
	}
	if len(a.TimeBuckets) != len(b.TimeBuckets) {
		return false
	}
	for k, v := range a.TimeBuckets {
		if b.TimeBuckets[k] != v {
			return false
		}
	}
	return true
}

func TestConcurrent_MatchesSequential(t *testing.T) {
	p := writeTempLog2(t, bigSample)
	seq := AnalyzeFiles([]string{p}, nil)
	ctx := context.Background()
	conc := AnalyzeConcurrently(ctx, makeFlagCtx(4, 8), []string{p}, nil)
	if !statsEqual(&seq, &conc) {
		t.Errorf("concurrent results differ from sequential\nseq=%+v\nconc=%+v", seq, conc)
	}
}

func TestConcurrent_WorkerCountIndependent(t *testing.T) {
	p := writeTempLog2(t, bigSample)
	var base *AnalysisStats
	for _, w := range []int{1, 2, 4, 8} {
		ctx := context.Background()
		got := AnalyzeConcurrently(ctx, makeFlagCtx(w, w*2), []string{p}, nil)
		if base == nil {
			tmp := got
			base = &tmp
			continue
		}
		if !statsEqual(base, &got) {
			t.Errorf("worker count %d differs from baseline", w)
		}
	}
}

func TestConcurrent_FilterApplied(t *testing.T) {
	p := writeTempLog2(t, bigSample)
	filters := []Filter{func(e parser.LogEntry) bool { return e.Level == parser.LevelError }}
	seq := AnalyzeFiles([]string{p}, filters)
	conc := AnalyzeConcurrently(context.Background(), makeFlagCtx(3, 6), []string{p}, filters)
	if !statsEqual(&seq, &conc) {
		t.Errorf("filtered concurrent results differ\nseq=%+v\nconc=%+v", seq, conc)
	}
}

func TestConcurrent_EmptyInput(t *testing.T) {
	p := writeTempLog2(t, "")
	ctx := context.Background()
	got := AnalyzeConcurrently(ctx, makeFlagCtx(2, 4), []string{p}, nil)
	if got.TotalLines != 0 || got.ParsedLines != 0 || got.MalformedLines != 0 {
		t.Errorf("empty input stats = %+v, want zeros", got)
	}
}

func TestConcurrent_Cancellation(t *testing.T) {
	// Build a large file so cancellation has something to interrupt.
	var sb strings.Builder
	for i := 0; i < 200000; i++ {
		sb.WriteString("2026-10-20 14:32:10 INFO [auth] User login successful\n")
	}
	p := writeTempLog2(t, sb.String())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	done := make(chan struct{})
	go func() {
		AnalyzeConcurrently(ctx, makeFlagCtx(2, 4), []string{p}, nil)
		close(done)
	}()
	select {
	case <-done:
		// returned (possibly after finishing quickly) — acceptable
	case <-time.After(5 * time.Second):
		t.Fatal("AnalyzeConcurrently did not return within 5s after cancellation")
	}
}

func TestConcurrent_MissingFileContinues(t *testing.T) {
	p := writeTempLog2(t, bigSample)
	ctx := context.Background()
	got := AnalyzeConcurrently(ctx, makeFlagCtx(2, 4), []string{"/no/such/file.log", p}, nil)
	if got.TotalLines != 8 {
		t.Errorf("TotalLines = %d, want 8", got.TotalLines)
	}
}
