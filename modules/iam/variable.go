package iam

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// SVC is a map of service account configurations required for various system roles.
// Each entry defines a specific service account with its attributes, including the following:
// - resourceNameSuffix: A suffix used to uniquely identify the service account in resource naming.
// - AccountId: The unique identifier for the service account.
// - DisplayName: A human-readable name for the service account, displayed in the GCP console.
// - Members: The member string format for the service account, allowing integration into IAM policies.
// - Description: A brief explanation of the purpose of the IAM role associated with the service account.
// - Title: The display title of the custom IAM role (if applicable).
// - IAMRoleId: The identifier of the custom IAM role for permissions (if applicable).
// - Permissions: A list of specific IAM permissions to be granted to the custom role, allowing the service account
//   to perform necessary actions in the GCP environment.
// - createRole: A boolean flag indicating whether a custom IAM role should be created for the service account.
//
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
		AccountId:          "svc-gke-admin",
		DisplayName:        "Admin Service Account",
		createRole:         false,
	},
	// "ConfigConnector": {
	// 	resourceNameSuffix: "config-connector",
	// 	AccountId: ,
	// 	DisplayName:"Config Connector IAM Service Account" ,
	// 	Members: "serviceAccount:%s.svc.id.goog[cnrm-system/cnrm-controller-manager]",
	// }
}
