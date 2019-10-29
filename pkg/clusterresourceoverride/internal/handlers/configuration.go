package handlers

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/reference"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/asset"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resource"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resourcewatch"
)

func NewConfigurationHandler(o *Options) *configuration {
	return &configuration{
		client:  o.Client.Kubernetes,
		ensurer: resource.NewConfigMapClient(o.Client.Dynamic),
		lister: o.KubeLister,
		asset:   asset.New(o.OperandContext),
	}
}

type configuration struct {
	client  kubernetes.Interface
	ensurer *resource.ConfigMapEnsurer
	asset   *asset.Asset
	lister *resourcewatch.Lister
}

func (c *configuration) Handle(context *ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) (out *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, stop bool, handleErr error) {
	out = in

	desired, err := c.NewConfiguration(context, in)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.ConfigurationCheckFailed, err)
		return
	}

	name := c.asset.Configuration().Name()
	current, err := c.lister.CoreV1ConfigMapLister().ConfigMaps(context.WebhookNamespace()).Get(name)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.InternalError, err)
			return
		}

		cm, err := c.ensurer.Ensure(desired)
		if err != nil {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.InternalError, err)
			return
		}

		current = cm
		klog.V(2).Infof("key=%s resource=%T/%s successfully created", in.Name, current, current.Name)
	}

	equal := false
	hash := in.Spec.Config.Hash()
	if hash == out.Status.Hash.Configuration {
		equal = true
	}

	if ref := out.Status.Resources.ConfigurationRef; equal && ref != nil && ref.ResourceVersion == current.ResourceVersion {
		klog.V(2).Infof("key=%s resource=%T/%s is in sync", in.Name, current, current.Name)
		return
	}

	if !equal {
		klog.V(2).Infof("key=%s resource=%T/%s configuration has drifted", in.Name, current, current.Name)

		cm, err := c.ensurer.Ensure(desired)
		if err != nil {
			handleErr = condition.NewInstallReadinessError(autoscalingv1.ConfigurationCheckFailed, err)
			return
		}

		current = cm
	}

	newRef, err := reference.GetReference(current)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.CannotSetReference, err)
		return
	}

	out.Status.Hash.Configuration = hash
	out.Status.Resources.ConfigurationRef = newRef

	klog.V(2).Infof("key=%s resource=%T/%s resource-version=%s setting object reference", in.Name, current, current.Name, newRef.ResourceVersion)
	return
}

func (c *configuration) NewConfiguration(context *ReconcileRequestContext, override *autoscalingv1.ClusterResourceOverride) (configuration *corev1.ConfigMap, err error) {
	bytes, err := yaml.Marshal(override.Spec.Config)
	if err != nil {
		return
	}

	configuration = c.asset.Configuration().New()

	// Set owner reference.
	context.ControllerSetter().Set(configuration, override)

	if len(configuration.Data) == 0 {
		configuration.Data = map[string]string{}
	}
	configuration.Data[c.asset.Configuration().Key()] = string(bytes)

	return
}

func (c *configuration) IsConfigurationEqual(current *corev1.ConfigMap, this *autoscalingv1.ClusterResourceOverrideConfig) (equal bool, err error) {
	observed := current.Data[c.asset.Configuration().Key()]

	other := &autoscalingv1.ClusterResourceOverrideConfig{}
	err = yaml.Unmarshal([]byte(observed), other)
	if err != nil {
		return
	}

	equal = equality.Semantic.DeepEqual(this, other)
	return
}
