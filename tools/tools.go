package tools

// This package contains import references to packages required only for the
// build process.
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

import (
	"github.com/kevinburke/go-bindata/go-bindata"
	"github.com/securego/gosec/cmd/gosec"
	"sigs.k8s.io/controller-tools/cmd/controller-gen"
	"k8s.io/code-generator"

	_ "github.com/openshift/library-go/alpha-build-machinery"
)
