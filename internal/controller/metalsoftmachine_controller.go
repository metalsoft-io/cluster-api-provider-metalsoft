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

// MetalsoftMachineReconciler reconciles a MetalsoftMachine object
type MetalsoftMachineReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	MetalSoftClient  *metalsoft.MetalSoftClient
	ReconcileTimeout time.Duration
	WatchFilterValue string
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=metalsoftmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=metalsoftmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=bootstrap.cluster.x-k8s.io,resources=kubeadmconfigs;kubeadmconfigs/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets;,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *MetalsoftMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx, cancel := context.WithTimeout(ctx, reconciler.DefaultedLoopTimeout(r.ReconcileTimeout))
	defer cancel()

	log := ctrl.LoggerFrom(ctx)
	metalsoftMachine := &infrav1.MetalsoftMachine{}
	err := r.Get(ctx, req.NamespacedName, metalsoftMachine)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("MetalsoftMachine resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to get MetalsoftMachine")
		return ctrl.Result{}, err
	}

	// Fetch the Machine
	machine, err := util.GetOwnerMachine(ctx, r.Client, metalsoftMachine.ObjectMeta)
	if err != nil {
		log.Error(err, "Failed to get owner Machine")
		return ctrl.Result{}, err
	}
	if machine == nil {
		log.Info("Machine Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("machine", machine.Name)

	// Fetch the Cluster
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		log.Info("Machine is missing cluster label or cluster does not exist")
		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(cluster, metalsoftMachine) {
		log.Info("MetalsoftMachine or linked Cluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)

	// Fetch the MetalsoftCluster
	metalsoftCluster := &infrav1.MetalsoftCluster{}
	metalsoftClusterNamespacedName := client.ObjectKey{
		Namespace: metalsoftMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Get(ctx, metalsoftClusterNamespacedName, metalsoftCluster); err != nil {
		log.Error(err, "MetalsoftCluster is not available yet")
		return ctrl.Result{}, err
	}

	// Create the cluster scope
	clusterScope, err := scope.NewClusterScope(ctx, scope.ClusterScopeParams{
		Client:           r.Client,
		Cluster:          cluster,
		MetalsoftCluster: metalsoftCluster,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Create the machine scope
	machineScope, err := scope.NewMachineScope(ctx, scope.MachineScopeParams{
		Client:           r.Client,
		Cluster:          cluster,
		Machine:          machine,
		MetalsoftCluster: metalsoftCluster,
		MetalsoftMachine: metalsoftMachine,
	})
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any MetalsoftMachine changes.
	defer func() {
		if err := machineScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted machines
	if !metalsoftMachine.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.reconcileDelete(ctx, machineScope, clusterScope)
	}

	// Handle non-deleted machines
	return r.reconcileNormal(ctx, machineScope, clusterScope)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MetalsoftMachineReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.MetalsoftMachine{}).
		Complete(r)
}

func (r *MetalsoftMachineReconciler) reconcileNormal(ctx context.Context, machineScope *scope.MachineScope, clusterScope *scope.ClusterScope) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling MetalsoftMachine")

	metalsoftmachine := machineScope.MetalsoftMachine

	if metalsoftmachine.Status.FailureReason != nil || metalsoftmachine.Status.FailureMessage != nil {
		log.Info("Error state detected, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	// If the MetalsoftMachine doesn't have our finalizer, add it
	controllerutil.AddFinalizer(metalsoftmachine, infrav1.MachineFinalizer)
	if err := machineScope.PatchObject(); err != nil {
		return ctrl.Result{}, err
	}

	if !machineScope.Cluster.Status.InfrastructureReady {
		log.Info("Cluster infrastructure is not ready yet")
		// conditions.MarkFalse(machineScope.MetalsoftMachine, ) // ToDo - add condition
		return ctrl.Result{}, nil
	}

	// Make sure bootstrap data secret is available and populated
	if machineScope.Machine.Spec.Bootstrap.DataSecretName == nil {
		log.Info("Bootstrap data secret is not yet available")
		// conditions.MarkFalse(machineScope.MetalsoftMachine, ) // ToDo - add condition
		return ctrl.Result{}, nil
	}

	// ToDo
	return ctrl.Result{}, nil
}

func (r *MetalsoftMachineReconciler) reconcileDelete(ctx context.Context, machineScope *scope.MachineScope, clusterScope *scope.ClusterScope) error {
	log := ctrl.LoggerFrom(ctx, "machine", machineScope.Machine.Name, "cluster", machineScope.Cluster.Name)
	log.Info("Reconciling Delete MetalsoftMachine")

	metalsoftmachine := machineScope.MetalsoftMachine

	// ToDo

	controllerutil.RemoveFinalizer(metalsoftmachine, infrav1.MachineFinalizer)
	record.Event(machineScope.MetalsoftMachine, "MetalsoftMachineReconcile", "Reconciled")
	return nil
}
