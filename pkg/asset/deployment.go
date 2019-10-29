package asset

import (
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Asset) Deployment() *deployment {
	return &deployment{
		context: a.context,
		asset:   a,
	}
}

type deployment struct {
	context runtime.OperandContext
	asset   *Asset
}

func (d *deployment) Name() string {
	return d.context.WebhookName()
}

func (d *deployment) New() *appsv1.DaemonSet {
	tolerationSeconds := int64(120)
	return &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: d.context.WebhookNamespace(),
			Name:      d.Name(),
			Labels: map[string]string{
				d.context.WebhookName(): "true",
			},
		},
		Spec: appsv1.DaemonSetSpec{
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
					HostNetwork: true,
					NodeSelector: map[string]string{
						"node-role.kubernetes.io/master": "",
					},
					ServiceAccountName: d.context.WebhookName(),
					Containers: []corev1.Container{
						{
							Name:            d.context.WebhookName(),
							Image:           d.context.OperandImage(),
							ImagePullPolicy: corev1.PullAlways,
							Args: []string{
								"--secure-port=9400",
								"--bind-address=127.0.0.1",
								"--audit-log-path=-",
								"--tls-cert-file=/var/serving-cert/tls.crt",
								"--tls-private-key-file=/var/serving-cert/tls.key",
								"--v=8",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "CONFIGURATION_PATH",
									Value: "/etc/clusterresourceoverride/config/override.yaml",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 9400,
									HostPort:      9400,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "serving-cert",
									MountPath: "/var/serving-cert",
								},
								{
									Name:      "configuration",
									MountPath: "/etc/clusterresourceoverride/config/override.yaml",
									SubPath:   d.asset.Configuration().Key(),
								},
							},
							// with hostnetwork:true probe fails
							//ReadinessProbe: &corev1.Probe{
							//	Handler: corev1.Handler{
							//		HTTPGet: &corev1.HTTPGetAction{
							//			Path:   "/healthz",
							//			Port:   intstr.FromInt(9400),
							//			Scheme: corev1.URISchemeHTTPS,
							//		},
							//	},
							//},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "serving-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: d.asset.ServiceServingSecret().Name(),
									DefaultMode: func() *int32 {
										v := int32(420)
										return &v
									}(),
								},
							},
						},

						{
							Name: "configuration",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: d.asset.Configuration().Name(),
									},
								},
							},
						},
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "node-role.kubernetes.io/master",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
						{
							Key:               "node.kubernetes.io/unreachable",
							Operator:          corev1.TolerationOpExists,
							Effect:            corev1.TaintEffectNoExecute,
							TolerationSeconds: &tolerationSeconds,
						},
						{
							Key:               "node.kubernetes.io/not-ready",
							Operator:          corev1.TolerationOpExists,
							Effect:            corev1.TaintEffectNoExecute,
							TolerationSeconds: &tolerationSeconds,
						},
					},
				},
			},
		},
	}
}
