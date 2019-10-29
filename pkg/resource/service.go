package resource

import (
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/dynamic"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewDynamicServiceClient(client dynamic.Client) *Service {
	return &Service{
		client: client,
	}
}

type Service struct {
	client dynamic.Client
}

func (s *Service) Ensure(service *corev1.Service) (current *corev1.Service, err error) {
	unstructured, errGot := s.client.Ensure(string(corev1.ResourceServices), service)
	if errGot != nil {
		err = errGot
		return
	}

	current = &corev1.Service{}
	if conversionErr:= runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.UnstructuredContent(), current); conversionErr != nil {
		err = conversionErr
		return
	}

	return
}
