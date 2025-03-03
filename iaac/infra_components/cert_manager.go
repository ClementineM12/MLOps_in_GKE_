package infracomponents

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func deployCertManager(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	namespace string,
	k8sProvider *kubernetes.Provider,
	infraComponents InfraComponents,
	opts ...pulumi.ResourceOption,
) (*helm.Release, error) {

	resourceName := fmt.Sprintf("%s-cert-manager", projectConfig.ResourceNamePrefix)
	certManagerRelease, err := helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Name:            pulumi.String("cert-manager"),
		Namespace:       pulumi.String(CertManagerNamespace),
		CreateNamespace: pulumi.Bool(true),
		Chart:           pulumi.String(CertManagerHelmChart),
		Version:         pulumi.String(CertManagerHelmChartVersion),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String(CertManagerHelmChartRepo),
		},
		// Do not skip installing CRDs
		SkipCrds: pulumi.Bool(false),
		// Set the installCRDs value explicitly
		Values: pulumi.Map{
			"installCRDs": pulumi.Bool(true),
		},
		Timeout: pulumi.Int(300),
	}, append(opts, pulumi.Provider(k8sProvider))...)

	if err != nil {
		return nil, err
	}
	if err := configGroup(ctx, projectConfig, namespace, certManagerRelease, k8sProvider, infraComponents); err != nil {
		return nil, err
	}
	return certManagerRelease, nil
}
