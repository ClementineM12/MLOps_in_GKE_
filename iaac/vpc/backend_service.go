package vpc

// Package vpc provides functionality for configuring a Global Load Balancer (GLB) in Google Cloud Platform (GCP) using Pulumi.
// It includes the creation of backend services and health checks necessary for handling traffic efficiently and ensuring
// the availability of backend resources.
//
// The resources created in this package include:
//
// 1. **Load Balancer Backend Service (`createLoadBalancerBackendService`)**:
//    - Creates a backend service that defines how traffic is distributed across backend instances or groups.
//    - Associates a health check with the backend service to ensure traffic is only sent to healthy instances.
//    - Serves as a key component of the Global Load Balancer to manage backend resources effectively.
//
// 2. **TCP Health Checks (`createLoadBalancerTCPHealthChecks`)**:
//    - Configures health checks that monitor backend services through TCP connections on a specified port (default: 80).
//    - Ensures that only healthy backend instances receive traffic.
//    - Uses parameters like `CheckIntervalSec`, `HealthyThreshold`, and `UnhealthyThreshold` to define how health checks operate.
//
// 3. **Backend Service (`createLoadbalancerBackendService`)**:
//    - Defines the behavior of the Global Load Balancer, including connection draining and CDN caching policies.
//    - Uses backend configurations to distribute traffic based on load balancing policies.
//    - Tied to health checks to avoid routing requests to unhealthy instances.
//
// These resources work together to facilitate a **scalable, resilient, and highly available** architecture by enabling
// global load balancing across distributed backend services in Google Cloud. Proper use of health checks ensures that
// only responsive and healthy backends receive traffic, improving the reliability and performance of applications.

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateLoadBalancerBackendService sets up the backend service for a Global Load Balancer.
// It creates a backend service that will handle incoming traffic routed by the load balancer.
// This function also sets up the health checks used by the backend service.
func createLoadBalancerBackendService(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
) (*compute.BackendService, error) {

	gcpGLBTCPHealthCheck, err := createLoadBalancerHTTPSHealthCheck(ctx, projectConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Load Balancer Health check: %w", err)
	}
	gcpBackendService, err := createLoadbalancerBackendService(ctx, projectConfig, gcpGLBTCPHealthCheck.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to create Load Balancer Backend Service: %w", err)
	}

	return gcpBackendService, nil
}

// createLoadBalancerTCPHealthChecks creates the TCP health checks that the Global Load Balancer will use
// to verify the health of backend services. The health checks monitor the health of services
// through TCP connections on a specified port.
func createLoadBalancerHTTPSHealthCheck(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
) (*compute.HealthCheck, error) {

	resourceName := fmt.Sprintf("%s-glb-https-hc", projectConfig.ResourceNamePrefix)
	gcpGLBHealthCheck, err := compute.NewHealthCheck(ctx, resourceName, &compute.HealthCheckArgs{
		Project:          pulumi.String(projectConfig.ProjectId),
		CheckIntervalSec: pulumi.Int(10),
		Description:      pulumi.String("HTTPS Health Check for Istio"),
		HealthyThreshold: pulumi.Int(3),
		HttpHealthCheck: &compute.HealthCheckHttpHealthCheckArgs{
			Port:        pulumi.Int(15021), // Istio Gateway Health Check Port
			RequestPath: pulumi.String("/healthz/ready"),
			ProxyHeader: pulumi.String("NONE"),
		},
		TimeoutSec:         pulumi.Int(5),
		UnhealthyThreshold: pulumi.Int(5),
	})

	return gcpGLBHealthCheck, err
}

// createLoadbalancerBackendService creates a backend service for the Global Load Balancer
// to route traffic from the GCP Load Balancer to Istio Ingress Gateway.
func createLoadbalancerBackendService(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	gcpGLBHealthCheck pulumi.StringInput,
) (*compute.BackendService, error) {

	// 🔹 Fetch the AutoNEG-managed NEG dynamically
	// negSelfLink := getAutoNegNetworkEndpointGroup(ctx, projectConfig)

	resourceName := fmt.Sprintf("%s-glb-backend", projectConfig.ResourceNamePrefix)

	gcpBackendService, err := compute.NewBackendService(ctx, resourceName, &compute.BackendServiceArgs{
		Project:                      pulumi.String(projectConfig.ProjectId),
		Name:                         pulumi.String(fmt.Sprintf("%s-backend-svc", projectConfig.ResourceNamePrefix)),
		Description:                  pulumi.String("Global Load Balancer - Backend Service"),
		EnableCdn:                    pulumi.Bool(false),
		ConnectionDrainingTimeoutSec: pulumi.Int(10),
		HealthChecks:                 gcpGLBHealthCheck,
		Protocol:                     pulumi.String("HTTPS"),
		LoadBalancingScheme:          pulumi.String("EXTERNAL"),

		// Backend routing to Istio Ingress Gateway
		Backends: compute.BackendServiceBackendArray{},
	})
	return gcpBackendService, err
}

// func getAutoNegNetworkEndpointGroup(
// 	ctx *pulumi.Context,
// 	projectConfig global.ProjectConfig,
// ) pulumi.StringOutput {
// 	// Construct the expected AutoNEG-managed NEG name
// 	negName := pulumi.Sprintf("k8s1-%s-istio-ingressgateway", projectConfig.ProjectId)

// 	// Retrieve the NEG dynamically
// 	neg := compute.LookupNetworkEndpointGroupOutput(ctx, compute.LookupNetworkEndpointGroupOutputArgs{
// 		Name:    negName,
// 		Project: pulumi.String(projectConfig.ProjectId),
// 		Zone:    pulumi.String("europe-west8"),
// 	})
// 	return pulumi.StringOutput(neg.SelfLink())
// }
