package e2e

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"

	"github.com/openshift/cluster-resource-override-admission-operator/pkg/dynamic"
	"github.com/stretchr/testify/require"
)

func TestDynamicClient(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		original runtime.Object
		object   runtime.Object
	}{
		{
			name:     "WithNamespacedObjectCreation",
			resource: "serviceaccounts",
			original: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: options.namespace,
					Name:      "test-sa",
					Annotations: map[string]string{
						"should-not-be-deleted": "true",
					},
				},
			},
			object: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: options.namespace,
					Name:      "test-sa",
					Annotations: map[string]string{
						"should-be-added": "true",
					},
				},
			},
		},
	}

	dynamic, err := dynamic.NewForConfig(options.config)
	require.NoError(t, err)
	require.NotNil(t, dynamic)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.original != nil {
				// create the object first.
				// objectGot, errGot := dynamic.Ensure(test.resource, test.original)
				service, ok := test.original.(*corev1.ServiceAccount)
				require.True(t, ok)

				objectGot, errGot := options.client.CoreV1().ServiceAccounts(options.namespace).Create(service)
				require.NoError(t, errGot)
				require.NotNil(t, objectGot)

				defer func() {
					// options.client.CoreV1().ServiceAccounts(options.namespace).Delete(service.Name, &metav1.DeleteOptions{})
				}()
			}

			objectGot, errGot := dynamic.Ensure(test.resource, test.object)
			require.NoError(t, errGot)
			require.NotNil(t, objectGot)

			current := &corev1.ServiceAccount{}
			conversionErr := runtime.DefaultUnstructuredConverter.FromUnstructured(objectGot.UnstructuredContent(), current)
			require.NoError(t, conversionErr)

			assert.Equal(t, objectGot.GetResourceVersion(), current.GetResourceVersion())
		})
	}
}
