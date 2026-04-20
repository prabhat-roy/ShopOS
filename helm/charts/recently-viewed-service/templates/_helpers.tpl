{{/*
ShopOS — recently-viewed-service Helm helper templates
*/}}

{{- define "recently-viewed-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "recently-viewed-service.fullname" -}}
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

{{- define "recently-viewed-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "recently-viewed-service.labels" -}}
helm.sh/chart: {{ include "recently-viewed-service.chart" . }}
{{ include "recently-viewed-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: customer-experience
{{- end }}

{{- define "recently-viewed-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "recently-viewed-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
