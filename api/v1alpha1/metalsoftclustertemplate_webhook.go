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
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var metalsoftclustertemplatelog = logf.Log.WithName("metalsoftclustertemplate-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *MetalsoftClusterTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1alpha1-metalsoftclustertemplate,mutating=true,failurePolicy=fail,sideEffects=None,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=metalsoftclustertemplates,verbs=create;update,versions=v1alpha1,name=default.metalsoftclustertemplate.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &MetalsoftClusterTemplate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *MetalsoftClusterTemplate) Default() {
	metalsoftclustertemplatelog.Info("default", "name", r.Name)

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1alpha1-metalsoftclustertemplate,mutating=false,failurePolicy=fail,sideEffects=None,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=metalsoftclustertemplates,verbs=create;update,versions=v1alpha1,name=validation.metalsoftclustertemplate.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Validator = &MetalsoftClusterTemplate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MetalsoftClusterTemplate) ValidateCreate() (admission.Warnings, error) {
	metalsoftclustertemplatelog.Info("validate create", "name", r.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MetalsoftClusterTemplate) ValidateUpdate(oldRaw runtime.Object) (admission.Warnings, error) {
	metalsoftclustertemplatelog.Info("validate update", "name", r.Name)
	old, ok := oldRaw.(*MetalsoftClusterTemplate)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected an MetalsoftClusterTemplate but got a %T", oldRaw))
	}

	if !reflect.DeepEqual(r.Spec, old.Spec) {
		return nil, apierrors.NewBadRequest("MetalsoftClusterTemplate.Spec is immutable")
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MetalsoftClusterTemplate) ValidateDelete() (admission.Warnings, error) {
	metalsoftclustertemplatelog.Info("validate delete", "name", r.Name)

	return nil, nil
}
