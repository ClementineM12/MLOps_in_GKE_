package vpc

import (
	"fmt"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateVPC Function is function is responsible for the creation of a VPC Network using the createVPCNetwork function.
// Also, it creates firewall rules for health checks and inbound traffic via the createFirewallRuleHealthChecks and createFirewallInbound functions respectively.
func CreateVPC(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpDependencies []pulumi.Resource,
) (*compute.Network, error) {
	gcpVPCNetwork, err := createVPCNetwork(ctx, resourceNamePrefix, gcpProjectId, gcpDependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to create VPC Network: %w", err)
	}
	err = createFirewallRuleHealthChecks(ctx, resourceNamePrefix, gcpProjectId, gcpVPCNetwork.Name)
	if err != nil {
		return nil, err
	}
	err = createFirewallInbound(ctx, resourceNamePrefix, gcpProjectId, gcpVPCNetwork.Name)
	if err != nil {
		return nil, err
	}
	return gcpVPCNetwork, nil
}

// createVPCNetwork creates a VPC Network in Google Cloud.
func createVPCNetwork(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpDependencies []pulumi.Resource,
) (*compute.Network, error) {
	resourceName := fmt.Sprintf("%s-vpc", resourceNamePrefix)
	gcpNetwork, err := compute.NewNetwork(ctx, resourceName, &compute.NetworkArgs{
		Project:               pulumi.String(gcpProjectId),
		Name:                  pulumi.String(resourceName),
		Description:           pulumi.String("Global VPC Network"),
		AutoCreateSubnetworks: pulumi.Bool(false),
	}, pulumi.DependsOn(gcpDependencies))

	return gcpNetwork, err
}

// createFirewallRuleHealthChecks creates a firewall rule that allows incoming TCP traffic (ports 80, 8080, 443) for health checks used by services like load balancers.
// The allowed source ranges are from IP blocks 35.191.0.0/16 and 130.211.0.0/22, which are Google Cloud’s health check sources (https://cloud.google.com/load-balancing/docs/health-check-concepts#ip-ranges).
func createFirewallRuleHealthChecks(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpNetworkName pulumi.StringOutput,
) error {
	resourceName := fmt.Sprintf("%s-fw-in-allow-health-checks", resourceNamePrefix)
	_, err := compute.NewFirewall(ctx, resourceName, &compute.FirewallArgs{
		Project:     pulumi.String(gcpProjectId),
		Name:        pulumi.String(resourceName),
		Description: pulumi.String("FW - Allow - Ingress - TCP Health Checks"),
		Network:     gcpNetworkName,
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
			pulumi.String("35.191.0.0/16"),
			pulumi.String("130.211.0.0/22"),
		},
	})
	return err
}

// createFirewallInbound creates a firewall rule to allow inbound TCP traffic on ports 80, 8080, and 443,
// typically used for load balancer communication with applications within the VPC.
func createFirewallInbound(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpNetworkName pulumi.StringOutput,
) error {
	resourceName := fmt.Sprintf("%s-fw-in-allow-cluster-app", resourceNamePrefix)
	_, err := compute.NewFirewall(ctx, resourceName, &compute.FirewallArgs{
		Project:     pulumi.String(gcpProjectId),
		Name:        pulumi.String(resourceName),
		Description: pulumi.String("FW - Allow - Ingress - Load Balancer to Application"),
		Network:     gcpNetworkName,
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
	return err
}

// CreateVPCSubnet reates a subnetwork (subnet) within the VPC.
// The subnet is created in a specific region defined by CloudRegion and the network is linked to the VPC network.
// It enables Private IP Google Access, which allows instances in the subnet to access Google APIs and services over private IPs.
func CreateVPCSubnet(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	region CloudRegion,
	gcpNetwork *compute.Network,
) (*compute.Subnetwork, error) {
	resourceName := fmt.Sprintf("%s-vpc-subnet-%s", resourceNamePrefix, region.Region)
	gcpSubnetwork, err := compute.NewSubnetwork(ctx, resourceName, &compute.SubnetworkArgs{
		Project:               pulumi.String(gcpProjectId),
		Name:                  pulumi.String(resourceName),
		Description:           pulumi.String(fmt.Sprintf("VPC Subnet - %s", region.Region)),
		IpCidrRange:           pulumi.String(region.SubnetIp),
		Region:                pulumi.String(region.Region),
		Network:               gcpNetwork.ID(),
		PrivateIpGoogleAccess: pulumi.Bool(true),
	})
	return gcpSubnetwork, err
}
