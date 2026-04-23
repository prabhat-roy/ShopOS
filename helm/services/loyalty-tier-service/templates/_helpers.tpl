{{/*
ShopOS — loyalty-tier-service Helm helper templates
*/}}

{{- define "loyalty-tier-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "loyalty-tier-service.fullname" -}}
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

{{- define "loyalty-tier-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "loyalty-tier-service.labels" -}}
helm.sh/chart: {{- include "loyalty-tier-service.chart" . }}
{{- include "loyalty-tier-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: customer-experience
{{- end }}

{{- define "loyalty-tier-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "loyalty-tier-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
