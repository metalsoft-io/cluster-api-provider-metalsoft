package scope

import (
	"context"

	infrav1 "github.com/metalsoft-io/cluster-api-provider-metalsoft/api/v1alpha1"
	"github.com/pkg/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type ClusterScopeParams struct {
	Client           client.Client
	Cluster          *clusterv1.Cluster
	MetalsoftCluster *infrav1.MetalsoftCluster
}

// ClusterScope defines the basic context for an actuator to operate upon.
type ClusterScope struct {
	client      client.Client
	patchHelper *patch.Helper

	Cluster          *clusterv1.Cluster
	MetalsoftCluster *infrav1.MetalsoftCluster
}

// NewClusterScope creates a new ClusterScope from the supplied parameters.
// This is meant to be called for each reconcile iteration only on MetalsoftClusterReconciler.
func NewClusterScope(ctx context.Context, params ClusterScopeParams) (*ClusterScope, error) {
	if params.Cluster == nil {
		return nil, errors.New("Cluster is required when creating a ClusterScope")
	}
	if params.MetalsoftCluster == nil {
		return nil, errors.New("MetalsoftCluster is required when creating a ClusterScope")
	}

	helper, err := patch.NewHelper(params.MetalsoftCluster, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &ClusterScope{
		client:           params.Client,
		Cluster:          params.Cluster,
		MetalsoftCluster: params.MetalsoftCluster,
		patchHelper:      helper,
	}, nil
}

// Name returns the cluster name.
func (s *ClusterScope) Name() string {
	return s.Cluster.Name
}

// Namespace returns the cluster namespace.
func (s *ClusterScope) Namespace() string {
	return s.Cluster.Namespace
}

// DatacenterName returns the current datacenter name.
func (s *ClusterScope) DatacenterName() string {
	return s.MetalsoftCluster.Spec.DatacenterName
}

// InfrastructureLabel returns the current infrastructure label.
func (s *ClusterScope) InfrastructureLabel() string {
	return s.MetalsoftCluster.Spec.InfrastructureLabel
}

// VipSubnetLabel returns the current VIP subnet label.
// func (s *ClusterScope) VipSubnetLabel() string {
// 	return s.MetalsoftCluster.Spec.VipSubnetLabel
// }

// InfrastructureID returns the current infrastructure ID.
func (s *ClusterScope) InfrastructureID() int {
	return s.MetalsoftCluster.Spec.InfrastructureID
}

// SubnetID returns the current subnet ID.
func (s *ClusterScope) SubnetID() int {
	return s.MetalsoftCluster.Spec.Network.SubnetID
}

// ControlPlaneEndpoint returns the cluster control-plane endpoint.
func (s *ClusterScope) ControlPlaneEndpoint() clusterv1.APIEndpoint {
	endpoint := s.MetalsoftCluster.Spec.ControlPlaneEndpoint
	endpoint.Port = 6443
	return endpoint
}

// SetReady sets the MetalsoftCluster Ready Status.
func (s *ClusterScope) SetReady() {
	s.MetalsoftCluster.Status.Ready = true
}

// SetControlPlaneEndpoint sets cluster control-plane endpoint.
func (s *ClusterScope) SetControlPlaneEndpoint(endpoint clusterv1.APIEndpoint) {
	s.MetalsoftCluster.Spec.ControlPlaneEndpoint = endpoint
}

// set infrastructure ID
func (s *ClusterScope) SetInfrastructureID(infrastructureID int) {
	s.MetalsoftCluster.Spec.InfrastructureID = infrastructureID
}

// set subnet ID
func (s *ClusterScope) SetSubnetID(subnetID int) {
	s.MetalsoftCluster.Spec.Network.SubnetID = subnetID
}

// PatchObject persists the cluster configuration and status.
func (s *ClusterScope) PatchObject() error {
	return s.patchHelper.Patch(context.TODO(), s.MetalsoftCluster)
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *ClusterScope) Close() error {
	return s.PatchObject()
}
