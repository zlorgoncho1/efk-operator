# EFK Stack Operator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21-blue.svg)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.24%2B-blue.svg)](https://kubernetes.io/)

A Kubernetes operator to easily deploy a production-ready EFK (Elasticsearch, Fluent Bit, Kibana) stack with flexible CRDs and Helm templates.

**Repository**: [https://github.com/zlorgoncho1/efk-operator](https://github.com/zlorgoncho1/efk-operator)

## Overview

This operator allows you to deploy a complete EFK stack by simply creating a custom `EFKStack` resource. The operator automatically manages the deployment, configuration, and updates of components.

## Architecture

- **Elasticsearch**: Log storage and indexing
- **Fluent Bit**: Log collection from all Kubernetes nodes
- **Kibana**: Visualization and analysis interface

## Quick Start

### For Users

See the [User Guide](docs/USER_GUIDE.md) for complete installation and usage instructions.

**Quick installation**:

```bash
# Install CRDs (from GitHub)
kubectl apply -f https://raw.githubusercontent.com/zlorgoncho1/efk-operator/main/efk/config/crd/bases/logging.efk.crds.io_efkstacks.yaml

# Deploy the operator (requires kustomize)
kubectl create namespace system
kubectl apply -k https://github.com/zlorgoncho1/efk-operator.git/efk/config/default

# Create an EFK stack (example)
kubectl apply -f https://raw.githubusercontent.com/zlorgoncho1/efk-operator/main/efk/config/samples/logging_v1_efkstack.yaml
```

### For Developers

See the [Getting Started Guide](docs/GETTING_STARTED.md) to set up the development environment.

**Quick start**:

```bash
# Build the Docker environment
make docker-build

# Quick tests
make test-quick
```

## Prerequisites

- Kubernetes 1.24+
- `kubectl` configured to access your cluster
- Docker and Docker Compose (for development)

## Installation

For detailed installation instructions, see the [User Guide](docs/USER_GUIDE.md#installation-de-lopérateur).

## Usage

### Create an EFK Stack

Create an `efkstack.yaml` file:

```yaml
apiVersion: logging.efk.crds.io/v1
kind: EFKStack
metadata:
  name: my-efk-stack
  namespace: efk-system
spec:
  version: "1.0.0"
  namespace: efk-system
  
  elasticsearch:
    version: "8.11.0"
    replicas: 3
    resources:
      requests:
        cpu: "2"
        memory: "4Gi"
      limits:
        cpu: "4"
        memory: "8Gi"
    storage:
      size: "100Gi"
      storageClassName: "standard"
    security:
      tlsEnabled: true
      authEnabled: true
  
  fluentBit:
    version: "2.2.0"
    resources:
      requests:
        cpu: "100m"
        memory: "128Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
  
  kibana:
    version: "8.11.0"
    replicas: 2
    resources:
      requests:
        cpu: "1"
        memory: "1Gi"
      limits:
        cpu: "2"
        memory: "2Gi"
    ingress:
      enabled: true
      host: "kibana.example.com"
      annotations:
        kubernetes.io/ingress.class: "nginx"
      tls:
        - hosts:
            - kibana.example.com
          secretName: kibana-tls
```

Apply the resource:

```bash
kubectl apply -f efkstack.yaml
```

### Check Status

```bash
kubectl get efkstack my-efk-stack -n efk-system
kubectl describe efkstack my-efk-stack -n efk-system
```

## Configuration

### Spécification EFKStack

#### Elasticsearch

- `version`: Elasticsearch version (e.g., "8.11.0")
- `replicas`: Number of replicas (minimum 1, recommended 3+ for production)
- `resources`: CPU and memory resources
- `storage`: Persistent storage configuration
- `security`: TLS and authentication configuration

#### Fluent Bit

- `version`: Fluent Bit version (e.g., "2.2.0")
- `resources`: CPU and memory resources
- `config`: Custom configuration (optional)

#### Kibana

- `version`: Kibana version (e.g., "8.11.0")
- `replicas`: Number of replicas (recommended 2+ for HA)
- `resources`: CPU and memory resources
- `ingress`: Ingress configuration for external access

## Development

### Project Structure

```
efk/
├── api/v1/              # CRD definitions
├── config/              # Kubernetes manifests
├── controllers/         # Reconciliation logic
├── helm-charts/         # Helm charts for components
└── internal/helm/       # Helm client
```

### Useful Commands

```bash
# Generate code
make generate

# Format code
make fmt

# Run tests
make test

# Build the operator
make build

# Run locally (outside cluster)
make run
```

### Tests

Tests use Ginkgo and Gomega:

```bash
make test
```

## Documentation

All documentation is available in the [`docs/`](docs/) directory:

### For Users
- **[User Guide](docs/USER_GUIDE.md)** - Installation, usage, and troubleshooting
- **[Architecture](docs/ARCHITECTURE.md)** - Project structure

### For Developers
- **[Getting Started Guide](docs/GETTING_STARTED.md)** - Development environment setup
- **[Testing Guide](docs/TESTING.md)** - Testing and validation
- **[Architecture](docs/ARCHITECTURE.md)** - Project structure

See [docs/README.md](docs/README.md) for the complete documentation index.

## Production

### Recommendations

1. **High Availability**: Use at least 3 replicas for Elasticsearch and 2+ for Kibana
2. **Security**: Enable TLS and authentication
3. **Storage**: Use high-performance storage classes (SSD)
4. **Monitoring**: Configure component monitoring
5. **Backup**: Implement backup strategies for Elasticsearch

### Production Configuration Example

See `efk/helm-charts/efk-stack/values-production.yaml` for a production-ready configuration example.

## Troubleshooting

### Check Operator Logs

```bash
kubectl logs -n system deployment/controller-manager
```

### Check Component Status

```bash
# Elasticsearch
kubectl get statefulset -n efk-system

# Fluent Bit
kubectl get daemonset -n efk-system

# Kibana
kubectl get deployment -n efk-system
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

- [Code of Conduct](CONTRIBUTING.md#code-of-conduct)
- [Development Setup](docs/GETTING_STARTED.md)
- [Coding Standards](CONTRIBUTING.md#coding-standards)
- [Pull Request Process](CONTRIBUTING.md#pull-request-process)

## Known Issues

### go vet Error (Non-blocking)

When running `go vet`, you may encounter an error related to `k8s.io/cli-runtime`:

```
k8s.io/cli-runtime@v0.28.4/pkg/genericclioptions/config_flags.go:335:53: 
not enough arguments in call to restmapper.NewShortcutExpander
```

**Cause**: This is a known compatibility issue between `helm.sh/helm/v3 v3.13.3` and `k8s.io/cli-runtime v0.28.4` (transitive dependency).

**Impact**: 
- `go vet` fails, but this does not affect the operator's functionality
- Our code does not directly use `cli-runtime`
- This is a transitive dependency from Helm

**Workaround**: This error can be safely ignored for now. The operator compiles and runs correctly. We're monitoring upstream fixes for this compatibility issue.

## License

Apache License 2.0

