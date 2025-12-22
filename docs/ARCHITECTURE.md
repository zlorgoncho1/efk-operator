# Architecture du Projet EFK Stack Operator

Ce document décrit l'organisation et la structure du projet.

## Structure des répertoires

```
efk/
├── api/                          # Définitions des API (CRDs)
│   └── v1/
│       ├── efkstack_types.go     # Types Go pour EFKStack CRD
│       ├── groupversion_info.go  # Informations de version du groupe API
│       └── zz_generated.*        # Code généré (deepcopy, etc.)
│
├── config/                       # Manifests Kubernetes
│   ├── crd/                     # Custom Resource Definitions
│   │   ├── bases/               # CRDs de base
│   │   └── patches/             # Patches pour les CRDs
│   ├── default/                 # Configuration par défaut
│   ├── manager/                 # Manifests pour le manager
│   ├── rbac/                    # RBAC (rôles, bindings, service accounts)
│   └── samples/                 # Exemples de ressources EFKStack
│
├── docker/                       # Configuration Docker
│   ├── Dockerfile               # Image de production
│   ├── Dockerfile.dev           # Image de développement
│   └── docker-compose.yml       # Compose pour l'environnement de dev
│
├── docs/                        # Documentation
│   ├── README.md                # Index de la documentation
│   ├── GETTING_STARTED.md       # Guide de démarrage
│   ├── USER_GUIDE.md            # Guide utilisateur
│   ├── TESTING.md               # Guide complet de test
│   └── ARCHITECTURE.md          # Ce fichier
│
├── hack/                        # Scripts et utilitaires
│   └── boilerplate.go.txt       # En-tête pour les fichiers générés
│
├── helm-charts/                 # Charts Helm pour les composants
│   └── efk-stack/
│       ├── Chart.yaml           # Métadonnées du chart principal
│       ├── values.yaml           # Valeurs par défaut
│       ├── values-development.yaml  # Valeurs pour développement
│       ├── values-production.yaml   # Valeurs pour production
│       ├── elasticsearch/        # Chart Elasticsearch
│       ├── fluentbit/            # Chart Fluent Bit
│       └── kibana/               # Chart Kibana
│
├── internal/                     # Code interne (non exposé)
│   ├── controller/              # Contrôleurs
│   │   ├── efkstack_controller.go      # Contrôleur principal
│   │   ├── efkstack_controller_test.go # Tests du contrôleur
│   │   └── suite_test.go              # Suite de tests
│   └── helm/                     # Client Helm
│       └── client.go             # Client Helm pour déploiement
│
├── scripts/                      # Scripts utilitaires
│   ├── test-all.sh              # Script de test (Linux/Mac)
│   └── test-all.ps1             # Script de test (Windows)
│
├── .gitignore                    # Fichiers ignorés par Git
├── .dockerignore                 # Fichiers ignorés par Docker
├── go.mod                        # Dépendances Go
├── go.sum                        # Checksums des dépendances
├── main.go                       # Point d'entrée de l'opérateur
├── Makefile                      # Makefile principal
├── Makefile.kubebuilder          # Makefile Kubebuilder
├── PROJECT                       # Configuration Kubebuilder
└── README.md                     # Documentation principale
```

## Description des répertoires

### `api/`
Contient les définitions des Custom Resource Definitions (CRDs) en Go. Les fichiers `zz_generated.*` sont générés automatiquement par `controller-gen`.

### `config/`
Manifests Kubernetes pour :
- **crd/** : Définitions des CRDs
- **rbac/** : Rôles et permissions
- **manager/** : Déploiement du manager
- **samples/** : Exemples d'utilisation

### `docker/`
Configuration Docker pour le développement et la production :
- **Dockerfile.dev** : Environnement de développement avec tous les outils
- **Dockerfile** : Image de production pour l'opérateur
- **docker-compose.yml** : Orchestration de l'environnement de dev

### `docs/`
Toute la documentation du projet :
- Guides de démarrage et d'utilisation
- Documentation technique
- Guides de test

### `helm-charts/`
Charts Helm pour déployer les composants EFK :
- Charts individuels pour chaque composant
- Fichiers de valeurs pour différents environnements

### `internal/`
Code interne non exposé :
- **controller/** : Logique de réconciliation
- **helm/** : Client Helm pour gérer les déploiements

### `scripts/`
Scripts utilitaires pour automatiser les tâches (tests, build, etc.)

## Fichiers importants

### `main.go`
Point d'entrée de l'opérateur. Configure le manager et démarre les contrôleurs.

### `Makefile`
Contient les commandes principales pour :
- Build et test
- Génération de code
- Déploiement
- Gestion Docker

### `go.mod` / `go.sum`
Gestion des dépendances Go.

### `PROJECT`
Configuration Kubebuilder (domaine, repo, etc.)

## Organisation des Charts Helm

Chaque composant (Elasticsearch, Fluent Bit, Kibana) a son propre chart dans `helm-charts/efk-stack/` :

```
helm-charts/efk-stack/
├── elasticsearch/
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
│       ├── statefulset.yaml
│       ├── service.yaml
│       └── ...
├── fluentbit/
│   └── ...
└── kibana/
    └── ...
```

## Conventions

- **Code Go** : Suit les conventions Go standard
- **Charts Helm** : Suit les conventions Helm
- **Documentation** : Markdown dans `docs/`
- **Scripts** : Bash pour Linux/Mac, PowerShell pour Windows

## Fichiers générés

Les fichiers suivants sont générés automatiquement et ne doivent **pas** être modifiés manuellement :

- `api/**/zz_generated.*` : Code généré par controller-gen
- `config/crd/bases/*.yaml` : CRDs générés
- `bin/` : Binaires compilés

## Pour en savoir plus

- [Guide de démarrage](GETTING_STARTED.md)
- [Guide de test](TESTING.md)
- [Guide utilisateur](USER_GUIDE.md)
- [README principal](../README.md)

