package handlers

import (
	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/reference"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/asset"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resource"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	apiregistrationclientset "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewApiserviceHandler(o *Options) *apiservice {
	return &apiservice{
		client:  o.Client.APIRegistration,
		dynamic: resource.NewAPIServiceClient(o.Client.Dynamic),
		asset:   asset.New(o.OperandContext),
	}
}

type apiservice struct {
	client apiregistrationclientset.Interface
	dynamic *resource.APIService
	asset   *asset.Asset
}

func (a *apiservice) Handle(context *ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) (out *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, stop bool, handleErr error) {
	out = in

	name := a.asset.APIService().Name()
	current, err := a.client.ApiregistrationV1().APIServices().Get(name, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.APIServiceNotAvailable, err)
			return
		}

		// No APIService object
		object := a.asset.APIService().New()
		object.Spec.CABundle = context.GetBundle().ServingCertCA
		context.ControllerSetter().Set(object, in)

		apiservice, err := a.dynamic.Ensure(object)
		if err != nil {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.APIServiceNotAvailable, err)
			return
		}

		current = apiservice
		klog.V(2).Infof("key=%s resource=%T/%s successfully created", in.Name, current, current.Name)
	}

	if ref := in.Status.Resources.APiServiceRef; ref != nil && ref.ResourceVersion == current.ResourceVersion {
		klog.V(2).Infof("key=%s resource=%T/%s is in sync", in.Name, current, current.Name)
		return
	}

	newRef, err := reference.GetReference(current)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
		return
	}

	klog.V(2).Infof("key=%s resource=%T/%s resource-version=%s setting object reference", in.Name, current, current.Name, newRef.ResourceVersion)
	out.Status.Resources.APiServiceRef = newRef

	return
}