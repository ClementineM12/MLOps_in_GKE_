package flyte

import (
	"fmt"
	"mlops/global"
	infracomponents "mlops/infra_components"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/sql"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	coreV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metaV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createKubernetesResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	infraComponents infracomponents.InfraComponents,
	k8sProvider *kubernetes.Provider,
	cloudSQL *sql.DatabaseInstance,
) ([]pulumi.Resource, error) {

	dependencies := []pulumi.Resource{}
	_, err := createFlyteNamespace(ctx, projectConfig, k8sProvider)
	if err != nil {
		return dependencies, err
	}
	dependencies, _, err = infracomponents.CreateInfraComponents(ctx, projectConfig, namespace, k8sProvider, infraComponents)
	if err != nil {
		return dependencies, err
	}

	dependsOn := append(dependencies,
		cloudSQL,
	)

	return dependsOn, nil
}

func createFlyteNamespace(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) (*coreV1.Namespace, error) {

	resourceName := fmt.Sprintf("%s-flyte-ns", projectConfig.ResourceNamePrefix)
	return coreV1.NewNamespace(ctx, resourceName, &coreV1.NamespaceArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Name: pulumi.String(namespace),
		},
	}, pulumi.Provider(k8sProvider))
}
