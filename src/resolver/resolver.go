package resolver

import (
	"log"
	"time"

	"github.com/miekg/dns"

	"fwdns/src/cache"
)

type Resolver struct {
	cache           *cache.Cache
	upstreamServers []string    // e.g. ["8.8.8.8", "1.1.1.1"], called in order
	client          *dns.Client // client for talking to upstreams
}

func New(c *cache.Cache, upstreamServers []string, timeout time.Duration) *Resolver {
	return &Resolver{
		cache:           c,
		upstreamServers: upstreamServers,
		client:          &dns.Client{Net: "udp", Timeout: timeout},
	}
}

func (r *Resolver) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	if len(req.Question) == 0 {
		_ = w.WriteMsg(errorReply(req))
		return
	}
	q := req.Question[0]
	key := cache.Key(q)
	name := q.Name
	qtype := dns.TypeToString[q.Qtype]

	if resp, ok := r.cache.Get(key); ok {
		resp.Id = req.Id
		resp.Question = req.Question
		log.Printf("lookup %s %s -> cache hit (%s)", name, qtype, dns.RcodeToString[resp.Rcode])
		_ = w.WriteMsg(resp)
		return
	}
	resp, err := r.forward(req)
	if err != nil || resp == nil {
		log.Printf("lookup %s %s -> upstream error: %v", name, qtype, err)
		_ = w.WriteMsg(errorReply(req))
		return
	}
	r.cache.Set(key, resp)
	log.Printf("lookup %s %s -> forwarded (%s, %d answers)", name, qtype, dns.RcodeToString[resp.Rcode], len(resp.Answer))
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

func errorReply(req *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetRcode(req, dns.RcodeServerFailure)
	return m
}
