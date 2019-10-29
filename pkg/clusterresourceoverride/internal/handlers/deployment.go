package handlers

import (
	"fmt"
	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/reference"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/asset"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	dynamicclient "github.com/openshift/cluster-resource-override-admission-operator/pkg/dynamic"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/ensurer"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/secondarywatch"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/status"
	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewDeploymentHandler(o *Options) *deploymentHandler {
	return &deploymentHandler{
		client:     o.Client.Kubernetes,
		dynamic:    o.Client.Dynamic,
		deployment: ensurer.NewDeploymentEnsurer(o.Client.Dynamic),
		asset:      asset.New(o.OperandContext),
		lister:     o.KubeLister,
	}
}

type deploymentHandler struct {
	client     kubernetes.Interface
	deployment *ensurer.Deployment
	dynamic    dynamicclient.Ensurer
	lister     *secondarywatch.Lister
	asset      *asset.Asset
}

func (c *deploymentHandler) Handle(context *ReconcileRequestContext, original *autoscalingv1.ClusterResourceOverride) (current *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, handleErr error) {
	current = original

	name := c.asset.Deployment().Name()
	object, err := c.lister.AppsV1DeploymentLister().Deployments(context.WebhookNamespace()).Get(name)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
			return
		}

		// no deployment, either it's a new install or the deployment object has been removed.
		cm, err := c.EnsureDeployment(context, original)
		if err != nil {
			handleErr = err
			return
		}

		object = cm
		klog.V(2).Infof("key=%s resource=%T/%s successfully created", original.Name, object, object.Name)
	}

	// We have an existing deployment.
	// If the cert has changed or the configuration has changed
	// then we should recreate the pod.
	annotations := object.GetAnnotations()
	redeploy := false
	switch {
	case annotations[context.GetConfigurationHashAnnotationKey()] != current.Status.Hash.Configuration:
		klog.V(2).Infof("key=%s resource=%T/%s configuration hash mismatch", original.Name, object, object.Name)
		redeploy = true

	case annotations[context.GetServingCertHashAnnotationKey()] != current.Status.Hash.ServingCert:
		klog.V(2).Infof("key=%s resource=%T/%s serving cert hash mismatch", original.Name, object, object.Name)
		redeploy = true

	case status.IsDeploymentFailedCreate(&object.Status):
		klog.V(2).Infof("key=%s resource=%T/%s Deployment is in 'FailedCreate' state", original.Name, object, object.Name)
		redeploy = true
	}

	if redeploy {
		klog.V(2).Infof("key=%s resource=%T/%s redeploying", original.Name, object, object.Name)
		cm, err := c.EnsureDeployment(context, original)
		if err != nil {
			handleErr = err
			return
		}

		object = cm
	}

	if ref := current.Status.Resources.DeploymentRef; ref != nil && ref.ResourceVersion == object.ResourceVersion {
		klog.V(2).Infof("key=%s resource=%T/%s is in sync", original.Name, object, object.Name)
		return
	}

	newRef, err := reference.GetReference(object)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
		return
	}

	klog.V(2).Infof("key=%s resource=%T/%s resource-version=%s setting object reference", original.Name, object, object.Name, newRef.ResourceVersion)
	current.Status.Resources.DeploymentRef = newRef

	return
}

func (c *deploymentHandler) EnsureDeployment(context *ReconcileRequestContext, cro *autoscalingv1.ClusterResourceOverride) (current *appsv1.Deployment, err error) {
	// If a pod dies then the Deployment object goes into the following state
	//     message: 'Internal error occurred: failed calling webhook "clusterresourceoverrides.admission.autoscaling.openshift.io":
	//      the server is currently unable to handle the request'
	//	   reason: FailedCreate
	//     status: "True"
	//     type: ReplicaFailure
	//
	// If the MutatingWebhookConfiguration object already exists the Deployment object gets into the above state
	// and the Pod never gets recreated.
	// TODO: Find out if there is a better way to handle this error.
	name := c.asset.NewMutatingWebhookConfiguration().Name()
	if deleteErr := c.client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(name, &metav1.DeleteOptions{}); deleteErr != nil && !k8serrors.IsNotFound(deleteErr) {
		err = fmt.Errorf("failed to delete MutatingWebhookConfiguration - %s", deleteErr.Error())
		return
	}

	if err = c.EnsureRBAC(context, cro); err != nil {
		return
	}

	object := c.asset.Deployment().New()
	if len(object.Annotations) == 0 {
		object.Annotations = map[string]string{}
	}
	if len(object.Spec.Template.Annotations) == 0 {
		object.Spec.Template.Annotations = map[string]string{}
	}

	object.Spec.Template.Annotations[context.GetOwnerAnnotationKey()] = cro.Name
	object.Spec.Template.Annotations[context.GetConfigurationHashAnnotationKey()] = cro.Status.Hash.Configuration
	object.Spec.Template.Annotations[context.GetServingCertHashAnnotationKey()] = cro.Status.Hash.ServingCert

	object.Annotations[context.GetConfigurationHashAnnotationKey()] = cro.Status.Hash.Configuration
	object.Annotations[context.GetServingCertHashAnnotationKey()] = context.GetBundle().Hash()

	context.ControllerSetter().Set(object, cro)

	// current, err = c.deployment.Ensure(object)
	return
}

func (c *deploymentHandler) EnsureRBAC(context *ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) error {
	list := c.asset.RBAC().New()
	for _, item := range list {
		context.ControllerSetter()(item.Object, in)

		current, err := c.dynamic.Ensure(item.Resource, item.Object)
		if err != nil {
			return fmt.Errorf("resource=%s failed to ensure RBAC - %s %v", item.Resource, err, item.Object)
		}

		klog.V(2).Infof("key=%s ensured RBAC resource %s", in.Name, current.GetName())
	}

	return nil
}
