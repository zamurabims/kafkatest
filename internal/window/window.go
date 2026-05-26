package window

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"wish1/internal/metrics"
)

type Entry struct {
	Query string `json:"query"`
	Count int    `json:"count"`
}

type bucket struct {
	mu       sync.Mutex
	counts   map[string]map[string]struct{} // query → set<session_id>
	startsAt time.Time
}

func newBucket(t time.Time) *bucket {
	return &bucket{
		counts:   make(map[string]map[string]struct{}),
		startsAt: t,
	}
}

func (b *bucket) record(query, sessionID string) {
	b.mu.Lock()
	if b.counts[query] == nil {
		b.counts[query] = make(map[string]struct{})
	}
	b.counts[query][sessionID] = struct{}{}
	b.mu.Unlock()
}

type Window struct {
	bucketDuration time.Duration
	bucketCount    int

	mu      sync.Mutex
	buckets []*bucket

	cachedTop unsafe.Pointer
}

func New(bucketCount int, bucketDuration time.Duration) *Window {
	w := &Window{
		bucketDuration: bucketDuration,
		bucketCount:    bucketCount,
		buckets:        make([]*bucket, 0, bucketCount),
	}
	empty := []Entry{}
	atomic.StorePointer(&w.cachedTop, unsafe.Pointer(&empty))
	return w
}

func (w *Window) Record(query, sessionID string, ts time.Time) {
	w.mu.Lock()
	w.evict(ts)
	cur := w.currentBucket(ts)
	w.mu.Unlock()

	cur.record(query, sessionID)
}

func (w *Window) currentBucket(ts time.Time) *bucket {
	slotStart := ts.Truncate(w.bucketDuration)

	if len(w.buckets) > 0 {
		last := w.buckets[len(w.buckets)-1]
		if last.startsAt.Equal(slotStart) {
			return last
		}
	}

	b := newBucket(slotStart)
	w.buckets = append(w.buckets, b)
	return b
}

func (w *Window) evict(now time.Time) {
	cutoff := now.Add(-w.bucketDuration * time.Duration(w.bucketCount))
	i := 0
	for i < len(w.buckets) && w.buckets[i].startsAt.Before(cutoff) {
		i++
	}
	w.buckets = w.buckets[i:]
}

func (w *Window) RebuildCache(now time.Time) {
	start := time.Now()

	w.mu.Lock()
	w.evict(now)
	snapshot := make([]*bucket, len(w.buckets))
	copy(snapshot, w.buckets)
	w.mu.Unlock()

	totals := make(map[string]int)
	for _, b := range snapshot {
		b.mu.Lock()
		for query, sessions := range b.counts {
			totals[query] += len(sessions)
		}
		b.mu.Unlock()
	}

	entries := make([]Entry, 0, len(totals))
	for q, c := range totals {
		entries = append(entries, Entry{Query: q, Count: c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count != entries[j].Count {
			return entries[i].Count > entries[j].Count
		}
		return entries[i].Query < entries[j].Query
	})

	atomic.StorePointer(&w.cachedTop, unsafe.Pointer(&entries))

	metrics.CacheSize.Set(float64(len(entries)))
	metrics.CacheRebuildDuration.Observe(time.Since(start).Seconds())
}

func (w *Window) Top(n int, filter func(string) bool) []Entry {
	p := atomic.LoadPointer(&w.cachedTop)
	all := *(*[]Entry)(p)

	result := make([]Entry, 0, n)
	for _, e := range all {
		if filter != nil && filter(e.Query) {
			continue
		}
		result = append(result, e)
		if len(result) == n {
			break
		}
	}
	return result
}
