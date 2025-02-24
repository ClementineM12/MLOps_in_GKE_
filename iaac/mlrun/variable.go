package mlrun

import (
	infracomponents "mlops/infra_components"
)

var infraComponents = infracomponents.InfraComponents{
	CertManager:  true,
	NginxIngress: true,
}
