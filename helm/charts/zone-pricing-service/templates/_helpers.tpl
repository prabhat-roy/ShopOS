{{- define "zone-pricing-service.fullname" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "zone-pricing-service.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/name: {{ include "zone-pricing-service.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: supply-chain
{{- end }}

{{- define "zone-pricing-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "zone-pricing-service.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
