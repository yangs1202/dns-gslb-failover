package config

import "testing"

func TestParseEndpoints(t *testing.T) {
	t.Parallel()

	endpoints, err := parseEndpoints("region-a=https://example-a.invalid/healthz,region-b=http://example-b.invalid/healthz")
	if err != nil {
		t.Fatalf("parseEndpoints returned error: %v", err)
	}

	if len(endpoints) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(endpoints))
	}
	if endpoints[0].RegionID != "region-a" {
		t.Fatalf("expected first region region-a, got %q", endpoints[0].RegionID)
	}
}

func TestParseEndpointsRejectsDuplicateRegions(t *testing.T) {
	t.Parallel()

	_, err := parseEndpoints("region-a=https://example-a.invalid/healthz,region-a=https://example-b.invalid/healthz")
	if err == nil {
		t.Fatal("expected duplicate region error")
	}
}

func TestParseEndpointsRejectsUnsupportedScheme(t *testing.T) {
	t.Parallel()

	_, err := parseEndpoints("region-a=tcp://example-a.invalid:443")
	if err == nil {
		t.Fatal("expected unsupported scheme error")
	}
}
