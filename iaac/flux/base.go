package flux

import (
	"fmt"
	"os/exec"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/container"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func DeployFlux(ctx *pulumi.Context, k8sProvider *kubernetes.Provider, NodePool *container.NodePool) error {
	githubRepo := config.Get(ctx, "ar:githubRepo")

	// Deploy FluxCD using Helm
	fluxHelmRelease, err := helm.NewRelease(ctx, "flux", &helm.ReleaseArgs{
		Chart:           pulumi.String("flux2"),
		Version:         pulumi.String("2.10.0"), // Update to the latest version
		Namespace:       pulumi.String("flux-system"),
		CreateNamespace: pulumi.Bool(true),
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
	},
		pulumi.Provider(k8sProvider),
		pulumi.DependsOn([]pulumi.Resource{NodePool}),
	)
	if err != nil {
		return err
	}
	// Output Flux Helm Release status
	ctx.Export("fluxHelmRelease", fluxHelmRelease.Status)

	// err = bootstrapFluxToGitHub(ctx, githubRepo)
	// if err != nil {
	// 	return err
	// }

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
