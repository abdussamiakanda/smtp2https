package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
)

const apiKeyHeader = "X-Api-Key"

// Route is the webhook target and optional auth for one domain.
type Route struct {
	Webhook string
	APIKey  string
}

// routeFlags collects domain=webhook pairs from repeated -route flags (no API key).
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

func validateWebhook(webhook string) error {
	_, err := url.ParseRequestURI(webhook)
	return err
}

func loadRoutes(configPath string, cliRoutes routeFlags) (map[string]Route, error) {
	routes := make(map[string]Route)

	if configPath != "" {
		fileRoutes, err := loadRoutesFile(configPath)
		if err != nil {
			return nil, err
		}
		for domain, route := range fileRoutes {
			routes[domain] = route
		}
	}

	for domain, webhook := range cliRoutes {
		routes[domain] = Route{Webhook: webhook}
	}

	if len(routes) == 0 {
		return nil, errors.New("no routes configured: set -config and/or -route")
	}

	return routes, nil
}

type routeConfig struct {
	Webhook string `json:"webhook"`
	APIKey  string `json:"api_key"`
}

func loadRoutesFile(path string) (map[string]Route, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read routes config: %w", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse routes config: %w", err)
	}

	routes := make(map[string]Route, len(raw))
	for domain, value := range raw {
		domain = normalizeDomain(domain)
		if domain == "" {
			return nil, errors.New("routes config contains an empty domain key")
		}

		route, err := parseRouteEntry(value)
		if err != nil {
			return nil, fmt.Errorf("domain %q: %w", domain, err)
		}
		routes[domain] = route
	}

	return routes, nil
}

func parseRouteEntry(value json.RawMessage) (Route, error) {
	var webhookURL string
	if err := json.Unmarshal(value, &webhookURL); err == nil {
		webhookURL = strings.TrimSpace(webhookURL)
		if err := validateWebhook(webhookURL); err != nil {
			return Route{}, fmt.Errorf("invalid webhook URL: %w", err)
		}
		return Route{Webhook: webhookURL}, nil
	}

	var cfg routeConfig
	if err := json.Unmarshal(value, &cfg); err != nil {
		return Route{}, errors.New("expected a webhook URL string or {\"webhook\":\"...\",\"api_key\":\"...\"}")
	}

	cfg.Webhook = strings.TrimSpace(cfg.Webhook)
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	if cfg.Webhook == "" {
		return Route{}, errors.New("webhook is required")
	}
	if err := validateWebhook(cfg.Webhook); err != nil {
		return Route{}, fmt.Errorf("invalid webhook URL: %w", err)
	}

	return Route{Webhook: cfg.Webhook, APIKey: cfg.APIKey}, nil
}

func normalizeDomain(domain string) string {
	return strings.ToLower(strings.TrimSpace(domain))
}

func routeForRecipient(routes map[string]Route, address string) (domain string, route Route, err error) {
	domain, err = recipientDomain(address)
	if err != nil {
		return "", Route{}, err
	}

	route, ok := routes[domain]
	if !ok {
		return domain, Route{}, fmt.Errorf("no webhook configured for domain %q", domain)
	}

	return domain, route, nil
}
