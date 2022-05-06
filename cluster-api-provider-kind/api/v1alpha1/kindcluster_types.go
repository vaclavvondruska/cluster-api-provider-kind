/*
Copyright 2022.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type KindClusterProtocol string
type KindClusterPropagation string
type KindClusterRole string
type KindClusterIPFamily string
type KindClusterKubeProxyMode string
type KindClusterState string

const (
	KindClusterProtocolTCP  = KindClusterProtocol("TCP")
	KindClusterProtocolUDP  = KindClusterProtocol("UDP")
	KindClusterProtocolSCTP = KindClusterProtocol("SCTP")

	KindClusterPropagationNone            = KindClusterPropagation("None")
	KindClusterPropagationHostToContainer = KindClusterPropagation("HostToContainer")
	KindClusterPropagationBidirectional   = KindClusterPropagation("Bidirectional")

	KindClusterRoleControlPlane = KindClusterRole("control-plane")
	KindClusterRoleWorker       = KindClusterRole("worker")

	KindCLusterIPFamilyV4 = KindClusterIPFamily("ipv4")
	KindCLusterIPFamilyV6 = KindClusterIPFamily("ipv6")

	KindClusterKubeProxyModeNone     = KindClusterKubeProxyMode("none")
	KindClusterKubeProxyModeIPTables = KindClusterKubeProxyMode("iptables")
	KindClusterKubeProxyModeIPVS     = KindClusterKubeProxyMode("ipvs")

	KindClusterStatePending  = KindClusterState("Pending")
	KindClusterStateRunning  = KindClusterState("Running")
	KindClusterStateDeleting = KindClusterState("Deleting")
	KindClusterStateFailed   = KindClusterState("Failed")

	KindClusterFinalizerName = "kindcluster.finalizers.infrastructure.cluster.x-k8s.io"
)

// KindClusterExtraPortMapping defines configuration of extra port mappings in KindClusterNode
type KindClusterExtraPortMapping struct {
	ContainerPort int                 `json:"containerPort,omitempty" yaml:"containerPort,omitempty"`
	HostPort      int                 `json:"hostPort,omitempty" yaml:"hostPort,omitempty"`
	ListenAddress string              `json:"listenAddress,omitempty" yaml:"listenAddress,omitempty"`
	Protocol      KindClusterProtocol `json:"protocol,omitempty" yaml:"protocol,omitempty"`
}

// KindClusterExtraMount defines configuration of extra mounts in KindClusterNode
type KindClusterExtraMount struct {
	HostPath       string                 `json:"hostPath,omitempty" yaml:"hostPath,omitempty"`
	ContainerPath  string                 `json:"containerPath,omitempty" yaml:"containerPath,omitempty"`
	ReadOnly       bool                   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	SELinuxRelabel bool                   `json:"selinuxRelabel,omitempty" yaml:"selinuxRelabel,omitempty"`
	Propagation    KindClusterPropagation `json:"propagation,omitempty" yaml:"propagation,omitempty"`
}

// KindClusterNode defines configuration a node in a Kind cluster
type KindClusterNode struct {
	Role                 KindClusterRole               `json:"role,omitempty" yaml:"role,omitempty"`
	Image                string                        `json:"image,omitempty" yaml:"image,omitempty"`
	KubeAdmConfigPatches []string                      `json:"kubeadmConfigPatches,omitempty" yaml:"kubeadmConfigPatches,omitempty"`
	ExtraMounts          []KindClusterExtraMount       `json:"extraMounts,omitempty" yaml:"extraMounts,omitempty"`
	ExtraPortMappings    []KindClusterExtraPortMapping `json:"extraPortMappings,omitempty" yaml:"extraPortMappings,omitempty"`
}

// KindClusterNetworking defines a networking configuration of a Kind cluster
type KindClusterNetworking struct {
	IPFamily          KindClusterIPFamily      `json:"ipFamily,omitempty" yaml:"ipFamily,omitempty"`
	APIServerAddress  string                   `json:"apiServerAddress,omitempty" yaml:"apiServerAddress,omitempty"`
	APIServerPort     int                      `json:"apiServerPort,omitempty" yaml:"apiServerPort,omitempty" `
	PodSubnet         string                   `json:"podSubnet,omitempty" yaml:"podSubnet,omitempty"`
	ServiceSubnet     string                   `json:"serviceSubnet,omitempty" yaml:"serviceSubnet,omitempty"`
	DisableDefaultCNI bool                     `json:"disableDefaultCNI,omitempty" yaml:"disableDefaultCNI,omitempty"`
	KubeProxyMode     KindClusterKubeProxyMode `json:"kubeProxyMode,omitempty" yaml:"kubeProxyMode,omitempty"`
}

// KindClusterControlPlaneEndpoint defines host and port of the cluster control plane
type KindClusterControlPlaneEndpoint struct {
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
	Port int    `json:"port,omitempty" yaml:"port,omitempty"`
}

// KindClusterSpec defines the desired state of KindCluster
type KindClusterSpec struct {
	FeatureGates         map[string]bool                 `json:"featureGates,omitempty" yaml:"featureGates,omitempty"`
	RuntimeConfig        map[string]string               `json:"runtimeConfig,omitempty" yaml:"runtimeConfig,omitempty"`
	Networking           KindClusterNetworking           `json:"networking,omitempty" yaml:"networking,omitempty"`
	Nodes                []KindClusterNode               `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	ControlPlaneEndpoint KindClusterControlPlaneEndpoint `json:"controlPlaneEndpoint,omitempty" yaml:"controlPlaneEndpoint,omitempty"`
}

// KindClusterStatus defines the observed state of KindCluster
type KindClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State KindClusterState `json:"state,omitempty"`
	Ready bool             `json:"ready,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KindCluster is the Schema for the kindclusters API
type KindCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KindClusterSpec   `json:"spec,omitempty"`
	Status KindClusterStatus `json:"status,omitempty"`
}

// IsBeingDeleted checks if the deletion timestamp is present
func (k *KindCluster) IsBeingDeleted() bool {
	return !k.ObjectMeta.DeletionTimestamp.IsZero()
}

// HasFinalizer checks if a specific finalizer is present
func (k *KindCluster) HasFinalizer(finalizerName string) bool {
	return containsString(k.ObjectMeta.Finalizers, finalizerName)
}

// AddFinalizer adds a new finalizer with the specified name to the KindCluster object
func (k *KindCluster) AddFinalizer(finalizerName string) {
	k.ObjectMeta.Finalizers = append(k.ObjectMeta.Finalizers, finalizerName)
}

// RemoveFinalizer removes a finalizer with the specified name from the KindCluster object
func (k *KindCluster) RemoveFinalizer(finalizerName string) {
	k.ObjectMeta.Finalizers = removeString(k.ObjectMeta.Finalizers, finalizerName)
}

// HasControlPlaneEndpoint checks if the KindCluster object has a non-empty KindClusterControlPlaneEndpoint
func (k *KindCluster) HasControlPlaneEndpoint() bool {
	return k.Spec.ControlPlaneEndpoint.Host != "" || k.Spec.ControlPlaneEndpoint.Port > 0
}

// AddControlPlaneEndpoint adds a KindClusterControlPlaneEndpoint with the specified host and port
func (k *KindCluster) AddControlPlaneEndpoint(host string, port int) {
	k.Spec.ControlPlaneEndpoint = KindClusterControlPlaneEndpoint{Host: host, Port: port}
}

//+kubebuilder:object:root=true

// KindClusterList contains a list of KindCluster
type KindClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KindCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KindCluster{}, &KindClusterList{})
}
