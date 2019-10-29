package resource

import (
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/dynamic"
	"k8s.io/apimachinery/pkg/runtime"

	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	//apiregistrationclientset "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	//k8serrors "k8s.io/apimachinery/pkg/api/errors"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewAPIServiceClient(client dynamic.Client) *APIService {
	return &APIService{
		client: client,
	}
}

type APIService struct {
	client dynamic.Client
	// registration apiregistrationclientset.Interface
}

//func (s *APIService) Ensure(apiservice *apiregistrationv1.APIService) (current *apiregistrationv1.APIService, err error) {
//	newObject, createErr := s.registration.ApiregistrationV1().APIServices().Create(apiservice)
//	if createErr == nil {
//		current = newObject
//		return
//	}
//
//	if !k8serrors.IsAlreadyExists(createErr) {
//		err = fmt.Errorf("failed to create APIService/%s - %s", apiservice.Name, createErr.Error())
//		return
//	}
//
//	original, getErr := s.registration.ApiregistrationV1().APIServices().Get(apiservice.Name, metav1.GetOptions{})
//	if getErr != nil {
//		err = fmt.Errorf("failed to retrieve APIService/%s - %s", apiservice.Name, getErr.Error())
//		return
//	}
//
//
//
//
//	return
//}

func (s *APIService) Ensure(apiservice *apiregistrationv1.APIService) (current *apiregistrationv1.APIService, err error) {
	unstructured, errGot := s.client.Ensure("apiservices", apiservice)
	if errGot != nil {
		err = errGot
		return
	}

	current = &apiregistrationv1.APIService{}
	if conversionErr:= runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.UnstructuredContent(), current); conversionErr != nil {
		err = conversionErr
		return
	}

	return
}