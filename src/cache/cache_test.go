package cache

import (
	"testing"
	"time"

	"github.com/miekg/dns"
)

// builds a simple A-record response with the given TTL
func msgWithTTL(ttl uint32) *dns.Msg {
	m := new(dns.Msg)
	m.Answer = append(m.Answer, &dns.A{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		A: []byte{1, 2, 3, 4},
	})
	return m
}

func TestKey(t *testing.T) {
	q := dns.Question{Name: "Example.COM.", Qtype: dns.TypeA, Qclass: dns.ClassINET}
	if got, want := Key(q), "example.com.|A|IN"; got != want {
		t.Errorf("Key() = %q, want %q", got, want)
	}
}

func TestSetAndGet(t *testing.T) {
	c := New()
	c.Set("k", msgWithTTL(60))

	out, ok := c.Get("k")
	if !ok {
		t.Fatal("Get() returned ok=false, want true")
	}
	if len(out.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(out.Answer))
	}
}

func TestGetMissing(t *testing.T) {
	c := New()
	if _, ok := c.Get("nope"); ok {
		t.Error("Get() on missing key returned ok=true, want false")
	}
}

func TestSetZeroTTLNotStored(t *testing.T) {
	c := New()
	c.Set("k", msgWithTTL(0))
	if c.Len() != 0 {
		t.Errorf("Len() = %d, want 0 (zero-TTL responses are not cached)", c.Len())
	}
}

func TestMinTTL(t *testing.T) {
	m := msgWithTTL(100)
	m.Answer = append(m.Answer, &dns.A{
		Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 30},
		A:   []byte{5, 6, 7, 8},
	})
	if got := MinTTL(m); got != 30 {
		t.Errorf("MinTTL() = %d, want 30", got)
	}
}

func TestDeleteExpired(t *testing.T) {
	c := New()
	c.entries["expired"] = entry{msg: msgWithTTL(1), expiresAt: time.Now().Add(-time.Second)}
	c.entries["fresh"] = entry{msg: msgWithTTL(1), expiresAt: time.Now().Add(time.Minute)}

	if removed := c.DeleteExpired(); removed != 1 {
		t.Errorf("DeleteExpired() removed %d, want 1", removed)
	}
	if c.Len() != 1 {
		t.Errorf("Len() = %d, want 1", c.Len())
	}
}
