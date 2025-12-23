# Testing Guide - EFK Stack Operator

This guide allows you to test that everything is operational before continuing.

## Prerequisites

- Docker and Docker Compose installed
- Access to a Kubernetes cluster (optional for basic tests)
- `kubectl` configured (optional)

## Tests to Perform

### 1. Project Structure Verification

```bash
cd efk
ls -la
```

You should see:
- `api/` - CRD definitions
- `config/` - Kubernetes manifests
- `docker/` - Docker configuration
- `helm-charts/` - Helm charts
- `internal/` - Internal code
- `main.go` - Entry point
- `go.mod` - Dependencies
- `Makefile` - Main Makefile

### 2. Docker Environment Test

```bash
# Build development image
make docker-build

# Verify that the image is created
docker images | grep efk-operator-dev

# Start environment
make dev-up

# Verify that the container is running
docker ps | grep efk-operator-dev

# Test a shell in the container
make dev-shell
# In the shell, test:
#   - kubebuilder version
#   - helm version
#   - kubectl version
#   - go version
# Exit with 'exit'
```

### 3. CRD Manifest Generation Test

```bash
# Generate CRD manifests
make manifests

# Verify that files are generated
ls -la config/crd/bases/

# You should see:
# logging.efk.crds.io_efkstacks.yaml
```

### 4. Code Generation Test

```bash
# Generate code (deepcopy, etc.)
make generate

# Verify that generated files are created
ls -la api/v1/zz_generated.*
```

### 5. Compilation Test

```bash
# Format code
make fmt

# Verify code
make vet

# Build binary
make build

# Verify that binary is created
ls -la bin/manager
```

### 6. Go Dependencies Test

```bash
# In Docker container
make dev-shell

# In the shell:
cd /workspace
go mod download
go mod verify
go mod tidy
exit
```

### 7. YAML File Validation Test

```bash
# Verify Kubernetes manifests
make dev-shell
# In the shell:
kubectl apply --dry-run=client -f config/crd/bases/
kubectl apply --dry-run=client -f config/rbac/
kubectl apply --dry-run=client -f config/manager/
exit
```

### 8. Helm Charts Test (validation)

```bash
make dev-shell
# In the shell:
cd helm-charts/efk-stack/elasticsearch
helm lint .
helm template . --debug
cd ../fluentbit
helm lint .
helm template . --debug
cd ../kibana
helm lint .
helm template . --debug
exit
```

### 9. Go Imports Test

```bash
make dev-shell
# In the shell:
go build ./...
go test ./... -v
exit
```

### 10. Complete Test (all checks)

```bash
# Run all checks
make all

# Or step by step:
make manifests
make generate
make fmt
make vet
make test
make build
```

## Validation Checklist

- [ ] Complete project structure
- [ ] Docker image builds without error
- [ ] Container starts correctly
- [ ] Tools available in container (kubebuilder, helm, kubectl, go)
- [ ] CRD manifests generated
- [ ] Code generated (deepcopy)
- [ ] Code compiles without error
- [ ] Code passes `go vet`
- [ ] Unit tests pass
- [ ] Helm charts valid (lint)
- [ ] Go imports correct
- [ ] Binary created and executable

## Troubleshooting

### Error: "docker-compose: command not found"
```bash
# Install Docker Compose or use:
docker compose -f docker/docker-compose.yml build
```

### Error: "go.mod: module path mismatch"
```bash
# Verify that go.mod uses the correct path:
# module github.com/zlorgoncho1/efk-operator
```

### Error: "import path not found"
```bash
# In the container:
go mod download
go mod tidy
```

### Error: "kubebuilder: command not found"
```bash
# Verify that Docker image is built with kubebuilder
make docker-build
```

## Advanced Tests (optional)

### Test with Local Kubernetes Cluster (kind)

```bash
# Install kind in container
make dev-shell
# In the shell:
kind create cluster --name efk-test
kubectl cluster-info

# Install CRDs
make install

# Create test EFKStack resource
kubectl apply -f config/samples/logging_v1_efkstack.yaml

# Verify
kubectl get efkstack

# Cleanup
kind delete cluster --name efk-test
exit
```
