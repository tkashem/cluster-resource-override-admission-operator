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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewDaemonSetHandler(o *Options) *daemonSetHandler {
	return &daemonSetHandler{
		client:     o.Client.Kubernetes,
		dynamic:    o.Client.Dynamic,
		deployment: ensurer.NewDaemonSetEnsurer(o.Client.Dynamic),
		asset:      asset.New(o.OperandContext),
		lister:     o.KubeLister,
	}
}

type daemonSetHandler struct {
	client     kubernetes.Interface
	deployment *ensurer.DaemonSetEnsurer
	dynamic    dynamicclient.Ensurer
	lister     *secondarywatch.Lister
	asset      *asset.Asset
}

type Deployer interface {
	Exists(namespace, name string) (object metav1.Object, err error)
}

func (c *daemonSetHandler) Handle(context *ReconcileRequestContext, original *autoscalingv1.ClusterResourceOverride) (current *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, handleErr error) {
	current = original
	ensure := false

	name := c.asset.Deployment().Name()
	object, getErr := c.lister.AppsV1DaemonSetLister().DaemonSets(context.WebhookNamespace()).Get(name)
	if getErr != nil && !k8serrors.IsNotFound(getErr) {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.InternalError, getErr)
		return
	}

	switch {
	case k8serrors.IsNotFound(getErr):
		ensure = true
	case object.GetAnnotations()[context.GetConfigurationHashAnnotationKey()] != current.Status.Hash.Configuration:
		klog.V(2).Infof("key=%s resource=%T/%s configuration hash mismatch", original.Name, object, object.Name)
		ensure = true
	case object.GetAnnotations()[context.GetServingCertHashAnnotationKey()] != current.Status.Hash.ServingCert:
		klog.V(2).Infof("key=%s resource=%T/%s serving cert hash mismatch", original.Name, object, object.Name)
		ensure = true
	}

	if ensure {
		cm, err := c.EnsureDeployment(context, original)
		if err != nil {
			handleErr = err
			return
		}

		object = cm
		klog.V(2).Infof("key=%s resource=%T/%s successfully ensured", original.Name, object, object.Name)
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

func (c *daemonSetHandler) EnsureDeployment(context *ReconcileRequestContext, cro *autoscalingv1.ClusterResourceOverride) (current *appsv1.DaemonSet, err error) {
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

	current, err = c.deployment.Ensure(object)
	return
}

func (c *daemonSetHandler) ApplyAnnotations(context *ReconcileRequestContext, cro *autoscalingv1.ClusterResourceOverride) func(object metav1.Object) {
	return func(object metav1.Object) {
		if len(object.GetAnnotations()) == 0 {
			object.SetAnnotations(map[string]string{})
		}

		object.GetAnnotations()[context.GetConfigurationHashAnnotationKey()] = cro.Status.Hash.Configuration
		object.GetAnnotations()[context.GetServingCertHashAnnotationKey()] = context.GetBundle().Hash()

		context.ControllerSetter().Set(object, cro)
	}
}

func (c *daemonSetHandler) ApplyPodTemplate(context *ReconcileRequestContext, cro *autoscalingv1.ClusterResourceOverride) func(object *corev1.PodTemplate) {
	return func(template *corev1.PodTemplate) {
		if len(template.Template.Annotations) == 0 {
			template.Template.Annotations = map[string]string{}
		}

		template.Template.Annotations[context.GetOwnerAnnotationKey()] = cro.Name
		template.Template.Annotations[context.GetConfigurationHashAnnotationKey()] = cro.Status.Hash.Configuration
		template.Template.Annotations[context.GetServingCertHashAnnotationKey()] = cro.Status.Hash.ServingCert
	}
}

func (c *daemonSetHandler) EnsureRBAC(context *ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) error {
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
