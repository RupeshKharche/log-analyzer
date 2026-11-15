package parser

import (
	"reflect"
	"testing"
	"time"
)

func TestParseLine_Valid(t *testing.T) {
	ts, _ := time.Parse("2006-01-02 15:04:05", "2026-10-20 14:32:10")
	cases := []struct {
		name string
		line string
		want LogEntry
	}{
		{
			name: "info with message",
			line: "2026-10-20 14:32:10 INFO [auth] User login successful",
			want: LogEntry{Timestamp: ts, Level: LevelInfo, Component: "auth", Message: "User login successful"},
		},
		{
			name: "error message with spaces",
			line: "2026-10-20 14:32:10 ERROR [db] Connection timeout to primary",
			want: LogEntry{Timestamp: ts, Level: LevelError, Component: "db", Message: "Connection timeout to primary"},
		},
		{
			name: "trimmed surrounding whitespace",
			line: "  2026-10-20 14:32:10 WARN [api] rate limited  ",
			want: LogEntry{Timestamp: ts, Level: LevelWarn, Component: "api", Message: "rate limited"},
		},
		{
			name: "empty message allowed",
			line: "2026-10-20 14:32:10 DEBUG [x] ",
			want: LogEntry{Timestamp: ts, Level: LevelDebug, Component: "x", Message: ""},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseLine(c.line)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestParseLine_AllLevels(t *testing.T) {
	ts, _ := time.Parse("2006-01-02 15:04:05", "2026-10-20 14:32:10")
	levels := []LogLevel{LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal}
	names := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	for i, lvl := range levels {
		line := "2026-10-20 14:32:10 " + names[i] + " [c] msg"
		got, err := ParseLine(line)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", names[i], err)
		}
		if got.Level != lvl {
			t.Errorf("%s: got level %v, want %v", names[i], got.Level, lvl)
		}
		if !got.Timestamp.Equal(ts) {
			t.Errorf("%s: timestamp mismatch", names[i])
		}
	}
}

func TestParseLine_Malformed(t *testing.T) {
	cases := []struct {
		name string
		line string
	}{
		{"empty", ""},
		{"too short", "2026-10-20 14:32"},
		{"bad timestamp", "2026-13-40 14:32:10 INFO [a] msg"},
		{"unknown level", "2026-10-20 14:32:10 TRACE [a] msg"},
		{"missing brackets", "2026-10-20 14:32:10 INFO auth msg"},
		{"only open bracket", "2026-10-20 14:32:10 INFO [auth msg"},
		{"missing fields", "2026-10-20 14:32:10 INFO"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseLine(c.line)
			if err == nil {
				t.Errorf("expected error for %q", c.line)
			}
		})
	}
}
