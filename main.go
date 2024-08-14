package main

import (
	"log"

	"github.com/haukened/uDNS/config"
	"github.com/miekg/dns"
)

func main() {
	// Create a new config object
	c, err := config.NewConfig("config.yaml")
	if err != nil {
		log.Fatalf("error creating config: %v", err)
	}

	// Create a new handler
	handler := NewHandler(c)

	// create a new DNS server
	server := &dns.Server{
		Addr:      c.ListenAddr,
		Net:       "udp",
		Handler:   handler,
		UDPSize:   65535,
		ReusePort: true,
	}

	// start the server
	log.Printf("Starting server on %s", c.ListenAddr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
