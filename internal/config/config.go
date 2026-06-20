package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	RegionID      string
	Endpoints     []Endpoint
	HealthTimeout time.Duration
	Cloudflare    CloudflareConfig
}

type Endpoint struct {
	RegionID string
	URL      string
}

type CloudflareConfig struct {
	APIToken   string
	ZoneID     string
	RecordID   string
	RecordName string
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		RegionID:      strings.TrimSpace(os.Getenv("GSLB_REGION_ID")),
		HealthTimeout: 2 * time.Second,
		Cloudflare: CloudflareConfig{
			APIToken:   os.Getenv("CLOUDFLARE_API_TOKEN"),
			ZoneID:     os.Getenv("CLOUDFLARE_ZONE_ID"),
			RecordID:   os.Getenv("CLOUDFLARE_RECORD_ID"),
			RecordName: os.Getenv("CLOUDFLARE_RECORD_NAME"),
		},
	}

	if cfg.RegionID == "" {
		return Config{}, fmt.Errorf("GSLB_REGION_ID is required")
	}

	if timeoutText := strings.TrimSpace(os.Getenv("GSLB_HEALTH_TIMEOUT")); timeoutText != "" {
		timeout, err := time.ParseDuration(timeoutText)
		if err != nil {
			return Config{}, fmt.Errorf("parse GSLB_HEALTH_TIMEOUT: %w", err)
		}
		if timeout <= 0 {
			return Config{}, fmt.Errorf("GSLB_HEALTH_TIMEOUT must be positive")
		}
		cfg.HealthTimeout = timeout
	}

	endpoints, err := parseEndpoints(os.Getenv("GSLB_REGION_ENDPOINTS"))
	if err != nil {
		return Config{}, err
	}
	cfg.Endpoints = endpoints

	return cfg, nil
}

func parseEndpoints(raw string) ([]Endpoint, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("GSLB_REGION_ENDPOINTS is required")
	}

	parts := strings.Split(raw, ",")
	endpoints := make([]Endpoint, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, part := range parts {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			return nil, fmt.Errorf("endpoint %q must use region_id=url format", part)
		}

		regionID := strings.TrimSpace(key)
		endpointURL := strings.TrimSpace(value)
		if regionID == "" || endpointURL == "" {
			return nil, fmt.Errorf("endpoint %q has empty region_id or url", part)
		}
		if _, exists := seen[regionID]; exists {
			return nil, fmt.Errorf("duplicate endpoint region_id %q", regionID)
		}

		parsedURL, err := url.ParseRequestURI(endpointURL)
		if err != nil {
			return nil, fmt.Errorf("parse endpoint url for %q: %w", regionID, err)
		}
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return nil, fmt.Errorf("endpoint %q must use http or https", regionID)
		}
		if parsedURL.Host == "" {
			return nil, fmt.Errorf("endpoint %q must include host", regionID)
		}

		seen[regionID] = struct{}{}
		endpoints = append(endpoints, Endpoint{
			RegionID: regionID,
			URL:      endpointURL,
		})
	}

	return endpoints, nil
}
