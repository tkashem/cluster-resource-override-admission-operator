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

func NewServiceCertSecretHandler(o *Options) *serviceCertSecretHandler {
	return &serviceCertSecretHandler{
		client:  o.Client.Kubernetes,
		dynamic: resource.NewDynamicServiceClient(o.Client.Dynamic),
		lister: o.KubeLister,
		asset:   asset.New(o.OperandContext),
	}
}

type serviceCertSecretHandler struct {
	client  kubernetes.Interface
	dynamic *resource.Service
	lister *resourcewatch.Lister
	asset   *asset.Asset
}

func (c *serviceCertSecretHandler) Handle(context *ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) (out *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, stop bool, handleErr error) {
	out = in

	// Make sure that we have all certs generated
	secretName := c.asset.ServiceServingSecret().Name()

	current, err := c.lister.CoreV1SecretLister().Secrets(context.WebhookNamespace()).Get(secretName)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)

		if k8serrors.IsNotFound(err) {
			// We are still waiting for the server serving Secret object object to be
			// created by the service-ca operator.
			// No further action in the handler chain until we have a secret object.
			klog.V(2).Infof("key=%s resource=%T/%s waiting for server serving secret to be created by service-ca operator", in.Name, current, secretName)
		}

		return
	}

	// we need to annotate this secret so that if it is deleted/updated
	// we can enqueue the CR
	if owner, ok := current.Annotations[context.GetOwnerAnnotationKey()]; !ok || owner != in.Name {
		copy := current.DeepCopy()
		if len(copy.Annotations) == 0 {
			copy.Annotations = map[string]string{}
		}

		copy.Annotations[context.GetOwnerAnnotationKey()] = in.Name
		updated, updateErr := c.client.CoreV1().Secrets(context.WebhookNamespace()).Update(copy)
		if updateErr != nil {
			handleErr = updateErr
			return
		}

		current = updated
	}

	if ref := out.Status.Resources.ServiceCertSecretRef; ref != nil && ref.ResourceVersion == current.ResourceVersion {
		klog.V(2).Infof("key=%s resource=%T/%s is in sync", in.Name, current, current.Name)
		return
	}

	newRef, err := reference.GetReference(current)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.CannotSetReference, err)
		return
	}

	klog.V(2).Infof("key=%s resource=%T/%s resource-version=%s setting object reference", in.Name, current, current.Name, newRef.ResourceVersion)

	out.Status.Resources.ServiceCertSecretRef = newRef
	return
}
