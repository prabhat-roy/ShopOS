{{- define "mobile-flutter-service.fullname" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "mobile-flutter-service.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/name: {{ include "mobile-flutter-service.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: web
{{- end }}

{{- define "mobile-flutter-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mobile-flutter-service.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
