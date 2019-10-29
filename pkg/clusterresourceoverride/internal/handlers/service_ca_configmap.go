package handlers

import (
	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/reference"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/asset"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resource"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resourcewatch"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewServiceCAConfigMapHandler(o *Options) *serviceCAConfigMapHandler {
	return &serviceCAConfigMapHandler{
		client:  o.Client.Kubernetes,
		dynamic: resource.NewConfigMapClient(o.Client.Dynamic),
		lister: o.KubeLister,
		asset:   asset.New(o.OperandContext),
	}
}

type serviceCAConfigMapHandler struct {
	client  kubernetes.Interface
	dynamic *resource.ConfigMapEnsurer
	lister *resourcewatch.Lister
	asset   *asset.Asset
}

func (c *serviceCAConfigMapHandler) Handle(context *ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) (out *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, stop bool, handleErr error) {
	out = in

	desired := c.asset.CABundleConfigMap().New()
	context.ControllerSetter().Set(desired, in)

	name := c.asset.CABundleConfigMap().Name()
	current, err := c.lister.CoreV1ConfigMapLister().ConfigMaps(context.WebhookNamespace()).Get(name)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
			return
		}

		cm, err := c.dynamic.Ensure(desired)
		if err != nil {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
			return
		}

		current = cm
		klog.V(2).Infof("key=%s resource=%T/%s successfully created", in.Name, current, current.Name)
	}

	if !verifyLabelsAndAnnotations(desired, current) {
		klog.V(2).Infof("key=%s resource=%T/%s resource has drifted", in.Name, current, current.Name)

		cm, err := c.dynamic.Ensure(desired)
		if err != nil {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
			return
		}

		current = cm
	}

	if ref := out.Status.Resources.ServiceCAConfigMapRef; ref != nil && ref.ResourceVersion == current.ResourceVersion {
		klog.V(2).Infof("key=%s resource=%T/%s is in sync", in.Name, current, current.Name)
		return
	}

	newRef, err := reference.GetReference(current)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
		return
	}

	klog.V(2).Infof("key=%s resource=%T/%s resource-version=%s setting object reference", in.Name, current, current.Name, newRef.ResourceVersion)
	out.Status.Resources.ServiceCAConfigMapRef = newRef
	return
}

func verifyLabelsAndAnnotations(desired, current *corev1.ConfigMap) bool {
	if len(current.Labels) == 0 || len(current.Annotations) == 0 {
		return false
	}

	for key, expected := range desired.Labels {
		if actual, ok := current.Labels[key]; !ok || expected != actual {
			return false
		}
	}

	for key, expected := range desired.Annotations {
		if actual, ok := current.Annotations[key]; !ok || expected != actual {
			return false
		}
	}

	return true
}

