{{/*
ShopOS — product-syndication-service Helm helper templates
*/}}

{{- define "product-syndication-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "product-syndication-service.fullname" -}}
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

{{- define "product-syndication-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "product-syndication-service.labels" -}}
helm.sh/chart: {{- include "product-syndication-service.chart" . }}
{{- include "product-syndication-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: marketplace
{{- end }}

{{- define "product-syndication-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "product-syndication-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
