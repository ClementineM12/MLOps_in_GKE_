package main

import (
	"errors"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func ConfigureProjectId(
	ctx *pulumi.Context,
) (string, error) {

	// Review Project ID Configuration
	gcpProjectId := config.Get(ctx, "gcp:project")
	if gcpProjectId == "" {
		return "", errors.New("[ CONFIGURATION ] - [gcp:project] - No GCP Project Set: Pulumi GCP Provider must have Project configured")
	}
	return gcpProjectId, nil
}

func ConfigureResourcePrefix(
	ctx *pulumi.Context,
) (string, error) {

	// Review Prefix Configuration
	resourceNamePrefix := config.Get(ctx, "project:prefix")
	if resourceNamePrefix == "" {
		return "", errors.New("[ CONFIGURATION ] - No Prefix has been provided; Please set a prefix (3-5 characters long), it is mandatory")
	} else {
		if len(resourceNamePrefix) > 5 {
			return "", fmt.Errorf("[ CONFIGURATION ] - Prefix: '%s' must be less than 5 characters in length", resourceNamePrefix)
		}
		fmt.Printf("[ CONFIGURATION ] - Prefix: %s has been provided; All Google Cloud resource names will be prefixed.\n", resourceNamePrefix)
	}
	return resourceNamePrefix, nil
}

func ConfigureSSL(
	domain string,
) bool {

	// Review Domain & SSL Configuration
	if domain != "" {
		fmt.Printf("[ CONFIGURATION ] - Domain: '%s' has been provided; SSL Certificates will be configured for this domain.\n", domain)
		fmt.Printf("[ CONFIGURATION ] - DNS: The DNS for the domain: '%s' must be configured to point to the IP Address of the Global Load Balancer.\n", domain)
		return true
	} else {
		fmt.Printf("[ CONFIGURATION ] - No Domain has been provided; Therefore HTTPS will not be enabled for this deployment.\n")
		return false
	}
}
