{{/*
ShopOS — oauth-client-service Helm helper templates
*/}}

{{- define "oauth-client-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "oauth-client-service.fullname" -}}
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

{{- define "oauth-client-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "oauth-client-service.labels" -}}
helm.sh/chart: {{- include "oauth-client-service.chart" . }}
{{- include "oauth-client-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: developer-platform
{{- end }}

{{- define "oauth-client-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "oauth-client-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
