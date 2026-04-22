{{/*
ShopOS — privacy-request-service Helm helper templates
*/}}

{{- define "privacy-request-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "privacy-request-service.fullname" -}}
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

{{- define "privacy-request-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "privacy-request-service.labels" -}}
helm.sh/chart: {{- include "privacy-request-service.chart" . }}
{{- include "privacy-request-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: compliance
{{- end }}

{{- define "privacy-request-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "privacy-request-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
