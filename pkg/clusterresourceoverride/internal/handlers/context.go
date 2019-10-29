package handlers

import (
	"fmt"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/apis/autoscaling"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/cert"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/resourcewatch"
	operatorruntime "github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
	autoscalingv1listers "github.com/openshift/cluster-resource-override-admission-operator/pkg/generated/listers/autoscaling/v1"
)

func NewReconcileRequestContext(oc operatorruntime.OperandContext) *ReconcileRequestContext {
	return &ReconcileRequestContext{
		OperandContext: oc,
		RequestContext:    &requestContext{},
	}
}

type Options struct {
	OperandContext operatorruntime.OperandContext
	Client *operatorruntime.Client
	CROLister autoscalingv1listers.ClusterResourceOverrideLister
	KubeLister *resourcewatch.Lister
}

type ReconcileRequestContext struct {
	operatorruntime.OperandContext
	RequestContext
}

// TODO: remove this interface
type RequestContext interface {
	SetBundle(*cert.Bundle)
	GetBundle() *cert.Bundle

	// TODO: remove, use operatorruntime.SetController directly?
	ControllerSetter() operatorruntime.SetControllerFunc
}

type requestContext struct {
	configurationHash string
	bundle *cert.Bundle
}

func (r *requestContext) SetBundle(bundle *cert.Bundle) {
	r.bundle = bundle
}

func (r *requestContext) GetBundle() *cert.Bundle {
	return r.bundle
}

func (r *requestContext) ControllerSetter() operatorruntime.SetControllerFunc {
	return operatorruntime.SetController
}

func (r *ReconcileRequestContext) GetConfigurationHashAnnotationKey() string {
	return fmt.Sprintf("%s.%s/configuration.hash", r.WebhookName(), autoscaling.GroupName)
}

func (r *ReconcileRequestContext) GetServingCertHashAnnotationKey() string {
	return fmt.Sprintf("%s.%s/servingcert.hash", r.WebhookName(), autoscaling.GroupName)
}

func (r *ReconcileRequestContext) GetOwnerAnnotationKey() string {
	return fmt.Sprintf("%s.%s/owner", r.WebhookName(), autoscaling.GroupName)
}
