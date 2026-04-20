{{/*
ShopOS — tax-reporting-service Helm helper templates
*/}}

{{- define "tax-reporting-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "tax-reporting-service.fullname" -}}
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

{{- define "tax-reporting-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "tax-reporting-service.labels" -}}
helm.sh/chart: {{ include "tax-reporting-service.chart" . }}
{{ include "tax-reporting-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: financial
{{- end }}

{{- define "tax-reporting-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tax-reporting-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
