{{- define "llm-ranking-chart.name" -}}
{{ .Chart.Name | quote }}
{{- end -}}

{{- define "llm-ranking-chart.fullname" -}}
{{ .Release.Name | quote }}
{{- end -}}

