package asset

import (
	"fmt"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Asset) CABundleConfigMap() *caBundleConfigMap {
	return &caBundleConfigMap{
		context: a.context,
	}
}

var (
	labels = map[string]string{
		autoscaling.GroupName: "true",
	}

	annotations = map[string]string{
		"service-serving": "true",
	}
)

type caBundleConfigMap struct {
	context runtime.OperandContext
}

func (c *caBundleConfigMap) LabelsAnnotations() (labels, annotations map[string]string) {
	return labels, annotations
}

func (c *caBundleConfigMap) Name() string {
	return fmt.Sprintf("%s-service-serving", c.context.WebhookName())
}

func (c *caBundleConfigMap) New() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.Name(),
			Namespace:   c.context.WebhookNamespace(),
			Labels:      labels,
			Annotations: annotations,
		},
	}
}
