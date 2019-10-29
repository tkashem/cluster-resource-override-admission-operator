package asset

import (
	"fmt"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

const (
	apiServiceVersion        = "v1"
	apiServiceGroup          = "admission.autoscaling.openshift.io"
	AutoRegisterManagedLabel = "kube-aggregator.kubernetes.io/automanaged"
)

func (a *Asset) APIService() *apiservice {
	return &apiservice{
		context: a.context,
		asset:   a,
	}
}

type apiservice struct {
	context runtime.OperandContext
	asset   *Asset
}

func (a *apiservice) Name() string {
	return fmt.Sprintf("%s.%s", apiServiceVersion, apiServiceGroup)
}

func (a *apiservice) New() *apiregistrationv1.APIService {
	return &apiregistrationv1.APIService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIService",
			APIVersion: "apiregistration.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: a.Name(),
			Labels: map[string]string{
				AutoRegisterManagedLabel: "onstart",
			},
		},
		Spec: apiregistrationv1.APIServiceSpec{
			Version:              apiServiceVersion,
			Group:                apiServiceGroup,
			GroupPriorityMinimum: 1000,
			VersionPriority:      15,
			Service: &apiregistrationv1.ServiceReference{
				Namespace: a.context.WebhookNamespace(),
				Name:      a.context.WebhookName(),
			},

			// CABundle will be injected at runtime.
			CABundle: nil,
		},
	}
}
