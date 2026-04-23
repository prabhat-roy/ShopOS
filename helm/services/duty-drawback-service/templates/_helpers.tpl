{{/*
ShopOS — duty-drawback-service Helm helper templates
*/}}

{{- define "duty-drawback-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "duty-drawback-service.fullname" -}}
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

{{- define "duty-drawback-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "duty-drawback-service.labels" -}}
helm.sh/chart: {{- include "duty-drawback-service.chart" . }}
{{- include "duty-drawback-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: supply-chain
{{- end }}

{{- define "duty-drawback-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "duty-drawback-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
