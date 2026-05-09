package resolver

import (
	"log"
	"time"

	"github.com/miekg/dns"

	"fwdns/src/cache"
	"fwdns/src/stats"
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
		_ = w.WriteMsg(errorReply(req))
		return
	}
	q := req.Question[0]
	key := cache.Key(q)

	if resp, ok := r.cache.Get(key); ok {
		resp.Id = req.Id
		resp.Question = req.Question
		r.record(q, true, false, cache.MinTTL(resp), start)
		_ = w.WriteMsg(resp)
		return
	}
	resp, err := r.forward(req)
	if err != nil || resp == nil {
		r.record(q, false, true, 0, start)
		_ = w.WriteMsg(errorReply(req))
		return
	}
	r.cache.Set(key, resp)
	r.record(q, false, false, cache.MinTTL(resp), start)
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

func (r *Resolver) record(q dns.Question, cached, isErr bool, ttl uint32, start time.Time) {
	qtype := dns.TypeToString[q.Qtype]
	latency := float64(time.Since(start).Microseconds()) / 1000.0

	if isErr == true {
		log.Printf("lookup |   %s   |   %s   |   ttl=%d   |   %fs  |  failed", q.Name, qtype, ttl, latency)
	} else if cached == false {
		log.Printf("lookup |   %s   |   %s   |   ttl=%d   |   %fs  |  cache miss", q.Name, qtype, ttl, latency)
	} else {
		log.Printf("lookup |   %s   |   %s   |   ttl=%d   |   %fs  |  cache hit", q.Name, qtype, ttl, latency)
	}

	r.stats.Record(stats.Query{
		Time:    time.Now(),
		Name:    q.Name,
		Type:    qtype,
		Cached:  cached,
		TTL:     ttl,
		Latency: latency,
	}, isErr)
}

func errorReply(req *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetRcode(req, dns.RcodeServerFailure)
	return m
}
