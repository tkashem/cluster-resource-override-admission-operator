package resourcewatch

import (
	listerscorev1 "k8s.io/client-go/listers/core/v1"
	listersappsv1 "k8s.io/client-go/listers/apps/v1"
	admissionregistrationv1beta1 "k8s.io/client-go/listers/admissionregistration/v1beta1"
)

type Lister struct {
	deployment listersappsv1.DeploymentLister
	pod listerscorev1.PodLister
	configmap listerscorev1.ConfigMapLister
	service listerscorev1.ServiceLister
	secret listerscorev1.SecretLister
	serviceaccount listerscorev1.ServiceAccountLister
	webhook admissionregistrationv1beta1.MutatingWebhookConfigurationLister
}

func (l *Lister) CoreV1ConfigMapLister() listerscorev1.ConfigMapLister {
	return l.configmap
}

func (l *Lister) CoreV1SecretLister() listerscorev1.SecretLister {
	return l.secret
}

func (l *Lister) CoreV1ServiceLister() listerscorev1.ServiceLister {
	return l.service
}

func (l *Lister) AppsV1DeploymentLister() listersappsv1.DeploymentLister {
	return l.deployment
}

func (l *Lister) AdmissionRegistrationV1beta1MutatingWebhookConfigurationLister() admissionregistrationv1beta1.MutatingWebhookConfigurationLister {
	return l.webhook
}
