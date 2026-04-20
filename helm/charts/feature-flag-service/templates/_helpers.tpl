{{/*
ShopOS — feature-flag-service Helm helper templates
*/}}

{{- define "feature-flag-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "feature-flag-service.fullname" -}}
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

{{- define "feature-flag-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "feature-flag-service.labels" -}}
helm.sh/chart: {{ include "feature-flag-service.chart" . }}
{{ include "feature-flag-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: platform
{{- end }}

{{- define "feature-flag-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "feature-flag-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
