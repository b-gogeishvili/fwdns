package tools

import (
	"strings"
)

func SplitUpstreams(servers string) []string {
	var res []string
	for _, addr := range strings.Split(servers, ",") {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		// if !strings.Contains(addr, ":") {
		//     addr += ":53"
		// }
		res = append(res, addr)
	}
	return res
}

