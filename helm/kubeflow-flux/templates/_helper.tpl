{{/*
patch.render loads and renders a patch template file from the chart’s files.
It expects a dictionary with:
  - patchPath: the relative path to the patch template file (e.g. "patches/istio/virtual-service.tpl")
  - Files: the global Files object (passed from the calling template)
The patch template must output valid YAML with target metadata:
  - kind
  - metadata.name
  - metadata.namespace (or defaults to "default")
*/}}
{{- define "patch.render" -}}
  {{- $rawContent := .ctx.Files.Get .patchPath }}
  {{- $patchContent := tpl $rawContent .ctx.Values | trim -}}
  {{- if eq $patchContent "" -}}
    {{- fail (printf "patch file %s rendered empty" .patchPath) -}}
  {{- end -}}
  {{- $patchData := fromYaml $patchContent -}}
patch: |- {{ $patchContent | nindent 4 }}
target:
  kind: {{ $patchData.kind }}
  name: {{ $patchData.metadata.name }}
  namespace: {{ $patchData.metadata.namespace | default "default" }}
{{- end -}}
