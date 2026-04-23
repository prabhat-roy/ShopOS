{{/*
ShopOS — dynamic-pricing-service Helm helper templates
*/}}

{{- define "dynamic-pricing-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "dynamic-pricing-service.fullname" -}}
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

{{- define "dynamic-pricing-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "dynamic-pricing-service.labels" -}}
helm.sh/chart: {{- include "dynamic-pricing-service.chart" . }}
{{- include "dynamic-pricing-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: commerce
{{- end }}

{{- define "dynamic-pricing-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "dynamic-pricing-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
