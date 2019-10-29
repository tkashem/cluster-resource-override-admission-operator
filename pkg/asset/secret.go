package asset

import (
	"fmt"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Asset) ServiceServingSecret() *serviceServingSecret {
	return &serviceServingSecret{
		context: a.context,
	}
}

const (
	SecretNamePrefix = "server-serving-cert"
)

type serviceServingSecret struct {
	context runtime.OperandContext
}

func (s *serviceServingSecret) Name() string {
	return fmt.Sprintf("%s-%s", SecretNamePrefix, s.context.WebhookName())
}

func (s *serviceServingSecret) New() *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.context.WebhookNamespace(),
			Name:      s.Name(),
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": nil,
			"tls.key": nil,
		},
	}
}
