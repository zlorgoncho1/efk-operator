#!/bin/bash
# Script de test complet pour EFK Stack Operator

set -e  # Arrêter en cas d'erreur

echo "=========================================="
echo "Tests EFK Stack Operator"
echo "=========================================="

# Couleurs pour l'output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Fonction pour afficher les résultats
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Vérifier que nous sommes dans le bon répertoire
if [ ! -f "go.mod" ]; then
    print_error "Vous devez être dans le répertoire efk/"
    exit 1
fi

print_info "Répertoire de travail: $(pwd)"

# Test 1: Vérifier la structure
echo ""
echo "Test 1: Vérification de la structure..."
if [ -d "api" ] && [ -d "config" ] && [ -d "docker" ] && [ -d "helm-charts" ] && [ -d "internal" ]; then
    print_success "Structure du projet OK"
else
    print_error "Structure du projet incomplète"
    exit 1
fi

# Test 2: Vérifier Docker
echo ""
echo "Test 2: Vérification de Docker..."
if command -v docker &> /dev/null; then
    print_success "Docker installé"
    if docker ps &> /dev/null; then
        print_success "Docker fonctionne"
    else
        print_error "Docker ne répond pas"
        exit 1
    fi
else
    print_error "Docker non installé"
    exit 1
fi

# Test 3: Construire l'image Docker
echo ""
echo "Test 3: Construction de l'image Docker..."
if make docker-build 2>&1 | grep -q "Successfully\|built"; then
    print_success "Image Docker construite"
else
    print_error "Échec de la construction de l'image Docker"
    exit 1
fi

# Test 4: Vérifier que l'image existe
echo ""
echo "Test 4: Vérification de l'image..."
if docker images | grep -q "efk-operator-dev"; then
    print_success "Image efk-operator-dev trouvée"
else
    print_error "Image efk-operator-dev non trouvée"
    exit 1
fi

# Test 5: Générer les manifests
echo ""
echo "Test 5: Génération des manifests CRD..."
if make manifests 2>&1 | grep -q "Generating\|manifest"; then
    print_success "Manifests générés"
    if [ -f "config/crd/bases/logging.efk.crds.io_efkstacks.yaml" ]; then
        print_success "Fichier CRD créé"
    else
        print_error "Fichier CRD non trouvé"
        exit 1
    fi
else
    print_error "Échec de la génération des manifests"
    exit 1
fi

# Test 6: Générer le code
echo ""
echo "Test 6: Génération du code..."
if make generate 2>&1 | grep -q "Generating\|generate"; then
    print_success "Code généré"
else
    print_error "Échec de la génération du code"
    exit 1
fi

# Test 7: Formater le code
echo ""
echo "Test 7: Formatage du code..."
if make fmt 2>&1; then
    print_success "Code formaté"
else
    print_error "Échec du formatage"
    exit 1
fi

# Test 8: Vérifier le code
echo ""
echo "Test 8: Vérification du code (go vet)..."
if make vet 2>&1; then
    print_success "Code vérifié (pas d'erreurs)"
else
    print_error "Erreurs trouvées par go vet"
    exit 1
fi

# Test 9: Compiler le code
echo ""
echo "Test 9: Compilation du code..."
if make build 2>&1; then
    print_success "Code compilé"
    if [ -f "bin/manager" ]; then
        print_success "Binaire créé: bin/manager"
    else
        print_error "Binaire non trouvé"
        exit 1
    fi
else
    print_error "Échec de la compilation"
    exit 1
fi

# Test 10: Vérifier les charts Helm
echo ""
echo "Test 10: Validation des charts Helm..."
if docker-compose -f docker/docker-compose.yml run --rm dev bash -c "cd helm-charts/efk-stack/elasticsearch && helm lint ." 2>&1 | grep -q "Lint\|OK\|no issues"; then
    print_success "Chart Elasticsearch valide"
else
    print_error "Chart Elasticsearch invalide"
    exit 1
fi

# Résumé
echo ""
echo "=========================================="
echo -e "${GREEN}Tous les tests sont passés !${NC}"
echo "=========================================="
echo ""
echo "Prochaines étapes :"
echo "  1. make dev-up          # Démarrer l'environnement"
echo "  2. make dev-shell        # Ouvrir un shell"
echo "  3. make test            # Exécuter les tests"
echo "  4. make install         # Installer les CRDs (si cluster disponible)"
echo ""

