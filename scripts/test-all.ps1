# Script de test complet pour EFK Stack Operator (PowerShell)
# Usage: .\scripts\test-all.ps1

$ErrorActionPreference = "Stop"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Tests EFK Stack Operator" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# Vérifier que nous sommes dans le bon répertoire
if (-not (Test-Path "go.mod")) {
    Write-Host "✗ Vous devez être dans le répertoire efk/" -ForegroundColor Red
    exit 1
}

Write-Host "ℹ Répertoire de travail: $(Get-Location)" -ForegroundColor Yellow

# Test 1: Vérifier la structure
Write-Host ""
Write-Host "Test 1: Vérification de la structure..." -ForegroundColor Cyan
$requiredDirs = @("api", "config", "docker", "helm-charts", "internal")
$allExist = $true
foreach ($dir in $requiredDirs) {
    if (Test-Path $dir) {
        Write-Host "  ✓ $dir existe" -ForegroundColor Green
    } else {
        Write-Host "  ✗ $dir manquant" -ForegroundColor Red
        $allExist = $false
    }
}
if ($allExist) {
    Write-Host "✓ Structure du projet OK" -ForegroundColor Green
} else {
    Write-Host "✗ Structure du projet incomplète" -ForegroundColor Red
    exit 1
}

# Test 2: Vérifier Docker
Write-Host ""
Write-Host "Test 2: Vérification de Docker..." -ForegroundColor Cyan
try {
    $dockerVersion = docker --version 2>&1
    Write-Host "✓ Docker installé: $dockerVersion" -ForegroundColor Green
    
    docker ps | Out-Null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Docker fonctionne" -ForegroundColor Green
    } else {
        Write-Host "✗ Docker ne répond pas" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "✗ Docker non installé ou non accessible" -ForegroundColor Red
    exit 1
}

# Test 3: Construire l'image Docker
Write-Host ""
Write-Host "Test 3: Construction de l'image Docker..." -ForegroundColor Cyan
Write-Host "  (Cela peut prendre quelques minutes...)" -ForegroundColor Yellow
try {
    make docker-build
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Image Docker construite" -ForegroundColor Green
    } else {
        Write-Host "✗ Échec de la construction de l'image Docker" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "✗ Erreur lors de la construction" -ForegroundColor Red
    exit 1
}

# Test 4: Vérifier que l'image existe
Write-Host ""
Write-Host "Test 4: Vérification de l'image..." -ForegroundColor Cyan
$imageExists = docker images | Select-String "efk-operator-dev"
if ($imageExists) {
    Write-Host "✓ Image efk-operator-dev trouvée" -ForegroundColor Green
} else {
    Write-Host "✗ Image efk-operator-dev non trouvée" -ForegroundColor Red
    exit 1
}

# Test 5: Générer les manifests
Write-Host ""
Write-Host "Test 5: Génération des manifests CRD..." -ForegroundColor Cyan
try {
    make manifests
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Manifests générés" -ForegroundColor Green
        if (Test-Path "config/crd/bases/logging.efk.crds.io_efkstacks.yaml") {
            Write-Host "✓ Fichier CRD créé" -ForegroundColor Green
        } else {
            Write-Host "✗ Fichier CRD non trouvé" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "✗ Échec de la génération des manifests" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "✗ Erreur lors de la génération" -ForegroundColor Red
    exit 1
}

# Test 6: Générer le code
Write-Host ""
Write-Host "Test 6: Génération du code..." -ForegroundColor Cyan
try {
    make generate
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Code généré" -ForegroundColor Green
    } else {
        Write-Host "✗ Échec de la génération du code" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "✗ Erreur lors de la génération" -ForegroundColor Red
    exit 1
}

# Test 7: Formater le code
Write-Host ""
Write-Host "Test 7: Formatage du code..." -ForegroundColor Cyan
try {
    make fmt
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Code formaté" -ForegroundColor Green
    } else {
        Write-Host "✗ Échec du formatage" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "✗ Erreur lors du formatage" -ForegroundColor Red
    exit 1
}

# Test 8: Vérifier le code
Write-Host ""
Write-Host "Test 8: Vérification du code (go vet)..." -ForegroundColor Cyan
try {
    make vet
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Code vérifié (pas d'erreurs)" -ForegroundColor Green
    } else {
        Write-Host "✗ Erreurs trouvées par go vet" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "✗ Erreur lors de la vérification" -ForegroundColor Red
    exit 1
}

# Test 9: Compiler le code
Write-Host ""
Write-Host "Test 9: Compilation du code..." -ForegroundColor Cyan
try {
    make build
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Code compilé" -ForegroundColor Green
        if (Test-Path "bin/manager") {
            Write-Host "✓ Binaire créé: bin/manager" -ForegroundColor Green
        } else {
            Write-Host "✗ Binaire non trouvé" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "✗ Échec de la compilation" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "✗ Erreur lors de la compilation" -ForegroundColor Red
    exit 1
}

# Résumé
Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Tous les tests sont passés !" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Prochaines étapes :" -ForegroundColor Yellow
Write-Host "  1. make dev-up          # Démarrer l'environnement"
Write-Host "  2. make dev-shell        # Ouvrir un shell"
Write-Host "  3. make test            # Exécuter les tests"
Write-Host "  4. make install         # Installer les CRDs (si cluster disponible)"
Write-Host ""

