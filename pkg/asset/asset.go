package asset

import (
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/runtime"
)

func New(context runtime.OperandContext) *Asset {
	return &Asset{
		context: context,
	}
}

type Asset struct {
	context runtime.OperandContext
}
