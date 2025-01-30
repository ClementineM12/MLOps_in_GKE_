package iam

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type svc struct {
	resourceNameSuffix string
	AccountId          string
	DisplayName        string
	Description        string
	Title              string
	IAMRoleId          string
	Permissions        pulumi.StringArray
	createRole         bool
}
