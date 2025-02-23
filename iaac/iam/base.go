package iam

import (
	"fmt"
	"mlops/global"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateServiceAccount creates a Google Cloud Service Account
func CreateIAMResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	IAM map[string]IAM,
) (map[string]ServiceAccountInfo, error) {

	// Validate each IAM configuration
	// for _, iamConfig := range IAM {
	// 	if err := iamConfig.Validate(ctx); err != nil {
	// 		return nil, err
	// 	}
	// }

	serviceAccounts, err := createServiceAccounts(ctx, projectConfig, IAM)
	if err != nil {
		return nil, fmt.Errorf("IAM service account: %w", err)
	}
	if err := createIAMBindings(ctx, projectConfig, IAM, serviceAccounts); err != nil {
		return nil, err
	}
	if err := createIAMPolicyMembers(ctx, projectConfig, IAM, serviceAccounts); err != nil {
		return nil, fmt.Errorf("IAM Role Policy Members: %w", err)
	}

	return serviceAccounts, nil
}

func createIAMBindings(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	IAM map[string]IAM,
	serviceAccounts map[string]ServiceAccountInfo,
) error {

	for roleName, iamInfo := range IAM {

		roleIDResourceNameSuffix := strings.ReplaceAll(roleName, "-", "_")

		if iamInfo.RoleBinding != "" && iamInfo.CreateServiceAccount {
			if _, err := createIAMRoleBinding(ctx, projectConfig, &iamInfo, roleName, serviceAccounts); err != nil {
				return err
			}
		}

		if !iamInfo.CreateRole {
			continue
		}

		newRole, err := createIAMRole(ctx, projectConfig, &iamInfo, roleName, roleIDResourceNameSuffix)
		if err != nil {
			return err
		}
		if _, err = createIAMServiceRoleBinding(ctx, projectConfig, roleName, serviceAccounts, newRole); err != nil {
			return err
		}

		if iamInfo.WorkloadIdentityBinding != nil {
			identityNamespace := pulumi.Sprintf("%s.svc.id.goog", projectConfig.ProjectId)

			resourceName := fmt.Sprintf("%s-%s-workload-identity-binding", projectConfig.ResourceNamePrefix, roleName)
			for _, svcBind := range iamInfo.WorkloadIdentityBinding {
				if _, err = serviceaccount.NewIAMBinding(ctx, resourceName, &serviceaccount.IAMBindingArgs{
					ServiceAccountId: serviceAccounts[roleName].ServiceAccount.ID(),
					Role:             pulumi.String("roles/iam.workloadIdentityUser"),
					Members: pulumi.StringArray{
						pulumi.Sprintf("serviceAccount:%s[%s]", identityNamespace, svcBind),
					},
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
