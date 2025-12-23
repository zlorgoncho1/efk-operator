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
	"fmt"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	loggingv1 "github.com/zlorgoncho1/efk-operator/api/v1"
	"github.com/zlorgoncho1/efk-operator/internal/helm"
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
//+kubebuilder:rbac:groups=apps,resources=statefulsets;deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps;namespaces;secrets;services;serviceaccounts;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete

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

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// reconcileElasticsearch handles Elasticsearch deployment
func (r *EFKStackReconciler) reconcileElasticsearch(ctx context.Context, efkStack *loggingv1.EFKStack, namespace string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Elasticsearch")

	releaseName := fmt.Sprintf("%s-elasticsearch", efkStack.Name)
	chartPath := filepath.Join("helm-charts", "efk-stack", "elasticsearch")

	// Prepare values for Helm chart
	values := map[string]interface{}{
		"version":  efkStack.Spec.Elasticsearch.Version,
		"replicas": efkStack.Spec.Elasticsearch.Replicas,
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
		efkStack.Status.Elasticsearch.State = "Error"
		r.Status().Update(ctx, efkStack)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Check release status
	status, err := r.HelmClient.GetReleaseStatus(releaseName)
	if err != nil {
		logger.Error(err, "Failed to get release status")
	}

	// Update status
	efkStack.Status.Elasticsearch.Version = efkStack.Spec.Elasticsearch.Version
	if status == "deployed" {
		efkStack.Status.Elasticsearch.State = "Ready"
		efkStack.Status.Elasticsearch.ReadyReplicas = efkStack.Spec.Elasticsearch.Replicas
	} else {
		efkStack.Status.Elasticsearch.State = "Deploying"
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
			"url": fmt.Sprintf("http://%s-elasticsearch:9200", efkStack.Name),
		},
	}

	// Deploy via Helm
	_, err := r.HelmClient.InstallOrUpgrade(ctx, releaseName, chartPath, values)
	if err != nil {
		efkStack.Status.FluentBit.State = "Error"
		r.Status().Update(ctx, efkStack)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Check release status
	status, err := r.HelmClient.GetReleaseStatus(releaseName)
	if err != nil {
		logger.Error(err, "Failed to get release status")
	}

	// Update status
	efkStack.Status.FluentBit.Version = efkStack.Spec.FluentBit.Version
	if status == "deployed" {
		efkStack.Status.FluentBit.State = "Ready"
	} else {
		efkStack.Status.FluentBit.State = "Deploying"
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
			"url": fmt.Sprintf("http://%s-elasticsearch:9200", efkStack.Name),
		},
	}

	if efkStack.Spec.Kibana.Ingress.Enabled {
		values["ingress"] = map[string]interface{}{
			"enabled":     true,
			"host":        efkStack.Spec.Kibana.Ingress.Host,
			"annotations": efkStack.Spec.Kibana.Ingress.Annotations,
			"tls":         efkStack.Spec.Kibana.Ingress.TLS,
		}
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
		efkStack.Status.Kibana.State = "Error"
		r.Status().Update(ctx, efkStack)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Check release status
	status, err := r.HelmClient.GetReleaseStatus(releaseName)
	if err != nil {
		logger.Error(err, "Failed to get release status")
	}

	// Update status
	efkStack.Status.Kibana.Version = efkStack.Spec.Kibana.Version
	if status == "deployed" {
		efkStack.Status.Kibana.State = "Ready"
		efkStack.Status.Kibana.ReadyReplicas = efkStack.Spec.Kibana.Replicas
		if efkStack.Spec.Kibana.Ingress.Enabled && efkStack.Spec.Kibana.Ingress.Host != "" {
			efkStack.Status.Kibana.URL = fmt.Sprintf("https://%s", efkStack.Spec.Kibana.Ingress.Host)
		}
	} else {
		efkStack.Status.Kibana.State = "Deploying"
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

// SetupWithManager sets up the controller with the Manager.
func (r *EFKStackReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&loggingv1.EFKStack{}).
		Complete(r)
}
