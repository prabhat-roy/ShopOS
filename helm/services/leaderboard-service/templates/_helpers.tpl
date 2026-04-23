{{/*
ShopOS — leaderboard-service Helm helper templates
*/}}

{{- define "leaderboard-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "leaderboard-service.fullname" -}}
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

{{- define "leaderboard-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "leaderboard-service.labels" -}}
helm.sh/chart: {{- include "leaderboard-service.chart" . }}
{{- include "leaderboard-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: gamification
{{- end }}

{{- define "leaderboard-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "leaderboard-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
