package handlers

import (
	"fmt"
	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/cert"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resourcewatch"
)

func NewCertReadyHandler(o *Options) *certReady {
	return &certReady{
		client:  o.Client.Kubernetes,
		lister: o.KubeLister,
	}
}

type certReady struct {
	client  kubernetes.Interface
	lister *resourcewatch.Lister
}

func (c *certReady) Handle(context *ReconcileRequestContext, in *autoscalingv1.ClusterResourceOverride) (out *autoscalingv1.ClusterResourceOverride, result controllerreconciler.Result, stop bool, handleErr error) {
	out = in
	resources := in.Status.Resources

	secret, err := c.lister.CoreV1SecretLister().Secrets(context.WebhookNamespace()).Get(resources.ServiceCertSecretRef.Name)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
		return
	}

	configmap, err := c.lister.CoreV1ConfigMapLister().ConfigMaps(context.WebhookNamespace()).Get(resources.ServiceCAConfigMapRef.Name)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
		return
	}

	clientCA, err := cert.GetClientCA(c.client)
	if err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, err)
		return
	}

	servingCertCA := []byte(configmap.Data["service-ca.crt"])
	bundle := &cert.Bundle{
		Serving: cert.Serving{
			ServiceKey:  secret.Data["tls.key"],
			ServiceCert: secret.Data["tls.crt"],
			ServingCertCA: servingCertCA,
		},
		KubeClintCA: clientCA,
	}

	if err := bundle.Validate(); err != nil {
		handleErr = condition.NewInstallReadinessError(autoscalingv1.CertNotAvailable, fmt.Errorf("certs not populated - %s", err.Error()))
		return
	}

	context.SetBundle(bundle)
	out.Status.Hash.ServingCert = bundle.Serving.Hash()

	klog.V(2).Infof("key=%s cert check passed", in.Name)
	return
}
