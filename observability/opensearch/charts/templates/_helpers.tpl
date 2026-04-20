{{- define "opensearch.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- define "opensearch.labels" -}}
app: {{ include "opensearch.fullname" . }}
{{- end }}
