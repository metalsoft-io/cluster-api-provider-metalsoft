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
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var metalsoftclusterlog = logf.Log.WithName("metalsoftcluster-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *MetalsoftCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1alpha1-metalsoftcluster,mutating=true,failurePolicy=fail,sideEffects=None,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=metalsoftclusters,verbs=create;update,versions=v1alpha1,name=default.metalsoftcluster.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &MetalsoftCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *MetalsoftCluster) Default() {
	metalsoftclusterlog.Info("default", "name", r.Name)

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1alpha1-metalsoftcluster,mutating=false,failurePolicy=fail,sideEffects=None,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=metalsoftclusters,verbs=create;update,versions=v1alpha1,name=validation.metalsoftcluster.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Validator = &MetalsoftCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MetalsoftCluster) ValidateCreate() (admission.Warnings, error) {
	metalsoftclusterlog.Info("validate create", "name", r.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MetalsoftCluster) ValidateUpdate(oldRaw runtime.Object) (admission.Warnings, error) {
	metalsoftclusterlog.Info("validate update", "name", r.Name)
	var allErrs field.ErrorList
	old := oldRaw.(*MetalsoftCluster)

	if !reflect.DeepEqual(r.Spec.DatacenterName, old.Spec.DatacenterName) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "DatacenterName"),
				r.Spec.DatacenterName, "field is immutable"),
		)
	}

	if !reflect.DeepEqual(r.Spec.InfrastructureLabel, old.Spec.InfrastructureLabel) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "InfrastructureLabel"),
				r.Spec.InfrastructureLabel, "field is immutable"),
		)
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(GroupVersion.WithKind("MetalsoftCluster").GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MetalsoftCluster) ValidateDelete() (admission.Warnings, error) {
	metalsoftclusterlog.Info("validate delete", "name", r.Name)

	return nil, nil
}
