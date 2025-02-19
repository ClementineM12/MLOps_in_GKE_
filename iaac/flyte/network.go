package flyte

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/servicenetworking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createServiceNetworking(
	ctx *pulumi.Context,
	network *compute.Network,
) (*compute.GlobalAddress, error) {

	return compute.NewGlobalAddress(ctx, "service-networking", &compute.GlobalAddressArgs{
		Name:         pulumi.String("cloudsql-vpc-service-networking"),
		Purpose:      pulumi.String("VPC_PEERING"),
		AddressType:  pulumi.String("INTERNAL"),
		PrefixLength: pulumi.Int(16),
		Network:      network.SelfLink,
	})
}

func createServiceNetworkingConnection(
	ctx *pulumi.Context,
	network *compute.Network,
	serviceNetworking *compute.GlobalAddress,
) (*servicenetworking.Connection, error) {

	return servicenetworking.NewConnection(ctx, "default", &servicenetworking.ConnectionArgs{
		Network:               network.ID(),
		Service:               pulumi.String("servicenetworking.googleapis.com"),
		ReservedPeeringRanges: pulumi.StringArray{serviceNetworking.Name},
	}, pulumi.DependsOn([]pulumi.Resource{serviceNetworking}))
}
