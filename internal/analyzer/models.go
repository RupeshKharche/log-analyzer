package analyzer

import (
	"time"
)

type AnalysisStats struct {
	TotalLines     int64
	ParsedLines    int64
	MalformedLines int64
	FilteredLines  int64

	LevelCounts     map[string]int64
	ComponentCounts map[string]int64
	ErrorCounts     map[string]int64
	TimeBuckets     map[string]int64

	ProcessingTime time.Duration
}

func NewStats() AnalysisStats {
	return AnalysisStats{
		LevelCounts:     make(map[string]int64),
		ComponentCounts: make(map[string]int64),
		ErrorCounts:     make(map[string]int64),
		TimeBuckets:     make(map[string]int64),
	}
}
