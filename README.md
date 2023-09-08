# Prometheus exporter for The Things Network

This collects data from The Things Stack API and exports it via HTTP for Prometheus.

Works with:
- The Things Network v3 Community edition - tested, default
- Things Industries - untested
- self-hosted Things Stack instances - untested

## Badges

[![Release](https://img.shields.io/github/release/juusujanar/ttn-exporter.svg?style=for-the-badge)](https://github.com/juusujanar/ttn-exporter/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE)
[![Build status](https://img.shields.io/github/actions/workflow/status/juusujanar/ttn-exporter/release.yml?style=for-the-badge&branch=main)](https://github.com/juusujanar/ttn-exporter/actions?workflow=Release)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/juusujanar/ttn-exporter)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)

## Quick start

To run the exporter:

```bash
export TTN_API_KEY=<key>
./ttn_exporter
```

If you want to use Docker:

```bash
docker run -p 9981:9981 -e TTN_API_KEY=<key> ghcr.io/juusujanar/ttn-exporter:v1.0.4
```

Help on flags:

```bash
./ttn_exporter --help
```

## Usage

### API keys

To use this exporter, you need to generate an API key and grant it the following rights:
- **List the gateways the organization is a collaborator of** - required when consuming organization gateways
- **View gateway status** - to get gateway status and metrics

### TTN URL

Specify custom URLs for which Things Stack instance to use using the `--ttn.uri`
flag. Exporter defaults to Things Network Community Edition in Europe (eu1.cloud.thethings.network).
For example, if you want to use enterprise Things Industries,

```bash
TTN_API_KEY=<key> ttn_exporter --ttn.uri="https://<tenant>.<region>.cloud.thethings.industries/"
```

### Docker

To run the exporter as a Docker container, run:

```bash
docker run -p 9981:9981 -e TTN_API_KEY=<key> ghcr.io/juusujanar/ttn-exporter:v1.0.4 --ttn.uri="https://<tenant>.<region>.cloud.thethings.industries/"
```

[docker hub]: https://hub.docker.com/r/janarj/ttn-exporter/
[github]: https://ghcr.io/repository/juusujanar/ttn-exporter
