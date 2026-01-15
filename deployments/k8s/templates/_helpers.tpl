{{- define "multiagent-social.name" -}}
{{- default .Chart.Name .Values.nameOverride -}}
{{- end -}}

{{- define "multiagent-social.fullname" -}}
{{- printf "%s-%s" (include "multiagent-social.name" .) .Release.Name -}}
{{- end -}}

