{{/*
ShopOS — click-tracking-service Helm helper templates
*/}}

{{- define "click-tracking-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "click-tracking-service.fullname" -}}
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

{{- define "click-tracking-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "click-tracking-service.labels" -}}
helm.sh/chart: {{- include "click-tracking-service.chart" . }}
{{- include "click-tracking-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: affiliate
{{- end }}

{{- define "click-tracking-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "click-tracking-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
