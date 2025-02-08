package vpc

// 📌 Think of it like this:
//
// Load Balancer = The "Traffic Controller"
// Backend Service = The "Traffic Distribution System"

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// CreateVPCResources provisions a VPC network along with necessary load balancing resources.
// It sets up the network (`createVPC`), backend service, global static IP, and optionally configures SSL certificates.
// The function also creates a URL map for HTTP traffic, ensuring proper request routing.
// This enables a secure and scalable infrastructure for cloud-based applications.
func CreateVPCResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	opts ...pulumi.ResourceOption,
) (*compute.Network, *compute.BackendService, error) {

	gcpNetwork, err := createVPC(ctx, projectConfig)
	if err != nil {
		return nil, nil, err
	}

	if config.GetBool(ctx, "vpc:loadBalancer") {
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
	return gcpNetwork, nil, nil
}

func CreateVPCSubnetResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	region global.CloudRegion,
	gcpNetwork pulumi.StringInput,
) (*compute.Subnetwork, error) {
	gcpSubnetwork, err := createVPCSubnet(ctx, projectConfig, region, gcpNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnetwork: %w", err)
	}

	if config.GetBool(ctx, "gke:privateNodes") {
		cloudRouter, err := createCloudRouter(ctx, projectConfig, region, gcpNetwork)
		if err != nil {
			return nil, fmt.Errorf("failed to create router: %w", err)
		}
		err = createCloudNAT(ctx, projectConfig, region, cloudRouter)
		if err != nil {
			return nil, fmt.Errorf("failed to create NAT: %w", err)
		}
		err = createFirewallEgress(ctx, projectConfig, gcpNetwork)
		if err != nil {
			return nil, fmt.Errorf("failed to create Firewall Egress: %w", err)
		}
		return gcpSubnetwork, nil
	}
	return gcpSubnetwork, nil
}
