{% define "developer-portal-ui.fullname" -%}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{% define "developer-portal-ui.labels" -}
helm.sh/chart: {{- include "developer-portal-ui.fullname" . }}
app.kubernetes.io/name: developer-portal-ui
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
