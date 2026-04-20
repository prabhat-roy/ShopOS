{{/*
ShopOS — tax-provider-integration Helm helper templates
*/}}

{{- define "tax-provider-integration.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "tax-provider-integration.fullname" -}}
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

{{- define "tax-provider-integration.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "tax-provider-integration.labels" -}}
helm.sh/chart: {{ include "tax-provider-integration.chart" . }}
{{ include "tax-provider-integration.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: integrations
{{- end }}

{{- define "tax-provider-integration.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tax-provider-integration.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
