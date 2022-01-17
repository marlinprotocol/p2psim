package core

import "testing"

func TestBasic(t *testing.T) {
	set := NewSet()

	set.Add(2, 4)
	if !set.Exists(2) {
		t.Error("2 must be in the set!")
	}
	if set.Exists(3) {
		t.Error("3 must not be in the set!")
	}
	if set.Len() != 2 {
		t.Error("Set must have two elements!")
	}

	set.Remove(2)
	if set.Exists(2) {
		t.Error("2 must not be in the set!")
	}
	if !set.Exists(4) {
		t.Error("4 must be in the set!")
	}
	if set.Len() != 1 {
		t.Error("Set must have one element!")
	}

	set.Remove(4, 8)
	if set.Exists(4) {
		t.Error("4 must not be in the set!")
	}
	if set.Len() != 0 {
		t.Error("Set must be empty!")
	}
}

func TestClear(t *testing.T) {
	set := NewSet(2, 4, 8)
	set.Clear()
	if set.Len() != 0 {
		t.Error("Set must be empty!")
	}
}
