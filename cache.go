package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DNSRecord represents a DNS record with its value and expiration time.
type DNSRecord struct {
	Name    string    // The name of the DNS record.
	Value   string    // The data of the DNS record.
	Type    uint16    // The type of the DNS record.
	Expires time.Time // The expiration time of the DNS record.
}

func (r DNSRecord) ToRR() (dns.RR, error) {
	strValue := fmt.Sprintf("%s\t%d\tIN\t%s\t%s", r.Name, int(time.Until(r.Expires).Seconds()), dns.TypeToString[r.Type], r.Value)
	return dns.NewRR(strValue)
}

// DNSCache represents a cache for DNS records.
type DNSCache struct {
	entries map[string][]DNSRecord
}

func canonicalizeKey(key string) string {
	if !strings.HasSuffix(key, ".") {
		key = key + "."
	}
	return key
}

func parseRR(rr dns.RR) (DNSRecord, error) {
	parts := strings.Split(rr.String(), "\t")
	if len(parts) != 5 {
		return DNSRecord{}, fmt.Errorf("unable to parse RR, expected 5 fields, got %d: %s", len(parts), rr.String())
	}
	return DNSRecord{
		Name:    parts[0],
		Value:   parts[4],
		Type:    rr.Header().Rrtype,
		Expires: time.Now().Add(time.Duration(rr.Header().Ttl) * time.Second),
	}, nil
}

// NewDNSCache creates a new DNS cache.
// It returns a pointer to a DNSCache object.
func NewDNSCache() *DNSCache {
	// implementation details...
	return &DNSCache{entries: make(map[string][]DNSRecord)}
}

// Add adds a new entry to the DNS cache with the specified key, query type, value, and time-to-live (TTL).
// The key is a string representing the cache entry key.
// The qtype is an unsigned 16-bit integer representing the query type.
// The value is a string representing the cache entry value.
// The ttl is an unsigned 32-bit integer representing the time-to-live in seconds for the cache entry.
func (c *DNSCache) Add(data dns.RR) {
	cKey := canonicalizeKey(data.Header().Name)
	// parse the RR
	record, err := parseRR(data)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	// check if the key exists
	records, ok := c.entries[cKey]
	if !ok {
		// if the key does not exist, create a new slice of DNSRecord
		records = []DNSRecord{}
		// append the new record to the slice
		records = append(records, record)
		// store the slice in the cache
		c.entries[cKey] = records
		return
	} else {
		// if the key exists, check if the record already exists
		found := false
		for idx, r := range records {
			if record.Value == r.Value {
				found = true
				// update the expiration time
				r.Expires = time.Now().Add(time.Duration(data.Header().Ttl) * time.Second)
				// update the record in the slice
				records[idx] = r
				// store the slice in the cache
				c.entries[cKey] = records
				return
			}
		}
		if !found {
			// if the record does not exist, append the new record to the slice
			records = append(records, record)
			// store the slice in the cache
			c.entries[cKey] = records
			return
		}
	}
}

// Get retrieves the DNS records associated with the specified key and query type.
// It returns a slice of DNSRecord and a boolean value indicating whether the records were found.
func (c *DNSCache) Get(key string, qtype uint16) (results []dns.RR, ok bool) {
	// ensure that the key has the canonical DNS trailing .
	cKey := canonicalizeKey(key)
	validResults := []DNSRecord{}
	rr, ok := c.entries[cKey]
	if !ok {
		return
	}
	// check expiration time
	for _, r := range rr {
		if r.Expires.After(time.Now()) {
			validResults = append(validResults, r)
		}
	}
	// if the number of valid results is different from the original number of results
	// update the cache
	if len(validResults) != len(rr) {
		// we don't need to store the expired records, so store valid results
		if len(validResults) == 0 {
			delete(c.entries, cKey)
			ok = false
			return
		} else {
			c.entries[cKey] = validResults
		}
	}
	// convert the DNSRecord to dns.RR
	for _, r := range validResults {
		rr, err := r.ToRR()
		if err != nil {
			log.Printf("unable to convert DNSRecord to RR: %v", err)
			ok = false
			return
		}
		results = append(results, rr)
	}

	return
}

// Delete removes the specified key from the DNSCache.
// It takes the key as a string and the qtype as a uint16.
// This function does not return any value.
func (c *DNSCache) Delete(key string, qtype uint16) {
	delete(c.entries, canonicalizeKey(key))
}

// Purge removes all entries from the DNS cache.
// This method clears the cache and removes all stored DNS records.
// It does not return any values.
func (c *DNSCache) Purge() {
	c.entries = make(map[string][]DNSRecord)
}

// Len returns the number of items in the DNSCache.
func (c *DNSCache) Len() int {
	return len(c.entries)
}
