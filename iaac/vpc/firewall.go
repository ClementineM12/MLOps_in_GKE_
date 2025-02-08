package vpc

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createFirewallRuleHealthChecks creates a firewall rule that allows incoming TCP traffic (ports 80, 8080, 443) for health checks used by services like load balancers.
// The allowed source ranges are from IP blocks 35.191.0.0/16 and 130.211.0.0/22, which are Google Cloud’s health check sources (https://cloud.google.com/load-balancing/docs/health-check-concepts#ip-ranges).
func createFirewallRuleHealthChecks(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
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
	projectConfig global.ProjectConfig,
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

func createFirewallEgress(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	gcpNetwork pulumi.StringInput,
) error {

	resourceName := fmt.Sprintf("%s-fw-allow-egress", projectConfig.ResourceNamePrefix)
	_, err := compute.NewFirewall(ctx, resourceName, &compute.FirewallArgs{
		Project:     pulumi.String(projectConfig.ProjectId),
		Name:        pulumi.String(resourceName),
		Description: pulumi.String("FW - Allow - Egress - Internet Access from GKE Nodes"),
		Network:     gcpNetwork,
		Direction:   pulumi.String("EGRESS"), // ✅ Allow outbound traffic
		DestinationRanges: pulumi.StringArray{
			pulumi.String("0.0.0.0/0"), // ✅ Allow access to the internet
		},
		Allows: compute.FirewallAllowArray{
			&compute.FirewallAllowArgs{
				Protocol: pulumi.String("all"), // ✅ Allow all outbound traffic
			},
		},
	})
	return err
}
