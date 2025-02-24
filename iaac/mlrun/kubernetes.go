package mlrun

import (
	"fmt"
	"mlops/global"
	infracomponents "mlops/infra_components"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	coreV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metaV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createKubernetesResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) ([]pulumi.Resource, error) {

	_, err := createMLRunNamespace(ctx, projectConfig, k8sProvider)
	if err != nil {
		return []pulumi.Resource{}, err
	}
	dependencies, err := infracomponents.CreateInfraComponents(ctx, projectConfig, k8sProvider, infraComponents)
	if err != nil {
		return []pulumi.Resource{}, err
	}

	return dependencies, nil
}

func createMLRunNamespace(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) (*coreV1.Namespace, error) {

	resourceName := fmt.Sprintf("%s-mlrun-ns", projectConfig.ResourceNamePrefix)
	return coreV1.NewNamespace(ctx, resourceName, &coreV1.NamespaceArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Name: pulumi.String("mlrun"),
		},
	}, pulumi.Provider(k8sProvider))
}
