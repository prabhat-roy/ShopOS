{% define "partner-portal.fullname" -%}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{% define "partner-portal.labels" -}
helm.sh/chart: {{- include "partner-portal.fullname" . }}
app.kubernetes.io/name: partner-portal
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
