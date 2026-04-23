{{- define "rate-limiter-core.fullname" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "rate-limiter-core.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/name: {{ include "rate-limiter-core.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: platform
{{- end }}

{{- define "rate-limiter-core.selectorLabels" -}}
app.kubernetes.io/name: {{ include "rate-limiter-core.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
