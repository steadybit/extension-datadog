{{/* vim: set filetype=mustache: */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "secret.name" -}}
{{- default "steadybit-extension-datadog" .Values.datadog.existingSecret -}}
{{- end -}}
