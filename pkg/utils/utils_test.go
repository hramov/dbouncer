package utils

import (
	"testing"
)

func TestMergeStatMaps(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"a": 3, "c": 4}

	MergeStatMaps(m1, m2)

	if m1["a"] != 4 {
		t.Errorf("m1[\"a\"] = %d, want 4", m1["a"])
	}

	if m1["b"] != 2 {
		t.Errorf("m1[\"b\"] = %d, want 2", m1["b"])
	}

	if m1["c"] != 4 {
		t.Errorf("m1[\"c\"] = %d, want 0", m1["c"])
	}
}
