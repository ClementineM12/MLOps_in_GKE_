package flyte

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/servicenetworking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createServiceNetworking(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	network *compute.Network,
) (*compute.GlobalAddress, error) {

	resourceName := fmt.Sprintf("%s-db-internal-address", projectConfig.ResourceNamePrefix)
	return compute.NewGlobalAddress(ctx, resourceName, &compute.GlobalAddressArgs{
		Name:         pulumi.String("cloudsql-vpc-service-networking"),
		Purpose:      pulumi.String("VPC_PEERING"),
		AddressType:  pulumi.String("INTERNAL"),
		PrefixLength: pulumi.Int(16),
		Network:      network.SelfLink,
	})
}

func createServiceNetworkingConnection(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	network *compute.Network,
	serviceNetworking *compute.GlobalAddress,
) (*servicenetworking.Connection, error) {

	resourceName := fmt.Sprintf("%s-db-netowrk", projectConfig.ResourceNamePrefix)
	return servicenetworking.NewConnection(ctx, resourceName, &servicenetworking.ConnectionArgs{
		Network:               network.ID(),
		Service:               pulumi.String("servicenetworking.googleapis.com"),
		ReservedPeeringRanges: pulumi.StringArray{serviceNetworking.Name},
	}, pulumi.DependsOn([]pulumi.Resource{serviceNetworking}))
}
