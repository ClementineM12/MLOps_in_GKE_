package infracomponents

import (
	"mlops/global"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateInfraComponents(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	namespace string,
	k8sProvider *kubernetes.Provider,
	infraComponents InfraComponents,
) ([]pulumi.Resource, error) {

	var (
		dependencies []pulumi.Resource
		err          error
	)
	var certManagerIssuer *helm.Release
	var nginxController *helm.Release

	if infraComponents.NginxIngress {
		nginxController, err = deployNginxController(ctx, projectConfig, k8sProvider)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, nginxController)
	}
	if infraComponents.CertManager {
		// If nginxController exists, add it as a dependency.
		var opts []pulumi.ResourceOption
		if nginxController != nil {
			opts = append(opts, pulumi.DependsOn([]pulumi.Resource{nginxController}))
		}
		certManagerIssuer, err = deployCertManager(ctx, projectConfig, namespace, k8sProvider, infraComponents, opts...)
		if err != nil {
			return nil, err
		}
		if infraComponents.Ingress {
			deployIngress(ctx, projectConfig, namespace, infraComponents)
		}
		dependencies = append(dependencies, certManagerIssuer)
	}

	return dependencies, nil
}
