package flyte

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AccountInfo struct {
	ServiceAccount *serviceaccount.Account
	Member         pulumi.StringArrayOutput
	Email          pulumi.StringOutput
}
