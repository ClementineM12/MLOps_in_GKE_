package infracomponents

var (
	NginxControllerNamespace        = "nginx-ingress"
	NginxControllerHelmChart        = "ingress-nginx"
	NginxControllerHelmChartVersion = "4.11.4"
	NginxControllerHelmChartRepo    = "https://kubernetes.github.io/ingress-nginx"

	CertManagerNamespace        = "cert-manager"
	CertManagerHelmChart        = "cert-manager"
	CertManagerHelmChartVersion = "v1.17.0"
	CertManagerHelmChartRepo    = "https://charts.jetstack.io"
)
