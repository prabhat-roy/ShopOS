{{/*
ShopOS — vendor-onboarding-service Helm helper templates
*/}}

{{- define "vendor-onboarding-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "vendor-onboarding-service.fullname" -}}
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

{{- define "vendor-onboarding-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "vendor-onboarding-service.labels" -}}
helm.sh/chart: {{- include "vendor-onboarding-service.chart" . }}
{{- include "vendor-onboarding-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: b2b
{{- end }}

{{- define "vendor-onboarding-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "vendor-onboarding-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
