# dns-gslb-failover

Simple DNS-based GSLB failover agent for three-region deployments.

The intended design is:

1. Each region runs the same agent.
2. Agents perform HTTP health checks against every regional endpoint.
3. A healthy endpoint must return `200 OK`.
4. Agents write observations to `etcd`.
5. Only the `etcd` leader with quorum may decide failover.
6. The leader updates Cloudflare DNS to point at the selected active VIP.

## Security posture

This repository is public and must not contain private infrastructure details.

- No real region names, public IPs, internal domains, Cloudflare tokens, or Vault paths.
- Cloudflare credentials are read from environment variables.
- Vault integration is intentionally out of scope for the public version.
- Example configuration uses `.invalid` domains only.

## Environment

```sh
GSLB_REGION_ID=region-a
GSLB_REGION_ENDPOINTS=region-a=https://example-a.invalid/healthz,region-b=https://example-b.invalid/healthz,region-c=https://example-c.invalid/healthz
GSLB_HEALTH_TIMEOUT=2s
CLOUDFLARE_API_TOKEN=...
CLOUDFLARE_ZONE_ID=...
CLOUDFLARE_RECORD_ID=...
CLOUDFLARE_RECORD_NAME=service.example.com
```

## Current status

Initial scaffold only:

- ENV-based configuration
- HTTP `200 OK` health checker
- agent entrypoint that prints regional health observations

Planned next steps:

- `etcd` observation storage
- `etcd` lease-based leader election
- quorum-gated failover decision
- Cloudflare DNS update client
