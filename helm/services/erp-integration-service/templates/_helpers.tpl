{{/*
ShopOS — erp-integration-service Helm helper templates
*/}}

{{- define "erp-integration-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "erp-integration-service.fullname" -}}
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

{{- define "erp-integration-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "erp-integration-service.labels" -}}
helm.sh/chart: {{ include "erp-integration-service.chart" . }}
{{ include "erp-integration-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: integrations
{{- end }}

{{- define "erp-integration-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "erp-integration-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
