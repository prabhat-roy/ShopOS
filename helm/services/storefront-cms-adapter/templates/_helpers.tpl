{{- define "storefront-cms-adapter.fullname" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "storefront-cms-adapter.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/name: {{ include "storefront-cms-adapter.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/domain: content
{{- end }}

{{- define "storefront-cms-adapter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "storefront-cms-adapter.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
