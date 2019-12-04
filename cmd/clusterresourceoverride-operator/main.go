package main

import (
	goflag "flag"
	"fmt"
	"k8s.io/component-base/logs"
	"math/rand"
	"github.com/openshift/cluster-resource-override-admission-operator/pkg/cmd/operator"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"time"
	utilflag "k8s.io/component-base/cli/flag"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	logs.InitLogs()
	defer logs.FlushLogs()

	command := &cobra.Command{
		Use:   "clusterresourceoverride-operator",
		Short: "OpenShift ClusterResourceOverride Mutating Admission Webhook Operator",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	command.AddCommand(operator.NewStartCommand())

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
