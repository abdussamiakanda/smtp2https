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

// Route is the webhook target and optional per-domain options.
// Extra keys in routes.json (besides webhook and api_key) are merged into the webhook JSON body.
type Route struct {
	Webhook string
	APIKey  string
	Extra   map[string]interface{}
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

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(value, &obj); err != nil {
		return Route{}, errors.New("expected a webhook URL string or {\"webhook\":\"...\", ...}")
	}

	webhookRaw, ok := obj["webhook"]
	if !ok {
		return Route{}, errors.New("webhook is required")
	}

	var webhook string
	if err := json.Unmarshal(webhookRaw, &webhook); err != nil {
		return Route{}, errors.New("webhook must be a string URL")
	}
	webhook = strings.TrimSpace(webhook)
	if webhook == "" {
		return Route{}, errors.New("webhook is required")
	}
	if err := validateWebhook(webhook); err != nil {
		return Route{}, fmt.Errorf("invalid webhook URL: %w", err)
	}

	var apiKey string
	if raw, ok := obj["api_key"]; ok {
		if err := json.Unmarshal(raw, &apiKey); err != nil {
			return Route{}, errors.New("api_key must be a string")
		}
		apiKey = strings.TrimSpace(apiKey)
	}

	extra := make(map[string]interface{})
	for key, raw := range obj {
		if key == "webhook" || key == "api_key" {
			continue
		}
		var v interface{}
		if err := json.Unmarshal(raw, &v); err != nil {
			return Route{}, fmt.Errorf("invalid value for field %q", key)
		}
		extra[key] = v
	}

	return Route{
		Webhook: webhook,
		APIKey:  apiKey,
		Extra:   extra,
	}, nil
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
