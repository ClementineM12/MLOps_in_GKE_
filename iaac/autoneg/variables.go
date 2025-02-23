package autoneg

import (
	"mlops/iam"

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

// - AutoNEG: The AutoNEG (Automatic Network Endpoint Groups) service account is specifically designed to support
// 	   GKE's Automatic Network Endpoint Group functionality. AutoNEG simplifies the configuration and
// 	   management of backend services for Kubernetes clusters by automatically managing network endpoint
// 	   group memberships. This service account requires precise permissions to perform its role effectively.
// - Admin: This is a general-purpose admin service account. No custom IAM role is created for this account.

var AutoNEGSystemIAM = map[string]iam.IAM{
	"autoneg": {
		ResourceNamePrefix: "autoneg",
		DisplayName:        "AutoNEG Service Account",
		Description:        "Custom IAM Role - GKE AutoNeg",
		Title:              "AutoNEG",
		IAMRoleId:          "autoneg_system",
		Permissions: pulumi.StringArray{
			pulumi.String("compute.backendServices.get"),
			pulumi.String("compute.backendServices.update"),
			pulumi.String("compute.regionBackendServices.get"),
			pulumi.String("compute.regionBackendServices.update"),
			pulumi.String("compute.networkEndpointGroups.use"),
			pulumi.String("compute.healthChecks.useReadOnly"),
			pulumi.String("compute.regionHealthChecks.useReadOnly"),
		},
		CreateRole:           true,
		CreateMember:         true,
		CreateServiceAccount: true,
		WorkloadIdentityBinding: []string{
			"autoneg-system/autoneg-controller-manager",
		},
	},
}
