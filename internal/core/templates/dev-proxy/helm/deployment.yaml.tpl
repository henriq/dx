apiVersion: apps/v1
kind: Deployment
metadata:
  name: dev-proxy
  labels:
    app: dev-proxy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: dev-proxy
  template:
    metadata:
      labels:
        app: dev-proxy
      annotations:
        checksum: {{ .Checksum }}
    spec:
      containers:
      - name: haproxy
        image: henriq/haproxy-{{ .Name }}
        imagePullPolicy: Never
        ports:
        - containerPort: 8080
        - containerPort: 8888
      - name: mitmproxy
        image: henriq/mitmproxy-{{ .Name }}
        imagePullPolicy: Never
        tty: true
        stdin: true
        command: ["mitmweb"]
        args:
          - --set
          - keep_host_header=true
          - --set
          - web_password={{ .Name }}
          - --set
          - onboarding=false
          - --set
          - web_open_browser=false
          - --set
          - showhost=true
          - --web-host=0.0.0.0
          - --web-port=8000
{{- range $key, $value := .Services }}
          - --mode=reverse:http://localhost:{{.FrontendPort}}@{{ .ProxyPort }}
{{- end }}

---

apiVersion: v1
kind: Service
metadata:
  name: dev-proxy
spec:
  selector:
    app: dev-proxy
  ports:
  - protocol: TCP
    port: 8888
    targetPort: 8888
    name: stats
  - protocol: TCP
    port: 8000
    targetPort: 8000
    name: mitmweb

---

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dev-proxy-haproxy
spec:
  rules:
  - host: stats.dev-proxy.{{ .Name }}.localhost
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: dev-proxy
            port:
              number: 8888

---

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dev-proxy-mitmweb
spec:
  rules:
  - host: dev-proxy.{{ .Name }}.localhost
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: dev-proxy
            port:
              number: 8000

---

{{- range $key, $value := .Services }}

apiVersion: v1
kind: Service
metadata:
  name: {{ .Name }}-srv
spec:
  selector:
{{ .Selector | toYaml | indent 4}}
  ports:
  - protocol: TCP
    port: {{ .KubernetesPort }}
    targetPort: {{ .KubernetesPort }}
    name: http

---

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ .Name }}-dx
spec:
  rules:
  - host: {{ .Name }}.{{ $.Name }}.localhost
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: {{ .Name }}
            port:
              number: {{ .KubernetesPort }}

---

{{- end}}