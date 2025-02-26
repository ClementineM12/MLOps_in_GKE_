package iam

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func transformRole(role string) string {
	// Split the string by "/" and get the last element
	parts := strings.Split(role, "/")
	if len(parts) < 2 {
		return ""
	}
	value := parts[len(parts)-1]
	// Replace periods with dashes
	result := strings.ReplaceAll(value, ".", "-")
	return result
}

func (iam IAM) Validate(ctx *pulumi.Context) error {
	// ResourceNamePrefix must be provided if either CreateMember or CreateServiceAccount is true.
	if (iam.CreateMember || iam.CreateServiceAccount) && iam.ResourceNamePrefix == "" {
		err := fmt.Errorf("field `ResourceNamePrefix` must be provided if `CreateMember` or `CreateServiceAccount` is true")
		ctx.Log.Error(err.Error(), nil)
		return err
	}

	// If CreateRole is true then Permissions must be provided.
	if iam.CreateRole && len(iam.Permissions) == 0 {
		err := fmt.Errorf("field `Permissions` must be provided if `CreateRole` is true")
		ctx.Log.Error(err.Error(), nil)
		return err
	}

	// If RoleBinding is not empty then CreateServiceAccount must be true.
	if iam.RoleBindings != nil && !iam.CreateServiceAccount {
		err := fmt.Errorf("field `CreateServiceAccount` must be true if `RoleBinding` is provided")
		ctx.Log.Error(err.Error(), nil)
		return err
	}

	// If WorkloadIdentityBinding is provided then CreateServiceAccount must be true.
	if len(iam.WorkloadIdentityBinding) > 0 && !iam.CreateServiceAccount {
		err := fmt.Errorf("field `CreateServiceAccount` must be true if `WorkloadIdentityBinding` is provided")
		ctx.Log.Error(err.Error(), nil)
		return err
	}

	// If CreateMember is true then Roles must be provided.
	if iam.CreateMember && len(iam.Roles) == 0 {
		err := fmt.Errorf("field `Roles` must be provided if `CreateMember` is true")
		ctx.Log.Error(err.Error(), nil)
		return err
	}

	return nil
}
