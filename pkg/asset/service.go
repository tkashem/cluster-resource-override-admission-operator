package asset

import (
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ServingCertSecretAnnotationName = "service.alpha.openshift.io/serving-cert-secret-name"
)

func (a *Asset) Service() *service {
	return &service{
		asset: a,
		context: a.context,
	}
}

type service struct {
	context runtime.OperandContext
	asset *Asset
}

func (s *service) Name() string {
	return s.context.WebhookName()
}

func (s *service) New() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: s.context.WebhookName(),
			Namespace: s.context.WebhookNamespace(),
			Labels: map[string]string{
				s.context.WebhookName(): "true",
			},
			Annotations: map[string]string{
				ServingCertSecretAnnotationName: s.asset.ServiceServingSecret().Name(),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				s.context.WebhookName(): "true",
			},
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port:       443,
					TargetPort: intstr.FromInt(8443),
				},
			},
		},
	}
}