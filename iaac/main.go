package main

import (
	"fmt"
	"os/exec"

	"mlops/autoneg"
	"mlops/gke"
	"mlops/global"
	"mlops/registry"
	"mlops/storage"
	"mlops/vpc"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metaV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
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
		err = deployFlux(ctx)
		if err != nil {
			return err
		}

	}
	return nil
}

func bootstrapFluxToGitHub(ctx *pulumi.Context, githubRepo string) error {
	// Run flux bootstrap command
	cmd := exec.Command("flux", "bootstrap", "github",
		fmt.Sprintf("--owner=%s", githubRepo),
		"--repository=kubeflow-flux-deploy",
		"--branch=main",
		"--path=./flux-system",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to bootstrap Flux: %v\n%s", err, string(output))
	}
	ctx.Log.Info("Flux successfully bootstrapped to GitHub", nil)
	return nil
}

func deployFlux(ctx *pulumi.Context) error {
	githubRepo := config.Get(ctx, "ar:githubRepo")

	fluxNamespace, err := v1.NewNamespace(ctx, "flux-system", &v1.NamespaceArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Name: pulumi.String("flux-system"),
		},
	})
	if err != nil {
		return err
	}

	// Deploy FluxCD using Helm
	fluxHelmRelease, err := helm.NewRelease(ctx, "flux", &helm.ReleaseArgs{
		Chart:     pulumi.String("flux2"),
		Version:   pulumi.String("2.10.0"), // Update to the latest version
		Namespace: fluxNamespace.Metadata.Name().Elem(),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://fluxcd-community.github.io/helm-charts"),
		},
		Values: pulumi.Map{
			"gitRepository": pulumi.Map{
				"url": pulumi.String(fmt.Sprintf("https://github.com/%s", githubRepo)),
				"ref": pulumi.Map{
					"branch": pulumi.String("main"),
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{fluxNamespace}))
	if err != nil {
		return err
	}
	// Output Flux Helm Release status
	ctx.Export("fluxNamespace", fluxNamespace.Metadata.Name())
	ctx.Export("fluxHelmRelease", fluxHelmRelease.Status)

	// err = bootstrapFluxToGitHub(ctx, githubRepo)
	// if err != nil {
	// 	return err
	// }

	return nil
}
