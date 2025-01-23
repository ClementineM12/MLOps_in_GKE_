package vpc

import (
	"fmt"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ConfigureSSLCertificate configures the SSL certificate and related resources for a Global Load Balancer (GLB) with HTTPS support.
// It creates a managed SSL certificate, a URL map for HTTPS traffic, a Target HTTPS Proxy, and a Global Forwarding Rule for HTTPS traffic.
// It ensures all necessary dependencies for the resources are set, including the SSL certificate and backend service.
func ConfigureSSLCertificate(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	domain string,
	gcpBackendService *compute.BackendService,
	gcpGlobalAddress *compute.GlobalAddress,
	gcpDependencies []pulumi.Resource,
) error {

	gcpGLBManagedSSLCert, err := createManagedSSLCertificate(ctx, resourceNamePrefix, gcpProjectId, domain, gcpDependencies)
	if err != nil {
		return fmt.Errorf("failed to create Managed SSL Certificate: %w", err)
	}
	gcpGLBURLMapHTTPS, err := createLoadbalancerURLMapHTTPS(ctx, resourceNamePrefix, gcpProjectId, gcpBackendService)
	if err != nil {
		return fmt.Errorf("failed to create Load Balancer URL Map HTPS: %w", err)
	}
	gcpGLBTargetHTTPSProxy, err := createLoadbalancerHTTPSProxy(ctx, resourceNamePrefix, gcpProjectId, gcpGLBURLMapHTTPS, gcpGLBManagedSSLCert)
	if err != nil {
		return fmt.Errorf("failed to create Load Balancer HTPS Proxy: %w", err)
	}
	err = createLoadBalancerForwardingRule(ctx, resourceNamePrefix, gcpProjectId, gcpGlobalAddress, nil, gcpGLBTargetHTTPSProxy)
	if err != nil {
		return fmt.Errorf("failed to create Load Balancer HTPS Forwarding rule: %w", err)
	}
	return nil
}

// createManagedSSLCertificate creates a Managed SSL Certificate resource for use with the Global Load Balancer.
// It takes in the domain name and project details, and sets up the SSL certificate for HTTPS traffic handling.
// This function requires the configuration of dependencies like the backend service to ensure it is correctly created before being used in the load balancer setup.
func createManagedSSLCertificate(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	domain string,
	gcpDependencies []pulumi.Resource,
) (*compute.ManagedSslCertificate, error) {

	resourceName := fmt.Sprintf("%s-glb-ssl-cert", resourceNamePrefix)
	gcpGLBManagedSSLCert, err := compute.NewManagedSslCertificate(ctx, resourceName, &compute.ManagedSslCertificateArgs{
		Project:     pulumi.String(gcpProjectId),
		Name:        pulumi.String(resourceName),
		Description: pulumi.String("Global Load Balancer - Managed SSL Certificate"),
		Type:        pulumi.String("MANAGED"),
		Managed: &compute.ManagedSslCertificateManagedArgs{
			Domains: pulumi.StringArray{
				pulumi.String(domain),
			},
		},
	}, pulumi.DependsOn(gcpDependencies))
	return gcpGLBManagedSSLCert, err
}

// createLoadbalancerURLMapHTTPS creates a URL Map resource for handling HTTPS traffic within a Global Load Balancer setup.
// The URL Map routes incoming requests to the appropriate backend service based on the URL pattern and the domain specified.
// This function helps define the routing behavior for SSL/TLS traffic, ensuring secure requests are directed to the correct backend.
func createLoadbalancerURLMapHTTPS(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpBackendService *compute.BackendService,
) (*compute.URLMap, error) {

	resourceName := fmt.Sprintf("%s-glb-url-map-https-domain", resourceNamePrefix)
	gcpGLBURLMapHTTPS, err := compute.NewURLMap(ctx, resourceName, &compute.URLMapArgs{
		Project:        pulumi.String(gcpProjectId),
		Name:           pulumi.String(fmt.Sprintf("%s-glb-urlmap-https", resourceNamePrefix)),
		Description:    pulumi.String("Global Load Balancer - HTTPS URL Map"),
		DefaultService: gcpBackendService.SelfLink,
	})
	return gcpGLBURLMapHTTPS, err
}

// createLoadbalancerHTTPSProxy creates a Target HTTPS Proxy resource to handle incoming HTTPS requests.
// The proxy uses the previously created URL Map and Managed SSL Certificate to process secure traffic.
// The HTTPS Proxy is essential for directing traffic securely through the Global Load Balancer and to the proper backend service.
func createLoadbalancerHTTPSProxy(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpGLBURLMapHTTPS *compute.URLMap,
	gcpGLBManagedSSLCert *compute.ManagedSslCertificate,
) (*compute.TargetHttpsProxy, error) {

	resourceName := fmt.Sprintf("%s-glb-https-proxy", resourceNamePrefix)
	gcpGLBTargetHTTPSProxy, err := compute.NewTargetHttpsProxy(ctx, resourceName, &compute.TargetHttpsProxyArgs{
		Project: pulumi.String(gcpProjectId),
		Name:    pulumi.String(resourceName),
		UrlMap:  gcpGLBURLMapHTTPS.SelfLink,
		SslCertificates: pulumi.StringArray{
			gcpGLBManagedSSLCert.SelfLink,
		},
	})
	return gcpGLBTargetHTTPSProxy, err
}
