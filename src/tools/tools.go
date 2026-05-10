package tools

import (
	"strings"

	"github.com/miekg/dns"
)

func SplitUpstreams(servers string) []string {
	var res []string
	for _, addr := range strings.Split(servers, ",") {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		if !strings.Contains(addr, ":") {
			addr += ":53"
		}
		res = append(res, addr)
	}
	return res
}

// ErrorReply builds a SERVFAIL response for the given request.
func ErrorReply(req *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetRcode(req, dns.RcodeServerFailure)
	return m
}
