package asset

import (
	"fmt"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Asset) NewMutatingWebhookConfiguration() *mutatingWebhookConfiguration {
	return &mutatingWebhookConfiguration{
		context: a.context,
		asset:   a,
	}
}

const (
	plural = "clusterresourceoverrides"
)

type mutatingWebhookConfiguration struct {
	context runtime.OperandContext
	asset   *Asset
}

func (m *mutatingWebhookConfiguration) Name() string {
	return fmt.Sprintf("%s.%s", plural, apiServiceGroup)
}

func (m *mutatingWebhookConfiguration) New() *admissionregistrationv1beta1.MutatingWebhookConfiguration {
	url := fmt.Sprintf("https://localhost:9400/apis/%s/%s/%s", apiServiceGroup, apiServiceVersion, plural)
	policy := admissionregistrationv1beta1.Fail
	matchPolicy := admissionregistrationv1beta1.Equivalent
	namespaceMatchLabelKey := fmt.Sprintf("%s.%s/enabled", plural, apiServiceGroup)
	timeoutSeconds := int32(5)
	sideEffects := admissionregistrationv1beta1.SideEffectClassNone
	return &admissionregistrationv1beta1.MutatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MutatingWebhookConfiguration",
			APIVersion: "admissionregistration.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: m.Name(),
			Labels: map[string]string{
				"clusterresourceoverride": "true",
			},
		},
		Webhooks: []admissionregistrationv1beta1.MutatingWebhook{
			{
				Name: m.Name(),
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						namespaceMatchLabelKey: "true",
					},
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "runlevel",
							Operator: metav1.LabelSelectorOpNotIn,
							Values: []string{
								"0",
								"1",
							},
						},
					},
				},
				MatchPolicy: &matchPolicy,
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					// CABundle will be injected at runtime
					CABundle: nil,
					URL:      &url,
				},
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1beta1.OperationType{
							admissionregistrationv1beta1.Create,
							admissionregistrationv1beta1.Update,
						},

						Rule: admissionregistrationv1beta1.Rule{
							APIGroups: []string{
								"",
							},
							APIVersions: []string{
								"v1",
							},
							Resources: []string{
								"pods",
							},
						},
					},
				},
				FailurePolicy:  &policy,
				TimeoutSeconds: &timeoutSeconds,
				SideEffects:    &sideEffects,
			},
		},
	}
}
