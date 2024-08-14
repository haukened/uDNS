package config

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config represents the configuration structure for the uDNS application.
type Config struct {
	ListenAddr   string            `koanf:"listen_addr"`
	Nameservers  []string          `koanf:"nameservers"`
	Forwarders   map[string]string `koanf:"forwarders"`
	LocalRecords map[string]string `koanf:"local"`
	f            *file.File
}

// NewConfig creates a new Config object with the provided file path.
// If the file path is empty, it uses the default file path "config.yaml".
// It returns the Config object and an error if any.
func NewConfig(filePath string) (*Config, error) {
	// Create a new config object
	c := &Config{}
	// if the file path is empty, use the default file path
	if filePath != "" {
		filePath = "config.yaml"
	}
	c.f = file.Provider(filePath)
	err := c.load()
	if err != nil {
		return nil, err
	}
	c.watch()
	return c, nil
}

func (c *Config) ensurePorts() {
	var err error
	for idx, ns := range c.Nameservers {
		c.Nameservers[idx], err = ensureIPAndPort(ns)
		if err != nil {
			log.Fatalf("unable to parse IP and port: %v", err)
		}
	}
	for k, v := range c.Forwarders {
		c.Forwarders[k], err = ensureIPAndPort(v)
		if err != nil {
			log.Fatalf("unable to parse IP and port: %v", err)
		}
	}
}

// load loads the config file, unmarshals it into the Config object,
// and sets up a watcher for config file changes.
// It returns an error if any.
func (c *Config) load() error {
	// empty any records that may exist, or initialize them
	c.Nameservers = []string{}
	c.Forwarders = map[string]string{}
	c.LocalRecords = map[string]string{}

	// Load the config file
	k := koanf.New(".")
	if err := k.Load(c.f, yaml.Parser()); err != nil {
		return err
	}

	// Unmarshal the config
	err := k.Unmarshal("uDNS", &c)
	if err != nil {
		log.Fatalf("error unmarshalling config: %v", err)
	}
	c.ensurePorts()
	return nil
}

func (c *Config) watch() {
	// Watch the config file for changes
	c.f.Watch(func(event interface{}, err error) {
		// Handle the error if there is one
		if err != nil {
			log.Printf("error watching config file: %v", err)
			return
		}
		log.Println("config file changed, performing hot reload")
		err = c.load()
		if err != nil {
			log.Printf("error reloading config: %v", err)
		}
	})
}

func ensureIPAndPort(s string) (string, error) {
	// creating holding variables for the IP address and port
	ipStr := ""
	portStr := "53"

	// check if the string contains a colon
	if !strings.Contains(s, ":") {
		// if the string does not contain a colon, we know it doesn't have a port
		ipStr = s
	} else {
		// split the string by colon
		parts := strings.Split(s, ":")

		// get the IP address
		ipStr = parts[0]
		// get the port
		portStr = parts[1]
	}

	// ensure that the IP address is valid by attempting to parse it
	if net.ParseIP(ipStr) == nil {
		return "", fmt.Errorf("invalid IP address: %s", ipStr)
	}

	// ensure that the port is valid by parsing to an int and checking bounds
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", err
	}
	if port < 1 || port > 65535 {
		return "", fmt.Errorf("invalid port number: %s", portStr)
	}

	// return the IP address and port
	return fmt.Sprintf("%s:%d", ipStr, port), nil
}
