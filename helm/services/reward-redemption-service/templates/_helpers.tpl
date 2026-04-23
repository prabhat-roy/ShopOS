{{/*
ShopOS — reward-redemption-service Helm helper templates
*/}}

{{- define "reward-redemption-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "reward-redemption-service.fullname" -}}
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

{{- define "reward-redemption-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "reward-redemption-service.labels" -}}
helm.sh/chart: {{- include "reward-redemption-service.chart" . }}
{{- include "reward-redemption-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: gamification
{{- end }}

{{- define "reward-redemption-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "reward-redemption-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
