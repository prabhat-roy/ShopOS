{{/*
ShopOS — webhook-ingestion-service Helm helper templates
*/}}

{{- define "webhook-ingestion-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "webhook-ingestion-service.fullname" -}}
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

{{- define "webhook-ingestion-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "webhook-ingestion-service.labels" -}}
helm.sh/chart: {{- include "webhook-ingestion-service.chart" . }}
{{- include "webhook-ingestion-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: integrations
{{- end }}

{{- define "webhook-ingestion-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "webhook-ingestion-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
