{{/*
ShopOS — data-pipeline-service Helm helper templates
*/}}

{{- define "data-pipeline-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "data-pipeline-service.fullname" -}}
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

{{- define "data-pipeline-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "data-pipeline-service.labels" -}}
helm.sh/chart: {{ include "data-pipeline-service.chart" . }}
{{ include "data-pipeline-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: analytics-ai
{{- end }}

{{- define "data-pipeline-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "data-pipeline-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
