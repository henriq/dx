# DX - A developer experience tool

DX is a development experience tool that packages experiences and opinions about how an efficient
local development environment should work.

DX has the following goals:

* It should be easy to create and destroy environments on demand
* It should be easy to update services
* It should be possible to run services in an IDE fully integrated with the local environment
* It should provide a secure secrets management to keep secrets out of text files
* It should provide a way to inspect HTTP traffic between services
* It should be easy to share configurations between teams

DX solves these goals by providing a standardized way of defining how applications are built and deployed. It uses
Kubernetes and Helm to run services, and builds images with Docker. Integration with services running locally is
provided by patching kubernetes services to intercept all traffic in a proxy server, that monitors local ports for
local versions of running services, and provides routing decisions based on health checks. The proxy server also
intercepts HTTP traffic between services and provides an interface for inspecting traffic.

## Prerequisites

To use DX, you'll need:

**Kubernetes Environment:**
* A local Kubernetes cluster (e.g., Docker Desktop or Rancher Desktop)
* Docker client connected to the cluster's Docker daemon to ensure built images are accessible

**Required Tools:**
The following command-line tools must be installed and available in your PATH:
* `kubectl`
* `helm`
* `git`
* `bash`
* `docker`

## Installation

### Homebrew

Install DX with Homebrew by tapping the homebrew cask with `brew tap henriq/dx`. Then install DX with `brew install dx`.

### Manual Installation

Download the latest release from the [releases page](https://github.com/henriq/dx/releases) and extract it to a directory in your PATH.

### Windows Considerations

* **Line Ending Management** - Prevent git line-ending conversion issues with:

```bash
git config --global core.autocrlf false
```

## Shell Completion

DX supports command completion for various shells. Set it up for your preferred shell:

### Bash

```bash
# Add to ~/.bashrc or ~/.bash_profile
source <(dx completion bash)
```

### Zsh

```zsh
# Add to ~/.zshrc
source <(dx completion zsh)
```

### Fish

```fish
# Add to ~/.config/fish/config.fish
dx completion fish | source
```

### PowerShell

```powershell
# Add to your PowerShell profile
dx completion powershell | Out-String | Invoke-Expression
```

## Usage

```bash
# Initialize a new configuration
dx initialize

# Switch between contexts
dx context set [context-name]

# Build Docker images
dx build [service-name]

# Install services to Kubernetes
dx install [service-name]

# Update deployed services
dx update [service-name]

# Uninstalls services from Kubernetes
dx uninstall [service-name]

# Run custom scripts
dx run [script-name]
```

### Monitoring traffic

DX's dev-proxy provides an interface for inspecting HTTP traffic between services.

```bash
# View context configuration and service status
dx context info
```

This command displays:

* Configuration overview for the current context
* Services being intercepted by the dev-proxy
* Links to monitoring interfaces:
  * HAProxy status page - Shows traffic routing and service health
  * Mitmweb interface - Displays HTTP traffic between services

## Configuration

DX uses a YAML configuration file at `~/.dx-config.yaml` organized into contexts (projects).

### Basic Structure

```yaml
contexts:
  - name: my-project
    services:
      - name: api-service
        # Service configuration...
    localServices:
      - name: frontend
        # Local service configuration...
    scripts:
      my-script: echo "Running my script"
```

### Services

Services define components that DX manages in Kubernetes:

```yaml
services:
  - name: my-service
    # Helm deployment configuration
    helmRepoPath: /path/to/helm/repo
    helmChartRelativePath: charts/my-service
    helmBranch: main
    helmArgs:
      - --set=image.tag=latest
    localPort: 8080

    # Docker image configuration
    dockerImages:
      - name: my-image
        dockerfilePath: Dockerfile
        buildContextRelativePath: .
        buildArgs:
          - VERSION=1.0.0
        gitRepoPath: /path/to/git/repo
        gitRef: main

    # External images to pull
    remoteImages:
      - postgres:latest

    # Optional deployment profiles
    profiles:
      - dev
```

### Profiles

Profiles let you organize services into groups for targeted operations:

```bash
# Build only services in the 'default' profile
dx build

# Install only services in the 'infra' profile
dx install -p infra

# Update all services regardless of profile
dx update -p all
```

**How profiles work:**

* Services must be **explicitly added** to profiles via configuration
* Services can belong to multiple profiles if needed
* Operations run only on services that match the specified profile
* The special `all` profile includes every service and can be targeted with `-p all`
* The special `default` profile is used if no profile is selected with `-p`

**Example scenario:**

Consider a project with four services configured with these profiles:

```yaml
services:
  - name: database
    profiles: [infra]
    # other configuration...
  - name: message-queue
    profiles: [infra]
    # other configuration...
  - name: backend-api
    profiles: [default]
    # other configuration...
  - name: frontend
    profiles: [default]
    # other configuration...
```

With this configuration:
* `dx build` operates only on backend-api and frontend
* `dx install -p infra` operates only on database and message-queue
* `dx update -p all` operates on all four services

### Local Services

Local services define how DX routes traffic between Kubernetes and your local machine:

```yaml
localServices:
  - name: my-local-service
    localPort: 8080
    kubernetesPort: 80
    healthCheckPath: /health
    selector:
      app: my-app
```

The name and selector of the local service must match the Kubernetes service that is being
intercepted for the local service to work.

### Custom Scripts

DX allows you to define custom scripts in your configuration file that can be executed with the `run` command:

```yaml
scripts:
  example: |
    echo "write something"
    read foo
    echo "you wrote $foo"
```

These scripts can read from stdin and interact with the user:

```bash
# Run a custom script named 'example'
dx run example
```

The example above would prompt the user to type something, read the input into a variable called `foo`, and then echo back what was typed.

### Secrets Management

DX provides secure secrets handling, with encryption using AES-GCM and keys stored in your system's keyring.

```bash
# Set a secret
dx secret set api-key "my-secret-value"

# Get a secret
dx secret get api-key

# List all secrets
dx secret list

# Delete a secret
dx secret delete api-key
```

You can reference secrets in Helm charts:

```yaml
helmArgs:
  - --set
  - mySecretValue={{.Secrets.api-key}}
```

### Context Switching

Manage multiple projects by switching contexts:

```bash
# List available contexts
dx context

# Switch to a specific context
dx context my-project
```

## Advanced Topics

### Configuration Sharing

Teams can share configurations by importing context files:

```yaml
contexts:
  - name: my-context
    import: /path/to/shared-context.yaml
```

You can override specific values from the imported configuration:

```yaml
contexts:
  - name: my-context
    import: /path/to/shared-context.yaml
    services:
      - name: my-service
        dockerImages:
          - name: my-image
            gitRef: custom-branch
```

### Development Workflow

A typical development workflow with DX:

1. Configure your services in `~/.dx-config.yaml`
2. Build images with `dx build`
3. Deploy to Kubernetes with `dx install`
4. Run specific services locally while testing
5. DX automatically routes traffic between local and Kubernetes services
6. Update deployments with `dx update` when needed

### Local File Storage

DX creates and manages several files in the `~/.dx` directory:

* **Repository Clones** - Local copies of all project repositories and Helm charts specified in your configuration
* **Encrypted Secrets** - Encrypted secrets for each project context
* **Dev Proxy Configuration** - Configuration files for the dev-proxy that handles traffic routing

### Uninstallation

To uninstall DX, remove all files under `~/.dx` and remove the `dx` binary from your system.
