{{/*
ShopOS — i18n-l10n-service Helm helper templates
*/}}

{{- define "i18n-l10n-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "i18n-l10n-service.fullname" -}}
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

{{- define "i18n-l10n-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "i18n-l10n-service.labels" -}}
helm.sh/chart: {{ include "i18n-l10n-service.chart" . }}
{{ include "i18n-l10n-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: content
{{- end }}

{{- define "i18n-l10n-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "i18n-l10n-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
