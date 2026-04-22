{{/*
ShopOS — developer-portal-backend Helm helper templates
*/}}

{{- define "developer-portal-backend.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "developer-portal-backend.fullname" -}}
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

{{- define "developer-portal-backend.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "developer-portal-backend.labels" -}}
helm.sh/chart: {{- include "developer-portal-backend.chart" . }}
{{- include "developer-portal-backend.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: developer-platform
{{- end }}

{{- define "developer-portal-backend.selectorLabels" -}}
app.kubernetes.io/name: {{- include "developer-portal-backend.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
