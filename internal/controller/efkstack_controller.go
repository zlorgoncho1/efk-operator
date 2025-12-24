/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	loggingv1 "github.com/zlorgoncho1/efk-operator/api/v1"
	"github.com/zlorgoncho1/efk-operator/internal/helm"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EFKStackReconciler reconciles a EFKStack object
type EFKStackReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	HelmClient *helm.Client
	RestConfig *rest.Config
	KubeClient kubernetes.Interface
}

//+kubebuilder:rbac:groups=logging.efk.crds.io,resources=efkstacks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=logging.efk.crds.io,resources=efkstacks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=logging.efk.crds.io,resources=efkstacks/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets;deployments;daemonsets;replicasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps;namespaces;pods;secrets;services;serviceaccounts;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// It compares the state specified by the EFKStack object against the actual cluster state,
// and performs operations to make the cluster state reflect the state specified by the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *EFKStackReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the EFKStack instance
	efkStack := &loggingv1.EFKStack{}
	if err := r.Get(ctx, req.NamespacedName, efkStack); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			logger.Info("EFKStack resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get EFKStack")
		return ctrl.Result{}, err
	}

	// Set default namespace if not specified
	namespace := efkStack.Spec.Namespace
	if namespace == "" {
		namespace = req.Namespace
		if namespace == "" {
			namespace = "default"
		}
	}

	// Initialize status if needed
	if efkStack.Status.Phase == "" {
		efkStack.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, efkStack); err != nil {
			logger.Error(err, "Failed to update EFKStack status")
			return ctrl.Result{}, err
		}
	}

	// Initialize Helm client if not already done
	if r.HelmClient == nil {
		helmClient, err := helm.NewClient(r.RestConfig, r.KubeClient, namespace)
		if err != nil {
			logger.Error(err, "Failed to create Helm client")
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err
		}
		r.HelmClient = helmClient
	}

	// Reconcile components in order: Elasticsearch -> Fluent Bit -> Kibana
	result, err := r.reconcileElasticsearch(ctx, efkStack, namespace)
	if err != nil {
		logger.Error(err, "Failed to reconcile Elasticsearch")
		return result, err
	}

	// Only proceed to Fluent Bit if Elasticsearch is ready
	if efkStack.Status.Elasticsearch.State == "Ready" {
		result, err = r.reconcileFluentBit(ctx, efkStack, namespace)
		if err != nil {
			logger.Error(err, "Failed to reconcile Fluent Bit")
			return result, err
		}
	}

	// Only proceed to Kibana if Elasticsearch is ready
	if efkStack.Status.Elasticsearch.State == "Ready" {
		result, err = r.reconcileKibana(ctx, efkStack, namespace)
		if err != nil {
			logger.Error(err, "Failed to reconcile Kibana")
			return result, err
		}
	}

	// Update overall phase
	if err := r.updatePhase(ctx, efkStack); err != nil {
		logger.Error(err, "Failed to update phase")
		return ctrl.Result{}, err
	}

	// Vérifier et mettre à jour les ConfigMaps/Secrets pour déclencher le redémarrage des pods
	if err := r.checkAndUpdateConfigMapsSecrets(ctx, efkStack, namespace); err != nil {
		logger.Error(err, "Failed to check ConfigMaps/Secrets")
		// Ne pas bloquer le reconcile pour cette erreur
	}

	// Optimiser l'intervalle de réconciliation selon l'état (réduit pour réactivité)
	if efkStack.Status.Phase == "Ready" {
		// Si tout est Ready, vérifier moins fréquemment mais toujours réactif
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}
	// Si en cours de déploiement, vérifier plus fréquemment
	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// reconcileElasticsearch handles Elasticsearch deployment
func (r *EFKStackReconciler) reconcileElasticsearch(ctx context.Context, efkStack *loggingv1.EFKStack, namespace string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Elasticsearch")

	releaseName := fmt.Sprintf("%s-elasticsearch", efkStack.Name)
	chartPath := filepath.Join("helm-charts", "efk-stack", "elasticsearch")

	// Determine mode (default to cluster if not specified)
	mode := efkStack.Spec.Elasticsearch.Mode
	if mode == "" {
		mode = "cluster"
	}

	// In singleton mode, force replicas to 1
	replicas := efkStack.Spec.Elasticsearch.Replicas
	if mode == "singleton" {
		replicas = 1
	}

	// Prepare values for Helm chart
	values := map[string]interface{}{
		"version":  efkStack.Spec.Elasticsearch.Version,
		"mode":     mode,
		"replicas": replicas,
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"cpu":    efkStack.Spec.Elasticsearch.Resources.Requests.Cpu().String(),
				"memory": efkStack.Spec.Elasticsearch.Resources.Requests.Memory().String(),
			},
			"limits": map[string]interface{}{
				"cpu":    efkStack.Spec.Elasticsearch.Resources.Limits.Cpu().String(),
				"memory": efkStack.Spec.Elasticsearch.Resources.Limits.Memory().String(),
			},
		},
		"storage": map[string]interface{}{
			"size":             efkStack.Spec.Elasticsearch.Storage.Size,
			"storageClassName": efkStack.Spec.Elasticsearch.Storage.StorageClassName,
			"path":             efkStack.Spec.Elasticsearch.Storage.Path,
		},
		"security": map[string]interface{}{
			"tlsEnabled":  efkStack.Spec.Elasticsearch.Security.TLSEnabled,
			"authEnabled": efkStack.Spec.Elasticsearch.Security.AuthEnabled,
		},
	}
	if len(efkStack.Spec.Elasticsearch.NodeSelector) > 0 {
		values["nodeSelector"] = efkStack.Spec.Elasticsearch.NodeSelector
	}
	if len(efkStack.Spec.Elasticsearch.Tolerations) > 0 {
		values["tolerations"] = efkStack.Spec.Elasticsearch.Tolerations
	}

	// Deploy via Helm
	_, err := r.HelmClient.InstallOrUpgrade(ctx, releaseName, chartPath, values)
	if err != nil {
		errorMsg := fmt.Sprintf("Helm install/upgrade failed: %v", err)
		logger.Error(err, "Failed to deploy Elasticsearch via Helm",
			"release", releaseName,
			"chartPath", chartPath,
			"namespace", namespace,
			"mode", mode,
			"replicas", replicas)
		
		// Essayer de récupérer le statut du release pour plus d'informations
		if releaseStatus, getErr := r.HelmClient.GetReleaseStatus(releaseName); getErr == nil {
			logger.Info("Current Helm release status", "status", releaseStatus)
			errorMsg = fmt.Sprintf("%s (Release status: %s)", errorMsg, releaseStatus)
		}
		
		efkStack.Status.Elasticsearch.State = "Error"
		efkStack.Status.Elasticsearch.Message = errorMsg
		r.Status().Update(ctx, efkStack)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	// Check release status
	status, err := r.HelmClient.GetReleaseStatus(releaseName)
	if err != nil {
		logger.Error(err, "Failed to get release status", "release", releaseName)
		efkStack.Status.Elasticsearch.Message = fmt.Sprintf("Failed to get release status: %v", err)
	} else {
		efkStack.Status.Elasticsearch.Message = "" // Clear error message on success
	}

	// Update status
	efkStack.Status.Elasticsearch.Version = efkStack.Spec.Elasticsearch.Version
	if status == "deployed" {
		efkStack.Status.Elasticsearch.State = "Ready"
		efkStack.Status.Elasticsearch.ReadyReplicas = efkStack.Spec.Elasticsearch.Replicas
		efkStack.Status.Elasticsearch.Message = ""
	} else {
		efkStack.Status.Elasticsearch.State = "Deploying"
		if status != "" {
			efkStack.Status.Elasticsearch.Message = fmt.Sprintf("Release status: %s", status)
		}
	}

	return ctrl.Result{}, r.Status().Update(ctx, efkStack)
}

// reconcileFluentBit handles Fluent Bit deployment
func (r *EFKStackReconciler) reconcileFluentBit(ctx context.Context, efkStack *loggingv1.EFKStack, namespace string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Fluent Bit")

	releaseName := fmt.Sprintf("%s-fluentbit", efkStack.Name)
	chartPath := filepath.Join("helm-charts", "efk-stack", "fluentbit")

	// Prepare values for Helm chart
	values := map[string]interface{}{
		"version": efkStack.Spec.FluentBit.Version,
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"cpu":    efkStack.Spec.FluentBit.Resources.Requests.Cpu().String(),
				"memory": efkStack.Spec.FluentBit.Resources.Requests.Memory().String(),
			},
			"limits": map[string]interface{}{
				"cpu":    efkStack.Spec.FluentBit.Resources.Limits.Cpu().String(),
				"memory": efkStack.Spec.FluentBit.Resources.Limits.Memory().String(),
			},
		},
		"elasticsearch": map[string]interface{}{
			"host":  fmt.Sprintf("%s-elasticsearch", efkStack.Name),
			"port":  9200,
			"index": "fluent-bit",
		},
	}

	// Ajouter nodeSelector si spécifié
	if len(efkStack.Spec.FluentBit.NodeSelector) > 0 {
		values["nodeSelector"] = efkStack.Spec.FluentBit.NodeSelector
	}

	// Ajouter tolerations si spécifiées, sinon utiliser des tolerations par défaut pour DaemonSet
	if len(efkStack.Spec.FluentBit.Tolerations) > 0 {
		values["tolerations"] = efkStack.Spec.FluentBit.Tolerations
	} else {
		// Tolerations par défaut pour un DaemonSet qui doit s'exécuter sur tous les nœuds
		values["tolerations"] = []corev1.Toleration{
			{
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			},
			{
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoExecute,
			},
			{
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectPreferNoSchedule,
			},
		}
	}

	// Deploy via Helm
	_, err := r.HelmClient.InstallOrUpgrade(ctx, releaseName, chartPath, values)
	if err != nil {
		errorMsg := fmt.Sprintf("Helm install/upgrade failed: %v", err)
		logger.Error(err, "Failed to deploy Fluent Bit via Helm",
			"release", releaseName,
			"chartPath", chartPath,
			"namespace", namespace)
		
		// Essayer de récupérer le statut du release pour plus d'informations
		if releaseStatus, getErr := r.HelmClient.GetReleaseStatus(releaseName); getErr == nil {
			logger.Info("Current Helm release status", "status", releaseStatus)
			errorMsg = fmt.Sprintf("%s (Release status: %s)", errorMsg, releaseStatus)
		}
		
		efkStack.Status.FluentBit.State = "Error"
		efkStack.Status.FluentBit.Message = errorMsg
		r.Status().Update(ctx, efkStack)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	// Check release status
	status, err := r.HelmClient.GetReleaseStatus(releaseName)
	if err != nil {
		logger.Error(err, "Failed to get release status", "release", releaseName)
		efkStack.Status.FluentBit.Message = fmt.Sprintf("Failed to get release status: %v", err)
	} else {
		efkStack.Status.FluentBit.Message = "" // Clear error message on success
	}

	// Update status
	efkStack.Status.FluentBit.Version = efkStack.Spec.FluentBit.Version
	if status == "deployed" {
		efkStack.Status.FluentBit.State = "Ready"
		efkStack.Status.FluentBit.Message = ""
	} else {
		efkStack.Status.FluentBit.State = "Deploying"
		if status != "" {
			efkStack.Status.FluentBit.Message = fmt.Sprintf("Release status: %s", status)
		}
	}

	return ctrl.Result{}, r.Status().Update(ctx, efkStack)
}

// reconcileKibana handles Kibana deployment
func (r *EFKStackReconciler) reconcileKibana(ctx context.Context, efkStack *loggingv1.EFKStack, namespace string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Kibana")

	releaseName := fmt.Sprintf("%s-kibana", efkStack.Name)
	chartPath := filepath.Join("helm-charts", "efk-stack", "kibana")

	// Prepare values for Helm chart
	values := map[string]interface{}{
		"version":  efkStack.Spec.Kibana.Version,
		"replicas": efkStack.Spec.Kibana.Replicas,
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"cpu":    efkStack.Spec.Kibana.Resources.Requests.Cpu().String(),
				"memory": efkStack.Spec.Kibana.Resources.Requests.Memory().String(),
			},
			"limits": map[string]interface{}{
				"cpu":    efkStack.Spec.Kibana.Resources.Limits.Cpu().String(),
				"memory": efkStack.Spec.Kibana.Resources.Limits.Memory().String(),
			},
		},
		"elasticsearch": map[string]interface{}{
			"hosts": []string{
				fmt.Sprintf("http://%s-elasticsearch:9200", efkStack.Name),
			},
		},
	}

	if efkStack.Spec.Kibana.Ingress.Enabled {
		// Convertir le host en format hosts attendu par le template
		ingressHosts := []map[string]interface{}{
			{
				"host": efkStack.Spec.Kibana.Ingress.Host,
				"paths": []map[string]interface{}{
					{
						"path":     "/",
						"pathType": "Prefix",
					},
				},
			},
		}

		// Convertir TLS au format attendu par le template
		ingressTLS := []map[string]interface{}{}
		for _, tls := range efkStack.Spec.Kibana.Ingress.TLS {
			ingressTLS = append(ingressTLS, map[string]interface{}{
				"hosts":      tls.Hosts,
				"secretName": tls.SecretName,
			})
		}

		ingressConfig := map[string]interface{}{
			"enabled": true,
			"hosts":   ingressHosts,
			"tls":     ingressTLS,
		}
		// Ne pas inclure les annotations si elles sont vides pour éviter le warning Helm
		if len(efkStack.Spec.Kibana.Ingress.Annotations) > 0 {
			ingressConfig["annotations"] = efkStack.Spec.Kibana.Ingress.Annotations
		}
		// Ajouter ingressClassName si présent dans les annotations
		if className, ok := efkStack.Spec.Kibana.Ingress.Annotations["kubernetes.io/ingress.class"]; ok {
			ingressConfig["className"] = className
		}
		values["ingress"] = ingressConfig
	}
	if len(efkStack.Spec.Kibana.NodeSelector) > 0 {
		values["nodeSelector"] = efkStack.Spec.Kibana.NodeSelector
	}
	if len(efkStack.Spec.Kibana.Tolerations) > 0 {
		values["tolerations"] = efkStack.Spec.Kibana.Tolerations
	}

	// Deploy via Helm
	_, err := r.HelmClient.InstallOrUpgrade(ctx, releaseName, chartPath, values)
	if err != nil {
		errorMsg := fmt.Sprintf("Helm install/upgrade failed: %v", err)
		logger.Error(err, "Failed to deploy Kibana via Helm",
			"release", releaseName,
			"chartPath", chartPath,
			"namespace", namespace,
			"replicas", efkStack.Spec.Kibana.Replicas,
			"ingressEnabled", efkStack.Spec.Kibana.Ingress.Enabled)
		
		// Essayer de récupérer le statut du release pour plus d'informations
		if releaseStatus, getErr := r.HelmClient.GetReleaseStatus(releaseName); getErr == nil {
			logger.Info("Current Helm release status", "status", releaseStatus)
			errorMsg = fmt.Sprintf("%s (Release status: %s)", errorMsg, releaseStatus)
		}
		
		// Log les valeurs importantes pour le debug (sans les secrets)
		logger.V(1).Info("Helm values used",
			"version", efkStack.Spec.Kibana.Version,
			"replicas", efkStack.Spec.Kibana.Replicas,
			"ingressEnabled", efkStack.Spec.Kibana.Ingress.Enabled)
		
		efkStack.Status.Kibana.State = "Error"
		efkStack.Status.Kibana.Message = errorMsg
		r.Status().Update(ctx, efkStack)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	// Check release status
	status, err := r.HelmClient.GetReleaseStatus(releaseName)
	if err != nil {
		logger.Error(err, "Failed to get release status", "release", releaseName)
		efkStack.Status.Kibana.Message = fmt.Sprintf("Failed to get release status: %v", err)
	} else {
		efkStack.Status.Kibana.Message = "" // Clear error message on success
	}

	// Update status
	efkStack.Status.Kibana.Version = efkStack.Spec.Kibana.Version
	if status == "deployed" {
		efkStack.Status.Kibana.State = "Ready"
		efkStack.Status.Kibana.ReadyReplicas = efkStack.Spec.Kibana.Replicas
		efkStack.Status.Kibana.Message = ""
		if efkStack.Spec.Kibana.Ingress.Enabled && efkStack.Spec.Kibana.Ingress.Host != "" {
			efkStack.Status.Kibana.URL = fmt.Sprintf("https://%s", efkStack.Spec.Kibana.Ingress.Host)
		}
	} else {
		efkStack.Status.Kibana.State = "Deploying"
		if status != "" {
			efkStack.Status.Kibana.Message = fmt.Sprintf("Release status: %s", status)
		}
	}

	return ctrl.Result{}, r.Status().Update(ctx, efkStack)
}

// updatePhase updates the overall phase of the EFKStack
func (r *EFKStackReconciler) updatePhase(ctx context.Context, efkStack *loggingv1.EFKStack) error {
	// Determine phase based on component states
	if efkStack.Status.Elasticsearch.State == "Ready" &&
		efkStack.Status.FluentBit.State == "Ready" &&
		efkStack.Status.Kibana.State == "Ready" {
		efkStack.Status.Phase = "Ready"
	} else if efkStack.Status.Elasticsearch.State == "Deploying" ||
		efkStack.Status.FluentBit.State == "Deploying" ||
		efkStack.Status.Kibana.State == "Deploying" {
		efkStack.Status.Phase = "Deploying"
	} else {
		efkStack.Status.Phase = "Pending"
	}

	// Update conditions
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: efkStack.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "Reconciling",
		Message: fmt.Sprintf("Elasticsearch: %s, FluentBit: %s, Kibana: %s",
			efkStack.Status.Elasticsearch.State,
			efkStack.Status.FluentBit.State,
			efkStack.Status.Kibana.State),
	}

	if efkStack.Status.Phase == "Ready" {
		condition.Status = metav1.ConditionTrue
		condition.Reason = "AllComponentsReady"
		condition.Message = "All components are ready"
	}

	// Update or add condition
	found := false
	for i, c := range efkStack.Status.Conditions {
		if c.Type == condition.Type {
			efkStack.Status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		efkStack.Status.Conditions = append(efkStack.Status.Conditions, condition)
	}

	return r.Status().Update(ctx, efkStack)
}

// checkAndUpdateConfigMapsSecrets vérifie les ConfigMaps et Secrets et met à jour les annotations
// des Deployments/DaemonSets/StatefulSets pour forcer le redémarrage des pods quand ils changent
func (r *EFKStackReconciler) checkAndUpdateConfigMapsSecrets(ctx context.Context, efkStack *loggingv1.EFKStack, namespace string) error {
	logger := log.FromContext(ctx)

	// Liste des releases Helm pour cet EFKStack avec leurs composants associés
	releases := []struct {
		name      string
		component string
	}{
		{fmt.Sprintf("%s-elasticsearch", efkStack.Name), "elasticsearch"},
		{fmt.Sprintf("%s-fluentbit", efkStack.Name), "fluentbit"},
		{fmt.Sprintf("%s-kibana", efkStack.Name), "kibana"},
	}

	for _, release := range releases {
		releaseName := release.name
		component := release.component

		// Récupérer tous les ConfigMaps liés à ce release
		configMaps := &corev1.ConfigMapList{}
		configMapErr := r.List(ctx, configMaps, client.InNamespace(namespace), client.MatchingLabels{
			"app.kubernetes.io/instance": releaseName,
		})
		if configMapErr != nil {
			logger.V(1).Info("Failed to list ConfigMaps for release, may not exist yet", "release", releaseName, "error", configMapErr)
			// Continuer même si les ConfigMaps n'existent pas encore
		}

		// Récupérer tous les Secrets liés à ce release
		secrets := &corev1.SecretList{}
		secretErr := r.List(ctx, secrets, client.InNamespace(namespace), client.MatchingLabels{
			"app.kubernetes.io/instance": releaseName,
		})
		if secretErr != nil {
			logger.V(1).Info("Failed to list Secrets for release, may not exist yet", "release", releaseName, "error", secretErr)
			// Continuer même si les Secrets n'existent pas encore
		}

		// Si ni ConfigMaps ni Secrets n'existent, passer au suivant
		if configMapErr != nil && secretErr != nil {
			logger.V(1).Info("No ConfigMaps or Secrets found for release, skipping", "release", releaseName, "component", component)
			continue
		}

		// Calculer le hash combiné de tous les ConfigMaps et Secrets
		// Même si la liste est vide, on calcule un hash pour permettre la mise à jour initiale
		hash := r.computeConfigHash(configMaps.Items, secrets.Items)

		// Mettre à jour les Deployments
		deployments := &appsv1.DeploymentList{}
		if err := r.List(ctx, deployments, client.InNamespace(namespace), client.MatchingLabels{
			"app.kubernetes.io/instance": releaseName,
		}); err == nil {
			for i := range deployments.Items {
				deployment := &deployments.Items[i]
				if r.updatePodTemplateAnnotations(&deployment.Spec.Template, hash) {
					if err := r.Update(ctx, deployment); err != nil {
						logger.Error(err, "Failed to update Deployment", "deployment", deployment.Name, "component", component)
					} else {
						logger.Info("Updated Deployment annotations to trigger pod restart", "deployment", deployment.Name, "component", component, "hash", hash)
					}
				}
			}
		}

		// Mettre à jour les DaemonSets
		daemonSets := &appsv1.DaemonSetList{}
		if err := r.List(ctx, daemonSets, client.InNamespace(namespace), client.MatchingLabels{
			"app.kubernetes.io/instance": releaseName,
		}); err == nil {
			for i := range daemonSets.Items {
				daemonSet := &daemonSets.Items[i]
				if r.updatePodTemplateAnnotations(&daemonSet.Spec.Template, hash) {
					if err := r.Update(ctx, daemonSet); err != nil {
						logger.Error(err, "Failed to update DaemonSet", "daemonset", daemonSet.Name, "component", component)
					} else {
						logger.Info("Updated DaemonSet annotations to trigger pod restart", "daemonset", daemonSet.Name, "component", component, "hash", hash)
					}
				}
			}
		}

		// Mettre à jour les StatefulSets
		statefulSets := &appsv1.StatefulSetList{}
		if err := r.List(ctx, statefulSets, client.InNamespace(namespace), client.MatchingLabels{
			"app.kubernetes.io/instance": releaseName,
		}); err == nil {
			for i := range statefulSets.Items {
				statefulSet := &statefulSets.Items[i]
				if r.updatePodTemplateAnnotations(&statefulSet.Spec.Template, hash) {
					if err := r.Update(ctx, statefulSet); err != nil {
						logger.Error(err, "Failed to update StatefulSet", "statefulset", statefulSet.Name, "component", component)
					} else {
						logger.Info("Updated StatefulSet annotations to trigger pod restart", "statefulset", statefulSet.Name, "component", component, "hash", hash)
					}
				}
			}
		}
	}

	return nil
}

// computeConfigHash calcule un hash SHA256 des ConfigMaps et Secrets
func (r *EFKStackReconciler) computeConfigHash(configMaps []corev1.ConfigMap, secrets []corev1.Secret) string {
	hasher := sha256.New()

	// Ajouter les ConfigMaps au hash
	for _, cm := range configMaps {
		hasher.Write([]byte(cm.Name))
		hasher.Write([]byte(cm.Namespace))
		for k, v := range cm.Data {
			hasher.Write([]byte(k))
			hasher.Write([]byte(v))
		}
		for k, v := range cm.BinaryData {
			hasher.Write([]byte(k))
			hasher.Write(v)
		}
	}

	// Ajouter les Secrets au hash
	for _, secret := range secrets {
		hasher.Write([]byte(secret.Name))
		hasher.Write([]byte(secret.Namespace))
		for k, v := range secret.Data {
			hasher.Write([]byte(k))
			hasher.Write(v)
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))[:16] // Utiliser les 16 premiers caractères
}

// updatePodTemplateAnnotations met à jour les annotations du pod template avec le hash
// Retourne true si une mise à jour est nécessaire
func (r *EFKStackReconciler) updatePodTemplateAnnotations(template *corev1.PodTemplateSpec, hash string) bool {
	if template.Annotations == nil {
		template.Annotations = make(map[string]string)
	}

	annotationKey := "efk.crds.io/config-hash"
	currentHash := template.Annotations[annotationKey]

	if currentHash != hash {
		template.Annotations[annotationKey] = hash
		// Ajouter aussi un timestamp pour forcer le redémarrage
		template.Annotations["efk.crds.io/config-updated"] = time.Now().Format(time.RFC3339)
		return true
	}

	return false
}

// mapConfigMapToEFKStack mappe un ConfigMap à un EFKStack
func (r *EFKStackReconciler) mapConfigMapToEFKStack(ctx context.Context, obj client.Object) []reconcile.Request {
	configMap := obj.(*corev1.ConfigMap)
	logger := log.FromContext(ctx)

	// Chercher l'EFKStack associé via les labels
	efkStackName := ""
	if instance, ok := configMap.Labels["app.kubernetes.io/instance"]; ok && len(instance) > 0 {
		// Le format est généralement: efkstack-name-elasticsearch, efkstack-name-fluentbit, etc.
		// Chercher tous les EFKStacks dans le namespace
		efkStacks := &loggingv1.EFKStackList{}
		if err := r.List(ctx, efkStacks, client.InNamespace(configMap.Namespace)); err == nil {
			for i := range efkStacks.Items {
				efkStack := &efkStacks.Items[i]
				// Vérifier si le ConfigMap appartient à cet EFKStack
				expectedInstances := []string{
					fmt.Sprintf("%s-elasticsearch", efkStack.Name),
					fmt.Sprintf("%s-fluentbit", efkStack.Name),
					fmt.Sprintf("%s-kibana", efkStack.Name),
				}
				for _, expectedInstance := range expectedInstances {
					if instance == expectedInstance {
						efkStackName = efkStack.Name
						break
					}
				}
				if efkStackName != "" {
					break
				}
			}
		} else {
			logger.V(1).Info("Failed to list EFKStacks when mapping ConfigMap", "error", err, "configmap", configMap.Name)
		}
	} else {
		// Si le label n'est pas présent, essayer de trouver via le nom du ConfigMap
		// Les ConfigMaps suivent généralement le pattern: {release-name}-config
		efkStacks := &loggingv1.EFKStackList{}
		if err := r.List(ctx, efkStacks, client.InNamespace(configMap.Namespace)); err == nil {
			for i := range efkStacks.Items {
				efkStack := &efkStacks.Items[i]
				// Vérifier si le nom du ConfigMap correspond à un pattern attendu
				expectedNames := []string{
					fmt.Sprintf("%s-elasticsearch-config", efkStack.Name),
					fmt.Sprintf("%s-fluentbit-config", efkStack.Name),
					fmt.Sprintf("%s-kibana-config", efkStack.Name),
					fmt.Sprintf("%s-elasticsearch-%s-config", efkStack.Name, efkStack.Name),
					fmt.Sprintf("%s-fluentbit-%s-config", efkStack.Name, efkStack.Name),
					fmt.Sprintf("%s-kibana-%s-config", efkStack.Name, efkStack.Name),
				}
				for _, expectedName := range expectedNames {
					if configMap.Name == expectedName {
						efkStackName = efkStack.Name
						break
					}
				}
				if efkStackName != "" {
					break
				}
			}
		}
	}

	if efkStackName == "" {
		return []reconcile.Request{}
	}

	logger.Info("ConfigMap changed, triggering reconcile", "configmap", configMap.Name, "namespace", configMap.Namespace, "efkstack", efkStackName)
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      efkStackName,
				Namespace: configMap.Namespace,
			},
		},
	}
}

// mapSecretToEFKStack mappe un Secret à un EFKStack
func (r *EFKStackReconciler) mapSecretToEFKStack(ctx context.Context, obj client.Object) []reconcile.Request {
	secret := obj.(*corev1.Secret)
	logger := log.FromContext(ctx)

	// Chercher l'EFKStack associé via les labels
	efkStackName := ""
	if instance, ok := secret.Labels["app.kubernetes.io/instance"]; ok && len(instance) > 0 {
		// Chercher tous les EFKStacks dans le namespace
		efkStacks := &loggingv1.EFKStackList{}
		if err := r.List(ctx, efkStacks, client.InNamespace(secret.Namespace)); err == nil {
			for i := range efkStacks.Items {
				efkStack := &efkStacks.Items[i]
				// Vérifier si le Secret appartient à cet EFKStack
				expectedInstances := []string{
					fmt.Sprintf("%s-elasticsearch", efkStack.Name),
					fmt.Sprintf("%s-fluentbit", efkStack.Name),
					fmt.Sprintf("%s-kibana", efkStack.Name),
				}
				for _, expectedInstance := range expectedInstances {
					if instance == expectedInstance {
						efkStackName = efkStack.Name
						break
					}
				}
				if efkStackName != "" {
					break
				}
			}
		} else {
			logger.V(1).Info("Failed to list EFKStacks when mapping Secret", "error", err, "secret", secret.Name)
		}
	} else {
		// Si le label n'est pas présent, essayer de trouver via le nom du Secret
		// Les Secrets suivent généralement des patterns spécifiques
		efkStacks := &loggingv1.EFKStackList{}
		if err := r.List(ctx, efkStacks, client.InNamespace(secret.Namespace)); err == nil {
			for i := range efkStacks.Items {
				efkStack := &efkStacks.Items[i]
				// Vérifier si le nom du Secret correspond à un pattern attendu
				// Les secrets peuvent avoir différents noms selon leur usage (TLS, auth, etc.)
				expectedPatterns := []string{
					fmt.Sprintf("%s-elasticsearch", efkStack.Name),
					fmt.Sprintf("%s-fluentbit", efkStack.Name),
					fmt.Sprintf("%s-kibana", efkStack.Name),
				}
				for _, pattern := range expectedPatterns {
					if len(secret.Name) >= len(pattern) && secret.Name[:len(pattern)] == pattern {
						efkStackName = efkStack.Name
						break
					}
				}
				if efkStackName != "" {
					break
				}
			}
		}
	}

	if efkStackName == "" {
		return []reconcile.Request{}
	}

	logger.Info("Secret changed, triggering reconcile", "secret", secret.Name, "namespace", secret.Namespace, "efkstack", efkStackName)
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      efkStackName,
				Namespace: secret.Namespace,
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *EFKStackReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&loggingv1.EFKStack{}).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.mapConfigMapToEFKStack),
		).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.mapSecretToEFKStack),
		).
		Complete(r)
}
