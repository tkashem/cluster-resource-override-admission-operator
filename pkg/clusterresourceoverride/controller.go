package clusterresourceoverride

import (
	"errors"
	"fmt"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resourcewatch"
	"k8s.io/apimachinery/pkg/types"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	controllerreconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"

	autoscalingv1 "github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/controller"
	autoscalingv1listers "github.com/openshift/cluster-resource-override-admission-operator/pkg/generated/listers/autoscaling/v1"
	listers "github.com/openshift/cluster-resource-override-admission-operator/pkg/generated/listers/autoscaling/v1"

	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/reconciler"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride/internal/handlers"
	operatorruntime "github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
)

type Options struct {
	ResyncPeriod   time.Duration
	Workers        int
	Client         *operatorruntime.Client
	RuntimeContext operatorruntime.OperandContext
	Lister	*resourcewatch.Lister
}

func New(options *Options) (c controller.Interface, enqueuer operatorruntime.Enqueuer, err error) {
	if options == nil || options.Client == nil || options.RuntimeContext == nil {
		err = errors.New("invalid input to controller.New")
		return
	}

	// Create a new ClusterResourceOverrides watcher
	client := options.Client.Operator
	watcher := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return client.AutoscalingV1().ClusterResourceOverrides().List(options)
		},

		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return client.AutoscalingV1().ClusterResourceOverrides().Watch(options)
		},
	}

	// We need a queue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the work queue to a cache with the help of an informer. This way we
	// make sure that whenever the cache is updated, the clusterserviceversion
	// key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might
	// see a newer version of the ClusterServiceVersion than the version which
	// was responsible for triggering the update.
	indexer, informer := cache.NewIndexerInformer(watcher, &autoscalingv1.ClusterResourceOverride{}, options.ResyncPeriod,
		controller.NewEventHandler(queue), cache.Indexers{})

	lister := listers.NewClusterResourceOverrideLister(indexer)

	reconciler := reconciler.NewReconciler(&handlers.Options{
		OperandContext: options.RuntimeContext,
		Client:         options.Client,
		CROLister:      lister,
		KubeLister:     options.Lister,
	})

	instance := &clusterResourceOverrideController{
		workers:    options.Workers,
		queue:      queue,
		informer:   informer,
		reconciler: reconciler,
		lister: lister,
	}

	c = instance
	enqueuer = instance
	return
}

type clusterResourceOverrideController struct {
	workers    int
	queue      workqueue.RateLimitingInterface
	informer   cache.Controller
	reconciler controllerreconciler.Reconciler
	lister 		autoscalingv1listers.ClusterResourceOverrideLister
}

func (c *clusterResourceOverrideController) Name() string {
	return "clusterresourceoverride"
}

func (c *clusterResourceOverrideController) WorkerCount() int {
	return c.workers
}

func (c *clusterResourceOverrideController) Queue() workqueue.RateLimitingInterface {
	return c.queue
}

func (c *clusterResourceOverrideController) Informer() cache.Controller {
	return c.informer
}

func (c *clusterResourceOverrideController) Reconciler() controllerreconciler.Reconciler {
	return c.reconciler
}

func (c *clusterResourceOverrideController) Enqueue(obj interface{}) error {
	metaObj, err := operatorruntime.GetMetaObject(obj)
	if err != nil {
		return err
	}

	ownerName := getOwnerName(metaObj)
	if ownerName == "" {
		return fmt.Errorf("could not find owner for %s/%s", metaObj.GetNamespace(), metaObj.GetName())
	}

	cro, err := c.lister.Get(ownerName)
	if err != nil {
		return fmt.Errorf("ignoring request to enqueue - %s", err.Error())
	}

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: cro.GetNamespace(),
			Name:      cro.GetName(),
		},
	}

	c.queue.Add(request)
	return nil
}

func getOwnerName(object metav1.Object) string {
	const (
		// TODO: use the function that returns this value.
		key = "clusterresourceoverride.operator.autoscaling.openshift.io/owner"
	)



	// We check for annotations and owner references
	// If both exist, owner references takes precedence.
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil && ownerRef.Kind == autoscalingv1.ClusterResourceOverrideKind {
		return ownerRef.Name
	}

	annotations := object.GetAnnotations()
	if len(annotations) > 0 {
		owner, ok := annotations[key]
		if ok && owner != "" {
			return owner
		}
	}

	return ""
}

