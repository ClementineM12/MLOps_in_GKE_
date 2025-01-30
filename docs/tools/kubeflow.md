# Kubeflow 

## Clusters

> Kubectl version: `v1.19`
### [Management ](https://googlecloudplatform.github.io/kubeflow-gke-docs/docs/deploy/management-setup/)

Configuration for the cluster; a basic GKE cluster with workload identity.

> [!IMPORTANT]
> While the management cluster can be deployed in the same project as your Kubeflow cluster, typically you will want to deploy it in a separate project used for administering one or more Kubeflow instances, because it will run with escalated permissions to create Google Cloud resources in the managed projects.

Verify Config Connector
After the deployment:

bash
Copy
Edit
kubectl wait -n cnrm-system --for=condition=Ready pod --all
If everything is installed correctly, the output will confirm that all Config Connector components are running.

### [Kubeflow](https://googlecloudplatform.github.io/kubeflow-gke-docs/dev/docs/deploy/deploy-cli/)

Before proceeding with the creation of the Kubeflow Cluster, it is important that we have first created the management cluster and installed Config Connector.

## [Tool Architecture](https://www.kubeflow.org/docs/started/architecture/)

## [Istio in Kubeflow](https://www.kubeflow.org/docs/concepts/multi-tenancy/istio/#why-kubeflow-needs-istio)


