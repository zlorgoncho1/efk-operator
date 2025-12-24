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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EFKStackSpec defines the desired state of EFKStack
type EFKStackSpec struct {
	// Version globale de la stack
	// +optional
	Version string `json:"version,omitempty"`

	// Namespace de déploiement
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Configuration Elasticsearch
	// +kubebuilder:validation:Required
	Elasticsearch ElasticsearchSpec `json:"elasticsearch"`

	// Configuration Fluent Bit
	// +kubebuilder:validation:Required
	FluentBit FluentBitSpec `json:"fluentBit"`

	// Configuration Kibana
	// +kubebuilder:validation:Required
	Kibana KibanaSpec `json:"kibana"`

	// Configuration globale
	// +optional
	Global GlobalSpec `json:"global,omitempty"`
}

// ElasticsearchSpec defines the Elasticsearch configuration
type ElasticsearchSpec struct {
	// Version d'Elasticsearch
	// +kubebuilder:validation:Required
	Version string `json:"version"`

	// Mode de déploiement : "singleton" (single node) ou "cluster" (multi-node)
	// +kubebuilder:validation:Enum=singleton;cluster
	// +kubebuilder:default=cluster
	// +optional
	Mode string `json:"mode,omitempty"`

	// Nombre de replicas (ignoré en mode singleton, forcé à 1)
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=3
	Replicas int32 `json:"replicas"`

	// Ressources (CPU, mémoire)
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Configuration du stockage
	// +optional
	Storage StorageSpec `json:"storage,omitempty"`

	// Configuration de sécurité
	// +optional
	Security SecuritySpec `json:"security,omitempty"`

	// Configuration additionnelle (clés-valeurs)
	// +optional
	Config map[string]string `json:"config,omitempty"`

	// NodeSelector pour planifier les pods sur des nœuds spécifiques
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations pour permettre le scheduling sur des nœuds avec des taints
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// FluentBitSpec defines the Fluent Bit configuration
type FluentBitSpec struct {
	// Version de Fluent Bit
	// +kubebuilder:validation:Required
	Version string `json:"version"`

	// Ressources (CPU, mémoire)
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Configuration Fluent Bit
	// +optional
	Config FluentBitConfig `json:"config,omitempty"`

	// NodeSelector pour planifier les pods sur des nœuds spécifiques
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations pour permettre le scheduling sur des nœuds avec des taints
	// Pour un DaemonSet, il est recommandé d'ajouter des tolerations pour tous les taints communs
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// FluentBitConfig defines Fluent Bit configuration options
type FluentBitConfig struct {
	// Input configuration (JSON string)
	// +optional
	Input string `json:"input,omitempty"`

	// Filter configuration (JSON string)
	// +optional
	Filter string `json:"filter,omitempty"`

	// Output configuration (JSON string)
	// +optional
	Output string `json:"output,omitempty"`

	// Service configuration (JSON string)
	// +optional
	Service string `json:"service,omitempty"`
}

// KibanaSpec defines the Kibana configuration
type KibanaSpec struct {
	// Version de Kibana
	// +kubebuilder:validation:Required
	Version string `json:"version"`

	// Nombre de replicas
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=2
	Replicas int32 `json:"replicas"`

	// Ressources (CPU, mémoire)
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Configuration Ingress
	// +optional
	Ingress IngressSpec `json:"ingress,omitempty"`

	// NodeSelector pour planifier les pods sur des nœuds spécifiques
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations pour permettre le scheduling sur des nœuds avec des taints
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// StorageSpec defines storage configuration
type StorageSpec struct {
	// Storage class name
	// +optional
	StorageClassName string `json:"storageClassName,omitempty"`

	// Taille du stockage
	// +kubebuilder:validation:Pattern=^[0-9]+(Gi|Mi)$
	// +optional
	Size string `json:"size,omitempty"`

	// Type de volume (persistentVolumeClaim, emptyDir, etc.)
	// +optional
	VolumeType string `json:"volumeType,omitempty"`

	// Chemin personnalisé pour EFS (e.g., /eyone-prod/elasticsearch)
	// +optional
	Path string `json:"path,omitempty"`
}

// SecuritySpec defines security configuration
type SecuritySpec struct {
	// Activer TLS
	// +optional
	// +kubebuilder:default=true
	TLSEnabled bool `json:"tlsEnabled,omitempty"`

	// Activer l'authentification
	// +optional
	// +kubebuilder:default=true
	AuthEnabled bool `json:"authEnabled,omitempty"`

	// Secret contenant les certificats TLS
	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty"`

	// Secret contenant les credentials
	// +optional
	AuthSecretName string `json:"authSecretName,omitempty"`
}

// IngressSpec defines Ingress configuration
type IngressSpec struct {
	// Activer l'Ingress
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// Hostname
	// +optional
	Host string `json:"host,omitempty"`

	// Annotations pour l'Ingress
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// TLS configuration
	// +optional
	TLS []IngressTLS `json:"tls,omitempty"`
}

// IngressTLS defines TLS configuration for Ingress
type IngressTLS struct {
	// Hosts
	// +optional
	Hosts []string `json:"hosts,omitempty"`

	// Secret name
	// +optional
	SecretName string `json:"secretName,omitempty"`
}

// GlobalSpec defines global configuration
type GlobalSpec struct {
	// Storage class par défaut
	// +optional
	StorageClass string `json:"storageClass,omitempty"`

	// Registry d'images
	// +optional
	ImageRegistry string `json:"imageRegistry,omitempty"`

	// Configuration TLS globale
	// +optional
	TLS TLSSpec `json:"tls,omitempty"`
}

// TLSSpec defines TLS configuration
type TLSSpec struct {
	// Activer TLS
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// Secret contenant les certificats
	// +optional
	SecretName string `json:"secretName,omitempty"`
}

// EFKStackStatus defines the observed state of EFKStack
type EFKStackStatus struct {
	// Conditions représentent l'état actuel de la stack
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Phase de la stack
	// +optional
	Phase string `json:"phase,omitempty"`

	// État d'Elasticsearch
	// +optional
	Elasticsearch ElasticsearchStatus `json:"elasticsearch,omitempty"`

	// État de Fluent Bit
	// +optional
	FluentBit FluentBitStatus `json:"fluentBit,omitempty"`

	// État de Kibana
	// +optional
	Kibana KibanaStatus `json:"kibana,omitempty"`
}

// ElasticsearchStatus defines Elasticsearch status
type ElasticsearchStatus struct {
	// État (Ready, NotReady, etc.)
	// +optional
	State string `json:"state,omitempty"`

	// Nombre de pods prêts
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Version déployée
	// +optional
	Version string `json:"version,omitempty"`

	// URL du cluster
	// +optional
	URL string `json:"url,omitempty"`

	// Message d'erreur ou d'information
	// +optional
	Message string `json:"message,omitempty"`
}

// FluentBitStatus defines Fluent Bit status
type FluentBitStatus struct {
	// État (Ready, NotReady, etc.)
	// +optional
	State string `json:"state,omitempty"`

	// Nombre de pods prêts
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Version déployée
	// +optional
	Version string `json:"version,omitempty"`

	// Message d'erreur ou d'information
	// +optional
	Message string `json:"message,omitempty"`
}

// KibanaStatus defines Kibana status
type KibanaStatus struct {
	// État (Ready, NotReady, etc.)
	// +optional
	State string `json:"state,omitempty"`

	// Nombre de pods prêts
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Version déployée
	// +optional
	Version string `json:"version,omitempty"`

	// URL d'accès
	// +optional
	URL string `json:"url,omitempty"`

	// Message d'erreur ou d'information
	// +optional
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Elasticsearch",type="string",JSONPath=".status.elasticsearch.state"
//+kubebuilder:printcolumn:name="FluentBit",type="string",JSONPath=".status.fluentBit.state"
//+kubebuilder:printcolumn:name="Kibana",type="string",JSONPath=".status.kibana.state"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// EFKStack is the Schema for the efkstacks API
type EFKStack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EFKStackSpec   `json:"spec,omitempty"`
	Status EFKStackStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EFKStackList contains a list of EFKStack
type EFKStackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EFKStack `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EFKStack{}, &EFKStackList{})
}
