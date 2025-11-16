global
    # Log to stdout
    log stdout format raw local0 info
    maxconn 4096

defaults
    log     global
    mode    http
    option  httplog
    option  dontlognull
    
    # Default timeouts
    timeout connect 5s
    timeout client  30s
    timeout server  30s
    # Use a long tunnel timeout to allow web sockets
    timeout tunnel  8h
    timeout http-keep-alive 2s
    
    # Default health check
    default-server inter 5s fall 1 rise 1
    
    # Enable logging of HTTP requests
    option  http-server-close
    option  forwardfor

# Stats frontend
listen stats
    bind *:8888
    mode http
    stats enable
    stats uri /
    stats refresh 5s
    stats show-legends
    stats show-node
    no log

{{ range $key, $value := .Services }}
# Frontend for {{ .Name }}
frontend {{ .Name }}
    bind *:{{ .FrontendPort }}
    mode http
    log global
    capture request header Host len 64
    use_backend be-{{ .Name }}

# Backend for {{ .Name }}
backend be-{{ .Name }}
    {{- if .HealthCheckPath }}
    option httpchk
    http-check send meth GET uri {{ .HealthCheckPath }} ver HTTP/1.1 hdr Host host.docker.internal
    {{- else }}
    option tcp-check
    {{- end }}
    {{- if gt .LocalPort 0 }}
    server local host.docker.internal:{{ .LocalPort }} check
    {{- end }}
    server k8s {{ .Name }}-srv:{{ .KubernetesPort }} check backup
{{ end }}