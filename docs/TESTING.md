# Guide de Test - EFK Stack Operator

Ce guide vous permet de tester que tout est opérationnel avant de continuer.

## Prérequis

- Docker et Docker Compose installés
- Accès à un cluster Kubernetes (optionnel pour les tests de base)
- `kubectl` configuré (optionnel)

## Tests à effectuer

### 1. Vérification de la structure du projet

```bash
cd efk
ls -la
```

Vous devriez voir :
- `api/` - Définitions CRD
- `config/` - Manifests Kubernetes
- `docker/` - Configuration Docker
- `helm-charts/` - Charts Helm
- `internal/` - Code interne
- `main.go` - Point d'entrée
- `go.mod` - Dépendances
- `Makefile` - Makefile principal

### 2. Test de l'environnement Docker

```bash
# Construire l'image de développement
make docker-build

# Vérifier que l'image est créée
docker images | grep efk-operator-dev

# Démarrer l'environnement
make dev-up

# Vérifier que le conteneur tourne
docker ps | grep efk-operator-dev

# Tester un shell dans le conteneur
make dev-shell
# Dans le shell, tester :
#   - kubebuilder version
#   - helm version
#   - kubectl version
#   - go version
# Sortir avec 'exit'
```

### 3. Test de la génération des manifests CRD

```bash
# Générer les manifests CRD
make manifests

# Vérifier que les fichiers sont générés
ls -la config/crd/bases/

# Vous devriez voir :
# logging.efk.crds.io_efkstacks.yaml
```

### 4. Test de la génération du code

```bash
# Générer le code (deepcopy, etc.)
make generate

# Vérifier que les fichiers générés sont créés
ls -la api/v1/zz_generated.*
```

### 5. Test de compilation

```bash
# Formater le code
make fmt

# Vérifier le code
make vet

# Construire le binaire
make build

# Vérifier que le binaire est créé
ls -la bin/manager
```

### 6. Test des dépendances Go

```bash
# Dans le conteneur Docker
make dev-shell

# Dans le shell :
cd /workspace
go mod download
go mod verify
go mod tidy
exit
```

### 7. Test de validation des fichiers YAML

```bash
# Vérifier les manifests Kubernetes
make dev-shell
# Dans le shell :
kubectl apply --dry-run=client -f config/crd/bases/
kubectl apply --dry-run=client -f config/rbac/
kubectl apply --dry-run=client -f config/manager/
exit
```

### 8. Test des charts Helm (validation)

```bash
make dev-shell
# Dans le shell :
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

### 9. Test des imports Go

```bash
make dev-shell
# Dans le shell :
go build ./...
go test ./... -v
exit
```

### 10. Test complet (tous les checks)

```bash
# Exécuter tous les checks
make all

# Ou étape par étape :
make manifests
make generate
make fmt
make vet
make test
make build
```

## Checklist de validation

- [ ] Structure du projet complète
- [ ] Image Docker se construit sans erreur
- [ ] Conteneur démarre correctement
- [ ] Outils disponibles dans le conteneur (kubebuilder, helm, kubectl, go)
- [ ] Manifests CRD générés
- [ ] Code généré (deepcopy)
- [ ] Code se compile sans erreur
- [ ] Code passe `go vet`
- [ ] Tests unitaires passent
- [ ] Charts Helm valides (lint)
- [ ] Imports Go corrects
- [ ] Binaire créé et exécutable

## Dépannage

### Erreur : "docker-compose: command not found"
```bash
# Installer Docker Compose ou utiliser :
docker compose -f docker/docker-compose.yml build
```

### Erreur : "go.mod: module path mismatch"
```bash
# Vérifier que go.mod utilise le bon chemin :
# module github.com/zlorgoncho1/efk-operator
```

### Erreur : "import path not found"
```bash
# Dans le conteneur :
go mod download
go mod tidy
```

### Erreur : "kubebuilder: command not found"
```bash
# Vérifier que l'image Docker est bien construite avec kubebuilder
make docker-build
```

## Tests avancés (optionnel)

### Test avec un cluster Kubernetes local (kind)

```bash
# Installer kind dans le conteneur
make dev-shell
# Dans le shell :
kind create cluster --name efk-test
kubectl cluster-info

# Installer les CRDs
make install

# Créer une ressource EFKStack de test
kubectl apply -f config/samples/logging_v1_efkstack.yaml

# Vérifier
kubectl get efkstack

# Nettoyer
kind delete cluster --name efk-test
exit
```
