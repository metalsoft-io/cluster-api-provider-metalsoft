package scope

import (
	"context"
	"errors"
	"fmt"

	infrav1 "github.com/metalsoft-io/cluster-api-provider-metalsoft/api/v1alpha1"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MachineScopeParams defines the input parameters used to create a new MachineScope.
type MachineScopeParams struct {
	Client           client.Client
	Cluster          *clusterv1.Cluster
	Machine          *clusterv1.Machine
	MetalsoftCluster *infrav1.MetalsoftCluster
	MetalsoftMachine *infrav1.MetalsoftMachine
}

// MachineScope defines a scope defined around a machine and its cluster.
type MachineScope struct {
	client      client.Client
	patchHelper *patch.Helper

	Cluster          *clusterv1.Cluster
	Machine          *clusterv1.Machine
	MetalsoftCluster *infrav1.MetalsoftCluster
	MetalsoftMachine *infrav1.MetalsoftMachine
}

// NewMachineScope creates a new MachineScope from the supplied parameters.
// This is meant to be called for each reconcile iteration
// both MetalsoftClusterReconciler and MetalsoftMachineReconciler.
func NewMachineScope(ctx context.Context, params MachineScopeParams) (*MachineScope, error) {
	if params.Client == nil {
		return nil, errors.New("client is required when creating a MachineScope")
	}
	if params.Machine == nil {
		return nil, errors.New("machine is required when creating a MachineScope")
	}
	if params.Cluster == nil {
		return nil, errors.New("cluster is required when creating a MachineScope")
	}
	if params.MetalsoftCluster == nil {
		return nil, errors.New("metalsoft cluster is required when creating a MachineScope")
	}
	if params.MetalsoftMachine == nil {
		return nil, errors.New("metalsoft machine is required when creating a MachineScope")
	}

	helper, err := patch.NewHelper(params.MetalsoftMachine, params.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to init patch helper: %w", err)
	}
	return &MachineScope{
		client:      params.Client,
		patchHelper: helper,

		Cluster:          params.Cluster,
		Machine:          params.Machine,
		MetalsoftCluster: params.MetalsoftCluster,
		MetalsoftMachine: params.MetalsoftMachine,
	}, nil
}

// Name returns the MetalsoftMachine name.
func (m *MachineScope) Name() string {
	return m.MetalsoftMachine.Name
}

// Namespace returns the namespace name.
func (m *MachineScope) Namespace() string {
	return m.MetalsoftMachine.Namespace
}

// IsControlPlane returns true if the machine is a control plane.
func (m *MachineScope) IsControlPlane() bool {
	return util.IsControlPlaneMachine(m.Machine)
}

// Role returns the machine role from the labels.
func (m *MachineScope) Role() string {
	if util.IsControlPlaneMachine(m.Machine) {
		return "control-plane"
	}

	return "node"
}

// GetProviderID returns the MetalsoftMachine providerID from the spec.
func (m *MachineScope) GetProviderID() string {
	if m.MetalsoftMachine.Spec.ProviderID != nil {
		return *m.MetalsoftMachine.Spec.ProviderID
	}
	return ""
}

// ToDo
// SetProviderID sets the MetalsoftMachine providerID in spec from droplet id.
func (m *MachineScope) SetProviderID() {
	pid := "ToDo"
	providerID := fmt.Sprintf("metalsoft://%s", pid)
	m.MetalsoftMachine.Spec.ProviderID = pointer.String(providerID)
}

// SetReady sets the MetalsoftMachine Ready Status.
func (m *MachineScope) SetReady() {
	m.MetalsoftMachine.Status.Ready = true
}

// SetAddresses sets the addresses field on the MetalsoftMachine.
func (m *MachineScope) SetAddresses(addrs []clusterv1.MachineAddress) {
	m.MetalsoftMachine.Status.Addresses = addrs
}

// PatchObject persists the cluster configuration and status.
func (m *MachineScope) PatchObject() error {
	return m.patchHelper.Patch(context.TODO(), m.MetalsoftMachine)
}

// Close closes the current scope persisting the cluster configuration and status.
func (m *MachineScope) Close() error {
	return m.PatchObject()
}
