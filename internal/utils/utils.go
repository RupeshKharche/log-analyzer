package utils

import (
	"cmp"
	"slices"
)

type kv struct {
	Key   string
	Value int64
}

func GetTopNEntries(m map[string]int64, n int) []kv {
	if n <= 0 || len(m) == 0 {
		return nil
	}
	entries := make([]kv, 0, len(m))
	for k, v := range m {
		entries = append(entries, kv{Key: k, Value: v})
	}
	slices.SortFunc(entries, func(a, b kv) int {
		if a.Value != b.Value {
			if a.Value > b.Value {
				return -1
			}
			return 1
		}
		return cmp.Compare(a.Key, b.Key) // deterministic tie-break
	})
	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n]
}
