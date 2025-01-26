package iam

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
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

	gcpServiceAccount, err := createServiceAccount(ctx, projectConfig, &selectedSVC)
	if err != nil {
		ctx.Log.Error(fmt.Sprintf("IAM Service Account: %s", err), nil)
		return nil
	}
	if selectedSVC.createRole {
		gcpIAMRole, err := createIAMRole(ctx, projectConfig, &selectedSVC)
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("IAM Role: %s", err), nil)
			return nil
		}
		_, err = createIAMRoleBinding(ctx, projectConfig, &selectedSVC, gcpIAMRole, gcpServiceAccount)
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("IAM Role Binding: %s", err), nil)
			return nil
		}
	}
	return gcpServiceAccount
}

// createServiceAccount handles the creation of a Service Account
func createServiceAccount(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	svc *svc,
) (*serviceaccount.Account, error) {

	resourceName := fmt.Sprintf("%s-svc-%s", projectConfig.ResourceNamePrefix, svc.resourceNameSuffix)
	gcpServiceAccount, err := serviceaccount.NewAccount(ctx, resourceName, &serviceaccount.AccountArgs{
		Project:     pulumi.String(projectConfig.ProjectId),
		AccountId:   pulumi.String(svc.AccountId),
		DisplayName: pulumi.String(svc.DisplayName),
	})
	return gcpServiceAccount, err
}

// CreateIAMRole creates a Custom IAM Role that will be used by the Kubernetes Deployment.
// If svc selected is AutoNEG => This Role allows the AutoNeg CRD to link the Istio Ingress Gateway Service Ip to Load Balancer NEGs ( GCLB )
func createIAMRole(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	svc *svc,
) (*projects.IAMCustomRole, error) {

	resourceName := fmt.Sprintf("%s-iam-custom-role-%s", projectConfig.ResourceNamePrefix, svc.resourceNameSuffix)
	gcpIAMRole, err := projects.NewIAMCustomRole(ctx, resourceName, &projects.IAMCustomRoleArgs{
		Project:     pulumi.String(projectConfig.ProjectId),
		Description: pulumi.String(svc.Description),
		Permissions: svc.Permissions,
		RoleId:      pulumi.String(fmt.Sprintf("%s_iam_role_%s", projectConfig.ResourceNamePrefix, svc.IAMRoleId)),
		Title:       pulumi.String(svc.Title),
	})
	return gcpIAMRole, err
}

// CreateIAMRoleBinding creates the IAM Role Binding to link to the Service Account to Custom Role.
func createIAMRoleBinding(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	svc *svc,
	gcpIAMRole *projects.IAMCustomRole,
	gcpServiceAccount *serviceaccount.Account,
) (*projects.IAMBinding, error) {

	resourceName := fmt.Sprintf("%s-iam-role-binding-%s", projectConfig.ResourceNamePrefix, svc.resourceNameSuffix)
	gcpIAMRoleBinding, err := projects.NewIAMBinding(ctx, resourceName, &projects.IAMBindingArgs{
		Members: pulumi.StringArray{
			pulumi.String(fmt.Sprintf(svc.Members, projectConfig.ProjectId)),
		},
		Project: pulumi.String(projectConfig.ProjectId),
		Role:    gcpIAMRole.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{gcpServiceAccount}))

	return gcpIAMRoleBinding, err
}

// Bind Kubernetes AutoNeg Service Account to Workload Identity
// func bindIAMRoleToSVC(
// 	ctx *pulumi.Context,
// 	projectConfig project.ProjectConfig,
// 	cloudRegion *vpc.CloudRegion,
// 	gcpServiceAccountAutoNeg
// ) {
// 	resourceName := fmt.Sprintf("%s-iam-svc-k8s-%s", projectConfig.ResourceNamePrefix, cloudRegion.Region)
// 	_, err := serviceaccount.NewIAMBinding(ctx, resourceName, &serviceaccount.IAMBindingArgs{
// 		ServiceAccountId: gcpServiceAccountAutoNeg.Name,
// 		Role:             pulumi.String("roles/iam.workloadIdentityUser"),
// 		Members: pulumi.StringArray{
// 			pulumi.String(fmt.Sprintf("serviceAccount:%s.svc.id.goog[autoneg-system/autoneg-controller-manager]", projectConfig.ProjectId)),
// 		},
// 	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{gcpGKENodePool, helmClusterOps}), pulumi.Parent(gcpGKENodePool))
// 	if err != nil {
// 		return err
// 	}
// }
