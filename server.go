package main

import (
	"log"
	"strings"

	"github.com/haukened/uDNS/config"
	"github.com/miekg/dns"
)

type dnsHandler struct {
	config *config.Config
	cache  *DNSCache
}

func NewHandler(c *config.Config) *dnsHandler {
	return &dnsHandler{config: c, cache: NewDNSCache()}
}

func (h *dnsHandler) getForwarder(q string) (string, bool) {
	// determine the root domain of the query
	for upDomain, upServer := range h.config.Forwarders {
		// ensure the upDomain has the canonical DNS trailing .
		if !strings.HasSuffix(upDomain, ".") {
			upDomain = upDomain + "."
		}
		// determine if the query is a subdomain of the upDomain
		if dns.IsSubDomain(upDomain, q) {
			return upServer, true
		}
	}
	return "", false
}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, question := range r.Question {
		log.Printf("Received query: %s\n", question.Name)
		server, ok := h.getForwarder(question.Name)
		if !ok {
			log.Printf("no forwarder found for %s, using default upstream", question.Name)
			server = h.config.Nameservers[0]
		} else {
			log.Printf("forwarding query for %s to %s", question.Name, server)
		}
		// send the query
		answers := resolve(question.Name, question.Qtype, server)
		msg.Answer = append(msg.Answer, answers...)

		// TBD set cache - probably needs rework
		for _, ans := range answers {
			h.cache.Add(ans)
		}
	}
	w.WriteMsg(msg)
}

func resolve(domain string, qtype uint16, server string) []dns.RR {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.RecursionDesired = true

	c := new(dns.Client)
	in, rtt, err := c.Exchange(m, server)
	if err != nil {
		log.Printf("error querying %s: %v", server, err)
		return []dns.RR{}
	}

	log.Printf("query executed in %s\n", rtt)
	for _, ans := range in.Answer {
		log.Println(ans)
	}

	return in.Answer
}
