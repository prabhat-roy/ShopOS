{% define "seller-portal.fullname" -%}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{% define "seller-portal.labels" -}
helm.sh/chart: {{- include "seller-portal.fullname" . }}
app.kubernetes.io/name: seller-portal
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
