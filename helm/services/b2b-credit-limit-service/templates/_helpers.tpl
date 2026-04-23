{{/*
ShopOS — b2b-credit-limit-service Helm helper templates
*/}}

{{- define "b2b-credit-limit-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "b2b-credit-limit-service.fullname" -}}
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

{{- define "b2b-credit-limit-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "b2b-credit-limit-service.labels" -}}
helm.sh/chart: {{ include "b2b-credit-limit-service.chart" . }}
{{ include "b2b-credit-limit-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: b2b
{{- end }}

{{- define "b2b-credit-limit-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "b2b-credit-limit-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
