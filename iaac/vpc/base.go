package vpc

// 📌 Think of it like this:
//
// Load Balancer = The "Traffic Controller"
// Backend Service = The "Traffic Distribution System"

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateVPCResources provisions a VPC network along with necessary load balancing resources.
// It sets up the network (`createVPC`), backend service, global static IP, and optionally configures SSL certificates.
// The function also creates a URL map for HTTP traffic, ensuring proper request routing.
// This enables a secure and scalable infrastructure for cloud-based applications.
func CreateVPCResources(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	opts ...pulumi.ResourceOption,
) (*compute.Network, *compute.BackendService, error) {

	gcpNetwork, err := createVPC(ctx, projectConfig)
	if err != nil {
		return nil, nil, err
	}
	gcpBackendService, err := createLoadBalancerBackendService(ctx, projectConfig, opts...)
	if err != nil {
		return nil, nil, err
	}
	gcpGlobalAddress, err := createLoadBalancerStaticIP(ctx, projectConfig, opts...)
	if err != nil {
		return nil, nil, err
	}
	if projectConfig.SSL {
		err = configureSSLCertificate(ctx, projectConfig, gcpBackendService, gcpGlobalAddress, opts...)
		if err != nil {
			return nil, nil, err
		}
	}
	err = createLoadBalancerURLMapHTTP(ctx, projectConfig, gcpGlobalAddress, gcpBackendService)
	if err != nil {
		return nil, nil, err
	}
	return gcpNetwork, gcpBackendService, nil
}

// CreateVPCSubnet reates a subnetwork (subnet) within the VPC.
// The subnet is created in a specific region defined by CloudRegion and the network is linked to the VPC network.
// It enables Private IP Google Access, which allows instances in the subnet to access Google APIs and services over private IPs.
func CreateVPCSubnet(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	region project.CloudRegion,
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
	return gcpSubnetwork, fmt.Errorf("failed to create subnetwork: %w", err)
}
