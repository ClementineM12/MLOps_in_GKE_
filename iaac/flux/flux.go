package flux

import (
	"fmt"
	"os/exec"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

var (
	namespace        = "flux-system"
	helmChart        = "flux2"
	helmChartVersion = "2.15.0"
	helmChartRepo    = "https://fluxcd-community.github.io/helm-charts"
)

func DeployFlux(
	ctx *pulumi.Context,
	k8sProvider *kubernetes.Provider,
) error {

	githubRepo := config.Get(ctx, "ar:githubRepo")

	// Deploy FluxCD using Helm
	fluxHelmRelease, err := helm.NewRelease(ctx, "flux", &helm.ReleaseArgs{
		Chart:           pulumi.String(helmChart),
		Version:         pulumi.String(helmChartVersion),
		Namespace:       pulumi.String(namespace),
		CreateNamespace: pulumi.Bool(true),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String(helmChartRepo),
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
		pulumi.Provider(k8sProvider))
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
