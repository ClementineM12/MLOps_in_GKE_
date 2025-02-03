package vpc

import (
	"fmt"
	"mlops/global"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createLoadBalancerForwardingRule creates a Global Forwarding Rule for directing HTTP or HTTPS traffic
// to the appropriate Target Proxy in a Global Load Balancer (GLB).
func createLoadBalancerForwardingRule(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
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
