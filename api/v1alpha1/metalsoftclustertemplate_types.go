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
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MetalsoftClusterTemplateSpec defines the desired state of MetalsoftClusterTemplate
type MetalsoftClusterTemplateSpec struct {
	Template MetalsoftClusterTemplateResource `json:"template"`
}

type MetalsoftClusterTemplateResource struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta capiv1beta1.ObjectMeta `json:"metadata,omitempty"`

	Spec MetalsoftClusterSpec `json:"spec"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=metalsoftclustertemplates,scope=Namespaced,categories=cluster-api,shortName=metalsoftct
// +kubebuilder:storageversion

// MetalsoftClusterTemplate is the Schema for the metalsoftclustertemplates API
type MetalsoftClusterTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MetalsoftClusterTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// MetalsoftClusterTemplateList contains a list of MetalsoftClusterTemplate
type MetalsoftClusterTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetalsoftClusterTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetalsoftClusterTemplate{}, &MetalsoftClusterTemplateList{})
}
