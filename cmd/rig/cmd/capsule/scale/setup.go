package scale

import (
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	disable             bool
	overwriteAutoscaler bool
	forceDeploy         bool
)

var (
	requestCPU    string
	requestMemory string
	limitCPU      string
	limitMemory   string
	gpuType       string
)

var (
	replicas              uint32
	utilizationPercentage uint32
	minReplicas           uint32
	maxReplicas           uint32
	gpuLimit              uint32
)

type Cmd struct {
	fx.In

	Rig rig.Client
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
}

func Setup(parent *cobra.Command) {
	scale := &cobra.Command{
		Use:               "scale",
		Short:             "Scale and inspect the resources of the capsule",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	scaleGet := &cobra.Command{
		Use:   "get",
		Short: "Displays the resources (container size) and replicas of the capsule",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.get),
	}
	scale.AddCommand(scaleGet)

	scaleVertical := &cobra.Command{
		Use:   "vertical",
		Short: "Vertically scaling the capsule (setting the container size)",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.vertical),
	}
	scaleVertical.Flags().StringVar(&requestCPU, "request-cpu", "", "Minimum CPU cores per container")
	scaleVertical.Flags().StringVar(&requestMemory, "request-memory", "", "Minimum memory per container")
	scaleVertical.Flags().StringVar(&limitCPU, "limit-cpu", "", "Maximum CPU cores per container")
	scaleVertical.Flags().StringVar(&limitMemory, "limit-memory", "", "Maximum memory per container")
	scaleVertical.Flags().Uint32Var(&gpuLimit, "limit-gpu", 0, "Maximum number of GPUs per container")
	scaleVertical.Flags().StringVar(&gpuType, "gpu-type", "", "GPU type")
	scaleVertical.MarkFlagsRequiredTogether("limit-gpu", "gpu-type")

	scaleVertical.Flags().BoolVarP(
		&forceDeploy,
		"force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes",
	)
	scale.AddCommand(scaleVertical)

	scaleHorizontal := &cobra.Command{
		Use:   "horizontal",
		Short: "Horizontally scaling the capsule (setting the number of replicas and configuring the autoscaler)",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.horizontal),
	}
	scaleHorizontal.Flags().Uint32VarP(&replicas, "replicas", "r", 0, "number of replicas to scale to")
	scaleHorizontal.Flags().BoolVarP(
		&overwriteAutoscaler, "overwrite-autoscaler", "a", false, "if the autoscaler is enabled, this flag is "+
			"necessary to set the replicas. It will disable the autoscaler.",
	)
	scaleHorizontal.Flags().BoolVarP(
		&forceDeploy,
		"force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes",
	)
	scale.AddCommand(scaleHorizontal)

	scaleHorizontalAuto := &cobra.Command{
		Use:   "autoscale",
		Short: "Configure the autoscaler for horizontal scaling",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.autoscale),
	}
	scaleHorizontalAuto.Flags().Uint32VarP(
		&utilizationPercentage,
		"utilization-percentage", "u", 0, "CPU utilization percentage for the autoscaler. 1 <= 100",
	)
	scaleHorizontalAuto.Flags().Uint32Var(&minReplicas, "min-replicas", 0, "minimum replicas")
	scaleHorizontalAuto.Flags().Uint32Var(&maxReplicas, "max-replicas", 0, "maximum replicas")
	scaleHorizontalAuto.Flags().BoolVarP(
		&disable,
		"disable", "d", false, "Disable the autoscaler, fixing the capsule with its current number of replicas",
	)
	scaleHorizontal.AddCommand(scaleHorizontalAuto)

	parent.AddCommand(scale)
}
