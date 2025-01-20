package vpc

import (
	"fmt"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type CloudRegion struct {
	Id             string
	Enabled        bool
	Region         string
	SubnetIp       string
	GKECluster     *container.Cluster
	GKEClusterName string
}

func CreateVPC(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpDependencies []pulumi.Resource,
) (*compute.Network, error) {
	gcpVPCNetwork, err := createVPCNetwork(ctx, resourceNamePrefix, gcpProjectId, gcpDependencies)
	if err != nil {
		return nil, err
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

// createVPCNetwork creates a Google Cloud VPC Network
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
		Description:           pulumi.String("GKE MLOps - Global VPC Network"),
		AutoCreateSubnetworks: pulumi.Bool(false),
	}, pulumi.DependsOn(gcpDependencies))

	return gcpNetwork, err
}

// createFirewallRuleHealthChecks creates the Firewall Rules Health Checks (Network Endpoints within Load Balancer)
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
		Description: pulumi.String("GKE at Scale - FW - Allow - Ingress - TCP Health Checks"),
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

// createFirewallInbound creates the Firewall Rules - Inbound Cluster Access
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
		Description: pulumi.String("GKE at Scale - FW - Allow - Ingress - Load Balancer to Application"),
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
		Description:           pulumi.String(fmt.Sprintf("GKE at Scale - VPC Subnet - %s", region.Region)),
		IpCidrRange:           pulumi.String(region.SubnetIp),
		Region:                pulumi.String(region.Region),
		Network:               gcpNetwork.ID(),
		PrivateIpGoogleAccess: pulumi.Bool(true),
	})
	return gcpSubnetwork, err
}
