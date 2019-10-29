package asset

import (
	"fmt"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Asset) Configuration() *configuration {
	return &configuration{
		context: a.context,
	}
}

type configuration struct {
	context runtime.OperandContext
}

func (c *configuration) Key() string {
	return "configuration.yaml"
}

func (c *configuration) Name() string {
	return fmt.Sprintf("%s-configuration", c.context.WebhookName())
}

func (c *configuration) New() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name(),
			Namespace: c.context.WebhookNamespace(),
			Labels: map[string]string{
				autoscaling.GroupName: "true",
			},
		},
		Data: map[string]string{
			// The configuration will get injected from the `spec` of the Custom Resource
			// So we are leaving it empty.
			c.Key() : "",
		},
	}
}