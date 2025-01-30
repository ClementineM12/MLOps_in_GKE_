{{/*
Generate a list of sync options for ArgoCD applications.
*/}}
{{- define "syncOptions" -}}
{{- range .syncPolicy.syncOptions }}
  - {{ . }}
{{- end -}}
{{- end }}
