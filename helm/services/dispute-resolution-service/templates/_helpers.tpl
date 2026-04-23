{{/*
ShopOS — dispute-resolution-service Helm helper templates
*/}}

{{- define "dispute-resolution-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "dispute-resolution-service.fullname" -}}
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

{{- define "dispute-resolution-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "dispute-resolution-service.labels" -}}
helm.sh/chart: {{- include "dispute-resolution-service.chart" . }}
{{- include "dispute-resolution-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: marketplace
{{- end }}

{{- define "dispute-resolution-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "dispute-resolution-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
