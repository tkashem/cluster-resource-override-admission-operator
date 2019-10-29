package asset

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Asset) NewServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind: "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: a.context.WebhookName(),
			Namespace: a.context.WebhookNamespace(),
		},
	}
}

//func (s *sa) Ensure(context Context, request *corev1.ServiceAccount) (object *corev1.ServiceAccount, err error) {
//	object, createErr := s.client.CoreV1().ServiceAccounts(context.Namespace()).Create(request)
//	if err == nil {
//		return
//	}
//
//	if !k8serrors.IsAlreadyExists(createErr) {
//		err = fmt.Errorf("failed to create ServiceAccount - %s", createErr.Error())
//		return
//	}
//
//	original, getErr := s.client.CoreV1().ServiceAccounts(context.Namespace()).Get(request.GetName(), metav1.GetOptions{} )
//	if getErr != nil {
//		err = fmt.Errorf("failed to retrieve ServiceAccount - %s", getErr.Error())
//		return
//	}
//
//	patchBytes, patchErr := Patch(original, request)
//	if patchErr != nil {
//		err = fmt.Errorf("failed to generate patch for ServiceAccount - %s", patchErr.Error())
//		return
//	}
//
//	return s.client.CoreV1().ServiceAccounts(context.Namespace()).Patch(request.GetName(), types.StrategicMergePatchType, patchBytes)
//}