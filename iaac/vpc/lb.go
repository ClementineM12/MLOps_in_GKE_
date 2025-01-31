package vpc

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateLoadBalancerStaticIP creates a static IP address for a Global Load Balancer in GCP.
// It provisions a new GlobalAddress resource that will be used by the load balancer's forwarding rule.
// This IP address will be external and IPv4-based, specifically designed for load balancing.
func CreateLoadBalancerStaticIP(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	gcpDependencies []pulumi.Resource,
) (*compute.GlobalAddress, error) {

	// Create Global Load Balancer Static IP Address
	resourceName := fmt.Sprintf("%s-glb-ip-address", projectConfig.ResourceNamePrefix)
	gcpGlobalAddress, err := compute.NewGlobalAddress(ctx, resourceName, &compute.GlobalAddressArgs{
		Project:     pulumi.String(projectConfig.ProjectId),
		Name:        pulumi.String(resourceName),
		AddressType: pulumi.String("EXTERNAL"),
		IpVersion:   pulumi.String("IPV4"),
		Description: pulumi.String("Global Load Balancer - Static IP Address"),
	}, pulumi.DependsOn(gcpDependencies))
	if err != nil {
		return nil, err
	}
	ctx.Export(resourceName, gcpGlobalAddress.Address)
	// Return the Global Load Balancer IP Address
	return gcpGlobalAddress, err
}

// CreateLoadBalancerBackendService sets up the backend service for a Global Load Balancer.
// It creates a backend service that will handle incoming traffic routed by the load balancer.
// This function also sets up the health checks used by the backend service.
func CreateLoadBalancerBackendService(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	gcpDependencies []pulumi.Resource,
) (*compute.BackendService, error) {

	gcpGLBTCPHealthCheck, err := createLoadBalancerTCPHealthChecks(ctx, projectConfig, gcpDependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to create Load Balancer Health check: %w", err)
	}
	gcpBackendService, err := createLoadbalancerBackendService(ctx, projectConfig, gcpGLBTCPHealthCheck)
	if err != nil {
		return nil, fmt.Errorf("failed to create Load Balancer Backend Service: %w", err)
	}

	return gcpBackendService, nil
}

// createLoadBalancerTCPHealthChecks creates the TCP health checks that the Global Load Balancer will use
// to verify the health of backend services. The health checks monitor the health of services
// through TCP connections on a specified port.
func createLoadBalancerTCPHealthChecks(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	gcpDependencies []pulumi.Resource,
) (*compute.HealthCheck, error) {

	resourceName := fmt.Sprintf("%s-glb-tcp-hc", projectConfig.ResourceNamePrefix)
	gcpGLBTCPHealthCheck, err := compute.NewHealthCheck(ctx, resourceName, &compute.HealthCheckArgs{
		Project:          pulumi.String(projectConfig.ProjectId),
		CheckIntervalSec: pulumi.Int(1),
		Description:      pulumi.String("TCP Health Check"),
		HealthyThreshold: pulumi.Int(4),
		TcpHealthCheck: &compute.HealthCheckTcpHealthCheckArgs{
			Port:        pulumi.Int(80),
			ProxyHeader: pulumi.String("NONE"),
		},
		TimeoutSec:         pulumi.Int(1),
		UnhealthyThreshold: pulumi.Int(5),
	}, pulumi.DependsOn(gcpDependencies))

	return gcpGLBTCPHealthCheck, err
}

// createLoadbalancerBackendService creates a backend service for a Global Load Balancer.
// The backend service defines the behavior of the load balancer, such as connection draining, health checks, and other settings.
// It does not directly handle traffic but instead controls how traffic is distributed to backend instances or groups.
func createLoadbalancerBackendService(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	gcpGLBTCPHealthCheck *compute.HealthCheck,
) (*compute.BackendService, error) {

	var backendServiceBackendArray = compute.BackendServiceBackendArray{}

	resourceName := fmt.Sprintf("%s-glb-bs", projectConfig.ResourceNamePrefix)
	gcpBackendService, err := compute.NewBackendService(ctx, resourceName, &compute.BackendServiceArgs{
		Project:     pulumi.String(projectConfig.ProjectId),
		Name:        pulumi.String(fmt.Sprintf("%s-bes", projectConfig.ResourceNamePrefix)),
		Description: pulumi.String("Global Load Balancer - Backend Service"),
		CdnPolicy: &compute.BackendServiceCdnPolicyArgs{
			ClientTtl:  pulumi.Int(5),
			DefaultTtl: pulumi.Int(5),
			MaxTtl:     pulumi.Int(5),
		},
		ConnectionDrainingTimeoutSec: pulumi.Int(10),
		Backends:                     backendServiceBackendArray,
		HealthChecks:                 gcpGLBTCPHealthCheck.ID(),
	})
	return gcpBackendService, err
}

// CreateLoadBalancerURLMapHTTP creates a URL map for HTTP traffic for a Global Load Balancer.
// The URL map is responsible for mapping HTTP requests to the appropriate backend services based on the requested URL, host, and path.
// It also sets up any necessary routing rules, including redirects and path-based routing.
func CreateLoadBalancerURLMapHTTP(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	gcpGlobalAddress *compute.GlobalAddress,
	gcpBackendService *compute.BackendService,
) error {

	var gcpGLBURLMapHTTP *compute.URLMap
	var err error

	if projectConfig.Domain == "" {
		gcpGLBURLMapHTTP, err = createLoadBalancerURLMapHTTPWithNoDomain(ctx, projectConfig, gcpBackendService)
		if err != nil {
			return fmt.Errorf("failed to create Load Balancer URL HTTP Map [ No Domain ]: %w", err)
		}
	} else {
		gcpGLBURLMapHTTP, err = createLoadBalancerURLMapHTTPWithDomain(ctx, projectConfig, gcpBackendService)
		if err != nil {
			return fmt.Errorf("failed to create Load Balancer URL HTTP Map [ Domain ]: %w", err)
		}
	}
	gcpGLBTargetHTTPProxy, err := createLoadBalancerHTTPProxy(ctx, projectConfig, gcpGLBURLMapHTTP)
	if err != nil {
		return fmt.Errorf("failed to create Load Balancer HTTP Proxy: %w", err)
	}
	err = createLoadBalancerForwardingRule(ctx, projectConfig, gcpGlobalAddress, gcpGLBTargetHTTPProxy, nil)
	if err != nil {
		return fmt.Errorf("failed to create Load Balancer HTTP Forwarding rule: %w", err)
	}
	return nil
}

// createLoadBalancerURLMapHTTPWithNoDomain creates a URL map for HTTP traffic with no domain specified.
// This URL map is used when no specific domain is provided, and all traffic is forwarded to the default backend service.
func createLoadBalancerURLMapHTTPWithNoDomain(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	gcpBackendService *compute.BackendService,
) (*compute.URLMap, error) {

	resourceName := fmt.Sprintf("%s-glb-url-map-http-no-domain", projectConfig.ResourceNamePrefix)
	gcpGLBURLMapHTTP, err := compute.NewURLMap(ctx, resourceName, &compute.URLMapArgs{
		Project:        pulumi.String(projectConfig.ProjectId),
		Name:           pulumi.String(fmt.Sprintf("%s-glb-urlmap-http", projectConfig.ResourceNamePrefix)),
		Description:    pulumi.String("Global Load Balancer - HTTP URL Map"),
		DefaultService: gcpBackendService.SelfLink,
	})
	return gcpGLBURLMapHTTP, err
}

// createLoadBalancerURLMapHTTPWithDomain creates a URL map for HTTP traffic with a specific domain.
// The URL map routes traffic based on the provided domain and any additional path-based routing rules.
func createLoadBalancerURLMapHTTPWithDomain(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	gcpBackendService *compute.BackendService,
) (*compute.URLMap, error) {

	resourceName := fmt.Sprintf("%s-glb-url-map-http-domain", projectConfig.ResourceNamePrefix)
	gcpGLBURLMapHTTP, err := compute.NewURLMap(ctx, resourceName, &compute.URLMapArgs{
		Project:     pulumi.String(projectConfig.ProjectId),
		Name:        pulumi.String(fmt.Sprintf("%s-glb-urlmap-http", projectConfig.ResourceNamePrefix)),
		Description: pulumi.String("Global Load Balancer - HTTP URL Map"),
		HostRules: &compute.URLMapHostRuleArray{
			&compute.URLMapHostRuleArgs{
				Hosts: pulumi.StringArray{
					pulumi.String(projectConfig.Domain),
				},
				PathMatcher: pulumi.String("all-paths"),
				Description: pulumi.String("Default Route All Paths"),
			},
		},
		PathMatchers: &compute.URLMapPathMatcherArray{
			&compute.URLMapPathMatcherArgs{
				Name:           pulumi.String("all-paths"),
				DefaultService: gcpBackendService.SelfLink,
				PathRules: &compute.URLMapPathMatcherPathRuleArray{
					&compute.URLMapPathMatcherPathRuleArgs{
						Paths: pulumi.StringArray{
							pulumi.String("/*"),
						},
						UrlRedirect: &compute.URLMapPathMatcherPathRuleUrlRedirectArgs{
							StripQuery: pulumi.Bool(false),
							// If Domain Configured and SSL Enabled
							HttpsRedirect: pulumi.Bool(projectConfig.SSL),
						},
					},
				},
			},
		},
		DefaultService: gcpBackendService.SelfLink,
	})
	return gcpGLBURLMapHTTP, err
}

// createLoadBalancerHTTPProxy creates a target HTTP proxy for the Global Load Balancer.
// The HTTP proxy is used to route HTTP requests to the appropriate URL map for further processing.
func createLoadBalancerHTTPProxy(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	gcpGLBURLMapHTTP *compute.URLMap,
) (*compute.TargetHttpProxy, error) {

	resourceName := fmt.Sprintf("%s-glb-http-proxy", projectConfig.ResourceNamePrefix)
	gcpGLBTargetHTTPProxy, err := compute.NewTargetHttpProxy(ctx, resourceName, &compute.TargetHttpProxyArgs{
		Project: pulumi.String(projectConfig.ProjectId),
		Name:    pulumi.String(resourceName),
		UrlMap:  gcpGLBURLMapHTTP.SelfLink,
	})
	return gcpGLBTargetHTTPProxy, err
}
