{{/*
ShopOS — fraud-prevention-affiliate-service Helm helper templates
*/}}

{{- define "fraud-prevention-affiliate-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "fraud-prevention-affiliate-service.fullname" -}}
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

{{- define "fraud-prevention-affiliate-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "fraud-prevention-affiliate-service.labels" -}}
helm.sh/chart: {{- include "fraud-prevention-affiliate-service.chart" . }}
{{- include "fraud-prevention-affiliate-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: affiliate
{{- end }}

{{- define "fraud-prevention-affiliate-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "fraud-prevention-affiliate-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
