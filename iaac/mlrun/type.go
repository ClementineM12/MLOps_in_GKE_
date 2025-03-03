package mlrun

type MLRunConfig struct {
	registryEndpoint   string
	registryURL        string
	gcsBucketName      string
	registrySecretName string
	domain             string
}
