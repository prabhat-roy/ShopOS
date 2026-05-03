{{- define "chart.fullname" -}}
{{- if .Values.fullnameOverride -}}{{ .Values.fullnameOverride }}{{- else -}}{{ .Chart.Name }}{{- end -}}
{{- end -}}
{{- define "chart.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}{{ include "chart.fullname" . }}{{- else -}}default{{- end -}}
{{- end -}}
