# Dockerfile - Image de production pour l'opérateur EFK
# Multi-stage build pour une image optimale

# Stage 1: Build
FROM golang:1.21-alpine AS builder

WORKDIR /workspace

# Installer les dépendances de build
RUN apk add --no-cache git make

# Copier les fichiers de dépendances
COPY go.mod go.mod
COPY go.sum go.sum

# Télécharger les dépendances
WORKDIR /workspace
RUN go mod download

# Copier le code source
COPY . .

# Construire l'opérateur
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -o manager main.go

# Stage 2: Runtime
FROM alpine:latest

# Installer les dépendances runtime
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copier le binaire depuis le stage de build
COPY --from=builder /workspace/manager .

# Utiliser un utilisateur non-root pour la sécurité
RUN addgroup -g 1000 operator && \
    adduser -D -u 1000 -G operator operator && \
    chown -R operator:operator /root

USER operator:operator

# Exposer le port des métriques
EXPOSE 8080

# Point d'entrée
ENTRYPOINT ["./manager"]

