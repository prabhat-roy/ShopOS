{{/*
ShopOS — ml-feature-store Helm helper templates
*/}}

{{- define "ml-feature-store.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "ml-feature-store.fullname" -}}
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

{{- define "ml-feature-store.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "ml-feature-store.labels" -}}
helm.sh/chart: {{ include "ml-feature-store.chart" . }}
{{ include "ml-feature-store.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: analytics-ai
{{- end }}

{{- define "ml-feature-store.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ml-feature-store.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
