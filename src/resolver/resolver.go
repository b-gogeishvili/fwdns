package resolver

import (
	"log"
	"time"

	"github.com/miekg/dns"

	"fwdns/src/cache"
	"fwdns/src/stats"
	"fwdns/src/tools"
)

type lookupStatus string

const (
	statusHit  lookupStatus = "HIT"
	statusMiss lookupStatus = "MISS"
	statusFail lookupStatus = "FAIL"
)

type Resolver struct {
	cache           *cache.Cache
	stats           *stats.Stats
	upstreamServers []string    // e.g. ["8.8.8.8", "1.1.1.1"], called in order
	client          *dns.Client // client for talking to upstreams
}

func New(c *cache.Cache, s *stats.Stats, upstreamServers []string, timeout time.Duration) *Resolver {
	return &Resolver{
		cache:           c,
		stats:           s,
		upstreamServers: upstreamServers,
		client:          &dns.Client{Net: "udp", Timeout: timeout},
	}
}

func (r *Resolver) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	start := time.Now()

	if len(req.Question) == 0 {
		_ = w.WriteMsg(tools.ErrorReply(req))
		return
	}
	q := req.Question[0]
	key := cache.Key(q)

	if resp, ok := r.cache.Get(key); ok {
		resp.Id = req.Id
		resp.Question = req.Question
		r.record(q, statusHit, int(cache.MinTTL(resp)), start)
		_ = w.WriteMsg(resp)
		return
	}
	resp, err := r.forward(req)
	if err != nil || resp == nil {
		r.record(q, statusFail, -1, start)
		_ = w.WriteMsg(tools.ErrorReply(req))
		return
	}
	if resp.Rcode != dns.RcodeSuccess {
		r.record(q, statusFail, -1, start)
		_ = w.WriteMsg(resp)
		return
	}
	r.cache.Set(key, resp)
	r.record(q, statusMiss, int(cache.MinTTL(resp)), start)
	_ = w.WriteMsg(resp)
}

func (r *Resolver) forward(req *dns.Msg) (*dns.Msg, error) {
	var lastErr error
	for _, up := range r.upstreamServers {
		resp, _, err := r.client.Exchange(req, up)
		if err != nil {
			lastErr = err
			continue
		}
		return resp, nil
	}
	return nil, lastErr
}

func (r *Resolver) record(q dns.Question, status lookupStatus, ttl int, start time.Time) {
	qtype := dns.TypeToString[q.Qtype]
	latencyMs := float64(time.Since(start).Microseconds()) / 1000.0

	log.Printf("lookup  %-4s  %-32s  %-6s  ttl=%-6d  latency=%7.2fms",
		status, q.Name, qtype, ttl, latencyMs)

	r.stats.Record(stats.Query{
		Time:    time.Now(),
		Name:    q.Name,
		Type:    qtype,
		Cached:  status == statusHit,
		TTL:     ttl,
		Latency: latencyMs,
	}, status == statusFail)
}
