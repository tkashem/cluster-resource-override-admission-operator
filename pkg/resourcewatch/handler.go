package resourcewatch

import (
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	"github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
)

func newResourceEventHandler(enqueuer runtime.Enqueuer) *resourceEventHandler {
	return &resourceEventHandler{
		enqueuer: enqueuer,
	}
}

var _ cache.ResourceEventHandler = &resourceEventHandler{}

type resourceEventHandler struct {
	enqueuer runtime.Enqueuer
}

func (r *resourceEventHandler) OnAdd(obj interface{}) {
	metaObj, err := runtime.GetMetaObject(obj)
	if err != nil {
		klog.Errorf("resourcewatch.OnAdd: invalid object, type=%T", obj)
		return
	}

	if err := r.enqueuer.Enqueue(metaObj); err != nil {
		klog.V(3).Infof("resourcewatch.OnAdd: failed to enqueue owner - %s", err.Error())
	}
}

// OnUpdate creates UpdateEvent and calls Update on EventHandler
func (r *resourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldMetaObj, err := runtime.GetMetaObject(oldObj)
	if err != nil {
		klog.Errorf("resourcewatch.OnUpdate: invalid object, type=%T", oldObj)
		return
	}

	newMetaObj, err := runtime.GetMetaObject(newObj)
	if err != nil {
		klog.Errorf("resourcewatch.OnUpdate: invalid object, type=%T", newObj)
		return
	}

	if oldMetaObj.GetResourceVersion() == newMetaObj.GetResourceVersion() {
		klog.V(3).Infof("resourcewatch.OnUpdate: resource version has not changed, not going to enqueue owner, type=%T resource-version=%s", newMetaObj, newMetaObj.GetResourceVersion())
		return
	}

	if err := r.enqueuer.Enqueue(newMetaObj); err != nil {
		klog.V(3).Infof("resourcewatch.OnUpdate: failed to enqueue owner - %s", err.Error())
	}
}

func (r *resourceEventHandler) OnDelete(obj interface{}) {
	metaObj, err := runtime.GetMetaObject(obj)
	if err != nil {
		klog.Errorf("resourcewatch.OnDelete: invalid object, type=%T", obj)
		return
	}

	if err := r.enqueuer.Enqueue(metaObj); err != nil {
		klog.V(3).Infof("resourcewatch.OnDelete: failed to enqueue owner - %s", err.Error())
	}
}