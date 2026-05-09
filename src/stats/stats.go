package stats

import (
	"sync"
	"time"
)

// will need query struct for http dashboard
type Query struct {
	Time    time.Time `json:"time"`
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Cached  bool      `json:"cached"`
	TTL     uint32    `json:"ttl"`
	Latency float64   `json:"latency"`
}

type Stats struct {
	mu      sync.Mutex
	total   uint64
	hits    uint64
	misses  uint64
	errors  uint64
	recent  []Query
	maxKeep int
}

func New(maxRecent int) *Stats {
	return &Stats{
		recent:  make([]Query, 0, maxRecent),
		maxKeep: maxRecent,
	}
}

func (s *Stats) Record(q Query, isError bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.total++
	switch {
	case isError:
		s.errors++
	case q.Cached:
		s.hits++
	default:
		s.misses++
	}

	s.recent = append(s.recent, q)
	if len(s.recent) > s.maxKeep {
		s.recent = s.recent[1:]
	}
}
