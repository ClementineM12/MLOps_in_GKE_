package vpc

// Package vpc provides functionality for creating and managing a Virtual Private Cloud (VPC)
// in Google Cloud Platform (GCP) using Pulumi. It includes the creation of a VPC network,
// as well as associated firewall rules to manage traffic flow.
//
// The resources created in this package include:
//
// 1. **VPC Network (`createVPCNetwork`)**:
//    - Creates a custom GCP Virtual Private Cloud (VPC) network with subnet auto-creation disabled.
//    - Ensures network segmentation and isolation for workloads.
//
// 2. **Firewall Rule for Health Checks (`createFirewallRuleHealthChecks`)**:
//    - Allows inbound TCP traffic on ports 80, 8080, and 443.
//    - Permits connections from Google's health check IP ranges (35.191.0.0/16, 130.211.0.0/22).
//    - Ensures that Google-managed services (such as Load Balancers) can verify the health of
//      services within the VPC.
//
// 3. **Firewall Rule for Inbound Traffic (`createFirewallInbound`)**:
//    - Allows inbound TCP traffic on ports 80, 8080, and 443 from any source (0.0.0.0/0).
//    - Primarily used for enabling external access to applications deployed inside the VPC.
//    - Uses `gke-app-access` as a target tag to selectively allow traffic to instances
//      that match this tag.
//
// These resources are intended to facilitate the deployment of cloud-native applications,
// ensuring secure network communication while enabling scalability and health monitoring.
//
// Docs: https://cloud.google.com/load-balancing/docs/https/setup-global-ext-https-external-backend

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	GoogleCloudIPRange = pulumi.StringArray{
		pulumi.String("35.191.0.0/16"),
		pulumi.String("130.211.0.0/22"),
	}
)

// createVPC Function is function is responsible for the creation of a VPC Network using the createVPCNetwork function.
// Also, it creates firewall rules for health checks and inbound traffic via the createFirewallRuleHealthChecks and createFirewallInbound functions respectively.
func createVPC(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	opts ...pulumi.ResourceOption,
) (*compute.Network, error) {
	gcpVPCNetwork, err := createVPCNetwork(ctx, projectConfig, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create VPC Network: %w", err)
	}
	firewallHealthCheck, err := createFirewallRuleHealthChecks(ctx, projectConfig, gcpVPCNetwork.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to create firewall rule for Health checks: %w", err)
	}
	firewallInbound, err := createFirewallInbound(ctx, projectConfig, gcpVPCNetwork.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to create firewall rule for Inbound traffic: %w", err)
	}

	// Ensure CreateVPC does not return before firewall rules are applied
	pulumi.All(firewallHealthCheck, firewallInbound).ApplyT(func(_ []interface{}) error {
		return nil
	})
	return gcpVPCNetwork, nil
}

// createVPCNetwork creates a VPC Network in Google Cloud.
func createVPCNetwork(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	opts ...pulumi.ResourceOption,
) (*compute.Network, error) {

	resourceName := fmt.Sprintf("%s-vpc", projectConfig.ResourceNamePrefix)
	gcpNetwork, err := compute.NewNetwork(ctx, resourceName, &compute.NetworkArgs{
		Project:               pulumi.String(projectConfig.ProjectId),
		Name:                  pulumi.String(resourceName),
		Description:           pulumi.String("Global VPC Network"),
		AutoCreateSubnetworks: pulumi.Bool(false),
	}, opts...)

	return gcpNetwork, err
}

// createVPCSubnet reates a subnetwork (subnet) within the VPC.
// The subnet is created in a specific region defined by CloudRegion and the network is linked to the VPC network.
// It enables Private IP Google Access, which allows instances in the subnet to access Google APIs and services over private IPs.
func createVPCSubnet(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	region global.CloudRegion,
	gcpNetwork pulumi.StringInput,
) (*compute.Subnetwork, error) {

	resourceName := fmt.Sprintf("%s-vpc-subnet-%s", projectConfig.ResourceNamePrefix, region.Region)
	gcpSubnetwork, err := compute.NewSubnetwork(ctx, resourceName, &compute.SubnetworkArgs{
		Project:               pulumi.String(projectConfig.ProjectId),
		Name:                  pulumi.String(resourceName),
		Description:           pulumi.String(fmt.Sprintf("VPC Subnet - %s", region.Region)),
		IpCidrRange:           pulumi.String(region.SubnetIp),
		Region:                pulumi.String(region.Region),
		Network:               gcpNetwork,
		PrivateIpGoogleAccess: pulumi.Bool(true),
	})
	return gcpSubnetwork, err
}
