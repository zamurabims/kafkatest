package window

import (
	"fmt"
	"testing"
	"time"
)

const (
	buckets  = 30
	duration = 10 * time.Second
)

func TestRecordAndTop(t *testing.T) {
	w := New(buckets, duration)
	now := time.Now()

	w.Record("nike", "s1", now)
	w.Record("nike", "s2", now)
	w.Record("adidas", "s3", now)

	w.RebuildCache(now)

	top := w.Top(10, nil)
	if len(top) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(top))
	}
	if top[0].Query != "nike" || top[0].Count != 2 {
		t.Fatalf("expected nike:2 first, got %+v", top[0])
	}
}

func TestDeduplicationPerBucket(t *testing.T) {
	w := New(buckets, duration)
	now := time.Now()

	for i := 0; i < 100; i++ {
		w.Record("spam", "bot-session", now)
	}
	w.Record("legit", "real-user", now)

	w.RebuildCache(now)

	top := w.Top(10, nil)
	found := map[string]int{}
	for _, e := range top {
		found[e.Query] = e.Count
	}
	if found["spam"] != 1 {
		t.Fatalf("expected spam count=1 (deduplicated), got %d", found["spam"])
	}
}

func TestEviction(t *testing.T) {
	w := New(2, duration)
	base := time.Now()

	w.Record("old", "s1", base)
	future := base.Add(3 * duration)
	w.Record("new", "s2", future)

	w.RebuildCache(future)

	top := w.Top(10, nil)
	for _, e := range top {
		if e.Query == "old" {
			t.Fatal("old entry should have been evicted")
		}
	}
}

func TestFilter(t *testing.T) {
	w := New(buckets, duration)
	now := time.Now()

	w.Record("nike", "s1", now)
	w.Record("banned", "s2", now)

	w.RebuildCache(now)

	top := w.Top(10, func(q string) bool { return q == "banned" })
	for _, e := range top {
		if e.Query == "banned" {
			t.Fatal("banned word should be filtered")
		}
	}
}

func TestTopLimit(t *testing.T) {
	w := New(buckets, duration)
	now := time.Now()

	for i := 0; i < 20; i++ {
		w.Record(fmt.Sprintf("query-%d", i), fmt.Sprintf("s%d", i), now)
	}
	w.RebuildCache(now)

	top := w.Top(5, nil)
	if len(top) != 5 {
		t.Fatalf("expected 5, got %d", len(top))
	}
}
