package handlers

import (
	"fmt"
	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/secondarywatch"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/status"
	"k8s.io/klog"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewDeploymentStatusHandler(o *Options) *deploymentStatusHandler {
	return &deploymentStatusHandler{
		lister: o.KubeLister,
	}
}

type deploymentStatusHandler struct {
	lister *secondarywatch.Lister
}

func (c *deploymentStatusHandler) Handle(context *ReconcileRequestContext, original *autoscalingv1.ClusterResourceOverride) (current *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, handleErr error) {
	current = original

	name := original.Status.Resources.DeploymentRef.Name
	object, err := c.lister.AppsV1DeploymentLister().Deployments(context.WebhookNamespace()).Get(name)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.DeploymentNotReady, err)
		return
	}

	done, err := status.GetDeploymentStatus(object)
	if done {
		klog.V(2).Infof("key=%s resource=%T/%s deployment is ready", original.Name, object, object.Name)

		condition.NewBuilderWithStatus(&current.Status).WithInstallReady()
		current.Status.Version = context.OperandVersion()
		current.Status.Image = context.OperandImage()
		return
	}

	klog.V(2).Infof("key=%s resource=%T/%s deployment is not ready", original.Name, object, object.Name)

	if err == nil {
		err = fmt.Errorf("name=%s waiting for deployment to complete", object.Name)
	}

	handleErr = condition.NewInstallReadinessError(autoscalingv1.DeploymentNotReady, err)
	return
}
