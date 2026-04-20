{{- define "elasticsearch.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- define "elasticsearch.labels" -}}
app: {{ include "elasticsearch.fullname" . }}
{{- end }}
