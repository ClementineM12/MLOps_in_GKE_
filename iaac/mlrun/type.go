package mlrun

type MLRunConfig struct {
	registryEndpoint   string
	registryURL        string
	gcsBucketName      string
	registrySecretName string
	domain             string
}

// IngressPathConfig holds configuration for an individual path rule.
type IngressPathConfig struct {
	Path    string // subpath, e.g., "pipelines" (empty means root)
	Service string // name of the backend service
	Port    int    // port number on the service
}

// IngressConfig holds configuration for an entire Ingress resource.
type IngressConfig struct {
	DNS   string              // DNS prefix; for example, "grafana" leads to "grafana.example.com"
	Paths []IngressPathConfig // one or more path rules for the Ingress
}
