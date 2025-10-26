#!/usr/bin/env bash

cd $(mktemp -d)
cat <&0 > all.yaml
cat > force-recreation-of-pods.yaml << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: not-important-will-match-all-deployments
spec:
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/recreatedAt: $(date +\"%Y-%m-%dT%H:%M:%S\")
EOF

cat > kustomization.yaml << EOF
resources:
  - all.yaml
patches:
  - path: force-recreation-of-pods.yaml
    target:
      kind: Deployment
 {{- range $key, $value := .Services }}
  - patch: |-
      - op: replace
        path: /spec/selector/app
        value: dev-proxy
      - op: replace
        path: /spec/ports/0/targetPort
        value: {{ .ProxyPort }}
    target:
      kind: Service
      name: {{ .Name }}
{{- end }}
EOF

kubectl kustomize . && rm all.yaml force-recreation-of-pods.yaml kustomization.yaml