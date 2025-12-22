# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project structure
- EFKStack CRD definition
- Controller implementation with Helm integration
- Helm charts for Elasticsearch, Fluent Bit, and Kibana
- Docker-based development environment
- Comprehensive documentation
- User guide for installation and usage
- CI/CD workflow

### Changed
- N/A

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- N/A

## [0.1.0] - 2025-12-22

### Added
- Initial release
- EFKStack Custom Resource Definition
- Kubernetes operator for managing EFK stack deployments
- Helm charts for:
  - Elasticsearch (StatefulSet with persistent storage)
  - Fluent Bit (DaemonSet for log collection)
  - Kibana (Deployment with Ingress support)
- Production-ready configurations:
  - High availability support
  - TLS and authentication
  - Resource management
  - Storage configuration
  - Network policies
- Docker-based development environment
- Comprehensive documentation:
  - Getting started guide
  - User guide
  - Architecture documentation
  - Testing guide

[Unreleased]: https://github.com/zlorgoncho1/efk-operator/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/zlorgoncho1/efk-operator/releases/tag/v0.1.0

