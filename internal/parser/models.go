package parser

import (
	LogAnalyzer "loganalyzer/v2"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Component string
	Message   string
}

func (e LogEntry) String() string {
	var b strings.Builder
	b.WriteString(e.Timestamp.Format(LogAnalyzer.TimestampFormat))
	b.WriteString(e.Level.String())
	b.WriteString(e.Component)
	b.WriteString(e.Component)
	b.WriteString(e.Message)
	return b.String()
}

type RawLine struct {
	File   string
	Number int
	Text   string
}

func NewRawLine(file string, number int, text string) *RawLine {
	return &RawLine{
		File:   file,
		Number: number,
		Text:   text,
	}
}

type ParseResult struct {
	Entry    LogEntry
	Valid    bool
	RawLine  RawLine
	Filtered bool
}
