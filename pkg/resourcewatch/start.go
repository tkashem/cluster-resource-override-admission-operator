package resourcewatch

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/client-go/informers"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
	"time"
)

type Options struct {
	Client *runtime.Client
	ResyncPeriod   time.Duration
	Namespace string
}

type StarterFunc func(enqueuer runtime.Enqueuer, shutdown context.Context) error

func (s StarterFunc) Start(enqueuer runtime.Enqueuer, shutdown context.Context) error {
	return s(enqueuer, shutdown)
}

func New(options *Options) (lister *Lister, startFunc StarterFunc) {
	option := informers.WithNamespace(options.Namespace)
	factory := informers.NewSharedInformerFactoryWithOptions(options.Client.Kubernetes, options.ResyncPeriod, option)

	deployment := factory.Apps().V1().Deployments()
	pod := factory.Core().V1().Pods()
	configmap := factory.Core().V1().ConfigMaps()
	service := factory.Core().V1().Services()
	secret := factory.Core().V1().Secrets()
	serviceaccount := factory.Core().V1().ServiceAccounts()
	webhook := factory.Admissionregistration().V1beta1().MutatingWebhookConfigurations()

	startFunc = func(enqueuer runtime.Enqueuer, shutdown context.Context) error {
		handler := newResourceEventHandler(enqueuer)

		deployment.Informer().AddEventHandler(handler)
		pod.Informer().AddEventHandler(handler)
		configmap.Informer().AddEventHandler(handler)
		service.Informer().AddEventHandler(handler)
		secret.Informer().AddEventHandler(handler)
		serviceaccount.Informer().AddEventHandler(handler)
		webhook.Informer().AddEventHandler(handler)

		factory.Start(shutdown.Done())
		status := factory.WaitForCacheSync(shutdown.Done())
		if names := check(status); len(names) > 0 {
			return fmt.Errorf("WaitForCacheSync did not successfully complete resources=%s", names)
		}

		return nil
	}

	lister = &Lister{
		deployment: deployment.Lister(),
		pod: pod.Lister(),
		configmap: configmap.Lister(),
		service: service.Lister(),
		secret: secret.Lister(),
		serviceaccount: serviceaccount.Lister(),
		webhook: webhook.Lister(),
	}

	return
}

func check(status map[reflect.Type]bool) []string {
	names := make([]string, 0)

	for objType, synced := range status {
		if !synced {
			names = append(names,objType.Name())
		}
	}

	return names
}