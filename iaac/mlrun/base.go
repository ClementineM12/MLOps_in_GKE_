package mlrun

import (
	"fmt"
	"mlops/global"
	"mlops/storage"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	deploy = true
)

func CreateFlyteResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcpNetwork *compute.Network,
) error {

	gcsBucket := storage.CreateObjectStorage(ctx, projectConfig, "mlrun-project-bucket-01")
	dependencies, err := createKubernetesResources(ctx, projectConfig, k8sProvider)
	if err != nil {
		return err
	}
	if deploy {
		if err = deployFlyteCore(ctx, projectConfig, k8sProvider, gcsBucket, dependencies); err != nil {
			return err
		}
	}

	return nil
}

func deployFlyteCore(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	gcsBucket pulumi.StringInput,
	dependencies []pulumi.Resource,
) error {

	// Path to the values.yaml file.
	valuesFilePath := "../helm/mlrun/values/values.yaml"
	// Build the replacement map using resolved strings.
	userSettings := map[string]interface{}{}

	// Get the substituted values map.
	valuesMap, err := global.GetValues(valuesFilePath, userSettings)
	if err != nil {
		return err
	}

	// Deploy the Helm release for Flyte-Core.
	resourceName := fmt.Sprintf("%s-mlrun", projectConfig.ResourceNamePrefix)
	_, err = helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Name:            pulumi.String("mlrun"),
		Namespace:       pulumi.String("mlrun"),
		CreateNamespace: pulumi.Bool(true),
		Version:         pulumi.String("v1.7.2"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://mlrun.github.io/ce"),
		},
		Chart:  pulumi.String("mlrun"),
		Values: valuesMap,
	},
		pulumi.DependsOn(dependencies),
		pulumi.Provider(k8sProvider),
	)
	if err != nil {
		return fmt.Errorf("failed to deploy MLRun Helm chart: %w", err)
	}
	return nil
}
