{% define "storefront.fullname" -%}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{% define "storefront.labels" -}
helm.sh/chart: {{- include "storefront.fullname" . }}
app.kubernetes.io/name: storefront
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
