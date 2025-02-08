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
		AccountId:          "svc-gke-admin",
		DisplayName:        "GKE Admin",
		Description:        "Custom IAM Role - GKE Admin",
		Title:              "GKEAdmin",
		IAMRoleId:          "gke_admin",
		Permissions: pulumi.StringArray{
			// Kubernetes Engine Permissions
			pulumi.String("container.clusters.create"),
			pulumi.String("container.clusters.update"),
			pulumi.String("container.clusters.get"),
			pulumi.String("container.clusters.delete"),
			pulumi.String("container.nodes.create"),
			pulumi.String("container.nodes.list"),
			pulumi.String("container.nodes.update"),
			pulumi.String("container.nodes.get"),
			pulumi.String("container.services.get"),
			pulumi.String("container.services.list"),
			// Networking & Compute Permissions
			pulumi.String("compute.instances.create"),
			pulumi.String("compute.instances.delete"),
			pulumi.String("compute.instances.setMetadata"),
			pulumi.String("compute.instanceGroups.update"),
			pulumi.String("compute.networks.use"),
			pulumi.String("compute.subnetworks.use"),
			pulumi.String("compute.disks.create"),
			pulumi.String("compute.disks.setLabels"),
			// Storage
			pulumi.String("storage.buckets.create"),
			pulumi.String("storage.buckets.delete"),
			// IAM
			pulumi.String("iam.serviceAccounts.actAs"),
			// Monitoring
			pulumi.String("monitoring.timeSeries.list"),
			pulumi.String("monitoring.metricDescriptors.list"),
			pulumi.String("logging.logEntries.list"),
			pulumi.String("logging.logMetrics.get"),
			pulumi.String("logging.logServices.list"),
		},
		createRole: true,
	},
}
