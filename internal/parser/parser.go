package parser

import (
	"fmt"
	LogAnalyzer "loganalyzer/v2"
	"strings"
	"time"
)

type LogLevel int

const (
	LevelUnknown          = -1
	LevelDebug   LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var levels = map[LogLevel]string{
	LevelUnknown: "Unknown",
	LevelDebug:   "DEBUG",
	LevelInfo:    "INFO",
	LevelWarn:    "WARN",
	LevelError:   "ERROR",
	LevelFatal:   "FATAL",
}

var revLevels = map[string]LogLevel{
	"Unknown": LevelUnknown,
	"DEBUG":   LevelDebug,
	"INFO":    LevelInfo,
	"WARN":    LevelWarn,
	"ERROR":   LevelError,
	"FATAL":   LevelFatal,
}

func (l LogLevel) String() string {
	val, ok := levels[l]
	if !ok {
		return "Unknown"
	}
	return val
}

func ParseLogLevel(s string) (LogLevel, bool) {
	val, ok := revLevels[s]
	if !ok {
		val = LevelUnknown
		ok = false
	}
	return val, ok
}

func ParseLine(line string) (LogEntry, error) {
	line = strings.TrimSpace(line)

	tokens := strings.Split(line, " ")
	emptyLogEntry := LogEntry{}
	logEntry := LogEntry{}

	if len(line) < 19 {
		return emptyLogEntry, fmt.Errorf("parse line: line too short")
	}
	if len(tokens) < 4 {
		return emptyLogEntry, fmt.Errorf("parse line: not enough fields")
	}

	lengthWithoutMsg := 0

	// Parse date and time
	timestampLen := 19
	timestamp := line[:timestampLen]
	parsedTimestamp, err := time.Parse(LogAnalyzer.TimestampFormat, timestamp)
	if err != nil {
		return emptyLogEntry, fmt.Errorf("parse line: %w", err)
	}
	logEntry.Timestamp = parsedTimestamp
	lengthWithoutMsg += timestampLen

	// Parse log level
	logLevel := tokens[2]
	parsedLevel, ok := ParseLogLevel(logLevel)
	if !ok {
		return emptyLogEntry, fmt.Errorf("parse line: unknown level %q", logLevel)
	}
	logEntry.Level = parsedLevel
	lengthWithoutMsg += len(logLevel) + 1 // +1 for space before the log level

	// Parse component
	component := tokens[3]
	if !(strings.HasPrefix(component, "[") && strings.HasSuffix(component, "]")) {
		return emptyLogEntry, fmt.Errorf("invalid component: %s", component)
	}
	logEntry.Component = strings.Trim(component, "[]")
	lengthWithoutMsg += len(component) + 1 // +1 for space before the component

	// Parse msg
	lengthWithoutMsg = min(lengthWithoutMsg+1, len(line))
	msg := line[lengthWithoutMsg:]
	msg = strings.TrimSpace(msg)
	logEntry.Message = msg

	return logEntry, nil
}
