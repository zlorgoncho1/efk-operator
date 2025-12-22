# Guide Utilisateur - EFK Stack Operator

Guide complet pour installer et utiliser l'opérateur EFK Stack dans votre cluster Kubernetes.

## Table des matières

1. [Prérequis](#prérequis)
2. [Installation de l'opérateur](#installation-de-lopérateur)
3. [Utilisation basique](#utilisation-basique)
4. [Configuration avancée](#configuration-avancée)
5. [Dépannage](#dépannage)
6. [Mise à jour et maintenance](#mise-à-jour-et-maintenance)

## Prérequis

### Exigences du cluster

- Kubernetes 1.24 ou supérieur
- `kubectl` configuré et connecté à votre cluster
- Accès administrateur au cluster (pour installer les CRDs et l'opérateur)
- StorageClass configurée (pour les volumes persistants d'Elasticsearch)

### Outils requis

- `kubectl` (version compatible avec votre cluster)
- `helm` v3.0+ (optionnel, pour certaines méthodes d'installation)

## Installation de l'opérateur

### Méthode 1 : Installation via Manifests (Recommandée)

#### Étape 1 : Installer les CRDs

Vous pouvez installer directement depuis GitHub sans cloner le repository :

```bash
kubectl apply -f https://raw.githubusercontent.com/zlorgoncho1/efk-operator/main/efk/config/crd/bases/logging.efk.crds.io_efkstacks.yaml
```

**Alternative** : Si vous préférez cloner le repository pour développement :

```bash
git clone https://github.com/zlorgoncho1/efk-operator.git
cd efk-operator/efk
kubectl apply -f config/crd/bases/logging.efk.crds.io_efkstacks.yaml
```

Vérifier l'installation :

```bash
kubectl get crd efkstacks.logging.efk.crds.io
```

#### Étape 2 : Déployer l'opérateur

**Option A : Depuis GitHub (sans cloner)**

```bash
# Créer le namespace pour l'opérateur
kubectl create namespace system

# Déployer l'opérateur (nécessite kustomize)
kubectl apply -k https://github.com/zlorgoncho1/efk-operator.git/efk/config/default
```

**Option B : Après avoir cloné le repository**

```bash
# Créer le namespace pour l'opérateur
kubectl create namespace system

# Déployer l'opérateur
kubectl apply -k config/default
```

**Note** : Avant de déployer, vous devez mettre à jour l'image de l'opérateur dans `config/default/manager_image_patch.yaml` avec votre image Docker, ou utiliser `kustomize edit set image` pour modifier l'image.

#### Étape 5 : Vérifier le déploiement

```bash
# Vérifier que l'opérateur est déployé
kubectl get deployment -n system controller-manager

# Vérifier les pods
kubectl get pods -n system

# Vérifier les logs
kubectl logs -n system deployment/controller-manager -f
```

### Méthode 2 : Installation via Helm (Si chart disponible)

```bash
# Ajouter le repository Helm (quand disponible)
helm repo add efk-operator https://zlorgoncho1.github.io/efk-operator

# Installer l'opérateur
helm install efk-operator efk-operator/efk-operator \
  --namespace system \
  --create-namespace
```

## Utilisation basique

### Créer votre première stack EFK

#### Exemple minimal

Créez un fichier `my-efk-stack.yaml` :

```yaml
apiVersion: logging.efk.crds.io/v1
kind: EFKStack
metadata:
  name: my-efk-stack
  namespace: efk-system
spec:
  elasticsearch:
    version: "8.11.0"
    replicas: 1
    resources:
      requests:
        cpu: "1"
        memory: "2Gi"
      limits:
        cpu: "2"
        memory: "4Gi"
    storage:
      size: "50Gi"
      storageClassName: "standard"
  
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
    replicas: 1
    resources:
      requests:
        cpu: "500m"
        memory: "512Mi"
      limits:
        cpu: "1"
        memory: "1Gi"
```

#### Appliquer la ressource

```bash
# Créer le namespace (si nécessaire)
kubectl create namespace efk-system

# Appliquer la ressource
kubectl apply -f my-efk-stack.yaml
```

#### Vérifier le statut

```bash
# Voir le statut de la ressource
kubectl get efkstack my-efk-stack -n efk-system

# Voir les détails
kubectl describe efkstack my-efk-stack -n efk-system

# Vérifier les composants déployés
kubectl get all -n efk-system
```

### Exemple production

Voir `config/samples/logging_v1_efkstack.yaml` pour un exemple complet avec :
- Haute disponibilité (3 replicas Elasticsearch, 2 Kibana)
- Configuration de sécurité (TLS, authentification)
- Ingress pour Kibana
- Ressources optimisées

## Configuration avancée

### Options de configuration Elasticsearch

```yaml
spec:
  elasticsearch:
    version: "8.11.0"              # Version d'Elasticsearch
    replicas: 3                    # Nombre de replicas (minimum 1, recommandé 3+)
    resources:
      requests:
        cpu: "2"
        memory: "4Gi"
      limits:
        cpu: "4"
        memory: "8Gi"
    storage:
      size: "100Gi"                # Taille du stockage (format: 100Gi, 500Mi)
      storageClassName: "fast-ssd" # StorageClass à utiliser
      volumeType: "persistentVolumeClaim"
    security:
      tlsEnabled: true             # Activer TLS
      authEnabled: true            # Activer l'authentification
      tlsSecretName: "es-tls"      # Secret contenant les certificats TLS
      authSecretName: "es-auth"    # Secret contenant les credentials
    config:                        # Configuration additionnelle
      discovery.type: "kubernetes"
      xpack.security.enabled: "true"
```

### Options de configuration Fluent Bit

```yaml
spec:
  fluentBit:
    version: "2.2.0"
    resources:
      requests:
        cpu: "100m"
        memory: "128Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
    config:
      input: |
        [INPUT]
          Name tail
          Path /var/log/containers/*.log
          Parser docker
      filter: |
        [FILTER]
          Name kubernetes
          Match kube.*
      output: |
        [OUTPUT]
          Name es
          Match *
          Host elasticsearch
          Port 9200
```

### Options de configuration Kibana

```yaml
spec:
  kibana:
    version: "8.11.0"
    replicas: 2                    # Nombre de replicas (recommandé 2+)
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
        cert-manager.io/cluster-issuer: "letsencrypt-prod"
        nginx.ingress.kubernetes.io/ssl-redirect: "true"
      tls:
        - hosts:
            - kibana.example.com
          secretName: kibana-tls
```

### Configuration globale

```yaml
spec:
  global:
    storageClass: "fast-ssd"       # StorageClass par défaut
    imageRegistry: "registry.example.com"  # Registry d'images personnalisé
    tls:
      enabled: true
      secretName: "global-tls"
```

## Dépannage

### Vérifier le statut de l'opérateur

```bash
# Vérifier que l'opérateur fonctionne
kubectl get pods -n system -l control-plane=controller-manager

# Voir les logs de l'opérateur
kubectl logs -n system deployment/controller-manager -f
```

### Vérifier le statut de la stack EFK

```bash
# Voir le statut de la ressource EFKStack
kubectl get efkstack -n efk-system

# Voir les détails (phase, état des composants)
kubectl describe efkstack my-efk-stack -n efk-system

# Voir les événements
kubectl get events -n efk-system --sort-by='.lastTimestamp'
```

### Vérifier les composants individuels

```bash
# Elasticsearch
kubectl get statefulset -n efk-system
kubectl get pods -n efk-system -l app=elasticsearch
kubectl logs -n efk-system -l app=elasticsearch -f

# Fluent Bit
kubectl get daemonset -n efk-system
kubectl get pods -n efk-system -l app=fluent-bit
kubectl logs -n efk-system -l app=fluent-bit -f

# Kibana
kubectl get deployment -n efk-system
kubectl get pods -n efk-system -l app=kibana
kubectl logs -n efk-system -l app=kibana -f
```

### Problèmes courants

#### La stack reste en phase "Pending" ou "Deploying"

**Causes possibles** :
- L'opérateur n'est pas démarré
- Problème de permissions RBAC
- Problème de stockage (StorageClass non disponible)

**Solutions** :
```bash
# Vérifier l'opérateur
kubectl get pods -n system

# Vérifier les permissions
kubectl describe role -n system manager-role

# Vérifier les StorageClasses
kubectl get storageclass
```

#### Elasticsearch ne démarre pas

**Causes possibles** :
- Problème de stockage (PVC non créé)
- Ressources insuffisantes
- Problème de configuration

**Solutions** :
```bash
# Vérifier les PVCs
kubectl get pvc -n efk-system

# Vérifier les événements
kubectl describe statefulset -n efk-system

# Vérifier les ressources disponibles
kubectl top nodes
```

#### Fluent Bit ne collecte pas les logs

**Causes possibles** :
- Elasticsearch n'est pas accessible
- Configuration Fluent Bit incorrecte
- Problème de permissions

**Solutions** :
```bash
# Vérifier la connexion à Elasticsearch
kubectl exec -n efk-system <fluent-bit-pod> -- curl http://<elasticsearch-service>:9200

# Vérifier les logs Fluent Bit
kubectl logs -n efk-system -l app=fluent-bit

# Vérifier la configuration
kubectl get configmap -n efk-system
```

#### Kibana n'est pas accessible

**Causes possibles** :
- Ingress non configuré ou mal configuré
- Elasticsearch non accessible depuis Kibana
- Problème de certificats TLS

**Solutions** :
```bash
# Vérifier l'Ingress
kubectl get ingress -n efk-system

# Vérifier les services
kubectl get svc -n efk-system

# Tester la connexion interne
kubectl port-forward -n efk-system svc/kibana 5601:5601
# Puis accéder à http://localhost:5601
```

### Commandes utiles

```bash
# Voir tous les releases Helm créés par l'opérateur
helm list -n efk-system

# Voir les détails d'un release Helm
helm status <release-name> -n efk-system

# Voir l'historique d'un release
helm history <release-name> -n efk-system

# Voir les ressources créées
kubectl get all -n efk-system

# Voir les secrets créés
kubectl get secrets -n efk-system

# Voir les configmaps
kubectl get configmaps -n efk-system
```

## Mise à jour et maintenance

### Mettre à jour l'opérateur

```bash
# 1. Récupérer la nouvelle version
git pull origin main

# 2. Régénérer les manifests (si nécessaire)
make manifests

# 3. Mettre à jour les CRDs
kubectl apply -f config/crd/bases/logging.efk.crds.io_efkstacks.yaml

# 4. Mettre à jour l'opérateur
kubectl apply -k config/default

# 5. Vérifier le déploiement
kubectl rollout status deployment/controller-manager -n system
```

### Mettre à jour une stack EFK

Pour mettre à jour une stack existante, modifiez simplement la ressource EFKStack :

```bash
# Modifier la ressource
kubectl edit efkstack my-efk-stack -n efk-system

# Ou appliquer un nouveau fichier
kubectl apply -f my-efk-stack-updated.yaml
```

L'opérateur détectera les changements et mettra à jour les composants via Helm.

### Backup et restauration

#### Backup d'Elasticsearch

```bash
# Créer un snapshot (nécessite un repository configuré)
kubectl exec -n efk-system <elasticsearch-pod> -- \
  curl -X PUT "localhost:9200/_snapshot/my_backup/snapshot_1?wait_for_completion=true"
```

#### Restauration

```bash
# Restaurer depuis un snapshot
kubectl exec -n efk-system <elasticsearch-pod> -- \
  curl -X POST "localhost:9200/_snapshot/my_backup/snapshot_1/_restore"
```

### Désinstallation

#### Supprimer une stack EFK

```bash
# Supprimer la ressource EFKStack
kubectl delete efkstack my-efk-stack -n efk-system

# L'opérateur supprimera automatiquement tous les composants
# Vérifier que tout est supprimé
kubectl get all -n efk-system
```

#### Désinstaller l'opérateur

```bash
# Supprimer l'opérateur
kubectl delete -k config/default

# Supprimer les CRDs (ATTENTION : supprime aussi toutes les ressources EFKStack)
kubectl delete -f config/crd/bases/logging.efk.crds.io_efkstacks.yaml
```

## Bonnes pratiques

### Production

1. **Haute disponibilité** :
   - Utilisez au moins 3 replicas pour Elasticsearch
   - Utilisez au moins 2 replicas pour Kibana
   - Configurez Pod Disruption Budgets

2. **Sécurité** :
   - Activez toujours TLS
   - Activez l'authentification
   - Utilisez des secrets Kubernetes pour les certificats et credentials
   - Configurez Network Policies

3. **Stockage** :
   - Utilisez des StorageClasses performantes (SSD)
   - Planifiez la taille du stockage selon vos besoins de rétention
   - Configurez des snapshots réguliers

4. **Monitoring** :
   - Surveillez les ressources (CPU, mémoire, stockage)
   - Configurez des alertes sur les composants
   - Surveillez les logs de l'opérateur

5. **Backup** :
   - Configurez des snapshots Elasticsearch réguliers
   - Testez la restauration régulièrement
   - Stockez les backups hors du cluster

### Développement/Test

- Utilisez 1 replica pour Elasticsearch et Kibana
- Désactivez TLS pour simplifier (non recommandé en production)
- Utilisez des StorageClasses moins performantes
- Réduisez les ressources allouées

## Exemples de configurations

### Configuration minimale (développement)

Voir `config/samples/logging_v1_efkstack.yaml` pour un exemple complet.

### Configuration production

Voir `helm-charts/efk-stack/values-production.yaml` pour un exemple de configuration production-ready.

## Support

- **Issues** : [GitHub Issues](https://github.com/zlorgoncho1/efk-operator/issues)
- **Documentation** : [docs/](docs/)
- **Contributions** : Voir [CONTRIBUTING.md](../CONTRIBUTING.md)

