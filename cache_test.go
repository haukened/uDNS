package main

import (
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestDNSCache_AddGetValid(t *testing.T) {
	cache := NewDNSCache()

	rr := &dns.A{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeA,
			Ttl:    3600,
		},
		A: net.ParseIP("192.0.2.1"),
	}

	cache.Add(rr)

	records, ok := cache.Get("example.com", dns.TypeA)
	if !ok {
		t.Errorf("Expected to find DNS records for key 'example.com' and query type 'A', but not found")
	}

	if len(records) != 1 {
		t.Errorf("Expected to find 1 DNS record, but found %d", len(records))
	}
}

func TestDNSCache_AddGetInvalid(t *testing.T) {
	cache := NewDNSCache()

	rr := &dns.A{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeA,
			Ttl:    1,
		},
		A: net.ParseIP("192.0.2.1"),
	}

	cache.Add(rr)

	time.Sleep(2 * time.Second)

	records, ok := cache.Get("example.com", dns.TypeA)
	if ok {
		t.Errorf("Expected to find no DNS records, but found %d", len(records))
	}
}

func TestDNSCache_ValidAndInvalid(t *testing.T) {
	cache := NewDNSCache()

	rr := &dns.A{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeA,
			Ttl:    1,
		},
		A: net.ParseIP("192.0.2.1"),
	}
	cache.Add(rr)
	rr2 := &dns.A{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeA,
			Ttl:    10,
		},
		A: net.ParseIP("192.0.2.2"),
	}
	cache.Add(rr2)

	time.Sleep(2 * time.Second)

	records, ok := cache.Get("example.com", dns.TypeA)

	if !ok {
		t.Errorf("Expected to find DNS records for key 'example.com' and query type 'A', but not found")
	}

	if len(records) != 1 {
		t.Errorf("Expected to find 1 DNS record, but found %d", len(records))
	}

	if records[0].Value != "192.0.2.2" {
		t.Errorf("Cache returned incorrect record, expected '192.0.2.2' got %s", records[0].Value)
	}
}

func TestDNSCache_UpdateExpiration(t *testing.T) {
	cache := NewDNSCache()

	rr := &dns.A{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeA,
			Ttl:    10,
		},
		A: net.ParseIP("192.0.2.1"),
	}

	cache.Add(rr)

	rr.Hdr.Ttl = 20
	expected := time.Now().Add(20 * time.Second)
	cache.Add(rr)

	records, _ := cache.Get("example.com", dns.TypeA)

	if len(records) != 1 {
		t.Errorf("Expected to find 1 DNS record, but found %d", len(records))
	}

	if records[0].Expires.Sub(expected) > 1*time.Second {
		t.Errorf("Expected expiration time to be %v, but got %v", expected, records[0].Expires)
	}
}

func TestDNSCache_Delete(t *testing.T) {
	cache := NewDNSCache()

	rr := &dns.A{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeA,
			Ttl:    3600,
		},
		A: net.ParseIP("192.0.2.1"),
	}

	cache.Add(rr)

	cache.Delete("example.com", dns.TypeA)

	_, ok := cache.Get("example.com", dns.TypeA)
	if ok {
		t.Errorf("Expected DNS records for key 'example.com' and query type 'A' to be deleted, but found")
	}
}

func TestDNSCache_Purge(t *testing.T) {
	cache := NewDNSCache()

	rr := &dns.A{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeA,
			Ttl:    3600,
		},
		A: net.ParseIP("192.0.2.1"),
	}

	cache.Add(rr)

	cache.Purge()

	if cache.Len() != 0 {
		t.Errorf("Expected DNS cache to be empty after purging, but found %d entries", cache.Len())
	}
}

func TestDNSCache_Len(t *testing.T) {
	cache := NewDNSCache()

	if cache.Len() != 0 {
		t.Errorf("Expected DNS cache to be empty, but found %d entries", cache.Len())
	}

	rr := &dns.A{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeA,
			Ttl:    3600,
		},
		A: net.ParseIP("192.0.2.1"),
	}

	cache.Add(rr)

	if cache.Len() != 1 {
		t.Errorf("Expected DNS cache to have 1 entry, but found %d entries", cache.Len())
	}
}
func TestParseRR_ValidRR(t *testing.T) {
	rr := &dns.A{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeA,
			Ttl:    3600,
		},
		A: net.ParseIP("192.0.2.1"),
	}

	record, err := parseRR(rr)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if record.Name != "example.com." {
		t.Errorf("Expected record name to be 'example.com.', but got: %s", record.Name)
	}

	if record.Value != "192.0.2.1" {
		t.Errorf("Expected record value to be '192.0.2.1', but got: %s", record.Value)
	}

	if record.Type != dns.TypeA {
		t.Errorf("Expected record type to be dns.TypeA, but got: %d", record.Type)
	}

	// Calculate the expected expiration time based on the TTL
	expectedExpiration := time.Now().Add(3600 * time.Second)
	if record.Expires.Sub(expectedExpiration) > 1*time.Second {
		t.Errorf("Expected expiration time to be %v, but got %v", expectedExpiration, record.Expires)
	}
}
