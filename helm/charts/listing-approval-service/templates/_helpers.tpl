{{/*
ShopOS — listing-approval-service Helm helper templates
*/}}

{{- define "listing-approval-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "listing-approval-service.fullname" -}}
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

{{- define "listing-approval-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "listing-approval-service.labels" -}}
helm.sh/chart: {{- include "listing-approval-service.chart" . }}
{{- include "listing-approval-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: marketplace
{{- end }}

{{- define "listing-approval-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "listing-approval-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
