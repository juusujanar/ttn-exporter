# ttn-exporter

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.5.0](https://img.shields.io/badge/AppVersion-1.5.0-informational?style=flat-square)

Prometheus exporter for The Things Network (TTN) v3 / The Things Stack gateway metrics

Works with:
- The Things Network v3 Community edition - tested, default
- Things Industries - untested
- self-hosted Things Stack instances - untested

See [docs/ARCHITECTURE.md](https://github.com/juusujanar/ttn-exporter/blob/main/docs/ARCHITECTURE.md)
in the main repository for how the exporter fetches data and the full list of exported metrics.

**Homepage:** <https://github.com/juusujanar/ttn-exporter>

## Prerequisites

- Kubernetes 1.24+
- Helm 3.8+ (needed for `oci://` chart installs)
- Prometheus Operator CRDs, only if you plan to set `serviceMonitor.enabled=true`

## Getting a TTN API key

To use this exporter, you need to generate a user or an organization API key. Gateway API keys
do not work, because the exporter needs permissions to dynamically load the list of gateways.

When you create an organization or user API key, grant it these rights:
- **List the gateways the organization/user is a collaborator of** - required to get a list of gateways
- **View gateway status** - to get gateway status and metrics

## Installing the chart

This chart is published as an OCI artifact to GitHub Container Registry.

```bash
helm install ttn-exporter oci://ghcr.io/juusujanar/charts/ttn-exporter \
  --version 0.1.0 \
  --set apiKey.value=<key>
```

Or, from a local checkout of the source repository:

```bash
git clone https://github.com/juusujanar/ttn-exporter.git
helm install ttn-exporter ./ttn-exporter/charts/ttn-exporter --set apiKey.value=<key>
```

## Supplying the API key

For a quick test, pass the key directly and let the chart create the Secret:

```bash
helm install ttn-exporter oci://ghcr.io/juusujanar/charts/ttn-exporter \
  --set apiKey.value=<key>
```

For production, create the Secret yourself and point the chart at it, so the key never appears
in `helm history`/release state as plain values:

```bash
kubectl create secret generic ttn-api-key --from-literal=TTN_API_KEY=<key>

helm install ttn-exporter oci://ghcr.io/juusujanar/charts/ttn-exporter \
  --set apiKey.existingSecret=ttn-api-key
```

## Connecting to a different Things Stack instance

The chart defaults to the Things Network Community Edition (`eu1.cloud.thethings.network`). To
point at another region or a self-hosted/Things Industries instance:

```bash
helm install ttn-exporter oci://ghcr.io/juusujanar/charts/ttn-exporter \
  --set apiKey.value=<key> \
  --set ttn.uri="https://<tenant>.<region>.cloud.thethings.industries/"
```

## Enabling Prometheus Operator scraping

```bash
helm install ttn-exporter oci://ghcr.io/juusujanar/charts/ttn-exporter \
  --set apiKey.value=<key> \
  --set serviceMonitor.enabled=true
```

Each scrape triggers a synchronous call chain against the TTN API (list gateways, then one stats
call per gateway, bounded by `ttn.concurrency`). For large gateway fleets, raise
`serviceMonitor.scrapeTimeout`/`serviceMonitor.interval` and/or `ttn.concurrency` together if you
see scrape timeouts.

## Upgrading

```bash
helm upgrade ttn-exporter oci://ghcr.io/juusujanar/charts/ttn-exporter --reuse-values
```

`image.tag` defaults to the chart's `appVersion`, so a chart upgrade alone can also bump the
running exporter version — check the chart's release notes before upgrading in production.

## Uninstalling

```bash
helm uninstall ttn-exporter
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity rules for pod assignment. |
| apiKey.existingSecret | string | `""` | Name of a pre-existing Secret containing the API key. Leave empty to let the chart manage the Secret via apiKey.value. |
| apiKey.existingSecretKey | string | `"TTN_API_KEY"` | Key within apiKey.existingSecret that holds the API key. |
| apiKey.value | string | `""` | TTN_API_KEY value. Ignored when apiKey.existingSecret is set. Required when secret.create is true. |
| commonLabels | object | `{}` | Extra labels added to all chart resources. |
| containerPort | int | `9981` | Port the exporter's HTTP server (landing page + metrics) listens on inside the container. |
| extraArgs | list | `[]` | Additional raw CLI arguments appended after the generated flags. |
| extraEnv | list | `[]` | Additional environment variables, as a list of full EnvVar objects. |
| extraEnvFrom | list | `[]` | Additional envFrom entries, as a list of full EnvFromSource objects. |
| extraVolumeMounts | list | `[]` | Additional volumeMounts on the container. |
| extraVolumes | list | `[]` | Additional volumes on the pod. |
| fullnameOverride | string | `""` | Override the fully qualified app name. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy. |
| image.repository | string | `"ghcr.io/juusujanar/ttn-exporter"` | Image repository. Also published to janarj/ttn-exporter on Docker Hub; override to use that instead. |
| image.tag | string | `""` | Image tag. Defaults to "v" + Chart.yaml appVersion (e.g. "v1.5.0") when empty. |
| imagePullSecrets | list | `[]` | References to secrets for pulling the image from a private registry. |
| livenessProbe | object | `{"failureThreshold":3,"httpGet":{"path":"/","port":"http"},"initialDelaySeconds":5,"periodSeconds":15,"timeoutSeconds":3}` | Liveness probe. Targets "/" (the static landing page), never the metrics path, since scraping /metrics triggers a real synchronous call to the TTN API. |
| log.format | string | `"logfmt"` | Log format (logfmt or json). Maps to the log.format flag. |
| log.level | string | `"info"` | Log level. Maps to the log.level flag. |
| nameOverride | string | `""` | Override the chart name used in generated resource names. |
| networkPolicy.egress.enabled | bool | `true` | Allow all egress. Required in practice since the TTN API host can't be scoped by CIDR; set to false to leave egress ungoverned by this policy instead. |
| networkPolicy.enabled | bool | `false` | Create a NetworkPolicy restricting ingress to the exporter port. |
| networkPolicy.from | list | `[]` | Ingress `from` peers. Empty (default) allows traffic from any source (only the port is restricted); set this to scope scraping to your monitoring namespace/pods. |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| podAnnotations | object | `{}` | Annotations to add to the pod. |
| podDisruptionBudget.enabled | bool | `false` | Create a PodDisruptionBudget. |
| podDisruptionBudget.maxUnavailable | string | `nil` | Maximum number of pods that can be unavailable. Takes effect only when minAvailable is null. |
| podDisruptionBudget.minAvailable | int | `1` | Minimum number of pods that must remain available. Mutually exclusive with maxUnavailable. |
| podLabels | object | `{}` | Extra labels to add to the pod. |
| podSecurityContext | object | `{"fsGroup":65534,"runAsGroup":65534,"runAsNonRoot":true,"runAsUser":65534,"seccompProfile":{"type":"RuntimeDefault"}}` | Pod-level security context. Defaults match the container image, which runs as the "nobody" (65534) user. |
| priorityClassName | string | `""` | Priority class name for the pod. |
| readinessProbe | object | `{"failureThreshold":3,"httpGet":{"path":"/","port":"http"},"initialDelaySeconds":5,"periodSeconds":15,"timeoutSeconds":3}` | Readiness probe. See livenessProbe for why it targets "/". |
| replicaCount | int | `1` | Number of replicas. |
| resources | object | `{"limits":{"cpu":"100m","memory":"64Mi"},"requests":{"cpu":"10m","memory":"32Mi"}}` | Container resource requests/limits. |
| secret.annotations | object | `{}` | Annotations to add to the chart-managed Secret. |
| secret.create | bool | `true` | Create a Secret from apiKey.value. Only relevant when apiKey.existingSecret is empty. |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true,"runAsGroup":65534,"runAsNonRoot":true,"runAsUser":65534}` | Container-level security context. |
| service.annotations | object | `{}` | Annotations to add to the Service (e.g. classic prometheus.io/scrape annotations). |
| service.port | int | `9981` | Service port. Also used as the target for the named "http" container port. |
| service.type | string | `"ClusterIP"` | Kubernetes Service type. |
| serviceAccount.annotations | object | `{}` | Annotations to add to the ServiceAccount. |
| serviceAccount.automount | bool | `false` | Whether to automount the ServiceAccount token. The exporter never calls the Kubernetes API, so this is false by default. |
| serviceAccount.create | bool | `true` | Whether a ServiceAccount should be created. |
| serviceAccount.name | string | `""` | Name of the ServiceAccount to use. If not set and create is true, a name is generated. |
| serviceMonitor.enabled | bool | `false` | Create a Prometheus Operator ServiceMonitor. Requires the monitoring.coreos.com CRDs. |
| serviceMonitor.honorLabels | bool | `false` |  |
| serviceMonitor.interval | string | `"60s"` | Scrape interval. Increase alongside scrapeTimeout for large gateway fleets. |
| serviceMonitor.labels | object | `{}` | Extra labels on the ServiceMonitor (e.g. to match a Prometheus Operator selector). |
| serviceMonitor.metricRelabelings | list | `[]` |  |
| serviceMonitor.namespace | string | `""` | Namespace to create the ServiceMonitor in. Defaults to the release namespace. |
| serviceMonitor.relabelings | list | `[]` |  |
| serviceMonitor.scrapeTimeout | string | `"30s"` | Scrape timeout. Each scrape is a synchronous call chain bounded by ttn.concurrency; size this together with ttn.timeout and gateway count. |
| startupProbe.enabled | bool | `false` | Enable a startup probe. Usually unnecessary; the binary starts near-instantly. |
| startupProbe.failureThreshold | int | `30` |  |
| startupProbe.httpGet.path | string | `"/"` |  |
| startupProbe.httpGet.port | string | `"http"` |  |
| startupProbe.periodSeconds | int | `2` |  |
| tolerations | list | `[]` | Tolerations for pod assignment. |
| topologySpreadConstraints | list | `[]` | Topology spread constraints for the pod. |
| ttn.concurrency | int | `5` | Maximum number of concurrent per-gateway stats requests. Maps to the ttn.concurrency flag. |
| ttn.sslVerify | bool | `true` | Verify the TLS certificate of the Things Stack API endpoint. Maps to the ttn.ssl-verify flag. |
| ttn.timeout | string | `"5s"` | Timeout for each request to the Things Stack API. Maps to the ttn.timeout flag. |
| ttn.uri | string | `"https://eu1.cloud.thethings.network/"` | Things Stack base URI. Maps to the ttn.uri flag. |
| web.configFile | string | `""` | Path to an exporter-toolkit web.config.file (TLS/basic-auth for the exporter's own HTTP server). Advanced; if set, mount the file via extraVolumes/extraVolumeMounts. |
| web.telemetryPath | string | `"/metrics"` | Path under which metrics are exposed. Maps to the web.telemetry-path flag. |

