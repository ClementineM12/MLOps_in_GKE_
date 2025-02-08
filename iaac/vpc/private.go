package vpc

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createCloudRouter(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	region global.CloudRegion,
	networkID pulumi.StringInput,
) (*compute.Router, error) {

	routerName := fmt.Sprintf("%s-cloud-router", projectConfig.ResourceNamePrefix)
	cloudRouter, err := compute.NewRouter(ctx, routerName, &compute.RouterArgs{
		Name:    pulumi.String(routerName),
		Network: networkID,
		Region:  pulumi.String(region.Region),
		Project: pulumi.String(projectConfig.ProjectId),
	})
	if err != nil {
		return nil, err
	}
	return cloudRouter, nil
}

func createCloudNAT(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	region global.CloudRegion,
	router *compute.Router,
) error {

	natName := fmt.Sprintf("%s-cloud-nat", projectConfig.ResourceNamePrefix)
	_, err := compute.NewRouterNat(ctx, natName, &compute.RouterNatArgs{
		Name:                          pulumi.String(natName),
		Router:                        router.Name,
		Region:                        pulumi.String(region.Region),
		Project:                       pulumi.String(projectConfig.ProjectId),
		NatIpAllocateOption:           pulumi.String("AUTO_ONLY"), // Automatically allocate external IPs
		SourceSubnetworkIpRangesToNat: pulumi.String("ALL_SUBNETWORKS_ALL_IP_RANGES"),
	})
	if err != nil {
		return err
	}
	return nil
}
