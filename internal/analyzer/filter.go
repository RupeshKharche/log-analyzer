package analyzer

import (
	"loganalyzer/v2/internal/parser"
	"strings"
	"time"
)

// Filter Contract for defining custom filtering logic for each log line
type Filter func(parser.LogEntry) bool

// KeepAll Default filter keeping all log lines
func KeepAll(entry parser.LogEntry) bool {
	return true
}

// LevelFilter allows all entries with loglevel >= the given loglevel
func LevelFilter(min parser.LogLevel) Filter {
	return func(e parser.LogEntry) bool {
		return e.Level >= min
	}
}

// ComponentFilter allows all entries which have matching component
func ComponentFilter(comp string) Filter {
	return func(e parser.LogEntry) bool {
		return e.Component == comp
	}
}

// SinceFilter allows all entries which have timestamp >= since
func SinceFilter(since time.Time) Filter {
	return func(e parser.LogEntry) bool {
		// Compare returns 0 for equal, 1 for after
		return e.Timestamp.Compare(since) >= 0
	}
}

// UntilFilter allows all entries which have timestamp <= until
func UntilFilter(until time.Time) Filter {
	return func(e parser.LogEntry) bool {
		// Compare returns 0 for equal, -1 for before
		return e.Timestamp.Compare(until) <= 0
	}
}

// ContainsFilter allows all entries which contain the given string
func ContainsFilter(contains string) Filter {
	return func(e parser.LogEntry) bool {
		return strings.Contains(strings.ToLower(e.Message), strings.ToLower(contains))
	}
}
