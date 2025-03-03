package infracomponents

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func deployNginxController(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) (*helm.Release, error) {

	resourceName := fmt.Sprintf("%s-nginx-controller", projectConfig.ResourceNamePrefix)
	return helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Name:            pulumi.String("ingress-nginx"),
		Namespace:       pulumi.String(NginxControllerNamespace),
		CreateNamespace: pulumi.Bool(true),
		Chart:           pulumi.String(NginxControllerHelmChart),
		Version:         pulumi.String(NginxControllerHelmChartVersion),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String(NginxControllerHelmChartRepo),
		},
		Values: pulumi.Map{
			"controller": pulumi.Map{
				"service": pulumi.Map{
					"externalTrafficPolicy": pulumi.String("Local"),
				},
			},
		},
	}, pulumi.Provider(k8sProvider))
}
