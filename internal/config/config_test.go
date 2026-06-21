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

func TestParseDNSTargets(t *testing.T) {
	t.Parallel()

	targets, err := parseDNSTargets("region-a=region-a.example.invalid,region-b=region-b.example.invalid.")
	if err != nil {
		t.Fatalf("parseDNSTargets returned error: %v", err)
	}

	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	if targets[1].Name != "region-b.example.invalid" {
		t.Fatalf("expected trailing dot to be trimmed, got %q", targets[1].Name)
	}
}

func TestParseDNSTargetsRejectsURLs(t *testing.T) {
	t.Parallel()

	_, err := parseDNSTargets("region-a=https://region-a.example.invalid")
	if err == nil {
		t.Fatal("expected URL rejection error")
	}
}

func TestValidateRegionSetsRequiresMatchingRegions(t *testing.T) {
	t.Parallel()

	err := validateRegionSets(
		[]Endpoint{{RegionID: "region-a", URL: "https://example-a.invalid/healthz"}},
		[]DNSTarget{{RegionID: "region-b", Name: "region-b.example.invalid"}},
	)
	if err == nil {
		t.Fatal("expected mismatched region error")
	}
}

func TestParseRegionPriority(t *testing.T) {
	t.Parallel()

	priority, err := parseRegionPriority("region-a,region-b,region-c")
	if err != nil {
		t.Fatalf("parseRegionPriority returned error: %v", err)
	}

	if len(priority) != 3 {
		t.Fatalf("expected 3 priority entries, got %d", len(priority))
	}
	if priority[0] != "region-a" {
		t.Fatalf("expected first priority region-a, got %q", priority[0])
	}
}

func TestValidateRegionPriorityRequiresAllRegions(t *testing.T) {
	t.Parallel()

	err := validateRegionPriority(
		[]string{"region-a"},
		[]Endpoint{
			{RegionID: "region-a", URL: "https://example-a.invalid/healthz"},
			{RegionID: "region-b", URL: "https://example-b.invalid/healthz"},
		},
	)
	if err == nil {
		t.Fatal("expected missing priority region error")
	}
}

func TestParseDNSNames(t *testing.T) {
	t.Parallel()

	names, err := parseDNSNames("app.example.invalid.,api.example.invalid", "TEST_DNS_NAMES")
	if err != nil {
		t.Fatalf("parseDNSNames returned error: %v", err)
	}

	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	if names[0] != "app.example.invalid" {
		t.Fatalf("expected trailing dot to be trimmed, got %q", names[0])
	}
}

func TestConfigSupportsThreeRegionCNAMEScenario(t *testing.T) {
	t.Setenv("DNS_FAILOVER_REGION_ID", "region-a")
	t.Setenv("DNS_FAILOVER_REGION_ENDPOINTS", "region-a=https://region-a.example.invalid/ncm-cgi/health,region-b=https://region-b.example.invalid/ncm-cgi/health,region-c=https://region-c.example.invalid/ncm-cgi/health")
	t.Setenv("DNS_FAILOVER_REGION_DNS_TARGETS", "region-a=region-a.example.invalid,region-b=region-b.example.invalid,region-c=region-c.example.invalid")
	t.Setenv("DNS_FAILOVER_REGION_PRIORITY", "region-a,region-b,region-c")
	t.Setenv("DNS_FAILOVER_SERVICE_RECORDS", "app.example.invalid")
	t.Setenv("CLOUDFLARE_RECORD_NAME", "vip.example.invalid")
	t.Setenv("CLOUDFLARE_RECORD_TYPE", "CNAME")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}

	if cfg.Cloudflare.RecordName != "vip.example.invalid" {
		t.Fatalf("expected vip CNAME record, got %q", cfg.Cloudflare.RecordName)
	}
	if cfg.Endpoints[0].URL != "https://region-a.example.invalid/ncm-cgi/health" {
		t.Fatalf("expected configured health check path, got %q", cfg.Endpoints[0].URL)
	}
	if cfg.RegionPriority[0] != "region-a" {
		t.Fatalf("expected master priority region-a, got %q", cfg.RegionPriority[0])
	}
	if cfg.ServiceRecords[0] != "app.example.invalid" {
		t.Fatalf("expected service alias, got %q", cfg.ServiceRecords[0])
	}
}
