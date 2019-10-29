package handlers

import (
	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewValidationHandler(o *Options) *validation {
	return &validation{}
}

type validation struct {
}

func (c *validation) Handle(context *ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) (out *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, stop bool, handleErr error) {
	out = in

	validationErr := in.Spec.Config.Validate()
	if validationErr != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.InvalidParameters, validationErr)
	}

	return
}
