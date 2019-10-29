package resource

import (
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/dynamic"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewMutatingWebhookConfigurationClient(client dynamic.Client) *MutatingWebhookConfiguration {
	return &MutatingWebhookConfiguration{
		client: client,
	}
}

type MutatingWebhookConfiguration struct {
	client dynamic.Client
}

func (m *MutatingWebhookConfiguration) Ensure(configuration *admissionregistrationv1beta1.MutatingWebhookConfiguration) (current *admissionregistrationv1beta1.MutatingWebhookConfiguration, err error) {
	unstructured, errGot := m.client.Ensure("mutatingwebhookconfigurations", configuration)
	if errGot != nil {
		err = errGot
		return
	}

	current = &admissionregistrationv1beta1.MutatingWebhookConfiguration{}
	if conversionErr:= runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.UnstructuredContent(), current); conversionErr != nil {
		err = conversionErr
		return
	}

	return
}