package autoneg

import (
	"mlops/global"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func EnableAutoNEGController(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) (pulumi.Output, error) {

	gcpIAMAccountMember, err := createAutoNegIAMResources(ctx, projectConfig)
	if err != nil {
		return nil, err
	}
	negDeployment, err := createAutoNEGKubernetesResources(ctx, projectConfig, k8sProvider, gcpIAMAccountMember)
	if err != nil {
		return nil, err
	}
	return negDeployment, nil
}
