package iam

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type IAM struct {
	ResourceNamePrefix      string
	DisplayName             string
	Title                   string
	Description             string
	IAMRoleId               string
	Permissions             pulumi.StringArray
	CreateRole              bool
	CreateMember            bool
	CreateServiceAccount    bool
	CreateKey               bool
	RoleBindings            []string
	Roles                   []string
	WorkloadIdentityBinding []string
}

type ServiceAccountInfo struct {
	ServiceAccount *serviceaccount.Account
	Member         pulumi.StringArrayOutput
	Email          pulumi.StringOutput
	Key            *serviceaccount.Key
}
