package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
)

// routeFlags collects domain=webhook pairs from repeated -route flags.
type routeFlags map[string]string

func (r *routeFlags) String() string {
	return "domain=webhookURL"
}

func (r *routeFlags) Set(value string) error {
	domain, webhook, err := parseRoutePair(value)
	if err != nil {
		return err
	}
	(*r)[domain] = webhook
	return nil
}

func parseRoutePair(value string) (domain, webhook string, err error) {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return "", "", errors.New("expected domain=webhookURL")
	}

	domain = normalizeDomain(parts[0])
	if domain == "" {
		return "", "", errors.New("domain must not be empty")
	}

	webhook = strings.TrimSpace(parts[1])
	if _, err := url.ParseRequestURI(webhook); err != nil {
		return "", "", fmt.Errorf("invalid webhook URL for %q: %w", domain, err)
	}

	return domain, webhook, nil
}

func loadRoutes(configPath string, cliRoutes routeFlags) (map[string]string, error) {
	routes := make(map[string]string)

	if configPath != "" {
		fileRoutes, err := loadRoutesFile(configPath)
		if err != nil {
			return nil, err
		}
		for domain, webhook := range fileRoutes {
			routes[domain] = webhook
		}
	}

	for domain, webhook := range cliRoutes {
		routes[domain] = webhook
	}

	if len(routes) == 0 {
		return nil, errors.New("no routes configured: set -config and/or -route")
	}

	return routes, nil
}

func loadRoutesFile(path string) (map[string]string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read routes config: %w", err)
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse routes config: %w", err)
	}

	routes := make(map[string]string, len(raw))
	for domain, webhook := range raw {
		domain = normalizeDomain(domain)
		if domain == "" {
			return nil, errors.New("routes config contains an empty domain key")
		}
		webhook = strings.TrimSpace(webhook)
		if _, err := url.ParseRequestURI(webhook); err != nil {
			return nil, fmt.Errorf("invalid webhook URL for domain %q: %w", domain, err)
		}
		routes[domain] = webhook
	}

	return routes, nil
}

func normalizeDomain(domain string) string {
	return strings.ToLower(strings.TrimSpace(domain))
}

func webhookForRecipient(routes map[string]string, address string) (domain, webhook string, err error) {
	domain, err = recipientDomain(address)
	if err != nil {
		return "", "", err
	}

	webhook, ok := routes[domain]
	if !ok {
		return domain, "", fmt.Errorf("no webhook configured for domain %q", domain)
	}

	return domain, webhook, nil
}
