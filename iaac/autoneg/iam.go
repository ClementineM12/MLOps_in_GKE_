package autoneg

import (
	"fmt"

	"mlops/global"
	gcpIAM "mlops/iam"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateAutoNegResources provisions the required IAM roles, Service Account, and Workload Identity.
func createAutoNegIAMResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
) (pulumi.StringArrayOutput, error) {

	// Create AutoNEG Service Account
	gcpAutoNEGServiceAccount, gcpIAMAccountMember, err := gcpIAM.CreateServiceAccount(ctx, projectConfig, "AutoNEG")
	if err != nil {
		return pulumi.StringArrayOutput{}, fmt.Errorf("failed to configure IAM access for Auto NEG Controller end-to-end: %w", err)
	}
	// Assign Workload Identity IAM Role (if enabled)
	err = assignWorkloadIdentity(ctx, projectConfig, gcpAutoNEGServiceAccount)
	if err != nil {
		return pulumi.StringArrayOutput{}, fmt.Errorf("failed to assign IAM Workload Intentity to Auto NEG Controller: %w", err)
	}
	return gcpIAMAccountMember, nil
}

// Step 6: Assign Workload Identity IAM Role (if enabled)
func assignWorkloadIdentity(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	gcpAutoNEGServiceAccount *serviceaccount.Account,
) error {

	_, err := serviceaccount.NewIAMMember(ctx, "autoneg-workload-identity", &serviceaccount.IAMMemberArgs{
		ServiceAccountId: gcpAutoNEGServiceAccount.Name,
		Role:             pulumi.String("roles/iam.workloadIdentityUser"),
		Member: pulumi.Sprintf(
			"serviceAccount:%s.svc.id.goog[autoneg-system/autoneg-controller-manager]", projectConfig.ProjectId,
		),
	})
	return err
}
