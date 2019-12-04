package resource

import (
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/dynamic"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type clusterrolebinding struct {
	client dynamic.Client
}

func (c *clusterrolebinding) Ensure(role *rbacv1.ClusterRoleBinding) (current *rbacv1.ClusterRoleBinding, err error) {
	unstructured, errGot := c.client.Ensure("clusterrolebindings", role)
	if errGot != nil {
		err = errGot
		return
	}

	current = &rbacv1.ClusterRoleBinding{}
	if conversionErr:= runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.UnstructuredContent(), current); conversionErr != nil {
		err = conversionErr
		return
	}

	return
}
