{{- define "financial-rules-engine.fullname" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "financial-rules-engine.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/name: {{ include "financial-rules-engine.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: financial
{{- end }}

{{- define "financial-rules-engine.selectorLabels" -}}
app.kubernetes.io/name: {{ include "financial-rules-engine.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
