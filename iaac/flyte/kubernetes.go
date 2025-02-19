package flyte

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	coreV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metaV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createKubernetesResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) ([]pulumi.Resource, error) {
	if err := createFlyteNamespace(ctx, projectConfig, k8sProvider); err != nil {
		return []pulumi.Resource{}, err
	}
	nginxController, err := deployNginxController(ctx, projectConfig)
	if err != nil {
		return []pulumi.Resource{}, err
	}
	certManagerIssuer, err := deployCertManager(ctx, projectConfig)
	if err != nil {
		return []pulumi.Resource{}, err
	}

	dependsOn := []pulumi.Resource{
		certManagerIssuer,
		nginxController,
	}

	if err := createTLSSecret(ctx, projectConfig); err != nil {
		return []pulumi.Resource{}, err
	}
	return dependsOn, nil
}

func createFlyteNamespace(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) error {

	resourceName := fmt.Sprintf("%s-flyte-ns", projectConfig.ResourceNamePrefix)
	_, err := coreV1.NewNamespace(ctx, resourceName, &coreV1.NamespaceArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Name: pulumi.String("flyte"),
		},
	}, pulumi.Provider(k8sProvider))

	return err
}

func deployNginxController(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
) (*helm.Release, error) {

	resourceName := fmt.Sprintf("%s-nginx-controller", projectConfig.ResourceNamePrefix)
	return helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Chart:           pulumi.String("nginx-controller"),
		Namespace:       pulumi.String("nginx-ingress"),
		CreateNamespace: pulumi.Bool(true),
		Version:         pulumi.String("4.11.4"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://kubernetes.github.io/ingress-nginx"),
		},
	})
}

func deployCertManager(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
) (*helm.Release, error) {

	resourceName := fmt.Sprintf("%s-cert-manager", projectConfig.ResourceNamePrefix)
	return helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Name:            pulumi.String("cert-manager"),
		Namespace:       pulumi.String("cert-manager"),
		CreateNamespace: pulumi.Bool(true),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://charts.jetstack.io"),
		},
		Chart:   pulumi.String("cert-manager"),
		Version: pulumi.String("v1.17.0"),
	})
}

func createTLSSecret(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
) error {

	resourceName := fmt.Sprintf("%s-flyte-tls-k8s-secret", projectConfig.ResourceNamePrefix)
	_, err := coreV1.NewSecret(ctx, resourceName, &coreV1.SecretArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Name:      pulumi.String("flyte-secret-tls"),
			Namespace: pulumi.String("flyte"),
		},
	})
	return err
}
