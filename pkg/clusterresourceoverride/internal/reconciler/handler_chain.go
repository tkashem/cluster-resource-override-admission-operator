package reconciler

import (
	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/handlers"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Handler interface {
	Handle(context *handlers.ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) (out *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, stop bool, err error)
}

var _ Handler = HandlerChain{}

type HandlerChain []Handler

func (h HandlerChain) Handle(context *handlers.ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) (out *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, stop bool, err error) {
	for _, handler := range h {
		// Invoke the handler.
		out, result, stop, err = handler.Handle(context, in)
		if err != nil {
			// set status.conditions based on reconciliation error
			out = out.DeepCopy()
			condition.NewBuilderWithStatus(&out.Status).WithError(err)

			// if there was an error, we stop further processing.
			// and let the object be requeued.
			return
		}

		// no error, but we have been asked to stop further processing.
		// TODO: I don't think we need it. let the handler set result.Requeue to true
		if stop == true {
			return
		}

		if result.Requeue || result.RequeueAfter > 0 {
			// the handler has asked to requeue the object.
			return
		}

		in = out
	}

	return
}
