package iam

import (
	"errors"
	"fmt"
	"mlops/global"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createServiceAccount handles the creation of a Service Account
func createServiceAccounts(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	IAM map[string]IAM,
) (map[string]ServiceAccountInfo, error) {

	serviceAccounts := make(map[string]ServiceAccountInfo)

	var serviceAccountKey *serviceaccount.Key

	for roleName, iamInfo := range IAM {
		if !iamInfo.CreateServiceAccount {
			continue
		}

		resourceName := fmt.Sprintf("%s-%s-iam-svc", projectConfig.ResourceNamePrefix, roleName)
		IAMServiceAccount, err := serviceaccount.NewAccount(ctx, resourceName, &serviceaccount.AccountArgs{
			AccountId:   pulumi.String(fmt.Sprintf("%s-%s", iamInfo.ResourceNamePrefix, roleName)),
			Project:     pulumi.String(projectConfig.ProjectId),
			DisplayName: pulumi.String(iamInfo.DisplayName),
		})
		if err != nil {
			return nil, err
		}
		if iamInfo.CreateKey {
			serviceAccountKey, err = createServiceAccountKey(ctx, IAMServiceAccount, resourceName)
			if err != nil {
				return nil, err
			}
		}
		if IAMServiceAccount == nil {
			ctx.Log.Error("IAMServiceAccount is nil", nil)
			return nil, errors.New("IAMServiceAccount is nil")
		}
		member := IAMServiceAccount.Email.ApplyT(func(email string) []string {
			if email == "" {
				ctx.Log.Error("IAMServiceAccount.Email resolved as empty", nil)
			}
			return []string{fmt.Sprintf("serviceAccount:%s", email)}
		}).(pulumi.StringArrayOutput)

		email := IAMServiceAccount.AccountId.ApplyT(func(id string) string {
			if id == "" {
				ctx.Log.Error("IAMServiceAccount.AccountId resolved as empty", nil)
			}
			return fmt.Sprintf("%s@%s.iam.gserviceaccount.com", id, projectConfig.ProjectId)
		}).(pulumi.StringOutput)

		serviceAccounts[roleName] = ServiceAccountInfo{
			ServiceAccount: IAMServiceAccount,
			Member:         member,
			Email:          email,
			Key:            serviceAccountKey,
		}
	}
	return serviceAccounts, nil
}

// createServiceAccountKey creates a key for the provided service account.
// The key's JSON will later be used as the "password" for registry authentication.
func createServiceAccountKey(
	ctx *pulumi.Context,
	sa *serviceaccount.Account,
	serviceAccountResourceName string,
) (*serviceaccount.Key, error) {

	resourceName := fmt.Sprintf("%s-key", serviceAccountResourceName)
	saKey, err := serviceaccount.NewKey(ctx, resourceName, &serviceaccount.KeyArgs{
		ServiceAccountId: sa.Name,
	})
	if err != nil {
		return nil, err
	}
	return saKey, nil
}

// CreateIAMRole creates a Custom IAM Role that will be used by the Kubernetes Deployment.
// If svc selected is AutoNEG => This Role allows the AutoNeg CRD to link the Istio Ingress Gateway Service Ip to Load Balancer NEGs ( GCLB )
func createIAMRole(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	iamInfo *IAM,
	roleName string,
	roleIDResourceNameSuffix string,
) (*projects.IAMCustomRole, error) {

	roleIDResourceNamePrefix := strings.ReplaceAll(projectConfig.ResourceNamePrefix, "-", "_") // It must match regexp "^[a-zA-Z0-9_\\.]{3,64}$"

	resourceName := fmt.Sprintf("%s-%s-iam-role", projectConfig.ResourceNamePrefix, roleName)
	return projects.NewIAMCustomRole(ctx, resourceName, &projects.IAMCustomRoleArgs{
		Title:       pulumi.String(roleName),
		Permissions: iamInfo.Permissions,
		Project:     pulumi.String(projectConfig.ProjectId),
		RoleId:      pulumi.String(fmt.Sprintf("%s_iam_role_%s", roleIDResourceNamePrefix, roleIDResourceNameSuffix)),
	})
}

// CreateIAMServiceRoleBinding creates the IAM Role Binding to link to the Service Account to Custom Role.
func createIAMServiceCustomRoleBinding(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	roleName string,
	serviceAccounts map[string]ServiceAccountInfo,
	IAMRole *projects.IAMCustomRole,
) (*projects.IAMBinding, error) {

	resourceName := fmt.Sprintf("%s-%s-iam-role-binding", projectConfig.ResourceNamePrefix, roleName)
	return projects.NewIAMBinding(ctx, resourceName, &projects.IAMBindingArgs{
		Project: pulumi.String(projectConfig.ProjectId),
		Role:    IAMRole.ID(),
		Members: serviceAccounts[roleName].Member,
	})
}

func createIAMRoleBinding(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	roleBinding string,
	roleName string,
	serviceAccounts map[string]ServiceAccountInfo,
) (*projects.IAMBinding, error) {

	resourceName := fmt.Sprintf("%s-iam-role-binding-%s", projectConfig.ResourceNamePrefix, transformRole(roleBinding))
	return projects.NewIAMBinding(ctx, resourceName, &projects.IAMBindingArgs{
		Project: pulumi.String(projectConfig.ProjectId),
		Role:    pulumi.String(roleBinding),
		Members: serviceAccounts[roleName].Member,
	})
}

func createIAMPolicyMembers(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	IAM map[string]IAM,
	serviceAccounts map[string]ServiceAccountInfo,
) error {

	for roleName, iamInfo := range IAM {
		if iamInfo.CreateMember {
			roles := iamInfo.Roles
			accountId := fmt.Sprintf("%s-%s-%s", projectConfig.ResourceNamePrefix, iamInfo.ResourceNamePrefix, roleName)

			for _, role := range roles {
				sanitizedRoleName := strings.ReplaceAll(strings.Split(role, "/")[1], ".", "-")
				_, err := projects.NewIAMMember(ctx, fmt.Sprintf("%s-%s", accountId, sanitizedRoleName), &projects.IAMMemberArgs{
					Project: pulumi.String(projectConfig.ProjectId),
					Role:    pulumi.String(role),
					Member:  serviceAccounts[roleName].Member.Index(pulumi.Int(0)),
				})
				if err != nil {
					return err
				}
			}
			return nil
		}
	}
	return nil
}
