package mlrun

import (
	"fmt"
	"mlops/global"
	"mlops/iam"
	infracomponents "mlops/infra_components"
	"mlops/registry"
	"mlops/storage"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	deploy = true

	application      = "mlrun"
	domainPrefix     = "mlrun"
	namespace        = "mlrun"
	helmChart        = "mlrun-ce"
	helmChartVersion = "0.7.3"
	helmChartRepo    = "https://mlrun.github.io/ce"
)

func CreateMLRunResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcpNetwork *compute.Network,
) error {
	registryEndpoint := fmt.Sprintf("%s-docker.pkg.dev", cloudRegion.Region)
	registryURL := fmt.Sprintf("%s/%s/%s", registryEndpoint, projectConfig.ProjectId, registryName)
	domain := fmt.Sprintf("%s.%s", domainPrefix, projectConfig.Domain)

	infraComponents := infracomponents.InfraComponents{
		CertManager:       true,
		NginxIngress:      true,
		CertManagerIssuer: true,
		Certificate:       false,
		Domain:            domain,
		Ingress:           false,
	}

	registry, err := registry.CreateArtifactRegistry(ctx, projectConfig, registryName, pulumi.DependsOn([]pulumi.Resource{}))
	if err != nil {
		return err
	}
	serviceAccounts, err := iam.CreateIAMResources(ctx, projectConfig, MLRunIAM)
	if err != nil {
		return err
	}
	err = createDockerRegistrySecret(ctx, projectConfig, serviceAccounts, registry, registryURL, k8sProvider)
	if err != nil {
		return err
	}

	gcsBucket := storage.CreateObjectStorage(ctx, projectConfig, bucketName)
	dependencies, LetsEncrypt, err := createKubernetesResources(ctx, projectConfig, infraComponents, k8sProvider)
	if err != nil {
		return err
	}

	dependencies = append(dependencies, gcsBucket)
	MLRunConfig := MLRunConfig{
		RegistryURL:        registryURL,
		GcsBucketName:      bucketName,
		RegistrySecretName: registrySecretName,
		Domain:             domain,
		LetsEncrypt:        LetsEncrypt,
	}
	if deploy {
		if err = deployMLRun(ctx, projectConfig, k8sProvider, MLRunConfig, dependencies); err != nil {
			return err
		}
	}

	return nil
}

func deployMLRun(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	MLRunConfig MLRunConfig,
	dependencies []pulumi.Resource,
) error {

	// Path to the values.yaml file.
	valuesFilePath := "../helm/mlrun/values/values.yaml"
	// Build the replacement map using resolved strings.
	userSettings := map[string]interface{}{
		"gcsbucket":          MLRunConfig.GcsBucketName,
		"hostName":           MLRunConfig.Domain,
		"registryURL":        MLRunConfig.RegistryURL,
		"registrySecretName": MLRunConfig.RegistrySecretName,
		"whitelistedIPs":     projectConfig.WhitelistedIPs,
		"minioRootPassword":  "minio123",
		"letsEncrypt":        MLRunConfig.LetsEncrypt,
	}

	// Get the substituted values map.
	valuesMap, err := global.GetValues(valuesFilePath, userSettings)
	if err != nil {
		return err
	}

	// Deploy the Helm release for Flyte-Core.
	resourceName := fmt.Sprintf("%s-mlrun", projectConfig.ResourceNamePrefix)
	_, err = helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Name:      pulumi.String(application),
		Namespace: pulumi.String(namespace),
		Version:   pulumi.String(helmChartVersion),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String(helmChartRepo),
		},
		Chart:   pulumi.String(helmChart),
		Values:  valuesMap,
		Timeout: pulumi.Int(600),
	},
		pulumi.DependsOn(dependencies),
		pulumi.Provider(k8sProvider),
	)
	if err != nil {
		return fmt.Errorf("failed to deploy MLRun Helm chart: %w", err)
	}
	return nil
}
