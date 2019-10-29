package dynamic

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"testing"
)

func TestPatch(t *testing.T) {
	tests := []struct {
		name     string
		modified runtime.Object
		original runtime.Object
	}{
		{
			name: "test",
			modified: &apiregistrationv1.APIService{
				TypeMeta: metav1.TypeMeta{
					Kind:       "APIService",
					APIVersion: "apiregistration.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "v1.admission.autoscaling.openshift.io",
					Labels: map[string]string{
						"kube-aggregator.kubernetes.io/automanaged": "false",
					},
					ResourceVersion: "1234",
					UID:             types.UID("abc"),
				},
				Spec: apiregistrationv1.APIServiceSpec{
					Version:              "v1",
					Group:                "admission.autoscaling.openshift.io",
					GroupPriorityMinimum: 1000,
					VersionPriority:      15,
					Service: &apiregistrationv1.ServiceReference{
						Namespace: "test",
						Name:      "clusterresourceoverride",
					},

					// CABundle will be injected at runtime.
					CABundle: nil,
				},
			},

			original: &apiregistrationv1.APIService{
				TypeMeta: metav1.TypeMeta{
					Kind:       "APIService",
					APIVersion: "apiregistration.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "v1.admission.autoscaling.openshift.io",
					Labels: map[string]string{
						"kube-aggregator.kubernetes.io/automanaged":           "onstart",
						"kube-aggregator.kubernetes.io/should-not-be-changed": "true",
					},
				},
				Spec: apiregistrationv1.APIServiceSpec{
					Version:              "v1",
					Group:                "admission.autoscaling.openshift.io",
					GroupPriorityMinimum: 1000,
					VersionPriority:      15,
					Service: &apiregistrationv1.ServiceReference{
						Namespace: "test",
						Name:      "clusterresourceoverride",
					},

					// CABundle will be injected at runtime.
					CABundle: nil,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//modified, err := ToUnstructured(test.modified)
			//require.NoError(t, err)
			//
			//original, err := ToUnstructured(test.original)
			//require.NoError(t, err)

			bytes, patchErr := PatchWithRuntimeObject(test.original, test.modified, test.original)
			assert.NoError(t, patchErr)

			t.Logf("patch bytes - %s", string(bytes))
		})
	}
}
