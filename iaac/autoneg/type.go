package autoneg

import (
	rbacV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/rbac/v1"
)

type ClusterRoleDefinition struct {
	Name string
	Bind bool
	RBAC rbacV1.PolicyRuleArray
}

type RoleDefinition struct {
	Name string
	Bind bool
	RBAC rbacV1.PolicyRuleArray
}
