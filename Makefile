# Makefile pour l'opérateur EFK
# Toutes les commandes s'exécutent dans Docker

# Variables
DOCKER_COMPOSE := docker-compose -f docker/docker-compose.yml
DOCKER_IMAGE := efk-operator-dev:latest
DOCKER_RUN := docker run --rm -v "$(PWD):/workspace" -w /workspace $(DOCKER_IMAGE)

.PHONY: help
help: ## Affiche cette aide
	@echo "Commandes disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: docker-build
docker-build: ## Construire l'image Docker de développement
	$(DOCKER_COMPOSE) build

.PHONY: dev-up
dev-up: ## Démarrer l'environnement de développement
	$(DOCKER_COMPOSE) up -d

.PHONY: dev-down
dev-down: ## Arrêter l'environnement de développement
	$(DOCKER_COMPOSE) down

.PHONY: dev-shell
dev-shell: ## Ouvrir un shell dans le conteneur de développement
	$(DOCKER_COMPOSE) exec dev bash

.PHONY: init
init: docker-build ## Initialiser le projet avec Kubebuilder (dans Docker)
	$(DOCKER_COMPOSE) run --rm dev bash -c "kubebuilder init --domain efk.crds.io --repo github.com/zlorgoncho1/efk-operator"

.PHONY: create-api
create-api: ## Créer l'API EFKStack (dans Docker)
	$(DOCKER_COMPOSE) run --rm dev bash -c "kubebuilder create api --group logging --version v1 --kind EFKStack --resource --controller"

.PHONY: manifests
manifests: ## Générer les manifests CRD
	$(DOCKER_COMPOSE) run --rm dev bash -c "make -f Makefile.kubebuilder manifests"

.PHONY: generate
generate: ## Générer le code (deepcopy, etc.)
	$(DOCKER_COMPOSE) run --rm dev bash -c "make -f Makefile.kubebuilder generate"

.PHONY: fmt
fmt: ## Formater le code
	$(DOCKER_COMPOSE) run --rm dev bash -c "make -f Makefile.kubebuilder fmt"

.PHONY: vet
vet: ## Vérifier le code avec go vet
	$(DOCKER_COMPOSE) run --rm dev bash -c "make -f Makefile.kubebuilder vet"

.PHONY: test
test: ## Exécuter les tests
	$(DOCKER_COMPOSE) run --rm dev bash -c "make -f Makefile.kubebuilder test"

.PHONY: build
build: ## Construire le binaire de l'opérateur
	$(DOCKER_COMPOSE) run --rm dev bash -c "make -f Makefile.kubebuilder build"

.PHONY: docker-build-operator
docker-build-operator: ## Construire l'image Docker de l'opérateur
	$(DOCKER_COMPOSE) run --rm dev bash -c "make -f Makefile.kubebuilder docker-build"

.PHONY: install
install: manifests ## Installer les CRDs dans le cluster
	$(DOCKER_COMPOSE) run --rm dev bash -c "kubectl apply -f config/crd/bases"

.PHONY: uninstall
uninstall: manifests ## Désinstaller les CRDs du cluster
	$(DOCKER_COMPOSE) run --rm dev bash -c "kubectl delete -f config/crd/bases"

.PHONY: deploy
deploy: manifests ## Déployer l'opérateur dans le cluster
	$(DOCKER_COMPOSE) run --rm dev bash -c "kubectl apply -k config/default"

.PHONY: undeploy
undeploy: ## Retirer l'opérateur du cluster
	$(DOCKER_COMPOSE) run --rm dev bash -c "kubectl delete -k config/default"

.PHONY: run
run: manifests generate ## Exécuter l'opérateur localement (hors cluster)
	$(DOCKER_COMPOSE) run --rm dev bash -c "make -f Makefile.kubebuilder run"

.PHONY: clean
clean: ## Nettoyer les fichiers générés
	$(DOCKER_COMPOSE) run --rm dev bash -c "make -f Makefile.kubebuilder clean"
	rm -rf bin/

.PHONY: test-quick
test-quick: ## Tests rapides (docker-build, manifests, build)
	@echo "=== Test 1: Construction de l'image Docker ==="
	$(DOCKER_COMPOSE) build
	@echo "=== Test 2: Génération des manifests ==="
	$(DOCKER_COMPOSE) run --rm dev bash -c "make -f Makefile.kubebuilder manifests"
	@echo "=== Test 3: Génération du code ==="
	$(DOCKER_COMPOSE) run --rm dev bash -c "make -f Makefile.kubebuilder generate"
	@echo "=== Test 4: Compilation ==="
	$(DOCKER_COMPOSE) run --rm dev bash -c "make -f Makefile.kubebuilder build"
	@echo "=== Tests rapides terminés avec succès! ==="

.PHONY: all
all: manifests generate fmt vet test build ## Exécuter toutes les étapes de build

