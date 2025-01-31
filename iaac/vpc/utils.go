package vpc

import (
	"fmt"
	"mlops/project"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// // createFirewallRule is used to create a firewall rule
// func createFirewallRule(
// 	ctx *pulumi.Context,
// 	resourceNamePrefix string,
// 	resourceNameSuffix string,
// 	gcpProjectId string,
// 	gcpNetwork *compute.Network,
// 	description string,
// 	sourceRanges pulumi.StringArray,
// 	targetTags pulumi.StringArray,
// 	allowedPorts []string,
// ) error {
// 	resourceName := fmt.Sprintf("%s-fw-%s", resourceNamePrefix, resourceNameSuffix)

// 	// Prepare the allowed ports for the firewall rule
// 	var allowedFirewallArgs compute.FirewallAllowArray
// 	for _, port := range allowedPorts {
// 		allowedFirewallArgs = append(allowedFirewallArgs, &compute.FirewallAllowArgs{
// 			Protocol: pulumi.String("tcp"),
// 			Ports:    pulumi.StringArray{pulumi.String(port)},
// 		})
// 	}

// 	_, err := compute.NewFirewall(ctx, resourceName, &compute.FirewallArgs{
// 		Project:      pulumi.String(gcpProjectId),
// 		Name:         pulumi.String(resourceName),
// 		Description:  pulumi.String(description),
// 		Network:      gcpNetwork.Name,
// 		Allows:       allowedFirewallArgs,
// 		SourceRanges: sourceRanges,
// 		TargetTags:   targetTags,
// 	})

// 	return err
// }

func createLoadBalancerForwardingRule(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	gcpGlobalAddress *compute.GlobalAddress,
	gcpGLBTargetHTTPProxy *compute.TargetHttpProxy,
	gcpGLBTargetHTTPSProxy *compute.TargetHttpsProxy,
) error {
	// Declare variables outside the if-else block
	var (
		port                    string
		protocol                string
		targetHTTPProxySelfLink pulumi.StringInput
	)

	if gcpGLBTargetHTTPProxy != nil {
		port = "80"
		protocol = "HTTP"
		targetHTTPProxySelfLink = gcpGLBTargetHTTPProxy.SelfLink
	} else {
		port = "443"
		protocol = "HTTPS"
		targetHTTPProxySelfLink = gcpGLBTargetHTTPSProxy.SelfLink
	}

	resourceName := fmt.Sprintf("%s-glb-%s-fwd-rule", projectConfig.ResourceNamePrefix, strings.ToLower(protocol))

	// Create the GlobalForwardingRule resource
	_, err := compute.NewGlobalForwardingRule(ctx, resourceName, &compute.GlobalForwardingRuleArgs{
		Project:             pulumi.String(projectConfig.ProjectId),
		Target:              targetHTTPProxySelfLink,
		IpAddress:           gcpGlobalAddress.SelfLink,
		PortRange:           pulumi.String(port),
		LoadBalancingScheme: pulumi.String("EXTERNAL"),
	})
	return err
}
