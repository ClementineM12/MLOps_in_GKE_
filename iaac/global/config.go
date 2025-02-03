package global

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func GenerateProjectConfig(
	ctx *pulumi.Context,
) ProjectConfig {

	domain := config.Get(ctx, "project:domain")
	// Validate
	validateArtifactRegistryConfig(ctx)

	return ProjectConfig{
		ResourceNamePrefix: configureResourcePrefix(ctx),
		ProjectId:          configureProjectId(ctx),
		Domain:             domain,
		SSL:                configureSSL(ctx, domain),
		EnabledRegions:     configureRegions(ctx),
		Target:             configureTarget(ctx),
	}
}

func configureProjectId(
	ctx *pulumi.Context,
) string {

	// Review Project ID Configuration
	gcpProjectId := config.Get(ctx, "gcp:project")
	if gcpProjectId == "" {
		ctx.Log.Error("No GCP Project defined; Pulumi GCP Provider must have Project configured.", nil)
		return ""
	}
	return gcpProjectId
}

func configureResourcePrefix(
	ctx *pulumi.Context,
) string {

	// Review Prefix Configuration
	resourceNamePrefix := config.Get(ctx, "project:prefix")
	if resourceNamePrefix == "" {
		ctx.Log.Error("No Prefix has been provided; Please set a prefix (3-5 characters long), it is mandatory.", nil)
		return ""
	} else {
		if len(resourceNamePrefix) > 5 {
			ctx.Log.Error(fmt.Sprintf("Prefix '%s' must be less than 5 characters in length.", resourceNamePrefix), nil)
			return ""
		}
		fmt.Printf("\033[1;32m[INFO] - Prefix '%s' has been provided; All Google Cloud resource names will be prefixed.\n\033[0m", resourceNamePrefix)
	}
	return resourceNamePrefix
}

func configureSSL(
	ctx *pulumi.Context,
	domain string,
) bool {

	// Review Domain & SSL Configuration
	if domain != "" {
		fmt.Printf("\033[1;32m[INFO] - Domain '%s' has been provided; SSL Certificates will be configured for this domain.\n\033[0m", domain)
		fmt.Printf("\033[1;32m[INFO] - The DNS for the domain: '%s' must be configured to point to the IP Address of the Global Load Balancer.\n\033[0m", domain)
		return true
	} else {
		ctx.Log.Warn("No Domain has been provided; HTTPS will not be enabled for this deployment.", nil)
		return false
	}
}

// configureRegions separates regions into enabled and not enabled
func configureRegions(
	ctx *pulumi.Context,
) []CloudRegion {

	var enabledRegions []CloudRegion

	enabledRegionIds := strings.Split(config.Get(ctx, "vpc:regions"), ",")
	for _, regionId := range enabledRegionIds {
		found := false
		for _, cloudRegion := range CloudRegions {
			if cloudRegion.Id == regionId {
				enabledRegions = append(enabledRegions, cloudRegion)
				found = true
			}
		}
		if !found {
			ctx.Log.Warn(fmt.Sprintf("Region ID %s does not exist in predefined Cloud Regions.", regionId), nil)
		}
	}
	fmt.Printf("\033[1;32m[INFO] - Processing Cloud Regions: [ %s ]\n\033[0m", formatRegions(enabledRegions))

	return enabledRegions
}

// configureTarget retrieves and validates the "gke:target" configuration value.
// The target must be set to either "management" or "deployment". If the target
// is missing or invalid, the function logs an error and returns an empty string.
func configureTarget(
	ctx *pulumi.Context,
) string {

	target := config.Get(ctx, "gke:target")
	if target == "" {
		ctx.Log.Error("Target must be provided: set to either 'deployment' or 'management'.", nil)
	} else {
		if target != "management" && target != "deployment" {
			ctx.Log.Error("Target must be set to either 'deployment' or 'management'; Provide a valid target.", nil)
		}
	}
	return target
}

func validateArtifactRegistryConfig(
	ctx *pulumi.Context,
) {

	create := config.GetBool(ctx, "ar:create")
	if create {
		githubRepo := config.Get(ctx, "ar:githubRepo")
		if githubRepo == "" {
			ctx.Log.Error("Artifact Registry is enabled; provide a GitHub Repository configuration 'ar:githubRepo'.", nil)
		}
	}
}
