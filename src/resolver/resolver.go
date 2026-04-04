package resolver

import (
	"time"

	"github.com/miekg/dns"
)

type Resolver struct {
	upstreamServers []string    // e.g. ["8.8.8.8", "1.1.1.1"], called in order
	client          *dns.Client // client for talking to upstreams
}

func New(upstreamServers []string, timeout time.Duration) *Resolver {
	return &Resolver {
		upstreamServers: upstreamServers,
		client:          &dns.Client{ Net: "udp", Timeout: timeout },
	}
}

func (r *Resolver) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	resp, err := r.forward(req)
	if err != nil || resp == nil {
		_ = w.WriteMsg(errorReply(req))
		return
	}
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

