package asset

import (
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (a *Asset) Deployment() *deployment {
	return &deployment{
		context: a.context,
		asset: a,
	}
}

type deployment struct {
	context runtime.OperandContext
	asset *Asset
}

func (d *deployment) Name() string {
	return d.context.WebhookName()
}

func (d *deployment) New() *appsv1.Deployment {
	var replicas int32 = 1

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: d.context.WebhookNamespace(),
			Name:      d.Name(),
			Labels: map[string]string{
				d.context.WebhookName(): "true",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					d.context.WebhookName(): "true",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: d.context.WebhookName(),
					Labels: map[string]string{
						d.context.WebhookName(): "true",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: d.context.WebhookName(),
					Containers: []corev1.Container{
						corev1.Container{
							Name:            d.context.WebhookName(),
							Image:           d.context.OperandImage(),
							ImagePullPolicy: corev1.PullAlways,
							Args: []string{
								"--secure-port=8443",
								"--audit-log-path=-",
								"--tls-cert-file=/var/serving-cert/tls.crt",
								"--tls-private-key-file=/var/serving-cert/tls.key",
								"--v=8",
							},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:      "CONFIGURATION_PATH",
									Value:     "/etc/clusterresourceoverride/config/override.yaml",
								},
							},
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 8443,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name: "serving-cert",
									MountPath: "/var/serving-cert",
								},
								corev1.VolumeMount{
									Name: "configuration",
									MountPath: "/etc/clusterresourceoverride/config/override.yaml",
									SubPath: d.asset.Configuration().Key(),
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:        "/healthz",
										Port:        intstr.FromInt(8443),
										Scheme:      corev1.URISchemeHTTPS,
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name:         "serving-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: d.asset.ServiceServingSecret().Name(),
									DefaultMode: func() *int32{
										v := int32(420)
										return &v
									}(),
								},
							},
						},

						corev1.Volume{
							Name:         "configuration",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: d.asset.Configuration().Name(),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
