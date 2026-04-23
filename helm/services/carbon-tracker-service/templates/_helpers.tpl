{{/*
ShopOS — carbon-tracker-service Helm helper templates
*/}}

{{- define "carbon-tracker-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "carbon-tracker-service.fullname" -}}
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

{{- define "carbon-tracker-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "carbon-tracker-service.labels" -}}
helm.sh/chart: {{- include "carbon-tracker-service.chart" . }}
{{- include "carbon-tracker-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: sustainability
{{- end }}

{{- define "carbon-tracker-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "carbon-tracker-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
