module github.com/openshift/cluster-resource-override-admission-operator

go 1.12

require (
	github.com/kevinburke/go-bindata v3.16.0+incompatible
	github.com/openshift/library-go v0.0.0-20191118102510-4e2c7112d252
	github.com/prometheus/client_golang v1.0.0
	github.com/securego/gosec v0.0.0-20191008095658-28c1128b7336
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.0.0-20191109101512-6d4d1612ba53
	k8s.io/apiextensions-apiserver v0.0.0-20191109110701-3fdecfd8e730
	k8s.io/apimachinery v0.0.0-20191109100837-dffb012825f2
	k8s.io/apiserver v0.0.0-20191109104008-f2672160bdbe
	k8s.io/client-go v0.0.0-20191109102209-3c0d1af94be5
	k8s.io/code-generator v0.0.0-20191109100332-a9a0d9c0b3aa
	k8s.io/component-base v0.0.0-20191109103431-7fd2da093d6d
	k8s.io/klog v1.0.0
	k8s.io/kube-aggregator v0.0.0-20191109104959-a1b02ed9435a
	sigs.k8s.io/controller-runtime v0.3.0
	sigs.k8s.io/controller-tools v0.2.2
	sigs.k8s.io/yaml v1.1.0
)
