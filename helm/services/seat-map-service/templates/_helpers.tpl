{{- define "seat-map-service.fullname" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "seat-map-service.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/name: {{ include "seat-map-service.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: events-ticketing
{{- end }}

{{- define "seat-map-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "seat-map-service.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
