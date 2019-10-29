package handlers

import (
	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/reference"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/asset"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resource"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resourcewatch"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewWebhookConfigurationHandlerHandler(o *Options) *webhookConfigurationHandler {
	return &webhookConfigurationHandler{
		client:  o.Client.Kubernetes,
		dynamic: resource.NewMutatingWebhookConfigurationClient(o.Client.Dynamic),
		lister: o.KubeLister,
		asset:   asset.New(o.OperandContext),
	}
}

type webhookConfigurationHandler struct {
	client  kubernetes.Interface
	dynamic *resource.MutatingWebhookConfiguration
	lister *resourcewatch.Lister
	asset   *asset.Asset
}

func (w *webhookConfigurationHandler) Handle(context *ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) (out *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, stop bool, handleErr error) {
	out = in

	name := w.asset.NewMutatingWebhookConfiguration().Name()
	current, err := w.lister.AdmissionRegistrationV1beta1MutatingWebhookConfigurationLister().Get(name)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
			return
		}

		// No MutatingWebhookConfiguration object!
		object := w.asset.NewMutatingWebhookConfiguration().New()
		context.ControllerSetter().Set(object, in)

		clientCA := context.GetBundle().KubeClintCA
		for i := range object.Webhooks {
			object.Webhooks[i].ClientConfig.CABundle = clientCA
		}

		apiservice, err := w.dynamic.Ensure(object)
		if err != nil {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
			return
		}

		current = apiservice
		klog.V(2).Infof("key=%s resource=%T/%s successfully created", in.Name, current, current.Name)
	}

	if ref := in.Status.Resources.MutatingWebhookConfigurationRef; ref != nil && ref.ResourceVersion == current.ResourceVersion {
		klog.V(2).Infof("key=%s resource=%T/%s is in sync", in.Name, current, current.Name)
		return
	}

	newRef, err := reference.GetReference(current)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
		return
	}

	klog.V(2).Infof("key=%s resource=%T/%s resource-version=%s setting object reference", in.Name, current, current.Name, newRef.ResourceVersion)

	out.Status.Resources.MutatingWebhookConfigurationRef = newRef
	return
}