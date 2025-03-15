package ml

import (
	"mlops/flux"
	"mlops/flyte"
	"mlops/global"
	"mlops/mlrun"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TargetMLOpTool(
	ctx *pulumi.Context,
	target string,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	gcpNetwork *compute.Network,
) error {

	var err error
	if target == "kubeflow" {
		if err = flux.DeployFlux(ctx, k8sProvider); err != nil {
			return err
		}
	}
	if target == "flyte" {
		if err = flyte.CreateFlyteResources(ctx, projectConfig, k8sProvider, gcpNetwork); err != nil {
			return err
		}
	}
	if target == "mlrun" {
		if err = mlrun.CreateMLRunResources(ctx, projectConfig, k8sProvider, gcpNetwork); err != nil {
			return err
		}
	}
	return nil
}
