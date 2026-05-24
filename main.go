package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/miekg/dns"

	"fwdns/src/cache"
	"fwdns/src/dashboard"
	"fwdns/src/resolver"
	"fwdns/src/stats"
	"fwdns/src/tools"
)

func main() {
	dnsAddr := flag.String("dns", ":5300", "port to serve DNS on")
	httpAddr := flag.String("http", ":8080", "port to serve HTTP on")
	upstream := flag.String("upstream", "9.9.9.9,1.1.1.1", "upstream dns servers, separated by comma")
	timeout := flag.Duration("timeout", 10*time.Second, "how long to wait for an upstream server (seconds)")
	cleanup := flag.Duration("cleanup", 60*time.Second, "how often to clean expired cache entries (seconds)")
	flag.Parse()

	upstreamServers := tools.SplitUpstreams(*upstream)

	if len(upstreamServers) == 0 {
		log.Fatal("At least one upstream server is required")
	}

	c := cache.New()
	s := stats.New(100)
	res := resolver.New(c, s, upstreamServers, *timeout)

	stopCleanup := c.StartCleanup(*cleanup)
	defer stopCleanup()

	dnsServer := &dns.Server{Addr: *dnsAddr, Net: "udp", Handler: res}
	go func() {
		log.Printf("DNS server is listening on %s (UDP); upstream servers: %s",
			*dnsAddr, strings.Join(upstreamServers, ", "))
		err := dnsServer.ListenAndServe()

		if err != nil {
			log.Fatalf("DNS server failed: %v", err)
		}
	}()

	webServer := &http.Server{Addr: *httpAddr, Handler: dashboard.New(s, c).Handler()}
	go func() {
		log.Printf("Web dashboard on http://localhost%s", *httpAddr)
		if err := webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("web server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("DNS Server is shutting down...")
	_ = dnsServer.Shutdown()
}
