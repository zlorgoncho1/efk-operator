# Getting Started - EFK Stack Operator

Quick start guide for developers.

## Prerequisites

- Docker and Docker Compose installed
- Access to a Kubernetes cluster (optional for basic tests)
- `kubectl` configured (optional)

## Quick Installation

### 1. Build Docker Environment

```bash
cd efk
make docker-build
```

**Estimated time**: 2-3 minutes (first time)

**Expected result**: `efk-operator-dev:latest` image built

### 2. Quick Automatic Tests

```bash
make test-quick
```

**Estimated time**: 1-2 minutes

This command executes:
- Docker image build
- CRD manifest generation
- Code generation (deepcopy)
- Binary compilation

### 3. Verify Installation

```bash
# Verify that files are created
ls config/crd/bases/
ls bin/manager
```

**Expected result**:
- ✅ Generated CRD file: `config/crd/bases/logging.efk.crds.io_efkstacks.yaml`
- ✅ Binary created: `bin/manager`

## Development Environment

### Start Environment

```bash
make dev-up
```

### Open Development Shell

```bash
make dev-shell
```

In the shell, you have access to:
- `go` - Go 1.21.13
- `helm` - Helm v3.19.4
- `kubectl` - kubectl (latest version)
- `controller-gen` - controller-gen v0.14.0
- `kustomize` - kustomize v5.8.0

### Development Commands

```bash
# Generate CRD manifests
make manifests

# Generate code (deepcopy)
make generate

# Format code
make fmt

# Verify code (with note on go vet)
make vet

# Compile
make build

# Run locally (outside cluster)
make run
```

## Quick Tests

### Manual Test in 3 Steps

1. **Build image**: `make docker-build`
2. **Automatic tests**: `make test-quick`
3. **Verify**: `ls config/crd/bases/ && ls bin/`

### Test Scripts (optional)

#### Linux/Mac:
```bash
chmod +x scripts/test-all.sh
./scripts/test-all.sh
```

#### Windows (PowerShell):
```powershell
.\scripts\test-all.ps1
```

## Common Issues

### Error: "docker-compose: command not found"
**Solution**: Use `docker compose` (without hyphen) or install Docker Compose

### Error: "make: command not found"
**Windows**: Install [Make for Windows](https://www.gnu.org/software/make/) or use WSL

### Docker build error
**Solution**: Verify that Docker Desktop is running

### go vet error (non-blocking)
See the [Known Issues](../README.md#known-issues) section in the main README.

## Next Steps

Once the environment is set up:

1. **Development**: Start modifying code in `internal/controller/`
2. **Tests**: See [TESTING.md](TESTING.md) for complete tests
3. **Architecture**: See [ARCHITECTURE.md](ARCHITECTURE.md) to understand the structure
4. **Usage**: See [USER_GUIDE.md](USER_GUIDE.md) to deploy in a cluster

## Complete Documentation

- **[TESTING.md](TESTING.md)** - Complete testing and validation guide
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Project architecture and structure
- **[USER_GUIDE.md](USER_GUIDE.md)** - User guide for installation/usage

