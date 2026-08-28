# Architecture

This document explains how `ttn-exporter` works internally: how it fetches data, what
metrics it produces, and how to configure it. For install/quick-start instructions, see
the main [README](../README.md).

## Overview

`ttn-exporter` is a Prometheus exporter for [The Things Stack](https://www.thethingsindustries.com/docs/)
(the LoRaWAN network server behind The Things Network). It polls the Things Stack v3 HTTP
API to list the gateways an API key can see, fetches each gateway's connection statistics,
and exposes the results as Prometheus metrics.

There is no TTN/LoRaWAN SDK dependency — the API client is a small hand-rolled
`net/http` + `encoding/json` layer in [`collector/`](../collector), talking to:

- `GET /api/v3/gateways` — list gateways
- `GET /api/v3/gs/gateways/{gateway_id}/connection/stats` — per-gateway connection stats

## Request flow

The exporter is **pull-based with no caching layer**: every scrape of `/metrics` triggers a
live round trip to the Things Stack API. There is no background polling goroutine — all
work happens synchronously inside the HTTP request handling the scrape.

```
Prometheus scrape
  → Exporter.Collect()                    ttn_exporter.go:133
    → Exporter.scrape()                   ttn_exporter.go:140
      → collector.GetInfo()               collector/collect.go:23
        → getGatewayList()                collector/gateways_list.go:29
            GET /api/v3/gateways (paginated, 100/page)
        → getGatewayStats() × N gateways  collector/gateway_stats.go:74
            GET /api/v3/gs/gateways/{id}/connection/stats
            (run concurrently, bounded by --ttn.concurrency)
```

1. **`main()`** (`ttn_exporter.go:215`) parses flags and the `TTN_API_KEY` environment
   variable, builds an `Exporter`, and registers it — together with the standard Go,
   process, and version collectors — on a fresh `prometheus.Registry`. It then serves that
   registry over HTTP.

2. **`getGatewayList`** (`collector/gateways_list.go:29`) calls `GET /api/v3/gateways` with
   `limit=100&field_mask=name`, incrementing `page` and appending results until a page
   comes back with fewer than 100 gateways. If this call fails (network error or non-200
   response), the whole scrape fails: `ttn_up` is reported as `0` and no gateway metrics
   are emitted for that scrape.

3. **`getGatewayStats`** (`collector/gateway_stats.go:74`) is called once per gateway to
   fetch `GET /api/v3/gs/gateways/{gateway_id}/connection/stats`. These calls run
   concurrently via `golang.org/x/sync/errgroup`, bounded by `g.SetLimit(concurrency)`
   (`collector/collect.go:31-32`), where `concurrency` defaults to 5 and is configurable
   via `--ttn.concurrency`. `http.Transport.MaxIdleConnsPerHost` is set to the same
   concurrency value (`ttn_exporter.go:109`) so the concurrent requests can reuse
   keep-alive connections instead of reconnecting each time.

   Unlike the gateway list, a single gateway's stats failing does **not** fail the whole
   scrape — that gateway is simply reported as disconnected (`ttn_gateway_connected 0`)
   and a warning is logged (`collector/collect.go:38-40`).

4. **Rate-limit handling** (`collector/ratelimit.go`): every outbound request (list and
   per-gateway) goes through `doWithRateLimitRetry`. On an HTTP `429` response, it retries
   up to 2 more times, waiting according to the `X-Rate-Limit-Retry` response header
   (seconds), defaulting to 1s if absent and capped at 10s, logging a warning on each
   retry. TTN's actual rate limits are undocumented; if you see these warnings in the
   logs, lower `--ttn.concurrency`. Requests retried this way must have a nil body — true
   for all requests this exporter makes, since it only performs GETs.

## Metrics

All custom metric descriptors live in `ttn_exporter.go:34-78` under the `ttn` namespace,
and are populated in `scrape()` / `parseVersion()` (`ttn_exporter.go:140-206`).

| Metric | Type | Labels | Description |
|---|---|---|---|
| `ttn_gateway_connected` | Gauge | `gateway_id`, `name`, `protocol` | `1` if the gateway is connected, else `0`. When disconnected, `protocol` is emitted as an empty string. |
| `ttn_gateway_uplink_count` | Counter | `gateway_id`, `name` | Number of uplink packets received by the gateway. Only emitted for connected gateways; skipped (with a logged warning) if the API's string value fails to parse as a number. |
| `ttn_gateway_downlink_count` | Counter | `gateway_id`, `name` | Number of downlink packets sent by the gateway. Same parsing/skip behavior as above. |
| `ttn_gateway_tx_acknowledgement_count` | Counter | `gateway_id`, `name` | Number of TX acknowledgements received by the gateway. Same parsing/skip behavior as above. |
| `ttn_gateway_round_trip_min` | Gauge | `gateway_id`, `name` | Minimum measured round-trip time between gateway and TTN, in seconds. |
| `ttn_gateway_round_trip_max` | Gauge | `gateway_id`, `name` | Maximum measured round-trip time, in seconds. |
| `ttn_gateway_round_trip_median` | Gauge | `gateway_id`, `name` | Median measured round-trip time, in seconds. |
| `ttn_gateway_round_trip_count` | Gauge | `gateway_id`, `name` | Number of round-trip measurements the above stats are based on. |
| `ttn_version_info` | Gauge | `version` | Always `1`; `version` is taken from the first gateway in the response that reports a `ttn-lw-gateway-server` version, falling back to `firmware` version if not present. |
| `ttn_up` | Gauge | none | `1` if the last scrape (gateway list + stats fetch) completed without error, else `0`. |
| `exporter_scrapes_total` | Counter | none | Total number of scrapes attempted. **Note:** unlike every other custom metric, this one is defined without the `ttn` namespace prefix (`ttn_exporter.go:112-115`), so it is exposed literally as `exporter_scrapes_total`, not `ttn_exporter_scrapes_total`. |

Round-trip metrics (`ttn_gateway_round_trip_*`) are only emitted when the Things Stack
response includes a `round_trip_times` object — some gateways/protocols may not report it.

In addition to the metrics above, the exporter also registers standard Prometheus
collectors on the same registry (`ttn_exporter.go:246-252`):

- `go_*` — Go runtime metrics (`collectors.NewGoCollector()`)
- `process_*` — process metrics such as CPU/memory/fds (`collectors.NewProcessCollector()`)
- `ttn_exporter_build_info` — build metadata (`version`, `revision`, `branch`, `goversion`
  labels) via `versioncollector.NewCollector("ttn_exporter")`

## Configuration

### Environment variable

| Variable | Required | Purpose |
|---|---|---|
| `TTN_API_KEY` | Yes | Bearer token sent as `Authorization: Bearer <key>` on every request to the Things Stack API. Deliberately an environment variable rather than a flag, so it doesn't leak via process listings. The process exits immediately if unset (`ttn_exporter.go:236-239`). |

The API key must be a **user or organization** key (a gateway key cannot list all
gateways) with these rights granted:
- List the gateways the user/organization is a collaborator of
- View gateway status

### Flags

| Flag | Default | Purpose |
|---|---|---|
| `--web.listen-address` | `:9981` | Address(es) to listen on for HTTP traffic. |
| `--web.config.file` | — | Path to an [exporter-toolkit web config](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md) for enabling TLS and/or basic auth on the exporter's own HTTP server. |
| `--web.telemetry-path` | `/metrics` | Path under which metrics are exposed. |
| `--ttn.uri` | `https://eu1.cloud.thethings.network/` | Base URI of the Things Stack instance to query. Change this to point at a Things Industries tenant or a self-hosted Things Stack. |
| `--ttn.ssl-verify` | `true` | Whether to verify the TLS certificate of the TTN API endpoint. |
| `--ttn.timeout` | `5s` | HTTP client timeout applied to each request made to the TTN API. |
| `--ttn.concurrency` | `5` | Maximum number of concurrent per-gateway stats requests. Lower this if the logs show TTN rate-limit warnings. |
| `--log.level`, `--log.format` | — | Standard `promslog` logging flags. |

Run `./ttn_exporter --help` for the full, current list (including standard
`exporter-toolkit`/`kingpin` flags such as `--web.systemd-socket` on Linux).

## HTTP endpoints

- `/` — a landing page (via `prometheus/exporter-toolkit/web.NewLandingPage`,
  `ttn_exporter.go:254-264`) showing the exporter name, version, and a link to the metrics
  path.
- `<web.telemetry-path>` (default `/metrics`) — the Prometheus metrics endpoint, served by
  `promhttp.HandlerFor`.
- `/debug/pprof/*` — Go's standard profiling endpoints, exposed as a side effect of
  importing `net/http/pprof` (`ttn_exporter.go:8`) on the default mux. Useful for
  debugging, but be aware it's reachable on the same listener as `/metrics`.

## Build & release

- Local builds go through the `Makefile` (`make build`, `make run`, `make compile` for
  cross-compilation).
- Releases are built by [GoReleaser](../.goreleaser.yaml): cross-platform binaries
  (linux/windows/darwin × amd64/arm/arm64), triggered by `.github/workflows/release.yml`
  on tag push.
- The [`Dockerfile`](../Dockerfile) does not compile Go code itself — it packages a
  pre-built, per-platform binary produced by GoReleaser into a `scratch` image that runs
  as the unprivileged `nobody` user and exposes port `9981`.
- Images are published to `ghcr.io/juusujanar/ttn-exporter` and `janarj/ttn-exporter`
  (Docker Hub).
