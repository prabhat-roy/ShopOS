{{/*
ShopOS — marketplace-connector-service Helm helper templates
*/}}

{{- define "marketplace-connector-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "marketplace-connector-service.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "marketplace-connector-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "marketplace-connector-service.labels" -}}
helm.sh/chart: {{ include "marketplace-connector-service.chart" . }}
{{ include "marketplace-connector-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: integrations
{{- end }}

{{- define "marketplace-connector-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "marketplace-connector-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
