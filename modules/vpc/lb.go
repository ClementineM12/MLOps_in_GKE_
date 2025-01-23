package vpc

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateGlobalLoadBalancerStaticIP creates the Global Load Balancer Static IP Address
func CreateLoadBalancerStaticIP(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpDependencies []pulumi.Resource,
) (*compute.GlobalAddress, error) {

	// Create Global Load Balancer Static IP Address
	resourceName := fmt.Sprintf("%s-glb-ip-address", resourceNamePrefix)
	gcpGlobalAddress, err := compute.NewGlobalAddress(ctx, resourceName, &compute.GlobalAddressArgs{
		Project:     pulumi.String(gcpProjectId),
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

// CreateLoadBalancerBackendService
func CreateLoadBalancerBackendService(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpDependencies []pulumi.Resource,
) (*compute.BackendService, error) {

	gcpGLBTCPHealthCheck, err := createLoadBalancerTCPHealthChecks(ctx, resourceNamePrefix, gcpProjectId, gcpDependencies)
	if err != nil {
		return nil, err
	}
	gcpBackendService, err := createLoadbalancerBackendService(ctx, resourceNamePrefix, gcpProjectId, gcpGLBTCPHealthCheck)

	return gcpBackendService, err
}

// createLoadBalancerTCPHealthChecks creates Health Checks (Network Endpoints within Load Balancer)
func createLoadBalancerTCPHealthChecks(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpDependencies []pulumi.Resource,
) (*compute.HealthCheck, error) {

	resourceName := fmt.Sprintf("%s-glb-tcp-hc", resourceNamePrefix)
	gcpGLBTCPHealthCheck, err := compute.NewHealthCheck(ctx, resourceName, &compute.HealthCheckArgs{
		Project:          pulumi.String(gcpProjectId),
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

// createLoadbalancerBackendService creates a Load Balancer Backend Service
func createLoadbalancerBackendService(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpGLBTCPHealthCheck *compute.HealthCheck,
) (*compute.BackendService, error) {

	var backendServiceBackendArray = compute.BackendServiceBackendArray{}

	resourceName := fmt.Sprintf("%s-glb-bes", resourceNamePrefix)
	gcpBackendService, err := compute.NewBackendService(ctx, resourceName, &compute.BackendServiceArgs{
		Project:     pulumi.String(gcpProjectId),
		Name:        pulumi.String(fmt.Sprintf("%s-bes", resourceNamePrefix)),
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

func CreateLoadBalancerURLMapHTTP(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	domain string,
	SSL bool,
	gcpGlobalAddress *compute.GlobalAddress,
	gcpBackendService *compute.BackendService,
) error {

	var gcpGLBURLMapHTTP *compute.URLMap
	var err error

	if domain == "" {
		gcpGLBURLMapHTTP, err = createLoadBalancerURLMapHTTPWithNoDomain(ctx, resourceNamePrefix, gcpProjectId, gcpBackendService)
		if err != nil {
			return nil
		}
	} else {
		gcpGLBURLMapHTTP, err = createLoadBalancerURLMapHTTPWithDomain(ctx, resourceNamePrefix, gcpProjectId, domain, SSL, gcpBackendService)
		if err != nil {
			return nil
		}
	}
	gcpGLBTargetHTTPProxy, err := createLoadBalancerHTTPProxy(ctx, resourceNamePrefix, gcpProjectId, gcpGLBURLMapHTTP)
	if err != nil {
		return err
	}
	err = createLoadBalancerHTTPRorwardingRule(ctx, resourceNamePrefix, gcpProjectId, gcpGlobalAddress, gcpGLBTargetHTTPProxy)
	return err
}

func createLoadBalancerURLMapHTTPWithNoDomain(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpBackendService *compute.BackendService,
) (*compute.URLMap, error) {

	resourceName := fmt.Sprintf("%s-glb-url-map-http-no-domain", resourceNamePrefix)
	gcpGLBURLMapHTTP, err := compute.NewURLMap(ctx, resourceName, &compute.URLMapArgs{
		Project:        pulumi.String(gcpProjectId),
		Name:           pulumi.String(fmt.Sprintf("%s-glb-urlmap-http", resourceNamePrefix)),
		Description:    pulumi.String("Global Load Balancer - HTTP URL Map"),
		DefaultService: gcpBackendService.SelfLink,
	})
	return gcpGLBURLMapHTTP, err
}

func createLoadBalancerURLMapHTTPWithDomain(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	domain string,
	SSL bool,
	gcpBackendService *compute.BackendService,
) (*compute.URLMap, error) {

	resourceName := fmt.Sprintf("%s-glb-url-map-http-domain", resourceNamePrefix)
	gcpGLBURLMapHTTP, err := compute.NewURLMap(ctx, resourceName, &compute.URLMapArgs{
		Project:     pulumi.String(gcpProjectId),
		Name:        pulumi.String(fmt.Sprintf("%s-glb-urlmap-http", resourceNamePrefix)),
		Description: pulumi.String("Global Load Balancer - HTTP URL Map"),
		HostRules: &compute.URLMapHostRuleArray{
			&compute.URLMapHostRuleArgs{
				Hosts: pulumi.StringArray{
					pulumi.String(domain),
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
							HttpsRedirect: pulumi.Bool(SSL),
						},
					},
				},
			},
		},
		DefaultService: gcpBackendService.SelfLink,
	})
	return gcpGLBURLMapHTTP, err
}

// Create Target HTTP Proxy
func createLoadBalancerHTTPProxy(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpGLBURLMapHTTP *compute.URLMap,
) (*compute.TargetHttpProxy, error) {

	resourceName := fmt.Sprintf("%s-glb-http-proxy", resourceNamePrefix)
	gcpGLBTargetHTTPProxy, err := compute.NewTargetHttpProxy(ctx, resourceName, &compute.TargetHttpProxyArgs{
		Project: pulumi.String(gcpProjectId),
		Name:    pulumi.String(resourceName),
		UrlMap:  gcpGLBURLMapHTTP.SelfLink,
	})
	return gcpGLBTargetHTTPProxy, err
}

// Create HTTP Global Forwarding Rule
func createLoadBalancerHTTPRorwardingRule(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpGlobalAddress *compute.GlobalAddress,
	gcpGLBTargetHTTPProxy *compute.TargetHttpProxy,
) error {

	resourceName := fmt.Sprintf("%s-glb-http-fwd-rule", resourceNamePrefix)
	_, err := compute.NewGlobalForwardingRule(ctx, resourceName, &compute.GlobalForwardingRuleArgs{
		Project:             pulumi.String(gcpProjectId),
		Target:              gcpGLBTargetHTTPProxy.SelfLink,
		IpAddress:           gcpGlobalAddress.SelfLink,
		PortRange:           pulumi.String("80"),
		LoadBalancingScheme: pulumi.String("EXTERNAL"),
	})
	return err
}
