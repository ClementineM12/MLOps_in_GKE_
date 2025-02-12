package main

import (
	"fmt"

	"mlops/autoneg"
	"mlops/gke"
	"mlops/global"
	"mlops/registry"
	"mlops/storage"
	"mlops/vpc"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		projectConfig := global.GenerateProjectConfig(ctx)
		gcpDependencies := global.EnableGCPServices(ctx, projectConfig)

		if config.GetBool(ctx, "storage:create") {
			storage.CreateObjectStorage(ctx, projectConfig)
		}
		if config.GetBool(ctx, "ar:create") {
			registry.CreateArtifactRegistry(ctx, projectConfig, pulumi.DependsOn(gcpDependencies))
		}

		err := CreateProjectResources(ctx, projectConfig)
		if err != nil {
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
		k8sProvider, _, err := gke.CreateGKEResources(ctx, projectConfig, &cloudRegion, gcpNetwork.ID(), gcpSubnetwork.ID())
		if err != nil {
			return err
		}

		if config.GetBool(ctx, "vpc.autoNEG") {
			negServiceAccount, err := autoneg.EnableAutoNEGController(ctx, projectConfig, k8sProvider)
			if err != nil {
				return err
			}
			// --------------------------- ArgoCD ----------------------------
			negServiceAccount.ID().ApplyT(func(_ string) error {
				if config.GetBool(ctx, "vpc:loadBalancer") {
					_, err = vpc.CreateBackendServiceResources(ctx, projectConfig)
					if err != nil {
						return err
					}
				}
				return nil
			})
		} else {
			if config.GetBool(ctx, "vpc:loadBalancer") {
				_, err = vpc.CreateBackendServiceResources(ctx, projectConfig)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
