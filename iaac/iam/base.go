package iam

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateServiceAccount creates a Google Cloud Service Account
func CreateServiceAccount(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	target string,
) (*serviceaccount.Account, pulumi.StringArrayOutput, error) {

	// Retrieve the corresponding service configuration from the map
	selectedSVC, found := SVC[target]
	if !found {
		ctx.Log.Error(fmt.Sprintf("service '%s' not found", target), nil)
		return nil, pulumi.StringArrayOutput{}, nil
	}

	gcpServiceAccount, serviceAccountMember, err := createServiceAccount(ctx, projectConfig, &selectedSVC)
	if err != nil {
		return nil, pulumi.StringArrayOutput{}, fmt.Errorf("IAM service account: %w", err)
	}
	if selectedSVC.createRole {
		gcpIAMRole, err := createIAMRole(ctx, projectConfig, &selectedSVC)
		if err != nil {
			return nil, pulumi.StringArrayOutput{}, fmt.Errorf("IAM Role: %w", err)
		}
		_, err = createIAMRoleBinding(ctx, projectConfig, &selectedSVC, gcpIAMRole, gcpServiceAccount, serviceAccountMember)
		if err != nil {
			return nil, pulumi.StringArrayOutput{}, fmt.Errorf("IAM Role Binding: %w", err)
		}
	}
	return gcpServiceAccount, serviceAccountMember, nil
}
