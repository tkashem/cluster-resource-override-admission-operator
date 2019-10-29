package dynamic

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	unstructuredv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

type Client interface {
	Ensure(resource string, object runtime.Object) (current *unstructuredv1.Unstructured, err error)
}

func ToUnstructured(object runtime.Object) (unstructured *unstructuredv1.Unstructured, err error) {
	raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return
	}

	unstructured = &unstructuredv1.Unstructured{
		Object: raw,
	}
	return
}


func GetGVR(resource string, unstructured *unstructuredv1.Unstructured) schema.GroupVersionResource {
	gvk := unstructured.GetObjectKind().GroupVersionKind()

	return 	schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: resource,
	}
}

// Patch generates a patch.
func Patch(original, modified *unstructured.Unstructured, datastruct interface{}) (bytes []byte, err error) {
	originalData, err := original.MarshalJSON()
	if err != nil {
		return
	}

	modifiedData, err := modified.MarshalJSON()
	if err != nil {
		return
	}

	return strategicpatch.CreateTwoWayMergePatch(originalData, modifiedData, datastruct)
}

func PatchNew(original, modified runtime.Object, datastruct interface{}) (bytes []byte, err error) {
	originalData, err := json.Marshal(original)
	if err != nil {
		return
	}

	modifiedData, err := json.Marshal(modified)
	if err != nil {
		return
	}

	return strategicpatch.CreateTwoWayMergePatch(originalData, modifiedData, datastruct)
}
