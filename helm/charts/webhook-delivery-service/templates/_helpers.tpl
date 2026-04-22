{{/*
ShopOS — webhook-delivery-service Helm helper templates
*/}}

{{- define "webhook-delivery-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "webhook-delivery-service.fullname" -}}
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

{{- define "webhook-delivery-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "webhook-delivery-service.labels" -}}
helm.sh/chart: {{- include "webhook-delivery-service.chart" . }}
{{- include "webhook-delivery-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: communications
{{- end }}

{{- define "webhook-delivery-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "webhook-delivery-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
