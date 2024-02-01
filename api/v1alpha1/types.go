/*
Copyright 2023 The Kubernetes Authors.

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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// MetalsoftMachineTemplateResource describes the data needed to create am MetalsoftMachine from a template.
type MetalsoftMachineTemplateResource struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta clusterv1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification of the desired behavior of the machine.
	Spec MetalsoftMachineSpec `json:"spec"`
}

// InstanceStatus describes the state of an Metalsoft instance.
type MetalsoftInstanceResourceStatus string

var (
	// InstanceStatusOrdered is the string representing an instance in a ordered state.
	InstanceStatusOrdered = MetalsoftInstanceResourceStatus("ordered")

	// InstanceStatusActive is the string representing an instance in a active state.
	InstanceStatusActive = MetalsoftInstanceResourceStatus("active")

	// InstanceStatusSuspended is the string representing an instance
	// that is suspended.
	InstanceStatusSuspended = MetalsoftInstanceResourceStatus("suspended")

	// InstanceStatusStopped is the string representing an instance
	// that has been stopped and can be started.
	InstanceStatusStopped = MetalsoftInstanceResourceStatus("stopped")

	// InstanceStatusDeleted is the string representing an instance that has been deleted.
	InstanceStatusDeleted = MetalsoftInstanceResourceStatus("deleted")
)

// NetworkSpec encapsulates all things related to a Metalsoft network.
type NetworkSpec struct {
	// Label is the label of the network to be used.
	// +optional
	Label *string `json:"label,omitempty"`

	// Type is the type of the network to be used.
	// +optional
	// +kubebuilder:validation:Enum=wan;lan;san
	// +kubebuilder:default=wan
	Type string `json:"type,omitempty"`

	// NetworkProfileID represents the network profile ID that can be applied to the network.
	// +optional
	// +kubebuilder:printcolumn:description="The default WAN network profile configured on the datacenter is applied if none is specified."
	NetworkProfileID int `json:"networkProfileID,omitempty"`

	// Subnets configuration.
	// +optional
	Subnets Subnets `json:"subnets,omitempty"`
}

// SubnetSpec configures an Metalsoft Subnet.
type SubnetSpec struct {
	// Label defines a unique identifier to reference this resource.
	Label string `json:"subnet_label,omitempty"`

	// Type is the type of the network to be used.
	// +kubebuilder:validation:Enum=ipv4;ipv6
	// +kubebuilder:default=ipv4
	Type string `json:"subnet_type,omitempty"`

	// Subnet prefix size, such as /30, /27, etc. For IPv4 subnets can be one of: 27, 28, 29, 30. For IPv6 subnet can only be 64.
	// +kubebuilder:validation:Enum=27;28;29;30;64
	// +kubebuilder:default=29
	PrefixSize int `json:"subnet_prefix_size,omitempty"`

	// Specifies if subnet will be used for allocating IP addresses via DHCP
	// +kube:conversion-gen=false
	AutomaticAllocation bool `json:"subnet_automatic_allocation,omitempty"`
}

// Subnets is a slice of Subnet.
type Subnets []SubnetSpec

// ToMap returns a map from label to subnet.
func (s Subnets) ToMap() map[string]*SubnetSpec {
	res := make(map[string]*SubnetSpec)
	for i := range s {
		x := s[i]
		res[x.Label] = &x
	}

	return res
}

// FindByLabel returns a single subnet matching the given label or nil.
func (s Subnets) FindByName(label string) *SubnetSpec {
	for _, x := range s {
		if x.Label == label {
			return &x
		}
	}

	return nil
}
