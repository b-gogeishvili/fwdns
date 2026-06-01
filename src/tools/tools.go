// provides helper functions
package tools

import (
	"strings"

	"github.com/miekg/dns"
)

// turn comma-separated IP addresses into a list
func SplitUpstreams(servers string) []string {
	var res []string
	for _, addr := range strings.Split(servers, ",") {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		// must specify the port, otherwise requests will fail
		if !strings.Contains(addr, ":") {
			addr += ":53"
		}
		res = append(res, addr)
	}
	return res
}

// builds a SERVFAIL response for the given request.
func ErrorReply(req *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetRcode(req, dns.RcodeServerFailure)
	return m
}
