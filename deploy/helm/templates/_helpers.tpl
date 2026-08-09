{{- define "cartlabs.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "cartlabs.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "cartlabs.name" . -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "cartlabs.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | quote }}
app.kubernetes.io/name: {{ include "cartlabs.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: cartlabs
{{- end -}}

{{- define "cartlabs.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cartlabs.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "cartlabs.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "cartlabs.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "cartlabs.image" -}}
{{- $root := index . 0 -}}
{{- $repository := index . 1 -}}
{{- if $root.Values.global.imageRegistry -}}
{{- printf "%s/%s:%s" ($root.Values.global.imageRegistry | trimSuffix "/") $repository $root.Values.global.imageTag -}}
{{- else -}}
{{- printf "%s:%s" $repository $root.Values.global.imageTag -}}
{{- end -}}
{{- end -}}

{{- define "cartlabs.componentLabels" -}}
{{ include "cartlabs.selectorLabels" .root }}
app.kubernetes.io/component: {{ .component }}
{{- end -}}
