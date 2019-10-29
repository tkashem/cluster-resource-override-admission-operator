package controller

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Options struct {
	Config 			*rest.Config
	ResyncPeriod   	time.Duration
}

type Interface interface {
	Name() string
	WorkerCount() int
	Queue() workqueue.RateLimitingInterface
	Informer() cache.Controller
	Reconciler() reconcile.Reconciler
}
