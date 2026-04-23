{{/*
ShopOS — route-optimization-service Helm helper templates
*/}}

{{- define "route-optimization-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "route-optimization-service.fullname" -}}
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

{{- define "route-optimization-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "route-optimization-service.labels" -}}
helm.sh/chart: {{- include "route-optimization-service.chart" . }}
{{- include "route-optimization-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: supply-chain
{{- end }}

{{- define "route-optimization-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "route-optimization-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
