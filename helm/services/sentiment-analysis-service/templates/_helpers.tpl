{{/*
ShopOS — sentiment-analysis-service Helm helper templates
*/}}

{{- define "sentiment-analysis-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "sentiment-analysis-service.fullname" -}}
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

{{- define "sentiment-analysis-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "sentiment-analysis-service.labels" -}}
helm.sh/chart: {{ include "sentiment-analysis-service.chart" . }}
{{ include "sentiment-analysis-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: analytics-ai
{{- end }}

{{- define "sentiment-analysis-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "sentiment-analysis-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
