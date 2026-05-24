package stats

import (
	"sync"
	"time"
)

type Query struct {
	Time    time.Time `json:"time"`
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Cached  bool      `json:"cached"`
	TTL     int       `json:"ttl"`
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

type Snapshot struct {
	Total     uint64  `json:"total"`
	Hits      uint64  `json:"hits"`
	Misses    uint64  `json:"misses"`
	Errors    uint64  `json:"errors"`
	HitRate   float64 `json:"hitRate"`
	CacheSize int     `json:"cacheSize"`
	Recent    []Query `json:"recent"`
}

func (s *Stats) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	var hitRate float64
	if answered := s.hits + s.misses; answered > 0 {
		hitRate = float64(s.hits) / float64(answered) * 100
	}

	recentCopy := make([]Query, len(s.recent))
	copy(recentCopy, s.recent)

	return Snapshot{
		Total:   s.total,
		Hits:    s.hits,
		Misses:  s.misses,
		Errors:  s.errors,
		HitRate: hitRate,
		Recent:  recentCopy,
	}
}
