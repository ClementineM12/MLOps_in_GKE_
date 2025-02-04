package autoneg

import (
	rbacV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var ClusterRoles = []ClusterRoleDefinition{
	{
		Name: "autoneg-manager-role",
		Bind: true,
		RBAC: rbacV1.PolicyRuleArray{
			&rbacV1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("")},
				Resources: pulumi.StringArray{pulumi.String("events")},
				Verbs:     pulumi.StringArray{pulumi.String("create"), pulumi.String("patch")},
			},
			&rbacV1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("")},
				Resources: pulumi.StringArray{pulumi.String("services")},
				Verbs:     pulumi.StringArray{pulumi.String("get"), pulumi.String("list"), pulumi.String("patch"), pulumi.String("update"), pulumi.String("watch")},
			},
			&rbacV1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("")},
				Resources: pulumi.StringArray{pulumi.String("services/finalizers")},
				Verbs:     pulumi.StringArray{pulumi.String("update")},
			},
			&rbacV1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("")},
				Resources: pulumi.StringArray{pulumi.String("services/status")},
				Verbs:     pulumi.StringArray{pulumi.String("get"), pulumi.String("patch"), pulumi.String("update")},
			},
		},
	},
	{
		Name: "autoneg-metrics-reader",
		RBAC: rbacV1.PolicyRuleArray{
			&rbacV1.PolicyRuleArgs{
				NonResourceURLs: pulumi.StringArray{pulumi.String("/metrics")},
				Verbs:           pulumi.StringArray{pulumi.String("get")},
			},
		},
	},
	{
		Name: "autoneg-proxy-role",
		Bind: true,
		RBAC: rbacV1.PolicyRuleArray{
			&rbacV1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("authentication.k8s.io")},
				Resources: pulumi.StringArray{pulumi.String("tokenreviews")},
				Verbs:     pulumi.StringArray{pulumi.String("create")},
			},
			&rbacV1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("authorization.k8s.io")},
				Resources: pulumi.StringArray{pulumi.String("subjectaccessreviews")},
				Verbs:     pulumi.StringArray{pulumi.String("create")},
			},
		},
	},
}

var Roles = []RoleDefinition{
	{
		Name: "autoneg-leader-election-role",
		Bind: true,
		RBAC: rbacV1.PolicyRuleArray{
			&rbacV1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("")},
				Resources: pulumi.StringArray{pulumi.String("configmaps")},
				Verbs:     pulumi.StringArray{pulumi.String("get"), pulumi.String("list"), pulumi.String("watch"), pulumi.String("create"), pulumi.String("update"), pulumi.String("patch"), pulumi.String("delete")},
			},
			&rbacV1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("coordination.k8s.io")},
				Resources: pulumi.StringArray{pulumi.String("leases")},
				Verbs:     pulumi.StringArray{pulumi.String("get"), pulumi.String("list"), pulumi.String("watch"), pulumi.String("create"), pulumi.String("update"), pulumi.String("patch"), pulumi.String("delete")},
			},
			&rbacV1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("")},
				Resources: pulumi.StringArray{pulumi.String("events")},
				Verbs:     pulumi.StringArray{pulumi.String("create")},
			},
		},
	},
}
