# EFK Stack Operator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21-blue.svg)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.24%2B-blue.svg)](https://kubernetes.io/)

Un opérateur Kubernetes pour déployer facilement une stack EFK (Elasticsearch, Fluent Bit, Kibana) production-ready avec des CRDs et des templates Helm flexibles.

**Repository**: [https://github.com/zlorgoncho1/efk-operator](https://github.com/zlorgoncho1/efk-operator)

## Vue d'ensemble

Cet opérateur permet de déployer une stack EFK complète en créant simplement une ressource `EFKStack` personnalisée. L'opérateur gère automatiquement le déploiement, la configuration et la mise à jour des composants.

## Architecture

- **Elasticsearch**: Stockage et indexation des logs
- **Fluent Bit**: Collecte des logs depuis tous les nœuds Kubernetes
- **Kibana**: Interface de visualisation et d'analyse

## Quick Start

### Pour les utilisateurs

Voir le [Guide Utilisateur](docs/USER_GUIDE.md) pour l'installation complète et l'utilisation.

**Installation rapide** :

```bash
# Installer les CRDs (depuis GitHub)
kubectl apply -f https://raw.githubusercontent.com/zlorgoncho1/efk-operator/main/efk/config/crd/bases/logging.efk.crds.io_efkstacks.yaml

# Déployer l'opérateur (nécessite kustomize)
kubectl create namespace system
kubectl apply -k https://github.com/zlorgoncho1/efk-operator.git/efk/config/default

# Créer une stack EFK (exemple)
kubectl apply -f https://raw.githubusercontent.com/zlorgoncho1/efk-operator/main/efk/config/samples/logging_v1_efkstack.yaml
```

### Pour les développeurs

Voir le [Guide de Démarrage](docs/GETTING_STARTED.md) pour configurer l'environnement de développement.

**Démarrage rapide** :

```bash
# Construire l'environnement Docker
make docker-build

# Tests rapides
make test-quick
```

## Prérequis

- Kubernetes 1.24+
- `kubectl` configuré pour accéder à votre cluster
- Docker et Docker Compose (pour le développement)

## Installation

Pour les instructions détaillées d'installation, consultez le [Guide Utilisateur](docs/USER_GUIDE.md#installation-de-lopérateur).

## Utilisation

### Créer une stack EFK

Créez un fichier `efkstack.yaml`:

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

Appliquez la ressource:

```bash
kubectl apply -f efkstack.yaml
```

### Vérifier le statut

```bash
kubectl get efkstack my-efk-stack -n efk-system
kubectl describe efkstack my-efk-stack -n efk-system
```

## Configuration

### Spécification EFKStack

#### Elasticsearch

- `version`: Version d'Elasticsearch (ex: "8.11.0")
- `replicas`: Nombre de replicas (minimum 1, recommandé 3+ pour production)
- `resources`: Ressources CPU et mémoire
- `storage`: Configuration du stockage persistant
- `security`: Configuration TLS et authentification

#### Fluent Bit

- `version`: Version de Fluent Bit (ex: "2.2.0")
- `resources`: Ressources CPU et mémoire
- `config`: Configuration personnalisée (optionnel)

#### Kibana

- `version`: Version de Kibana (ex: "8.11.0")
- `replicas`: Nombre de replicas (recommandé 2+ pour HA)
- `resources`: Ressources CPU et mémoire
- `ingress`: Configuration Ingress pour l'accès externe

## Développement

### Structure du projet

```
efk/
├── api/v1/              # Définitions CRD
├── config/              # Manifests Kubernetes
├── controllers/         # Logique de réconciliation
├── helm-charts/         # Charts Helm pour les composants
└── internal/helm/       # Client Helm
```

### Commandes utiles

```bash
# Générer le code
make generate

# Formater le code
make fmt

# Exécuter les tests
make test

# Construire l'opérateur
make build

# Exécuter localement (hors cluster)
make run
```

### Tests

Les tests utilisent Ginkgo et Gomega:

```bash
make test
```

## Documentation

Toute la documentation est disponible dans le dossier [`docs/`](docs/):

### Pour les utilisateurs
- **[Guide Utilisateur](docs/USER_GUIDE.md)** - Installation, utilisation et dépannage
- **[Architecture](docs/ARCHITECTURE.md)** - Structure du projet

### Pour les développeurs
- **[Guide de Démarrage](docs/GETTING_STARTED.md)** - Configuration de l'environnement de développement
- **[Guide de Test](docs/TESTING.md)** - Tests et validation
- **[Architecture](docs/ARCHITECTURE.md)** - Structure du projet

Voir [docs/README.md](docs/README.md) pour l'index complet de la documentation.

## Production

### Recommandations

1. **Haute disponibilité**: Utilisez au moins 3 replicas pour Elasticsearch et 2+ pour Kibana
2. **Sécurité**: Activez TLS et l'authentification
3. **Stockage**: Utilisez des storage classes performantes (SSD)
4. **Monitoring**: Configurez le monitoring des composants
5. **Backup**: Mettez en place des stratégies de backup pour Elasticsearch

### Exemple de configuration production

Voir `efk/helm-charts/efk-stack/values-production.yaml` pour un exemple de configuration production-ready.

## Dépannage

### Vérifier les logs de l'opérateur

```bash
kubectl logs -n system deployment/controller-manager
```

### Vérifier le statut des composants

```bash
# Elasticsearch
kubectl get statefulset -n efk-system

# Fluent Bit
kubectl get daemonset -n efk-system

# Kibana
kubectl get deployment -n efk-system
```

## Contribution

Les contributions sont les bienvenues! Veuillez consulter [CONTRIBUTING.md](CONTRIBUTING.md) pour les guidelines de contribution.

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

## Licence

Apache License 2.0

