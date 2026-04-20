{{/*
ShopOS — support-ticket-service Helm helper templates
*/}}

{{- define "support-ticket-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "support-ticket-service.fullname" -}}
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

{{- define "support-ticket-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "support-ticket-service.labels" -}}
helm.sh/chart: {{ include "support-ticket-service.chart" . }}
{{ include "support-ticket-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: customer-experience
{{- end }}

{{- define "support-ticket-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "support-ticket-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
