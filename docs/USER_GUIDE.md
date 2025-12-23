# User Guide - EFK Stack Operator

Complete guide to install and use the EFK Stack operator in your Kubernetes cluster.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Operator Installation](#operator-installation)
3. [Basic Usage](#basic-usage)
4. [Advanced Configuration](#advanced-configuration)
5. [Troubleshooting](#troubleshooting)
6. [Updates and Maintenance](#updates-and-maintenance)

## Prerequisites

### Cluster Requirements

- Kubernetes 1.24 or higher
- `kubectl` configured and connected to your cluster
- Administrator access to the cluster (to install CRDs and operator)
- StorageClass configured (for Elasticsearch persistent volumes)

### Required Tools

- `kubectl` (version compatible with your cluster)
- `helm` v3.0+ (optional, for certain installation methods)

## Operator Installation

### Method 1: Installation via Manifests (Recommended)

#### Step 1: Install CRDs

You can install directly from GitHub without cloning the repository:

```bash
kubectl apply -f https://raw.githubusercontent.com/zlorgoncho1/efk-operator/refs/heads/main/config/crd/bases/logging.efk.crds.io_efkstacks.yaml
```

**Alternative**: If you prefer to clone the repository for development:

```bash
git clone https://github.com/zlorgoncho1/efk-operator.git
cd efk-operator
kubectl apply -f config/crd/bases/logging.efk.crds.io_efkstacks.yaml
```

Verify installation:

```bash
kubectl get crd efkstacks.logging.efk.crds.io
```

#### Step 2: Deploy the Operator

**Note**: `kubectl apply -k` requires local access to the kustomization files, so you need to clone the repository first.

```bash
# Clone the repository
git clone https://github.com/zlorgoncho1/efk-operator.git
cd efk-operator

# Create namespace for operator
kubectl create namespace system

# Update the operator image (IMPORTANT: Replace with your Docker image)
# Edit config/default/manager_image_patch.yaml and set your image, or use:
# kustomize edit set image controller=your-registry/efk-operator:v0.1.0

# Deploy operator
kubectl apply -k config/default
```

**Alternative**: If you have `kustomize` installed, you can build the manifests and apply them:

```bash
# Clone the repository
git clone https://github.com/zlorgoncho1/efk-operator.git
cd efk-operator

# Build manifests with kustomize
kustomize build config/default | kubectl apply -f -
```

#### Step 3: Verify Deployment

```bash
# Verify that operator is deployed
kubectl get deployment -n system controller-manager

# Verify pods
kubectl get pods -n system

# Verify logs
kubectl logs -n system deployment/controller-manager -f
```

### Method 2: Installation via Helm (If chart available)

```bash
# Add Helm repository (when available)
helm repo add efk-operator https://zlorgoncho1.github.io/efk-operator

# Install operator
helm install efk-operator efk-operator/efk-operator \
  --namespace system \
  --create-namespace
```

## Basic Usage

### Create Your First EFK Stack

#### Minimal Example

Create a `my-efk-stack.yaml` file:

```yaml
apiVersion: logging.efk.crds.io/v1
kind: EFKStack
metadata:
  name: my-efk-stack
  namespace: efk-system
spec:
  elasticsearch:
    version: "8.11.0"
    replicas: 1
    resources:
      requests:
        cpu: "1"
        memory: "2Gi"
      limits:
        cpu: "2"
        memory: "4Gi"
    storage:
      size: "50Gi"
      storageClassName: "standard"
  
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
    replicas: 1
    resources:
      requests:
        cpu: "500m"
        memory: "512Mi"
      limits:
        cpu: "1"
        memory: "1Gi"
```

#### Apply the Resource

```bash
# Create namespace (if necessary)
kubectl create namespace efk-system

# Apply the resource
kubectl apply -f my-efk-stack.yaml
```

#### Check Status

```bash
# View resource status
kubectl get efkstack my-efk-stack -n efk-system

# View details
kubectl describe efkstack my-efk-stack -n efk-system

# Verify deployed components
kubectl get all -n efk-system
```

### Production Example

See `config/samples/logging_v1_efkstack.yaml` for a complete example with:
- High availability (3 Elasticsearch replicas, 2 Kibana)
- Security configuration (TLS, authentication)
- Ingress for Kibana
- Optimized resources

## Advanced Configuration

### Elasticsearch Configuration Options

```yaml
spec:
  elasticsearch:
    version: "8.11.0"              # Elasticsearch version
    replicas: 3                    # Number of replicas (minimum 1, recommended 3+)
    resources:
      requests:
        cpu: "2"
        memory: "4Gi"
      limits:
        cpu: "4"
        memory: "8Gi"
    storage:
      size: "100Gi"                # Storage size (format: 100Gi, 500Mi)
      storageClassName: "fast-ssd" # StorageClass to use
      volumeType: "persistentVolumeClaim"
    security:
      tlsEnabled: true             # Enable TLS
      authEnabled: true            # Enable authentication
      tlsSecretName: "es-tls"      # Secret containing TLS certificates
      authSecretName: "es-auth"    # Secret containing credentials
    config:                        # Additional configuration
      discovery.type: "kubernetes"
      xpack.security.enabled: "true"
```

### Fluent Bit Configuration Options

```yaml
spec:
  fluentBit:
    version: "2.2.0"
    resources:
      requests:
        cpu: "100m"
        memory: "128Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
    config:
      input: |
        [INPUT]
          Name tail
          Path /var/log/containers/*.log
          Parser docker
      filter: |
        [FILTER]
          Name kubernetes
          Match kube.*
      output: |
        [OUTPUT]
          Name es
          Match *
          Host elasticsearch
          Port 9200
```

### Kibana Configuration Options

```yaml
spec:
  kibana:
    version: "8.11.0"
    replicas: 2                    # Number of replicas (recommended 2+)
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
        cert-manager.io/cluster-issuer: "letsencrypt-prod"
        nginx.ingress.kubernetes.io/ssl-redirect: "true"
      tls:
        - hosts:
            - kibana.example.com
          secretName: kibana-tls
```

### Global Configuration

```yaml
spec:
  global:
    storageClass: "fast-ssd"       # Default StorageClass
    imageRegistry: "registry.example.com"  # Custom image registry
    tls:
      enabled: true
      secretName: "global-tls"
```

## Troubleshooting

### Check Operator Status

```bash
# Verify that operator is running
kubectl get pods -n system -l control-plane=controller-manager

# View operator logs
kubectl logs -n system deployment/controller-manager -f
```

### Check EFK Stack Status

```bash
# View EFKStack resource status
kubectl get efkstack -n efk-system

# View details (phase, component status)
kubectl describe efkstack my-efk-stack -n efk-system

# View events
kubectl get events -n efk-system --sort-by='.lastTimestamp'
```

### Check Individual Components

```bash
# Elasticsearch
kubectl get statefulset -n efk-system
kubectl get pods -n efk-system -l app=elasticsearch
kubectl logs -n efk-system -l app=elasticsearch -f

# Fluent Bit
kubectl get daemonset -n efk-system
kubectl get pods -n efk-system -l app=fluent-bit
kubectl logs -n efk-system -l app=fluent-bit -f

# Kibana
kubectl get deployment -n efk-system
kubectl get pods -n efk-system -l app=kibana
kubectl logs -n efk-system -l app=kibana -f
```

### Common Issues

#### Stack Stays in "Pending" or "Deploying" Phase

**Possible causes**:
- Operator is not started
- RBAC permission issue
- Storage issue (StorageClass not available)

**Solutions**:
```bash
# Verify operator
kubectl get pods -n system

# Verify permissions
kubectl describe role -n system manager-role

# Verify StorageClasses
kubectl get storageclass
```

#### Elasticsearch Won't Start

**Possible causes**:
- Storage issue (PVC not created)
- Insufficient resources
- Configuration issue

**Solutions**:
```bash
# Verify PVCs
kubectl get pvc -n efk-system

# Verify events
kubectl describe statefulset -n efk-system

# Verify available resources
kubectl top nodes
```

#### Fluent Bit Not Collecting Logs

**Possible causes**:
- Elasticsearch not accessible
- Incorrect Fluent Bit configuration
- Permission issue

**Solutions**:
```bash
# Verify connection to Elasticsearch
kubectl exec -n efk-system <fluent-bit-pod> -- curl http://<elasticsearch-service>:9200

# Verify Fluent Bit logs
kubectl logs -n efk-system -l app=fluent-bit

# Verify configuration
kubectl get configmap -n efk-system
```

#### Kibana Not Accessible

**Possible causes**:
- Ingress not configured or misconfigured
- Elasticsearch not accessible from Kibana
- TLS certificate issue

**Solutions**:
```bash
# Verify Ingress
kubectl get ingress -n efk-system

# Verify services
kubectl get svc -n efk-system

# Test internal connection
kubectl port-forward -n efk-system svc/kibana 5601:5601
# Then access http://localhost:5601
```

### Useful Commands

```bash
# View all Helm releases created by operator
helm list -n efk-system

# View Helm release details
helm status <release-name> -n efk-system

# View release history
helm history <release-name> -n efk-system

# View created resources
kubectl get all -n efk-system

# View created secrets
kubectl get secrets -n efk-system

# View configmaps
kubectl get configmaps -n efk-system
```

## Updates and Maintenance

### Update the Operator

```bash
# 1. Get the new version
git pull origin main

# 2. Regenerate manifests (if necessary)
make manifests

# 3. Update CRDs
kubectl apply -f config/crd/bases/logging.efk.crds.io_efkstacks.yaml

# 4. Update operator
kubectl apply -k config/default

# 5. Verify deployment
kubectl rollout status deployment/controller-manager -n system
```

### Update an EFK Stack

To update an existing stack, simply modify the EFKStack resource:

```bash
# Edit the resource
kubectl edit efkstack my-efk-stack -n efk-system

# Or apply a new file
kubectl apply -f my-efk-stack-updated.yaml
```

The operator will detect changes and update components via Helm.

### Backup and Restore

#### Elasticsearch Backup

```bash
# Create a snapshot (requires a configured repository)
kubectl exec -n efk-system <elasticsearch-pod> -- \
  curl -X PUT "localhost:9200/_snapshot/my_backup/snapshot_1?wait_for_completion=true"
```

#### Restore

```bash
# Restore from a snapshot
kubectl exec -n efk-system <elasticsearch-pod> -- \
  curl -X POST "localhost:9200/_snapshot/my_backup/snapshot_1/_restore"
```

### Uninstallation

#### Remove an EFK Stack

```bash
# Delete the EFKStack resource
kubectl delete efkstack my-efk-stack -n efk-system

# The operator will automatically remove all components
# Verify that everything is removed
kubectl get all -n efk-system
```

#### Uninstall the Operator

```bash
# Remove operator
kubectl delete -k config/default

# Remove CRDs (WARNING: also removes all EFKStack resources)
kubectl delete -f config/crd/bases/logging.efk.crds.io_efkstacks.yaml
```

## Best Practices

### Production

1. **High Availability**:
   - Use at least 3 replicas for Elasticsearch
   - Use at least 2 replicas for Kibana
   - Configure Pod Disruption Budgets

2. **Security**:
   - Always enable TLS
   - Enable authentication
   - Use Kubernetes secrets for certificates and credentials
   - Configure Network Policies

3. **Storage**:
   - Use high-performance StorageClasses (SSD)
   - Plan storage size according to your retention needs
   - Configure regular snapshots

4. **Monitoring**:
   - Monitor resources (CPU, memory, storage)
   - Configure alerts on components
   - Monitor operator logs

5. **Backup**:
   - Configure regular Elasticsearch snapshots
   - Test restoration regularly
   - Store backups outside the cluster

### Development/Test

- Use 1 replica for Elasticsearch and Kibana
- Disable TLS to simplify (not recommended in production)
- Use less performant StorageClasses
- Reduce allocated resources

## Configuration Examples

### Minimal Configuration (Development)

See `config/samples/logging_v1_efkstack.yaml` for a complete example.

### Production Configuration

See `helm-charts/efk-stack/values-production.yaml` for a production-ready configuration example.

## Support

- **Issues**: [GitHub Issues](https://github.com/zlorgoncho1/efk-operator/issues)
- **Documentation**: [docs/](docs/)
- **Contributions**: See [CONTRIBUTING.md](../CONTRIBUTING.md)

