package handlers

import (
	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/asset"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/secondarywatch"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/status"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewAvailabilityHandler(o *Options) *availabilityHandler {
	return &availabilityHandler{
		asset:  asset.New(o.OperandContext),
		lister: o.KubeLister,
	}
}

type availabilityHandler struct {
	asset  *asset.Asset
	lister *secondarywatch.Lister
}

func (a *availabilityHandler) Handle(context *ReconcileRequestContext, original *autoscalingv1.ClusterResourceOverride) (current *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, handleErr error) {
	current = original
	builder := condition.NewBuilderWithStatus(&current.Status)

	name := a.asset.Deployment().Name()
	object, err := a.lister.AppsV1DaemonSetLister().DaemonSets(context.WebhookNamespace()).Get(name)
	switch {
	case err == nil:
		done, statusErr := status.GetDaemonSetStatus(object)
		if done {
			builder.WithAvailable(corev1.ConditionTrue, "")
			return
		}

		builder.WithError(condition.NewAvailableError(autoscalingv1.AdmissionWebhookNotAvailable, statusErr))
	case k8serrors.IsNotFound(err):
		builder.WithError(condition.NewAvailableError(autoscalingv1.AdmissionWebhookNotAvailable, err))
	default:
		builder.WithError(condition.NewAvailableError(autoscalingv1.InternalError, err))
	}

	return
}
