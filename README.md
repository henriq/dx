# DX

**The problem:** You have multiple microservices in a local Kubernetes cluster. You want to develop one service on your machine while it communicates with the others. Setting up this routing manually is tedious and error-prone.

**What DX does:** DX connects your locally-running services to your Kubernetes cluster. Traffic meant for your service gets routed to your machine; everything else stays in the cluster. You can inspect HTTP traffic between services in a web UI.

## Quick Example

You have 5 services deployed to Kubernetes: `api`, `auth`, `users`, `payments`, and `notifications`.

To work on `api` locally:

```bash
# Deploy all services to your local Kubernetes cluster
dx install

# Start your api service locally on port 8080
./run-api-locally.sh

# DX detects your local service and routes cluster traffic to it
# Other services in Kubernetes now talk to your local api
```

Run `dx context info` to get the mitmweb (traffic inspection UI) URL where you can see every HTTP request flowing between services.

## How It Works

DX runs a proxy inside your Kubernetes cluster that intercepts service-to-service traffic:

```
┌─────────────────────────────────────────────────────────────────┐
│  Kubernetes Cluster                                             │
│                                                                 │
│  ┌──────────┐     ┌──────────┐     ┌──────────┐                │
│  │  auth    │     │  users   │     │ payments │                │
│  └────┬─────┘     └────┬─────┘     └────┬─────┘                │
│       │                │                │                       │
│       └────────────────┼────────────────┘                       │
│                        │                                        │
│                        ▼                                        │
│              ┌──────────────────┐                               │
│              │    dev-proxy     │◄─── Intercepts traffic        │
│              │                  │     to local services         │
│              └────────┬─────────┘                               │
│                       │                                         │
└───────────────────────┼─────────────────────────────────────────┘
                        │
                        ▼ Health check passes?
              ┌─────────────────────┐
              │  Your local machine │
              │                     │
              │  api service :8080  │◄─── Traffic routed here
              └─────────────────────┘
```

**The routing logic:**

1. DX patches Kubernetes services to route traffic through the proxy
2. The proxy checks if your local service is running (via health check)
3. If local service is healthy: traffic goes to your machine
4. If local service is down: traffic goes to the Kubernetes version
5. The proxy captures all HTTP traffic for inspection in mitmweb

## Prerequisites

DX assumes you already have a local Kubernetes cluster running.

**Kubernetes Environment:**
- A local Kubernetes cluster (Docker Desktop, Rancher Desktop, minikube, etc.)
- Docker client connected to the cluster's Docker daemon

**Required Tools:**
- `kubectl`
- `helm`
- `git`
- `docker`
- `bash`

## Installation

### Homebrew

```bash
brew tap henriq/dx
brew install dx
```

### Manual Installation

Download the latest release from the [releases page](https://github.com/henriq/dx/releases) and extract it to a directory in your PATH.

### Windows

When running on Windows, disable line ending conversion:

```bash
git config --global core.autocrlf false
```

## Quick Start

```bash
# 1. Create your configuration file
dx initialize

# 2. Edit ~/.dx-config.yaml to define your services
#    (see Configuration section below)

# 3. Build and deploy everything
dx update

# 4. View your environment status
dx context info
```

## Commands

### Core Workflow

| Command | Description |
|---------|-------------|
| `dx build [services...]` | Build Docker images for services |
| `dx install [services...]` | Deploy services to Kubernetes via Helm |
| `dx uninstall [services...]` | Remove services from Kubernetes |
| `dx update [services...]` | Build and deploy in one step |

All commands accept:
- **No arguments**: Operates on services in the default profile
- **Service names**: Operates only on specified services
- **`-p, --profile`**: Target a specific profile (`-p all` for everything)

### Context Management

Contexts let you work with multiple projects from a single configuration file.

```bash
dx context list              # Show all available contexts
dx context set <name>        # Switch to a different context
dx context info              # Display current context and monitoring URLs
dx context print             # Output current context as JSON
```

### Secrets

Secrets are encrypted with AES-GCM and stored per-context. Keys are kept in your system keyring.

```bash
dx secret set <key> <value>  # Store an encrypted secret
dx secret get <key>          # Retrieve a secret value
dx secret list               # List all secrets
dx secret delete <key>       # Remove a secret
```

### Utilities

```bash
dx initialize                # Create sample configuration file
dx run <script>              # Execute scripts defined in your context
dx gen-env-key               # Generate cluster verification key
dx version                   # Show version
```

## Common Workflows

### Daily Development

```bash
# Rebuild and redeploy a single service
dx update api

# Check what's running
dx context info
```

### Working with Profiles

Profiles group services for targeted operations:

```bash
# Deploy only infrastructure
dx install -p infra

# Work on application services
dx update -p default

# Tear down everything
dx uninstall -p all
```

### Multi-Project Setup

```bash
# List your configured projects
dx context list

# Switch projects
dx context set backend-services

# All commands now operate on the new context
dx update
```

## Configuration

DX uses `~/.dx-config.yaml` to define your development environment.

### Basic Structure

```yaml
contexts:
  - name: my-project
    services:
      - name: api
        # How to build and deploy this service
    localServices:
      - name: api
        # How to route traffic to your local version
    scripts:
      reset-db: kubectl delete pvc -l app=postgres
```

### Services

Services define what DX builds and deploys to Kubernetes:

```yaml
services:
  - name: api
    # Helm deployment
    helmRepoPath: /path/to/helm/repo
    helmChartRelativePath: charts/api
    helmBranch: main
    helmArgs:
      - --set=image.tag=latest

    # Docker build
    dockerImages:
      - name: api:latest
        dockerfilePath: Dockerfile
        buildContextRelativePath: .
        gitRepoPath: /path/to/source
        gitRef: main

    # Images to pull (not build)
    remoteImages:
      - postgres:15
      - redis:7

    # Which profiles include this service
    profiles:
      - default
```

### Local Services

Local services tell DX how to route traffic to your machine:

```yaml
localServices:
  - name: api
    localPort: 8080           # Where your local service runs
    kubernetesPort: 80        # The Kubernetes service port
    healthCheckPath: /health  # How to check if local is running
    selector:
      app: api                # Must match the Kubernetes service
```

When you run a service locally on `localPort`, DX automatically routes cluster traffic to your machine. If your local service stops responding to health checks, traffic falls back to the Kubernetes version.

### Profiles

Profiles let you operate on subsets of services:

```yaml
services:
  - name: postgres
    profiles: [infra]
  - name: redis
    profiles: [infra]
  - name: api
    profiles: [default]
  - name: frontend
    profiles: [default]
```

```bash
dx build              # Builds api, frontend (default profile)
dx install -p infra   # Deploys postgres, redis
dx update -p all      # Everything
```

### Dockerfile Override

Define Dockerfile content directly in configuration:

```yaml
dockerImages:
  - name: my-image
    dockerfileOverride: |
      FROM alpine:latest
      RUN apk add --no-cache curl
      COPY . /app
      CMD ["./run.sh"]
    buildContextRelativePath: .
    gitRepoPath: /path/to/source
    gitRef: main
```

Useful for testing build changes without modifying the source repository.

### Secrets in Configuration

Reference secrets in Helm arguments:

```yaml
helmArgs:
  - --set=database.password={{.Secrets.DB_PASSWORD}}
```

```bash
dx secret set DB_PASSWORD "s3cr3t"
dx install
```

### Custom Scripts

Define reusable commands:

```yaml
scripts:
  reset-db: |
    kubectl delete pvc -l app=postgres
    dx uninstall postgres
    dx install postgres
  logs: kubectl logs -f deployment/api
```

```bash
dx run reset-db
dx run logs
```

## Advanced Topics

### Configuration Sharing

Teams can share base configurations:

```yaml
contexts:
  - name: my-context
    import: /path/to/shared-context.yaml
    services:
      - name: api
        dockerImages:
          - name: api
            gitRef: my-feature-branch  # Override specific values
```

### Inspecting Traffic

The dev-proxy includes mitmproxy for HTTP inspection. Run `dx context info` to get the URL.

You can see:
- All HTTP requests between services
- Request/response headers and bodies
- Timing information

### Environment Verification

DX tracks which Kubernetes cluster/namespace you're working with:

```bash
# Generate a key for current cluster config
dx gen-env-key

# DX warns you if cluster config changes unexpectedly
```

### Shell Completion

```bash
# Bash
source <(dx completion bash)

# Zsh
source <(dx completion zsh)

# Fish
dx completion fish | source

# PowerShell
dx completion powershell | Out-String | Invoke-Expression
```

### Local File Storage

DX stores data in `~/.dx/`:
- Repository clones
- Encrypted secrets
- Dev-proxy configuration

### Uninstallation

Remove `~/.dx/` and the `dx` binary.
