package operator

import (
	"fmt"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resourcewatch"
	"k8s.io/klog"
	"net/http"
	"time"

	"github.com/openshift/cluster-resource-override-admission-operator/pkg/clusterresourceoverride"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/controller"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
)

func NewRunner() Interface {
	return &runner{
		done: make(chan struct{}, 0),
	}
}

type runner struct {
	done chan struct{}
}

func (r *runner) Run(config *Config, errorCh chan<- error) {
	defer func() {
		close(r.done)
		klog.V(1).Infof("controller(s) are done, operator is exiting")
	}()

	clients, err := runtime.NewClient(config.RestConfig)
	if err != nil {
		errorCh <- err
		return
	}

	context := runtime.NewOperandContext(config.Name, config.Namespace, "cluster", config.OperandImage, config.OperandVersion)

	// create lister(s) for secondary resources
	lister, starter := resourcewatch.New(&resourcewatch.Options{
		Client:       clients,
		ResyncPeriod: 60 * time.Minute,
		Namespace:    config.Namespace,
	})

	// start the controllers
	c, enqueuer, err := clusterresourceoverride.New(&clusterresourceoverride.Options{
		ResyncPeriod: 60 * time.Minute,
		Workers:      1,
		RuntimeContext: context,
		Client: clients,
		Lister: lister,
	})
	if err != nil {
		errorCh <- fmt.Errorf("failed to create controller - %s", err.Error())
		return
	}

	// Setup watches for secondary resources.
	if err := starter.Start(enqueuer, config.ShutdownContext); err != nil {
		errorCh <- fmt.Errorf("failed to start watch on secondary resources - %s", err.Error())
		return
	}

	runner := controller.NewRunner()
	runnerErrorCh := make(chan error, 0)
	go runner.Run(config.ShutdownContext, c, runnerErrorCh)
	if err := <-runnerErrorCh; err != nil {
		errorCh <- err
		return
	}

	// Serve a simple HTTP health check.
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	go http.ListenAndServe(":8080", healthMux)

	errorCh <- nil

	klog.V(1).Infof("operator is waiting for controller(s) to be done")
	<-runner.Done()
}

func (r *runner) Done() <-chan struct{} {
	return r.done
}