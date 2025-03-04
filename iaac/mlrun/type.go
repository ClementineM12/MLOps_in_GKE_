package mlrun

type MLRunConfig struct {
	RegistryEndpoint   string
	RegistryURL        string
	GcsBucketName      string
	RegistrySecretName string
	Domain             string
	LetsEncrypt        string
}
