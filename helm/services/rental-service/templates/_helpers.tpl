{{- define "rental-service.fullname" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "rental-service.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/name: {{ include "rental-service.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: rental
{{- end }}

{{- define "rental-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "rental-service.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
