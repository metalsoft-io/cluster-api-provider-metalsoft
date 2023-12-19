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

package controller

import (
	"context"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	// "k8s.io/client-go/tools/record"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/metalsoft-io/cluster-api-provider-metalsoft/api/v1alpha1"
	metalsoft "github.com/metalsoft-io/cluster-api-provider-metalsoft/pkg/cloud/metalsoft"
	"github.com/metalsoft-io/cluster-api-provider-metalsoft/pkg/cloud/metalsoft/scope"
	"github.com/metalsoft-io/cluster-api-provider-metalsoft/util/reconciler"
	"github.com/pkg/errors"
)

// MetalsoftClusterReconciler reconciles a MetalsoftCluster object
type MetalsoftClusterReconciler struct {
	client.Client
	// Recorder         record.EventRecorder
	Scheme           *runtime.Scheme
	MetalSoftClient  *metalsoft.MetalSoftClient
	ReconcileTimeout time.Duration
	WatchFilterValue string
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=metalsoftclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=metalsoftclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=metalsoftclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *MetalsoftClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx, cancel := context.WithTimeout(ctx, reconciler.DefaultedLoopTimeout(r.ReconcileTimeout))
	defer cancel()

	log := log.FromContext(ctx)
	metalsoftCluster := &infrav1.MetalsoftCluster{}
	err := r.Get(ctx, req.NamespacedName, metalsoftCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("MetalsoftCluster resource not found or already deleted")
			return ctrl.Result{}, nil
		}

		log.Error(err, "Unable to fetch MetalsoftCluster resource")
		return ctrl.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, metalsoftCluster.ObjectMeta)
	if err != nil {
		log.Error(err, "Failed to get owner cluster")
		return ctrl.Result{}, err
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(cluster, metalsoftCluster) {
		log.Info("MetalsoftCluster of linked Cluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	// Create the cluster scope
	clusterScope, err := scope.NewClusterScope(ctx, scope.ClusterScopeParams{
		Client:           r.Client,
		Cluster:          cluster,
		MetalsoftCluster: metalsoftCluster,
	})
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any MetalsoftMachine changes.
	defer func() {
		if err := clusterScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted clusters
	if !metalsoftCluster.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.reconcileDelete(ctx, clusterScope)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, clusterScope)
}

func (r *MetalsoftClusterReconciler) reconcileNormal(ctx context.Context, clusterScope *scope.ClusterScope) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling MetalsoftCluster")

	if !controllerutil.ContainsFinalizer(clusterScope.MetalsoftCluster, infrav1.ClusterFinalizer) {
		controllerutil.AddFinalizer(clusterScope.MetalsoftCluster, infrav1.ClusterFinalizer)
		if err := clusterScope.PatchObject(); err != nil {
			return ctrl.Result{}, err
		}
	}

	datacenterName := clusterScope.DatacenterName()
	infrastructureLabel := clusterScope.InfrastructureLabel()

	log.Info("DatacenterName: " + datacenterName)
	log.Info("InfrastructureLabel: " + infrastructureLabel)
	// Create the Metalsoft Infrastructure

	return ctrl.Result{}, nil
}

func (r *MetalsoftClusterReconciler) reconcileDelete(ctx context.Context, clusterScope *scope.ClusterScope) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling Delete MetalsoftCluster")

	// ToDo: Delete MetalsoftCluster

	controllerutil.RemoveFinalizer(clusterScope.MetalsoftCluster, infrav1.ClusterFinalizer)
	record.Event(clusterScope.MetalsoftCluster, "MetalsoftClusterReconcile", "Reconciled")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MetalsoftClusterReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.MetalsoftCluster{}).
		Complete(r)
}
