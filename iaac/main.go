package main

import (
	"fmt"

	"mlops/autoneg"
	"mlops/flux"
	"mlops/flyte"
	"mlops/gke"
	"mlops/global"
	"mlops/mlrun"
	"mlops/storage"
	"mlops/vpc"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		projectConfig := global.GenerateProjectConfig(ctx)
		global.EnableGCPServices(ctx, projectConfig)

		if config.GetBool(ctx, "storage:create") {
			bucketName := config.Get(ctx, "storage:name")
			storage.CreateObjectStorage(ctx, projectConfig, bucketName)
		}

		if err := CreateProjectResources(ctx, projectConfig); err != nil {
			return fmt.Errorf("failed to create Project resources end-to-end: %w", err)
		}
		return nil
	})
}

func CreateProjectResources(ctx *pulumi.Context, projectConfig global.ProjectConfig) error {
	// -------------------------- VPC -----------------------------
	gcpNetwork, err := vpc.CreateVPCResources(ctx, projectConfig)
	if err != nil {
		return nil
	}

	// Process Each Cloud Region
	for _, cloudRegion := range projectConfig.EnabledRegions {
		gcpSubnetwork, err := vpc.CreateVPCSubnetResources(ctx, projectConfig, cloudRegion, gcpNetwork.ID())
		if err != nil {
			return err
		}
		// --------------------------- GKE ----------------------------
		k8sProvider, NodePool, err := gke.CreateGKEResources(ctx, projectConfig, &cloudRegion, gcpNetwork.ID(), gcpSubnetwork.ID())
		if err != nil {
			return err
		}

		var negServiceAccount *serviceaccount.Account
		if config.GetBool(ctx, "vpc.autoNEG") {
			var err error
			negServiceAccount, err = autoneg.EnableAutoNEGController(ctx, projectConfig, k8sProvider)
			if err != nil {
				return err
			}
		}
		if config.GetBool(ctx, "vpc:loadBalancer") {
			// If AutoNEG is enabled, defer Load Balancer creation until it's ready
			if negServiceAccount != nil {
				negServiceAccount.ID().ApplyT(func(_ string) error {
					_, err := vpc.CreateBackendServiceResources(ctx, projectConfig)
					return err
				})
			} else {
				// If AutoNEG is not enabled, create Load Balancer immediately
				_, err := vpc.CreateBackendServiceResources(ctx, projectConfig)
				if err != nil {
					return err
				}
			}
		}

		NodePool.ID().ApplyT(func(_ string) error {
			if config.Get(ctx, "project:target") == "kubeflow" {
				if err = flux.DeployFlux(ctx, k8sProvider); err != nil {
					return err
				}
			}
			if config.Get(ctx, "project:target") == "flyte" {
				if err = flyte.CreateFlyteResources(ctx, projectConfig, &cloudRegion, k8sProvider, gcpNetwork); err != nil {
					return err
				}
			}
			if config.Get(ctx, "project:target") == "mlrun" {
				if err = mlrun.CreateMLRunResources(ctx, projectConfig, &cloudRegion, k8sProvider, gcpNetwork); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return nil
}
