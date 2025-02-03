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
	"mlops/project"

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
	projectConfig project.ProjectConfig,
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
	projectConfig project.ProjectConfig,
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

// createFirewallRuleHealthChecks creates a firewall rule that allows incoming TCP traffic (ports 80, 8080, 443) for health checks used by services like load balancers.
// The allowed source ranges are from IP blocks 35.191.0.0/16 and 130.211.0.0/22, which are Google Cloud’s health check sources (https://cloud.google.com/load-balancing/docs/health-check-concepts#ip-ranges).
func createFirewallRuleHealthChecks(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	gcpNetwork pulumi.StringInput,
) (*compute.Firewall, error) {

	resourceName := fmt.Sprintf("%s-fw-allow-health-checks", projectConfig.ResourceNamePrefix)
	firewallHealthCheck, err := compute.NewFirewall(ctx, resourceName, &compute.FirewallArgs{
		Project:     pulumi.String(projectConfig.ProjectId),
		Name:        pulumi.String(resourceName),
		Description: pulumi.String("FW - Allow - Ingress - TCP Health Checks"),
		Network:     gcpNetwork,
		Allows: compute.FirewallAllowArray{
			&compute.FirewallAllowArgs{
				Protocol: pulumi.String("tcp"),
				Ports: pulumi.StringArray{
					pulumi.String("80"),
					pulumi.String("8080"),
					pulumi.String("443"),
				},
			},
		},
		SourceRanges: GoogleCloudIPRange,
	})
	return firewallHealthCheck, err
}

// createFirewallInbound creates a firewall rule to allow inbound TCP traffic on ports 80, 8080, and 443,
// typically used for load balancer communication with applications within the VPC.
func createFirewallInbound(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	gcpNetwork pulumi.StringInput,
) (*compute.Firewall, error) {

	resourceName := fmt.Sprintf("%s-fw-allow-cluster-app", projectConfig.ResourceNamePrefix)
	firewallInbound, err := compute.NewFirewall(ctx, resourceName, &compute.FirewallArgs{
		Project:     pulumi.String(projectConfig.ProjectId),
		Name:        pulumi.String(resourceName),
		Description: pulumi.String("FW - Allow - Ingress - Load Balancer to Application"),
		Network:     gcpNetwork,
		Allows: compute.FirewallAllowArray{
			&compute.FirewallAllowArgs{
				Protocol: pulumi.String("tcp"),
				Ports: pulumi.StringArray{
					pulumi.String("80"),
					pulumi.String("8080"),
					pulumi.String("443"),
				},
			},
		},
		SourceRanges: pulumi.StringArray{
			pulumi.String("0.0.0.0/0"),
		},
		TargetTags: pulumi.StringArray{
			pulumi.String("gke-app-access"),
		},
	})
	return firewallInbound, err
}
