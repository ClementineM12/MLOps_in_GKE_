package iam

import (
	"fmt"
	"mlops/global"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createServiceAccount handles the creation of a Service Account
func createServiceAccount(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	svc *svc,
) (*serviceaccount.Account, pulumi.StringArrayOutput, error) {

	resourceName := fmt.Sprintf("%s-%s-iam-svc", projectConfig.ResourceNamePrefix, svc.resourceNameSuffix)
	gcpServiceAccount, err := serviceaccount.NewAccount(ctx, resourceName, &serviceaccount.AccountArgs{
		Project:     pulumi.String(projectConfig.ProjectId),
		AccountId:   pulumi.String(svc.AccountId),
		DisplayName: pulumi.String(svc.DisplayName),
	})
	serviceAccountMember := gcpServiceAccount.Email.ApplyT(func(email string) []string {
		return []string{fmt.Sprintf("serviceAccount:%s", email)}
	}).(pulumi.StringArrayOutput)
	return gcpServiceAccount, serviceAccountMember, err
}

// CreateIAMRole creates a Custom IAM Role that will be used by the Kubernetes Deployment.
// If svc selected is AutoNEG => This Role allows the AutoNeg CRD to link the Istio Ingress Gateway Service Ip to Load Balancer NEGs ( GCLB )
func createIAMRole(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	svc *svc,
) (*projects.IAMCustomRole, error) {

	roleIDResourceNamePrefix := strings.ReplaceAll(projectConfig.ResourceNamePrefix, "-", "_") // It must match regexp "^[a-zA-Z0-9_\\.]{3,64}$"

	resourceName := fmt.Sprintf("%s-iam-custom-role-%s", projectConfig.ResourceNamePrefix, svc.resourceNameSuffix)
	gcpIAMRole, err := projects.NewIAMCustomRole(ctx, resourceName, &projects.IAMCustomRoleArgs{
		Project:     pulumi.String(projectConfig.ProjectId),
		Description: pulumi.String(svc.Description),
		Permissions: svc.Permissions,
		RoleId:      pulumi.String(fmt.Sprintf("%s_iam_role_%s", roleIDResourceNamePrefix, svc.IAMRoleId)),
		Title:       pulumi.String(svc.Title),
	})
	return gcpIAMRole, err
}

// CreateIAMRoleBinding creates the IAM Role Binding to link to the Service Account to Custom Role.
func createIAMRoleBinding(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	svc *svc,
	gcpIAMRole *projects.IAMCustomRole,
	gcpServiceAccount *serviceaccount.Account,
	serviceAccountMember pulumi.StringArrayOutput,
) (*projects.IAMBinding, error) {

	resourceName := fmt.Sprintf("%s-iam-role-binding-%s", projectConfig.ResourceNamePrefix, svc.resourceNameSuffix)
	gcpIAMRoleBinding, err := projects.NewIAMBinding(ctx, resourceName, &projects.IAMBindingArgs{
		Members: serviceAccountMember,
		Project: pulumi.String(projectConfig.ProjectId),
		Role:    gcpIAMRole.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{gcpServiceAccount}))

	return gcpIAMRoleBinding, err
}

func createIAMPolicyMembers(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	svc *svc,
	serviceAccountMember pulumi.StringArrayOutput,
) error {
	roles := svc.Roles

	for _, role := range roles {
		sanitizedRoleName := strings.ReplaceAll(strings.Split(role, "/")[1], ".", "-")
		_, err := projects.NewIAMMember(ctx, fmt.Sprintf("%s-%s", svc.AccountId, sanitizedRoleName), &projects.IAMMemberArgs{
			Project: pulumi.String(projectConfig.ProjectId),
			Role:    pulumi.String(role),
			Member:  serviceAccountMember.Index(pulumi.Int(0)),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
