{{/*
ShopOS — seller-payout-service Helm helper templates
*/}}

{{- define "seller-payout-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "seller-payout-service.fullname" -}}
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

{{- define "seller-payout-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "seller-payout-service.labels" -}}
helm.sh/chart: {{- include "seller-payout-service.chart" . }}
{{- include "seller-payout-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{- .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{- .Release.Service }}
app.kubernetes.io/domain: marketplace
{{- end }}

{{- define "seller-payout-service.selectorLabels" -}}
app.kubernetes.io/name: {{- include "seller-payout-service.name" . }}
app.kubernetes.io/instance: {{- .Release.Name }}
{{- end }}
