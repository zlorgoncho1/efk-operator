# Dockerfile - Image de production pour l'opérateur EFK
# Multi-stage build pour une image optimale

# Stage 1: Build
FROM golang:1.21-alpine AS builder

WORKDIR /workspace

# Installer les dépendances de build
RUN apk add --no-cache git make

# Installer controller-gen pour générer le code
RUN go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0

# Copier les fichiers de dépendances
COPY go.mod go.mod
COPY go.sum go.sum

# Copier le code source (nécessaire pour go mod tidy avec le replace)
COPY . .

# Mettre à jour go.sum avec le replace de go.mod
WORKDIR /workspace
RUN go mod tidy

# Forcer la mise à jour de k8s.io/cli-runtime pour résoudre le problème de go.sum
RUN go get k8s.io/cli-runtime@v0.29.0 && go mod tidy

# Télécharger les dépendances
RUN go mod download

# Générer le code (DeepCopy, etc.) - limiter aux packages locaux uniquement pour éviter les erreurs avec Helm
RUN controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./api/..."

# Construire l'opérateur
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -o manager main.go

# Stage 2: Runtime
FROM alpine:latest

# Installer les dépendances runtime
RUN apk --no-cache add ca-certificates tzdata

# Définir le WORKDIR pour que les chemins relatifs fonctionnent
WORKDIR /workspace

# Copier le binaire vers un répertoire système standard
COPY --from=builder /workspace/manager /usr/local/bin/manager

# Copier les charts Helm (nécessaires pour le controller)
COPY --from=builder /workspace/helm-charts /workspace/helm-charts

# Utiliser un utilisateur non-root pour la sécurité
# Utiliser UID 65532 (nobody) pour correspondre à manager.yaml
RUN addgroup -g 65532 operator && \
    adduser -D -u 65532 -G operator operator && \
    chown operator:operator /usr/local/bin/manager && \
    chmod +x /usr/local/bin/manager && \
    chown -R operator:operator /workspace/helm-charts

USER operator:operator

# Exposer le port des métriques
EXPOSE 8080

# Point d'entrée
ENTRYPOINT ["/usr/local/bin/manager"]

