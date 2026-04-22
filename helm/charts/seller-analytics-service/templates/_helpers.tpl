{{/*
ShopOS — seller-analytics-service Helm helper templates
*/}}

{{- define "seller-analytics-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "seller-analytics-service.fullname" -}}
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

{{- define "seller-analytics-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "seller-analytics-service.labels" -}}
helm.sh/chart: {{- include "seller-analytics-service.chart" . }}
{{- include "seller-analytics-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: marketplace
{{- end }}

{{- define "seller-analytics-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "seller-analytics-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
