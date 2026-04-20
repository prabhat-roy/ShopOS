{{/*
ShopOS — review-rating-service Helm helper templates
*/}}

{{- define "review-rating-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "review-rating-service.fullname" -}}
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

{{- define "review-rating-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "review-rating-service.labels" -}}
helm.sh/chart: {{ include "review-rating-service.chart" . }}
{{ include "review-rating-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: customer-experience
{{- end }}

{{- define "review-rating-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "review-rating-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
