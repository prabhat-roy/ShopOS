{{/*
ShopOS — fraud-detection-service Helm helper templates
*/}}

{{- define "fraud-detection-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "fraud-detection-service.fullname" -}}
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

{{- define "fraud-detection-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "fraud-detection-service.labels" -}}
helm.sh/chart: {{ include "fraud-detection-service.chart" . }}
{{ include "fraud-detection-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: commerce
{{- end }}

{{- define "fraud-detection-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fraud-detection-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
