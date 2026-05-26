package stoplist

import (
	"testing"
)

func TestAddContainsRemove(t *testing.T) {
	sl := New()

	sl.Add("Nike")
	if !sl.Contains("nike") {
		t.Fatal("expected nike to be in stoplist")
	}
	if !sl.Contains("  NIKE  ") {
		t.Fatal("expected normalization to work")
	}

	sl.Remove("nike")
	if sl.Contains("nike") {
		t.Fatal("expected nike to be removed")
	}
}

func TestList(t *testing.T) {
	sl := New()
	sl.Add("foo")
	sl.Add("bar")
	words := sl.List()
	if len(words) != 2 {
		t.Fatalf("expected 2 words, got %d", len(words))
	}
}
