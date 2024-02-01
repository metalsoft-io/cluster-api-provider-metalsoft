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
	"sigs.k8s.io/cluster-api/errors"
)

const (
	// MachineFinalizer allows ReconcileMetalsoftMachine to clean up Metalsoft resources before
	// removing it from the apiserver.
	MachineFinalizer = "metalsoftmachine.infrastructure.cluster.x-k8s.io"
)

type InstanceBootMethod string

const (
	// InstanceBootMethodPxeIscsi is the string representing an instance that boots from an ISCSI disk through PXE.
	InstanceBootMethodPxeIscsi = InstanceBootMethod("pxe_iscsi")
	// InstanceBootMethodLocalDrive is the string representing an instance that boots from a local disk.
	InstanceBootMethodLocalDrive = InstanceBootMethod("local_drives")
)

type BlockStorage string

const (
	// BlockStorageDriveArray can be attached to a single instance.
	BlockStorageDriveArray = BlockStorage("drive_array")
	// BlockStorageSharedDrive can be attached to multiple instances.
	BlockStorageSharedDrive = BlockStorage("shared_drive")
)

type DiskType string

const (
	// DiskTypeSSD is the string representing a SSD disk type.
	DiskTypeSSD = DiskType("iscsi_ssd")
	// DiskTypeHDD is the string representing a HDD disk type.
	DiskTypeHDD = DiskType("iscsi_hdd")
)

type AttachedDiskSpec struct {
	// StorageType is the type of the storage, for now is just ISCSI block storage
	// of type DriveArray or SharedDrive.
	// Default is "drive_array".
	// +optional
	StorageType *BlockStorage `json:"storageType,omitempty"`
	// DiskType is the type of the storage.
	// Default is "iscsi_ssd".
	// +optional
	DiskType *DiskType `json:"diskType,omitempty"`
	// Size is the size of the disk in MB.
	// Default is 40960.
	// +optional
	Size *uint32 `json:"size,omitempty"`
	// DiskCount is the number of disks to be attached.
	// Default is 1.
	// +optional
	DiskCount int `json:"diskCount,omitempty"`
	// DiskLabel is the label that identifies the disk.
	// +optional
	DiskLabel string `json:"diskLabel,omitempty"`
}

// MetalsoftMachineSpec defines the desired state of MetalsoftMachine
type MetalsoftMachineSpec struct {
	// InstanceLabel is the label that identifies the instance.
	// +optional
	InstanceLabel string `json:"instanceLabel"`

	// InstanceBootMethod is the boot method of the instance.
	// Default is "local_drives".
	// +optional
	InstanceBootMethod *InstanceBootMethod `json:"instanceBootMethod,omitempty"`

	// InstanceServerTypeName is the type of the server instance to use.
	// +optional
	InstanceServerTypeName string `json:"instanceServerTypeName"`

	// InstanceServerTypeID is the ID of the server instance to use.
	// +optional
	InstanceServerTypeID int `json:"instanceServerTypeID"`

	// OS Template label to use for the Control Plane instance and the Worker instances.
	// +optional
	OSTemplateLabel string `json:"osTemplateLabel"`

	// OS Template ID to use for the Control Plane instance and the Worker instances.
	// +optional
	OSTemplateID int `json:"osTemplateID"`

	// AdditionalDisks is a list of additional disks to be attached to the instance.
	// +optional
	AdditionalDisks []AttachedDiskSpec `json:"additionalDisks,omitempty"`

	// ProviderID is the unique identifier as specified by the cloud provider.
	// +optional
	ProviderID *string `json:"providerID,omitempty"`
}

// MetalsoftMachineStatus defines the observed state of MetalsoftMachine
type MetalsoftMachineStatus struct {
	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// Addresses contains the Metalsoft instance associated addresses.
	Addresses []clusterv1.MachineAddress `json:"addresses,omitempty"`

	// InstanceState is the state of the Metalsoft instance for this machine.
	// +optional
	InstanceState *MetalsoftInstanceResourceStatus `json:"instanceState,omitempty"`

	// FailureReason will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a succinct value suitable
	// for machine interpretation.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the Machine object and/or logged in the
	// controller's output.
	// +optional
	FailureReason *errors.MachineStatusError `json:"failureReason,omitempty"`

	// FailureMessage will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a more verbose string suitable
	// for logging and human consumption.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the Machine object and/or logged in the
	// controller's output.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// Conditions defines current service state of the MetalsoftMachine.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=metalsoftmachines,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this MetalsoftMachine belongs"
// +kubebuilder:printcolumn:name="InstanceState",type="string",JSONPath=".status.instanceState",description="Metalsoft instance state"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Machine ready status"
// +kubebuilder:printcolumn:name="InstanceID",type="string",JSONPath=".spec.providerID",description="Metalsoft instance ID"
// +kubebuilder:printcolumn:name="Machine",type="string",JSONPath=".metadata.ownerReferences[?(@.kind==\"Machine\")].name",description="Machine object which owns with this MetalsoftMachine"

// MetalsoftMachine is the Schema for the metalsoftmachines API
type MetalsoftMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetalsoftMachineSpec   `json:"spec,omitempty"`
	Status MetalsoftMachineStatus `json:"status,omitempty"`
}

// GetConditions returns the observations of the operational state of the MetalsoftMachine resource.
func (r *MetalsoftMachine) GetConditions() clusterv1.Conditions {
	return r.Status.Conditions
}

// SetConditions sets the underlying service state of the MetalsoftMachine to the predescribed clusterv1.Conditions.
func (r *MetalsoftMachine) SetConditions(conditions clusterv1.Conditions) {
	r.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// MetalsoftMachineList contains a list of MetalsoftMachine
type MetalsoftMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetalsoftMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetalsoftMachine{}, &MetalsoftMachineList{})
}
