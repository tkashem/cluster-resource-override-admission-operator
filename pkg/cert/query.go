package cert

import (
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/api/core/v1"
)

type QueryFunc func() (bundle *Bundle, err error)

func NewOpenShiftQuerier(client  kubernetes.Interface, caConfigMap, serviceSecret *corev1.ObjectReference) QueryFunc {
	querier := &openshiftQuerier{
		client: client,
	}

	return func() (bundle *Bundle, err error) {
		return querier.Query(caConfigMap, serviceSecret)
	}
}

type openshiftQuerier struct {
	client  kubernetes.Interface
}

func (q *openshiftQuerier) Query(caConfigMap, serviceSecret *corev1.ObjectReference ) (bundle *Bundle, err error) {
	return
}