package vpc

// Package vpc provides functionality for configuring SSL/TLS encryption and HTTPS traffic handling
// for a Global Load Balancer (GLB) in Google Cloud Platform (GCP) using Pulumi.
// It includes the creation of a managed SSL certificate, URL mapping, HTTPS proxy, and forwarding rules
// to securely route traffic to backend services.
//
// The resources created in this package include:
//
// 1. **Managed SSL Certificate (`createManagedSSLCertificate`)**:
//    - Generates a **Google-managed SSL certificate** for HTTPS traffic on the Global Load Balancer.
//    - Automatically provisions and renews SSL certificates for the specified domain.
//    - Ensures secure communication by encrypting data in transit.
//
// 2. **HTTPS URL Map (`createLoadbalancerURLMapHTTPS`)**:
//    - Defines **URL routing rules** for the HTTPS Global Load Balancer.
//    - Ensures requests are forwarded to the appropriate backend service based on URL patterns.
//    - Directs all traffic securely to the backend service.
//
// 3. **Target HTTPS Proxy (`createLoadbalancerHTTPSProxy`)**:
//    - Acts as an intermediary that processes **HTTPS requests** before forwarding them to backend services.
//    - Uses the **managed SSL certificate** to terminate SSL and establish a secure connection.
//    - Ensures traffic is encrypted when reaching the load balancer.
//
// 4. **HTTPS Forwarding Rule (`createLoadBalancerForwardingRule`)**:
//    - Defines the **entry point** for HTTPS traffic in the Global Load Balancer.
//    - Routes external HTTPS requests to the Target HTTPS Proxy.
//    - Ensures that clients always connect using secure HTTPS.
//
// These resources work together to enable **secure, scalable, and highly available** HTTPS traffic
// management for applications running in Google Cloud. By integrating **Google-managed SSL certificates**,
// the setup ensures **automatic certificate renewal**, reducing operational overhead and improving security.

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// configureSSLCertificate configures the SSL certificate and related resources for a Global Load Balancer (GLB) with HTTPS support.
// It creates a managed SSL certificate, a URL map for HTTPS traffic, a Target HTTPS Proxy, and a Global Forwarding Rule for HTTPS traffic.
// It ensures all necessary dependencies for the resources are set, including the SSL certificate and backend service.
func configureSSLCertificate(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	gcpBackendService *compute.BackendService,
	gcpGlobalAddress *compute.GlobalAddress,
) error {

	gcpGLBManagedSSLCert, err := createManagedSSLCertificate(ctx, projectConfig)
	if err != nil {
		return fmt.Errorf("failed to create Managed SSL Certificate: %w", err)
	}
	gcpGLBURLMapHTTPS, err := createLoadbalancerURLMapHTTPS(ctx, projectConfig, gcpBackendService)
	if err != nil {
		return fmt.Errorf("failed to create Load Balancer URL Map HTPS: %w", err)
	}
	gcpGLBTargetHTTPSProxy, err := createLoadbalancerHTTPSProxy(ctx, projectConfig, gcpGLBURLMapHTTPS, gcpGLBManagedSSLCert)
	if err != nil {
		return fmt.Errorf("failed to create Load Balancer HTPS Proxy: %w", err)
	}
	err = createLoadBalancerForwardingRule(ctx, projectConfig, gcpGlobalAddress, nil, gcpGLBTargetHTTPSProxy) // utils
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
	projectConfig global.ProjectConfig,
) (*compute.ManagedSslCertificate, error) {

	resourceName := fmt.Sprintf("%s-glb-ssl-cert", projectConfig.ResourceNamePrefix)
	gcpGLBManagedSSLCert, err := compute.NewManagedSslCertificate(ctx, resourceName, &compute.ManagedSslCertificateArgs{
		Project:     pulumi.String(projectConfig.ProjectId),
		Name:        pulumi.String(resourceName),
		Description: pulumi.String("Global Load Balancer - Managed SSL Certificate"),
		Type:        pulumi.String("MANAGED"),
		Managed: &compute.ManagedSslCertificateManagedArgs{
			Domains: pulumi.StringArray{
				pulumi.String(projectConfig.Domain), // Uses the Domain provided
			},
		},
	})
	return gcpGLBManagedSSLCert, err
}

// createLoadbalancerURLMapHTTPS creates a URL Map resource for handling HTTPS traffic within a Global Load Balancer setup.
// The URL Map routes incoming requests to the appropriate backend service based on the URL pattern and the domain specified.
// This function helps define the routing behavior for SSL/TLS traffic, ensuring secure requests are directed to the correct backend.
func createLoadbalancerURLMapHTTPS(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	gcpBackendService *compute.BackendService,
) (*compute.URLMap, error) {

	resourceName := fmt.Sprintf("%s-glb-url-map-https-domain", projectConfig.ResourceNamePrefix)
	gcpGLBURLMapHTTPS, err := compute.NewURLMap(ctx, resourceName, &compute.URLMapArgs{
		Project:        pulumi.String(projectConfig.ProjectId),
		Name:           pulumi.String(fmt.Sprintf("%s-glb-urlmap-https", projectConfig.ResourceNamePrefix)),
		Description:    pulumi.String("Global Load Balancer - HTTPS URL Map"),
		DefaultService: gcpBackendService.SelfLink, // Points to the Backend service
	}, pulumi.DependsOn([]pulumi.Resource{gcpBackendService}))
	return gcpGLBURLMapHTTPS, err
}

// createLoadbalancerHTTPSProxy creates a Target HTTPS Proxy resource to handle incoming HTTPS requests.
// The proxy uses the previously created URL Map and Managed SSL Certificate to process secure traffic.
// The HTTPS Proxy is essential for directing traffic securely through the Global Load Balancer and to the proper backend service.
func createLoadbalancerHTTPSProxy(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	gcpGLBURLMapHTTPS *compute.URLMap,
	gcpGLBManagedSSLCert *compute.ManagedSslCertificate,
) (*compute.TargetHttpsProxy, error) {

	resourceName := fmt.Sprintf("%s-glb-https-proxy", projectConfig.ResourceNamePrefix)
	gcpGLBTargetHTTPSProxy, err := compute.NewTargetHttpsProxy(ctx, resourceName, &compute.TargetHttpsProxyArgs{
		Project: pulumi.String(projectConfig.ProjectId),
		Name:    pulumi.String(resourceName),
		UrlMap:  gcpGLBURLMapHTTPS.SelfLink, // Routing
		SslCertificates: pulumi.StringArray{
			gcpGLBManagedSSLCert.SelfLink, // Uses the Managed SSL Cert
		},
	}, pulumi.DependsOn([]pulumi.Resource{gcpGLBManagedSSLCert, gcpGLBURLMapHTTPS}))
	return gcpGLBTargetHTTPSProxy, err
}
