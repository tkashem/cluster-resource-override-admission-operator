package reconciler

import (
	"reflect"

	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/generated/clientset/versioned"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/condition"
)

type Updater struct {
	client versioned.Interface
}

func (u *Updater) Update(observed, desired *autoscalingv1.ClusterResourceOverride) error {
	o := condition.DeepCopyWithDefaultLastTransitionTime(&observed.Status)
	d := condition.DeepCopyWithDefaultLastTransitionTime(&desired.Status)

	if reflect.DeepEqual(o, d) {
		return nil
	}

	_, err := u.client.AutoscalingV1().ClusterResourceOverrides().UpdateStatus(desired)
	return err
}
