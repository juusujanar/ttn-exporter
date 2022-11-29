# Prometheus exporter for The Things Network

This collects data from The Things Stack API and exports it via HTTP for Prometheus.

Works with:
- The Things Network v3 Community edition - tested, default
- Things Industries - untested
- self-hosted Things Stack instances - untested

## Getting started

To run the exporter:

```bash
export TTN_API_KEY=<key>
./ttn_exporter [flags]
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
docker run -p 9981:9981 ghcr.io/juusujanar/ttn-exporter:v1.0.0 --ttn.uri="https://<tenant>.<region>.cloud.thethings.industries/" --ttn.api-key="<key>"
```

[docker hub]: https://hub.docker.com/r/janarj/ttn-exporter/
[github]: https://ghcr.io/repository/juusujanar/ttn-exporter

## License

MIT License, see [LICENSE](https://github.com/juusujanar/ttn-exporter/blob/master/LICENSE).
