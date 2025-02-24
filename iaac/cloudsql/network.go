package cloudsql

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
) ([]pulumi.Resource, error) {

	dependencies := []pulumi.Resource{}

	resourceName := fmt.Sprintf("%s-db-internal-address", projectConfig.ResourceNamePrefix)
	serviceNetworking, err := compute.NewGlobalAddress(ctx, resourceName, &compute.GlobalAddressArgs{
		Name:         pulumi.String("cloudsql-vpc-service-networking"),
		Purpose:      pulumi.String("VPC_PEERING"),
		AddressType:  pulumi.String("INTERNAL"),
		PrefixLength: pulumi.Int(16),
		Network:      network.SelfLink,
	},
		pulumi.DependsOn([]pulumi.Resource{network}))
	if err != nil {
		return dependencies, nil
	}

	resourceName = fmt.Sprintf("%s-db-network", projectConfig.ResourceNamePrefix)
	serviceNetworkConnection, err := servicenetworking.NewConnection(ctx, resourceName, &servicenetworking.ConnectionArgs{
		Network:               network.ID(),
		Service:               pulumi.String("servicenetworking.googleapis.com"),
		ReservedPeeringRanges: pulumi.StringArray{serviceNetworking.Name},
	},
		pulumi.DependsOn([]pulumi.Resource{serviceNetworking}),
	)
	if err != nil {
		return dependencies, err
	}

	dependencies = []pulumi.Resource{serviceNetworking, serviceNetworkConnection}

	return dependencies, nil
}
