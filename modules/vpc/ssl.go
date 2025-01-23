package vpc

import (
	"fmt"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ConfigureSSLCertificate
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
		return err
	}
	gcpGLBURLMapHTTPS, err := createLoadbalancerURLMapHTTPS(ctx, resourceNamePrefix, gcpProjectId, gcpBackendService)
	if err != nil {
		return err
	}
	gcpGLBTargetHTTPSProxy, err := createLoadbalancerHTTPSProxy(ctx, resourceNamePrefix, gcpProjectId, gcpGLBURLMapHTTPS, gcpGLBManagedSSLCert)
	if err != nil {
		return err
	}
	err = createLoadbalancerHTTPSForwardingRule(ctx, resourceNamePrefix, gcpProjectId, gcpGLBTargetHTTPSProxy, gcpGlobalAddress)
	if err != nil {
		return err
	}
	return nil
}

// createManagedSSLCertificate creates Managed SSL Certificate
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

// createLoadbalancerURLMapHTTPS creates URL Map
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

// createLoadbalancerHTTPSProxy creates Target HTTPS Proxy
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

// createLoadbalancerForwardingRule creates a Global Load Balancer Forwarding Rule for HTTPS Traffic.
func createLoadbalancerHTTPSForwardingRule(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	gcpGLBTargetHTTPSProxy *compute.TargetHttpsProxy,
	gcpGlobalAddress *compute.GlobalAddress,
) error {

	resourceName := fmt.Sprintf("%s-glb-https-fwd-rule", resourceNamePrefix)
	_, err := compute.NewGlobalForwardingRule(ctx, resourceName, &compute.GlobalForwardingRuleArgs{
		Project:             pulumi.String(gcpProjectId),
		Target:              gcpGLBTargetHTTPSProxy.SelfLink,
		IpAddress:           gcpGlobalAddress.SelfLink,
		PortRange:           pulumi.String("443"),
		LoadBalancingScheme: pulumi.String("EXTERNAL"),
	})
	return err
}
