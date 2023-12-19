/*
Copyright 2023.

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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// ClusterFinalizer allows ReconcileMetalsoftCluster to clean up Metalsoft resources associated with MetalsoftCluster before
	// removing it from the apiserver.
	ClusterFinalizer = "metalsoftcluster.infrastructure.cluster.x-k8s.io"
)

// MetalsoftClusterSpec defines the desired state of MetalsoftCluster
type MetalsoftClusterSpec struct {

	// DatacenterName represents the name of the datacenter where the cluster is deployed.
	DatacenterName string `json:"datacenterName"`

	// InfrastructureLabel represents the label used to identify the infrastructure.
	InfrastructureLabel string `json:"infrastructureLabel"`

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`
}

// MetalsoftClusterStatus defines the observed state of MetalsoftCluster
type MetalsoftClusterStatus struct {
	// Ready denotes that the cluster (infrastructure) is ready.
	// +optional
	Ready bool `json:"ready"`
}

// +kubebuilder:subresource:status
// +kubebuilder:resource:path=metalsoftclusters,scope=Namespaced,categories=cluster-api
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this MetalsoftCluster belongs"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="MetalsoftCluster ready status"

// MetalsoftCluster is the Schema for the metalsoftclusters API
type MetalsoftCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetalsoftClusterSpec   `json:"spec,omitempty"`
	Status MetalsoftClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MetalsoftClusterList contains a list of MetalsoftCluster
type MetalsoftClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetalsoftCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetalsoftCluster{}, &MetalsoftClusterList{})
}
