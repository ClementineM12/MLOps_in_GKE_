package main

import (
	"fmt"

	"mlops/argocd"
	"mlops/gke"
	"mlops/global"
	gcpIAM "mlops/iam"
	"mlops/istio"
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
	// -------------------------- IAM ------------------------------
	gcpServiceAccountAutoNeg := gcpIAM.CreateServiceAccount(ctx, projectConfig, "AutoNEG")
	// -------------------------- VPC -----------------------------
	gcpNetwork, gcpBackendService, err := vpc.CreateVPCResources(ctx, projectConfig, opts...)
	if err != nil {
		return nil
	}

	// Process Each Cloud Region
	for _, cloudRegion := range projectConfig.EnabledRegions {
		gcpSubnetwork, err := vpc.CreateVPCSubnet(ctx, projectConfig, cloudRegion, gcpNetwork.ID())
		if err != nil {
			return err
		}
		// --------------------------- GKE ----------------------------
		k8sProvider, gcpGKENodePool, err := gke.CreateGKEResources(ctx, projectConfig, &cloudRegion, gcpNetwork.ID(), gcpSubnetwork.ID())
		if err != nil {
			return err
		}
		// --------------------------- ArgoCD ----------------------------
		if config.GetBool(ctx, "argocd:create") {
			argocd.DeployArgoCD(ctx, projectConfig, k8sProvider)
		}
		// --------------------------- Istio ----------------------------
		if config.GetBool(ctx, "istio:create") {
			helmIstioBase, helmIstioD, err := istio.DeployIstio(ctx, projectConfig, cloudRegion, k8sProvider, gcpGKENodePool, gcpBackendService)
			if err != nil {
				return err
			}
			// Deploy Cluster Ops components for GKE AutoNeg and bind Service Account
			err = gcpIAM.ConfigurateAutoNeg(ctx, projectConfig, gcpServiceAccountAutoNeg.ID(), &cloudRegion, k8sProvider, gcpGKENodePool, helmIstioBase, helmIstioD)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
