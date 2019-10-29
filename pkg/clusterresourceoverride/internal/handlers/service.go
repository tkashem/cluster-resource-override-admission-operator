package handlers

import (
	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/reference"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/asset"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resource"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resourcewatch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewServiceHandler(o *Options) *serviceHandler {
	return &serviceHandler{
		client:  o.Client.Kubernetes,
		dynamic: resource.NewDynamicServiceClient(o.Client.Dynamic),
		lister: o.KubeLister,
		asset:   asset.New(o.OperandContext),
	}
}

type serviceHandler struct {
	client  kubernetes.Interface
	dynamic *resource.Service
	lister *resourcewatch.Lister
	asset   *asset.Asset
}

func (s *serviceHandler) Handle(context *ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) (out *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, stop bool, handleErr error) {
	out = in

	desired := s.asset.Service().New()
	context.ControllerSetter().Set(desired, in)

	name := s.asset.Service().Name()
	current, err := s.lister.CoreV1ServiceLister().Services(context.WebhookNamespace()).Get(name)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
			return
		}

		object, err := s.dynamic.Ensure(desired)
		if err != nil {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
			return
		}

		current = object
		klog.V(2).Infof("key=%s resource=%T/%s successfully created", in.Name, current, current.Name)
	}

	if ref := out.Status.Resources.ServiceRef; ref != nil && ref.ResourceVersion == current.ResourceVersion {
		klog.V(2).Infof("key=%s resource=%T/%s is in sync", in.Name, current, current.Name)
		return
	}

	// Has the Service object been modified?
	//if !s.Equal(desired, current) {
	//	object, err := s.dynamic.Ensure(desired)
	//	if err != nil {
	//		handleErr = croerrors.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
	//		return
	//	}
	//
	//	current = object
	//}

	newRef, err := reference.GetReference(current)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.CannotSetReference, err)
		return
	}

	klog.V(2).Infof("key=%s resource=%T/%s resource-version=%s setting object reference", in.Name, current, current.Name, newRef.ResourceVersion)

	out.Status.Resources.ServiceRef = newRef
	return
}

func (s *serviceHandler) Equal(this, that *corev1.Service) bool {
	return equality.Semantic.DeepDerivative(&this.Spec, &that.Spec) &&
		equality.Semantic.DeepDerivative(this.GetObjectMeta(), that.GetObjectMeta())
}