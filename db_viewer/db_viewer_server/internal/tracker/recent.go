package tracker

import (
	"sort"
	"sync"
	"time"
)

type RecentTracker struct {
	mu    sync.Mutex
	usage map[string]time.Time
}

func New() *RecentTracker {
	return &RecentTracker{usage: make(map[string]time.Time)}
}

func (r *RecentTracker) Touch(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.usage[key] = time.Now()
}

func (r *RecentTracker) SortByRecent(keys []string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	sorted := make([]string, len(keys))
	copy(sorted, keys)

	sort.SliceStable(sorted, func(i, j int) bool {
		ti := r.usage[sorted[i]]
		tj := r.usage[sorted[j]]
		return ti.After(tj)
	})

	return sorted
}
