package main

import (
	"fmt"

	"mlops/autoneg"
	"mlops/gke"
	"mlops/global"
	"mlops/ml"
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

	gcpSubnetwork, err := vpc.CreateVPCSubnetResources(ctx, projectConfig, gcpNetwork.ID())
	if err != nil {
		return err
	}
	// --------------------------- GKE ----------------------------
	k8sProvider, NodePool, err := gke.CreateGKEResources(ctx, projectConfig, gcpNetwork.ID(), gcpSubnetwork.ID())
	if err != nil {
		return err
	}

	var negServiceAccount *serviceaccount.Account
	if config.GetBool(ctx, "vpc.autoNEG") {
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
		target := config.Get(ctx, "project:target")
		if err = ml.TargetMLOpTool(ctx, target, projectConfig, k8sProvider, gcpNetwork); err != nil {
			return err
		}
		return nil
	})
	return nil
}
