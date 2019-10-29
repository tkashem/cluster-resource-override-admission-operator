package asset

import (
	"fmt"
	operatorruntime "github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Asset) RBAC() *rbac {
	return &rbac{
		context: a.context,
	}
}

type RBACItem struct {
	Resource string
	Object   operatorruntime.Object
}

type rbac struct {
	context operatorruntime.OperandContext
}

func (s *rbac) New() []*RBACItem {
	var (
		apiVersion = "rbac.authorization.k8s.io/v1"

		thisOperatorServiceAccount = rbacv1.Subject{
			Namespace: s.context.WebhookNamespace(),
			Kind:      "ServiceAccount",
			Name:      s.context.WebhookName(),
		}

		defaultClusterRoleName = fmt.Sprintf("default-aggregated-apiserver-%s", s.context.WebhookName())

		commonLabels = map[string]string{
			s.context.WebhookName(): "true",
		}
	)

	return []*RBACItem{
		// service account
		{
			Resource: "serviceaccounts",
			Object: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      s.context.WebhookName(),
					Namespace: s.context.WebhookNamespace(),
				},
			},
		},

		// to read the config for terminating authentication
		{
			Resource: "rolebindings",
			Object: &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: apiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("extension-server-authentication-reader-%s", s.context.WebhookName()),
					Namespace: "kube-system",
					Labels:    commonLabels,
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     "c",
				},
				Subjects: []rbacv1.Subject{
					thisOperatorServiceAccount,
				},
			},
		},

		// to let aggregated apiservers create admission reviews
		{
			Resource: "clusterroles",
			Object: &rbacv1.ClusterRole{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRole",
					APIVersion: apiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   fmt.Sprintf("system:%s-requester", s.context.WebhookName()),
					Labels: commonLabels,
				},
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{
							"autoscaling.openshift.io",
						},
						Resources: []string{
							s.context.WebhookName(),
						},
						Verbs: []string{
							"create",
						},
					},
				},
			},
		},

		// this should be a default for an aggregated apiserver
		{
			Resource: "clusterroles",
			Object: &rbacv1.ClusterRole{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRole",
					APIVersion: apiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   defaultClusterRoleName,
					Labels: commonLabels,
				},
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{
							"admissionregistration.k8s.io",
						},
						Resources: []string{
							"validatingwebhookconfigurations",
							"mutatingwebhookconfigurations",
						},
						Verbs: []string{
							"get",
							"list",
							"watch",
						},
					},
					{
						APIGroups: []string{
							"",
						},
						Resources: []string{
							"namespaces",
						},
						Verbs: []string{
							"get",
							"list",
							"watch",
						},
					},
				},
			},
		},

		// this should be a default for an aggregated apiserver
		{
			Resource: "clusterrolebindings",
			Object: &rbacv1.ClusterRoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRoleBinding",
					APIVersion: apiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   defaultClusterRoleName,
					Labels: commonLabels,
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     defaultClusterRoleName,
				},
				Subjects: []rbacv1.Subject{
					thisOperatorServiceAccount,
				},
			},
		},

		// to delegate authentication and authorization.
		{
			Resource: "clusterrolebindings",
			Object: &rbacv1.ClusterRoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRoleBinding",
					APIVersion: apiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   fmt.Sprintf("auth-delegator-%s", s.context.WebhookName()),
					Labels: commonLabels,
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     "system:auth-delegator",
				},
				Subjects: []rbacv1.Subject{
					thisOperatorServiceAccount,
				},
			},
		},

		// so that daemonset pods can use hostnetwork
		{
			Resource: "roles",
			Object: &rbacv1.Role{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Role",
					APIVersion: apiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-scc-hostnetwork-use", s.context.WebhookName()),
					Namespace: s.context.WebhookNamespace(),
					Labels:    commonLabels,
				},
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{
							"security.openshift.io",
						},
						Resources: []string{
							"securitycontextconstraints",
						},
						Verbs: []string{
							"use",
						},
						ResourceNames: []string{
							"hostnetwork",
						},
					},
				},
			},
		},
		{
			Resource: "rolebindings",
			Object: &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: apiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-scc-hostnetwork-use", s.context.WebhookName()),
					Namespace: s.context.WebhookNamespace(),
					Labels:    commonLabels,
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     fmt.Sprintf("%s-scc-hostnetwork-use", s.context.WebhookName()),
				},
				Subjects: []rbacv1.Subject{
					thisOperatorServiceAccount,
				},
			},
		},

		// so that kube-apiserver can directly call the webhook server
		{
			Resource: "clusterroles",
			Object: &rbacv1.ClusterRole{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRole",
					APIVersion: apiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   fmt.Sprintf("%s-anonymous-access", s.context.WebhookName()),
					Labels: commonLabels,
				},
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{
							"admission.autoscaling.openshift.io",
						},
						Resources: []string{
							"clusterresourceoverrides",
						},
						Verbs: []string{
							"create",
						},
					},
				},
			},
		},
		{
			Resource: "clusterrolebindings",
			Object: &rbacv1.ClusterRoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRoleBinding",
					APIVersion: apiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   fmt.Sprintf("%s-anonymous-access", s.context.WebhookName()),
					Labels: commonLabels,
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     fmt.Sprintf("%s-anonymous-access", s.context.WebhookName()),
				},
				Subjects: []rbacv1.Subject{
					{
						APIGroup: "rbac.authorization.k8s.io",
						Kind:     "User",
						Name:     "system:anonymous",
					},
				},
			},
		},
	}
}
