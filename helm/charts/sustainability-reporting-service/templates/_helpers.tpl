{{/*
ShopOS — sustainability-reporting-service Helm helper templates
*/}}

{{- define "sustainability-reporting-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "sustainability-reporting-service.fullname" -}}
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

{{- define "sustainability-reporting-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "sustainability-reporting-service.labels" -}}
helm.sh/chart: {{- include "sustainability-reporting-service.chart" . }}
{{- include "sustainability-reporting-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: sustainability
{{- end }}

{{- define "sustainability-reporting-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "sustainability-reporting-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
