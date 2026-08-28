{{/*
Expand the name of the chart.
*/}}
{{- define "ttn-exporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "ttn-exporter.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ttn-exporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels.
*/}}
{{- define "ttn-exporter.labels" -}}
helm.sh/chart: {{ include "ttn-exporter.chart" . }}
{{ include "ttn-exporter.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{/*
Selector labels. Kept minimal and stable since they back the immutable Deployment selector.
*/}}
{{- define "ttn-exporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ttn-exporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Name of the ServiceAccount to use.
*/}}
{{- define "ttn-exporter.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "ttn-exporter.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{/*
Fully qualified image reference. Falls back to "v" + Chart.AppVersion since the project's
image tags are v-prefixed (vX.Y.Z) while appVersion follows the un-prefixed Helm convention.
*/}}
{{- define "ttn-exporter.image" -}}
{{- $tag := .Values.image.tag | default (printf "v%s" .Chart.AppVersion) -}}
{{- printf "%s:%s" .Values.image.repository $tag -}}
{{- end -}}

{{/*
Name of the Secret holding TTN_API_KEY.
*/}}
{{- define "ttn-exporter.secretName" -}}
{{- if .Values.apiKey.existingSecret -}}
{{- .Values.apiKey.existingSecret -}}
{{- else -}}
{{- include "ttn-exporter.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
Key within the Secret holding TTN_API_KEY.
*/}}
{{- define "ttn-exporter.secretKey" -}}
{{- if .Values.apiKey.existingSecret -}}
{{- .Values.apiKey.existingSecretKey | default "TTN_API_KEY" -}}
{{- else -}}
TTN_API_KEY
{{- end -}}
{{- end -}}

{{/*
CLI arguments passed to the container, assembled from values.
*/}}
{{- define "ttn-exporter.args" -}}
- --web.listen-address=:{{ .Values.containerPort }}
- --web.telemetry-path={{ .Values.web.telemetryPath }}
- --ttn.uri={{ .Values.ttn.uri }}
- --ttn.ssl-verify={{ .Values.ttn.sslVerify }}
- --ttn.timeout={{ .Values.ttn.timeout }}
- --ttn.concurrency={{ .Values.ttn.concurrency }}
- --log.level={{ .Values.log.level }}
- --log.format={{ .Values.log.format }}
{{- if .Values.web.configFile }}
- --web.config.file={{ .Values.web.configFile }}
{{- end }}
{{- range .Values.extraArgs }}
- {{ . }}
{{- end }}
{{- end -}}
