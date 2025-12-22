---
name: Bug Report
about: Create a report to help us improve
title: '[BUG] '
labels: bug
assignees: ''
---

## Bug Description

A clear and concise description of what the bug is.

## Steps to Reproduce

1. Deploy EFKStack with configuration: '...'
2. Run command: '...'
3. See error: '...'

## Expected Behavior

A clear and concise description of what you expected to happen.

## Actual Behavior

A clear and concise description of what actually happened.

## Environment

- **Kubernetes Version**: 
- **Operator Version**: 
- **EFKStack Version**: 
- **Cloud Provider**: (e.g., AWS, GCP, Azure, on-premises)
- **Storage Class**: 

## Configuration

```yaml
# Paste your EFKStack configuration here (remove sensitive data)
```

## Logs

```
# Paste relevant logs here
```

### Operator Logs
```bash
kubectl logs -n system deployment/controller-manager
```

### Component Logs
```bash
# Elasticsearch
kubectl logs -n <namespace> -l app=elasticsearch

# Fluent Bit
kubectl logs -n <namespace> -l app=fluent-bit

# Kibana
kubectl logs -n <namespace> -l app=kibana
```

## Additional Context

Add any other context about the problem here.

## Checklist

- [ ] I have searched existing issues to ensure this is not a duplicate
- [ ] I have included all relevant information
- [ ] I have removed sensitive data from the configuration

