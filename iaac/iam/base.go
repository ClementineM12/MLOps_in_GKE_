package iam

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/container"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	helm "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
)

// CreateServiceAccount creates a Google Cloud Service Account
func CreateServiceAccount(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	target string,
) *serviceaccount.Account {

	// Retrieve the corresponding service configuration from the map
	selectedSVC, found := SVC[target]
	if !found {
		ctx.Log.Error(fmt.Sprintf("service '%s' not found", target), nil)
		return nil
	}

	gcpServiceAccount, serviceAccountMember, err := createServiceAccount(ctx, projectConfig, &selectedSVC)
	if err != nil {
		ctx.Log.Error(fmt.Sprintf("failed to create IAM service account: %s", err), nil)
		return nil
	}
	if selectedSVC.createRole {
		gcpIAMRole, err := createIAMRole(ctx, projectConfig, &selectedSVC)
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("IAM Role: %s", err), nil)
			return nil
		}
		_, err = createIAMRoleBinding(ctx, projectConfig, &selectedSVC, gcpIAMRole, gcpServiceAccount, serviceAccountMember)
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("IAM Role Binding: %s", err), nil)
			return nil
		}
	}
	return gcpServiceAccount
}

func ConfigurateAutoNeg(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	gcpServiceAccountAutoNeg pulumi.StringInput,
	cloudRegion *project.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcpGKENodePool *container.NodePool,
	helmIstioBase *helm.Release,
	helmIstioD *helm.Release,
) error {

	helmClusterOps, err := autoNegDeployClusterOps(ctx, projectConfig, cloudRegion, k8sProvider, gcpGKENodePool, helmIstioBase, helmIstioD)
	if err != nil {
		return fmt.Errorf("failed to deploy Clsuter Ops: %w", err)
	}
	err = autoNegServiceAccountBind(ctx, projectConfig, cloudRegion, k8sProvider, gcpGKENodePool, helmClusterOps, gcpServiceAccountAutoNeg)
	if err != nil {
		return fmt.Errorf("failed to bind Auto Neg Service account: %w", err)
	}
	return nil
}
