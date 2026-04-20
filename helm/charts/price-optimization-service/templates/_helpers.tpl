{{/*
ShopOS — price-optimization-service Helm helper templates
*/}}

{{- define "price-optimization-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "price-optimization-service.fullname" -}}
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

{{- define "price-optimization-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "price-optimization-service.labels" -}}
helm.sh/chart: {{ include "price-optimization-service.chart" . }}
{{ include "price-optimization-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: analytics-ai
{{- end }}

{{- define "price-optimization-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "price-optimization-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
