package vpc

// Package vpc provides functionality for configuring HTTP traffic management within a Global Load Balancer (GLB)
// in Google Cloud Platform (GCP) using Pulumi. This package defines the necessary components for handling HTTP requests,
// including URL mappings, proxies, and forwarding rules.
//
// The resources created in this package include:
//
// 1. **HTTP URL Map (`createLoadBalancerURLMapHTTP`)**:
//    - Defines the URL routing rules for HTTP traffic within the Global Load Balancer.
//    - Supports two configurations:
//      ✅ **With a domain (`createLoadBalancerURLMapHTTPWithDomain`)**: Routes traffic based on a specific domain name and path rules.
//      ✅ **Without a domain (`createLoadBalancerURLMapHTTPWithNoDomain`)**: Routes all HTTP traffic to a default backend service.
//    - Ensures correct request redirection and forwarding, including optional HTTPS redirection.
//
// 2. **Target HTTP Proxy (`createLoadBalancerHTTPProxy`)**:
//    - Acts as an intermediary between clients and the backend service by forwarding HTTP traffic.
//    - Works with the **URL Map** to determine how incoming requests should be processed.
//    - Essential for directing HTTP traffic within the Global Load Balancer.
//
// 3. **HTTP Forwarding Rule (`createLoadBalancerForwardingRule`)**:
//    - Configures a **public IP entry point** for HTTP requests to the Load Balancer.
//    - Routes external HTTP requests to the **Target HTTP Proxy**.
//    - Ensures traffic is correctly forwarded based on URL mappings.
//
// These resources work together to provide **scalable, resilient, and well-structured HTTP traffic management** within Google Cloud.
// The implementation also allows for optional HTTPS redirection when a domain and SSL certificate are provided.

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createLoadBalancerURLMapHTTP creates a URL map for HTTP traffic for a Global Load Balancer.
// The URL map is responsible for mapping HTTP requests to the appropriate backend services based on the requested URL, host, and path.
// It also sets up any necessary routing rules, including redirects and path-based routing.
func createLoadBalancerURLMapHTTP(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
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
	err = createLoadBalancerForwardingRule(ctx, projectConfig, gcpGlobalAddress, gcpGLBTargetHTTPProxy, nil) // utils
	if err != nil {
		return fmt.Errorf("failed to create Load Balancer HTTP Forwarding rule: %w", err)
	}
	return nil
}

// createLoadBalancerURLMapHTTPWithNoDomain creates a URL map for HTTP traffic with no domain specified.
// This URL map is used when no specific domain is provided, and all traffic is forwarded to the default backend service.
func createLoadBalancerURLMapHTTPWithNoDomain(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
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
	projectConfig global.ProjectConfig,
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
	projectConfig global.ProjectConfig,
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
