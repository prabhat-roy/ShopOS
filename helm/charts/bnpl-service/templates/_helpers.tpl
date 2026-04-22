{{/*
ShopOS — bnpl-service Helm helper templates
*/}}

{{/*
Expand the name of the chart.
*/}}
{{- define "bnpl-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "bnpl-service.fullname" -}}
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

{{/*
Create chart version label.
*/}}
{{- define "bnpl-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels applied to all resources.
*/}}
{{- define "bnpl-service.labels" -}}
helm.sh/chart: {{ include "bnpl-service.chart" . }}
{{ include "bnpl-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: shopos
app.kubernetes.io/domain: platform
{{- end }}

{{/*
Selector labels — used in matchLabels and Service selector.
*/}}
{{- define "bnpl-service.selectorLabels" -}}
app: {{ include "bnpl-service.name" . }}
app.kubernetes.io/name: {{ include "bnpl-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
