package handlers

import (
	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/asset"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resource"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/status"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiregistrationclientset "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewAvailabilityHandler(o *Options) *availability {
	return &availability{
		client:  o.Client.APIRegistration,
		kubeclient: o.Client.Kubernetes,
		dynamic: resource.NewAPIServiceClient(o.Client.Dynamic),
		asset:   asset.New(o.OperandContext),
	}
}

type availability struct {
	// client  kubernetes.Interface
	client apiregistrationclientset.Interface
	kubeclient kubernetes.Interface
	dynamic *resource.APIService
	asset   *asset.Asset
}

func (a *availability) Handle(context *ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) (out *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, stop bool, handleErr error) {
	out = in
	builder := condition.NewBuilderWithStatus(&out.Status)

	name := a.asset.APIService().Name()
	current, err := a.client.ApiregistrationV1().APIServices().Get(name, metav1.GetOptions{})
	switch {
	case err == nil:
		status, message := status.IsAPIServiceAvailable(current)
		builder.WithAvailable(status, message)
	case k8serrors.IsNotFound(err):
		builder.WithError(condition.NewAvailableError(autoscalingv1.APIServiceNotAvailable, err))
	default:
		builder.WithError(condition.NewAvailableError(autoscalingv1.InternalError, err))
	}

	return
}
