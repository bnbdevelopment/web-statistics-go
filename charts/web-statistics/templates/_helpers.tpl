{{- define "web-statistics.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "web-statistics.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}{{ .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}{{ printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}{{- end }}
{{- end }}
{{- end }}

{{- define "web-statistics.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "web-statistics.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{- define "web-statistics.selectorLabels" -}}
app.kubernetes.io/name: {{ include "web-statistics.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "web-statistics.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "web-statistics.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "web-statistics.databaseSecretName" -}}
{{- default (printf "%s-database" (include "web-statistics.fullname" .)) .Values.database.existingSecret }}
{{- end }}

{{- define "web-statistics.image" -}}
{{- printf "%s/%s:%s" .registry .repository .tag }}
{{- end }}

{{- define "web-statistics.backendChecksum" -}}
{{- include (print $.Template.BasePath "/secret-database.yaml") . | sha256sum }}
{{- end }}

{{- define "web-statistics.databaseEnv" -}}
- name: DB_HOST
  valueFrom:
    secretKeyRef:
      name: {{ include "web-statistics.databaseSecretName" . }}
      key: {{ .Values.database.secretKeys.host }}
- name: DB_PORT
  valueFrom:
    secretKeyRef:
      name: {{ include "web-statistics.databaseSecretName" . }}
      key: {{ .Values.database.secretKeys.port }}
- name: DB_NAME
  valueFrom:
    secretKeyRef:
      name: {{ include "web-statistics.databaseSecretName" . }}
      key: {{ .Values.database.secretKeys.name }}
- name: DB_USER
  valueFrom:
    secretKeyRef:
      name: {{ include "web-statistics.databaseSecretName" . }}
      key: {{ .Values.database.secretKeys.user }}
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "web-statistics.databaseSecretName" . }}
      key: {{ .Values.database.secretKeys.password }}
- name: DB_SSLMODE
  valueFrom:
    secretKeyRef:
      name: {{ include "web-statistics.databaseSecretName" . }}
      key: {{ .Values.database.secretKeys.sslMode }}
{{- end }}
