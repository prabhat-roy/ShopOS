{% define "admin-dashboard.fullname" -%}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{% define "admin-dashboard.labels" -}
helm.sh/chart: {{- include "admin-dashboard.fullname" . }}
app.kubernetes.io/name: admin-dashboard
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
