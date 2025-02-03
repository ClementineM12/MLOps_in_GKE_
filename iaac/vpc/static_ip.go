package vpc

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	GlobalAddressType      = "EXTERNAL"
	GlobalAddressIPVersion = "IPV4"
)

// createLoadBalancerStaticIP creates a static IP address for a Global Load Balancer in GCP.
// It provisions a new GlobalAddress resource that will be used by the load balancer's forwarding rule.
// This IP address will be external and IPv4-based, specifically designed for load balancing.
func createLoadBalancerStaticIP(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	opts ...pulumi.ResourceOption,
) (*compute.GlobalAddress, error) {

	resourceName := fmt.Sprintf("%s-glb-ip-address", projectConfig.ResourceNamePrefix)
	gcpGlobalAddress, err := compute.NewGlobalAddress(ctx, resourceName, &compute.GlobalAddressArgs{
		Project:     pulumi.String(projectConfig.ProjectId),
		Name:        pulumi.String(resourceName),
		AddressType: pulumi.String(GlobalAddressType),
		IpVersion:   pulumi.String(GlobalAddressIPVersion),
		Description: pulumi.String("Global Load Balancer - Static IP Address"),
	}, opts...)
	if err != nil {
		return nil, err
	}
	ctx.Export(resourceName, gcpGlobalAddress.Address)

	return gcpGlobalAddress, err
}
