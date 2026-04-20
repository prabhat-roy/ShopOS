{{/*
ShopOS — approval-workflow-service Helm helper templates
*/}}

{{- define "approval-workflow-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "approval-workflow-service.fullname" -}}
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

{{- define "approval-workflow-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "approval-workflow-service.labels" -}}
helm.sh/chart: {{ include "approval-workflow-service.chart" . }}
{{ include "approval-workflow-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: b2b
{{- end }}

{{- define "approval-workflow-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "approval-workflow-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
