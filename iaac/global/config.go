package global

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GenerateProjectConfig(
	ctx *pulumi.Context,
) ProjectConfig {

	domain := config.Get(ctx, "project:domain")
	whitelistedIPs := config.Get(ctx, "project:whitelistedIPs")
	if whitelistedIPs == "" {
		whitelistedIPs = "0.0.0.0/0"
	}

	ValidateMLOpsTarget(ctx)

	return ProjectConfig{
		ResourceNamePrefix: configureResourcePrefix(ctx),
		ProjectId:          configureProjectId(ctx),
		Domain:             domain,
		SSL:                configureSSL(ctx, domain),
		EnabledRegions:     configureRegions(ctx),
		CloudSQL:           getCloudSQLConfig(ctx),
		WhitelistedIPs:     whitelistedIPs,
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

func ConfigureArtifactRegistry(
	ctx *pulumi.Context,
	ArtifactRegistryConfig ArtifactRegistryConfig,
) ArtifactRegistryConfig {

	if ArtifactRegistryConfig.GithubServiceAccountCreate {
		githubRepo := config.Get(ctx, "project:githubRepo")
		if githubRepo == "" {
			ctx.Log.Error("Artifact Registry is enabled; provide a GitHub Repository configuration 'ar:githubRepo'.", nil)
		}
		ArtifactRegistryConfig.GithubRepo = githubRepo
	}
	return ArtifactRegistryConfig
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

func getCloudSQLConfig(
	ctx *pulumi.Context,
) *CloudSQLConfig {

	return &CloudSQLConfig{
		Create:             config.GetBool(ctx, "cloudsql:create"),
		User:               config.Get(ctx, "cloudsql:user"),
		Database:           config.Get(ctx, "cloudsql:database"),
		InstancePrefixName: config.Get(ctx, "cloudsql:instancePrefixName"),
	}
}

func ValidateEmail(
	ctx *pulumi.Context,
) string {

	email := config.Get(ctx, "project:email")
	if email == "" {
		ctx.Log.Error("Cert manager needs field `project:email` defined.", nil)
		return ""
	}
	return email
}

func ValidateMLOpsTarget(
	ctx *pulumi.Context,
) {

	target := config.Get(ctx, "project:target")
	if target != "" {
		if !listContains(MLOpsAllowedTargets, target) {
			ctx.Log.Error(fmt.Sprintf("Target MLOps tool is not included in the Allowlist: %s", formatListIntoString(MLOpsAllowedTargets)), nil)
		}
		caser := cases.Title(language.English)
		fmt.Printf("\033[1;32m[INFO] - MLOps tool targeted for deployment; %s\n\033[0m", caser.String(target))
	}
}
