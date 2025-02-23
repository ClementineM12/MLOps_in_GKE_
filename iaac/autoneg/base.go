package autoneg

import (
	"fmt"
	"mlops/global"
	"mlops/iam"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func EnableAutoNEGController(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) (*serviceaccount.Account, error) {

	// Create AutoNEG IAM Resources
	AutoNEGServiceAccount, err := iam.CreateIAMResources(ctx, projectConfig, AutoNEGSystemIAM)
	if err != nil {
		return nil, fmt.Errorf("failed to configure IAM access for Auto NEG Controller end-to-end: %w", err)
	}

	// Apply AutoNEG Kubernetes Deployment
	negDeployment := createAutoNEGKubernetesResources(ctx, projectConfig, k8sProvider, AutoNEGServiceAccount)

	// Ensure NEG Deployment is applied before returning
	negDeployment.ApplyT(func(_ interface{}) string {
		ctx.Log.Info("AutoNEG Deployment completed.", nil)
		return "AutoNEG Deployment successful"
	})

	return AutoNEGServiceAccount["autoneg"].ServiceAccount, nil
}
