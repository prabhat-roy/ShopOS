{% define "mobile-app.fullname" -%}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{% define "mobile-app.labels" -}
helm.sh/chart: {{- include "mobile-app.fullname" . }}
app.kubernetes.io/name: mobile-app
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
