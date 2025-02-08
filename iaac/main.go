package main

import (
	"fmt"

	"mlops/argocd"
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

		err := CreateProjectResources(ctx, projectConfig, pulumi.DependsOn(gcpDependencies))
		if err != nil {
			return fmt.Errorf("failed to create Project resources end-to-end: %w", err)
		}
		return nil
	})
}

func CreateProjectResources(ctx *pulumi.Context, projectConfig global.ProjectConfig, opts ...pulumi.ResourceOption) error {
	// -------------------------- VPC -----------------------------
	gcpNetwork, _, err := vpc.CreateVPCResources(ctx, projectConfig, opts...)
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
			negReady, err := autoneg.EnableAutoNEGController(ctx, projectConfig, k8sProvider)
			if err != nil {
				return err
			}
			// --------------------------- ArgoCD ----------------------------
			negReady.ApplyT(func(_ interface{}) error {
				if config.GetBool(ctx, "argocd:create") {
					err = argocd.DeployArgoCD(ctx, projectConfig, k8sProvider)
					if err != nil {
						return err
					}
				}
				return nil
			})
		}
		if config.GetBool(ctx, "argocd:create") {
			err = argocd.DeployArgoCD(ctx, projectConfig, k8sProvider)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
