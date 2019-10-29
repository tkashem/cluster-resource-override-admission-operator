package handlers

import (
	"fmt"
	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/asset"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	dynamicclient "github.com/openshift/cluster-resource-override-admission-operator/pkg/dynamic"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resource"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resourcewatch"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/status"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewDeploymentStatusHandler(o *Options) *deploymentStatusHandler {
	return &deploymentStatusHandler{
		client:  o.Client.Kubernetes,
		dynamic: o.Client.Dynamic,
		deployment: resource.NewDeploymentClient(o.Client.Dynamic),
		lister: o.KubeLister,
		asset:   asset.New(o.OperandContext),
	}
}

type deploymentStatusHandler struct {
	client  kubernetes.Interface
	deployment *resource.Deployment
	dynamic dynamicclient.Client
	lister *resourcewatch.Lister
	asset   *asset.Asset
}

func (c *deploymentStatusHandler) Handle(context *ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) (out *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, stop bool, handleErr error) {
	out = in

	name := in.Status.Resources.DeploymentRef.Name
	current, err := c.lister.AppsV1DeploymentLister().Deployments(context.WebhookNamespace()).Get(name)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.DeploymentNotReady, err)
		return
	}

	done, err := status.GetDeploymentStatus(current)
	if done {
		klog.V(2).Infof("key=%s resource=%T/%s deployment is ready", in.Name, current, current.Name)

		condition.NewBuilderWithStatus(&out.Status).WithInstallReady()
		out.Status.Version = context.OperandVersion()
		out.Status.Image = context.OperandImage()
		return
	}

	klog.V(2).Infof("key=%s resource=%T/%s deployment is not ready", in.Name, current, current.Name)

	if err == nil {
		err = fmt.Errorf("name=%s waiting for deployment to complete", current.Name)
	}

	handleErr = condition.NewInstallReadinessError(autoscalingv1.DeploymentNotReady, err)
	return
}
