package main

import (
	"fmt"

	"mlops/autoneg"
	"mlops/cloudsql"
	"mlops/flux"
	"mlops/flyte"
	"mlops/gke"
	"mlops/global"
	"mlops/registry"
	"mlops/storage"
	"mlops/vpc"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/sql"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		projectConfig := global.GenerateProjectConfig(ctx)
		gcpDependencies := global.EnableGCPServices(ctx, projectConfig)

		var gcsBucket pulumi.StringOutput
		if config.GetBool(ctx, "storage:create") {
			gcsBucket = storage.CreateObjectStorage(ctx, projectConfig)
		}
		if config.GetBool(ctx, "ar:create") {
			registry.CreateArtifactRegistry(ctx, projectConfig, pulumi.DependsOn(gcpDependencies))
		}

		if err := CreateProjectResources(ctx, projectConfig, gcsBucket); err != nil {
			return fmt.Errorf("failed to create Project resources end-to-end: %w", err)
		}
		return nil
	})
}

func CreateProjectResources(ctx *pulumi.Context, projectConfig global.ProjectConfig, gcsBucket pulumi.StringOutput) error {
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

		var cloudSQL *sql.DatabaseInstance
		if projectConfig.CloudSQL.Create {
			cloudSQL, err = cloudsql.DeployCloudSQL(ctx, projectConfig, &cloudRegion, gcpNetwork)
			if err != nil {
				return err
			}
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

		if config.Get(ctx, "project:target") == "kubeflow" {
			if err = flux.DeployFlux(ctx, k8sProvider, NodePool); err != nil {
				return err
			}
		}
		if config.Get(ctx, "project:target") == "flyte" {
			if err = flyte.CreateFlyteResources(ctx, projectConfig, &cloudRegion, k8sProvider, gcsBucket, cloudSQL); err != nil {
				return err
			}
		}
	}
	return nil
}
