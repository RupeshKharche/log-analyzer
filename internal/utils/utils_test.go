package utils

import "testing"

func TestGetTopNEntries_Basic(t *testing.T) {
	m := map[string]int64{"a": 5, "b": 3, "c": 5}
	got := GetTopNEntries(m, 2)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	// tie-break by key: "a" (5) before "c" (5)
	if got[0].Key != "a" || got[0].Value != 5 {
		t.Errorf("got[0] = %+v, want a/5", got[0])
	}
	if got[1].Key != "c" || got[1].Value != 5 {
		t.Errorf("got[1] = %+v, want c/5", got[1])
	}
}

func TestGetTopNEntries_NLargerThanLen(t *testing.T) {
	m := map[string]int64{"x": 1}
	got := GetTopNEntries(m, 10)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
}

func TestGetTopNEntries_EmptyOrZero(t *testing.T) {
	if got := GetTopNEntries(map[string]int64{}, 5); got != nil {
		t.Errorf("empty map: got %v, want nil", got)
	}
	if got := GetTopNEntries(map[string]int64{"a": 1}, 0); got != nil {
		t.Errorf("n=0: got %v, want nil", got)
	}
	if got := GetTopNEntries(map[string]int64{"a": 1}, -1); got != nil {
		t.Errorf("n<0: got %v, want nil", got)
	}
}
