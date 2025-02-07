package iam

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// - AutoNEG: The AutoNEG (Automatic Network Endpoint Groups) service account is specifically designed to support
// 	   GKE's Automatic Network Endpoint Group functionality. AutoNEG simplifies the configuration and
// 	   management of backend services for Kubernetes clusters by automatically managing network endpoint
// 	   group memberships. This service account requires precise permissions to perform its role effectively.
// - Admin: This is a general-purpose admin service account. No custom IAM role is created for this account.

var SVC = map[string]svc{
	"AutoNEG": {
		resourceNameSuffix: "autoneg",
		AccountId:          "svc-autoneg-system",
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
		createRole: true,
	},
	"Admin": {
		resourceNameSuffix: "admin",
		Permissions: pulumi.StringArray{
			pulumi.String("roles/container.clusterAdmin"),
		},
	},
}
