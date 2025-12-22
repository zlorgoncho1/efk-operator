# Getting Started - EFK Stack Operator

Guide de démarrage rapide pour les développeurs.

## Prérequis

- Docker et Docker Compose installés
- Accès à un cluster Kubernetes (optionnel pour les tests de base)
- `kubectl` configuré (optionnel)

## Installation Rapide

### 1. Construire l'environnement Docker

```bash
cd efk
make docker-build
```

**Temps estimé** : 2-3 minutes (première fois)

**Résultat attendu** : Image `efk-operator-dev:latest` construite

### 2. Tests automatiques rapides

```bash
make test-quick
```

**Temps estimé** : 1-2 minutes

Cette commande exécute :
- Construction de l'image Docker
- Génération des manifests CRD
- Génération du code (deepcopy)
- Compilation du binaire

### 3. Vérifier l'installation

```bash
# Vérifier que les fichiers sont créés
ls config/crd/bases/
ls bin/manager
```

**Résultat attendu** :
- ✅ Fichier CRD généré : `config/crd/bases/logging.efk.crds.io_efkstacks.yaml`
- ✅ Binaire créé : `bin/manager`

## Environnement de Développement

### Démarrer l'environnement

```bash
make dev-up
```

### Ouvrir un shell de développement

```bash
make dev-shell
```

Dans le shell, vous avez accès à :
- `go` - Go 1.21.13
- `helm` - Helm v3.19.4
- `kubectl` - kubectl (dernière version)
- `controller-gen` - controller-gen v0.14.0
- `kustomize` - kustomize v5.8.0

### Commandes de développement

```bash
# Générer les manifests CRD
make manifests

# Générer le code (deepcopy)
make generate

# Formater le code
make fmt

# Vérifier le code (avec note sur go vet)
make vet

# Compiler
make build

# Exécuter localement (hors cluster)
make run
```

## Tests Rapides

### Test manuel en 3 étapes

1. **Construire l'image** : `make docker-build`
2. **Tests automatiques** : `make test-quick`
3. **Vérifier** : `ls config/crd/bases/ && ls bin/`

### Scripts de test (optionnel)

#### Linux/Mac :
```bash
chmod +x scripts/test-all.sh
./scripts/test-all.sh
```

#### Windows (PowerShell) :
```powershell
.\scripts\test-all.ps1
```

## Problèmes Courants

### Erreur : "docker-compose: command not found"
**Solution** : Utilisez `docker compose` (sans tiret) ou installez Docker Compose

### Erreur : "make: command not found"
**Windows** : Installez [Make for Windows](https://www.gnu.org/software/make/) ou utilisez WSL

### Erreur lors du build Docker
**Solution** : Vérifiez que Docker Desktop est démarré

### Erreur go vet (non bloquant)
Voir la section [Known Issues](../README.md#known-issues) dans le README principal.

## Prochaines Étapes

Une fois l'environnement configuré :

1. **Développement** : Commencez à modifier le code dans `internal/controller/`
2. **Tests** : Consultez [TESTING.md](TESTING.md) pour les tests complets
3. **Architecture** : Consultez [ARCHITECTURE.md](ARCHITECTURE.md) pour comprendre la structure
4. **Utilisation** : Consultez [USER_GUIDE.md](USER_GUIDE.md) pour déployer dans un cluster

## Documentation Complète

- **[TESTING.md](TESTING.md)** - Guide complet de test et validation
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Architecture et structure du projet
- **[USER_GUIDE.md](USER_GUIDE.md)** - Guide utilisateur pour installation/utilisation

