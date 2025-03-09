package flyte

import (
	"fmt"
)

func createFlyteNamespaces() []string {

	// Create a slice to hold the resulting namespaces.
	var namespaces []string

	// Compute the cartesian product of flyteProjects and flyteDomains.
	for _, project := range flyteProjects {
		for _, domain := range flyteDomains {
			namespace := fmt.Sprintf("%s-%s", project, domain)
			namespaces = append(namespaces, namespace)
		}
	}
	return namespaces
}
