# MLOps Tool Comparison

## 🔹 Prerequisites

Ensure the following tools are installed before running the project:

- [Helm](https://helm.sh/) `v3.14.0`
- [Golang](https://go.dev/) `v1.22.6`
- [Pulumi](https://www.pulumi.com/) `v3.147.0`

---

## Getting Started

This repository contains **Infrastructure-as-Code (IaaC)** definitions using [Pulumi](https://www.pulumi.com/) to manage resources on **Google Cloud Platform (GCP)**.

The infrastructure includes:
- **GKE Cluster**
- **VPC Networks**
- **IAM Roles**
- **FluxCD for GitOps** -- used only for **Kubeflow**
- **GCS**
- **CloudSQL** -- Postgres

---

## Deploy Infrastructure

Create a configuration file for pulumi ( e.g. `Pulumi.development.yaml` )

First, create and deploy your **IaC resources**:
```yaml
config:
  gcp:project: <project_id>

  project:prefix: <prefix_for_resources>
  project:target: <mlop_tool_target_to_deploy>

  project:domain: <your_domain>
  project:email: <your_email>
  project:whitelistedIPs: <IPs_to_whitelist_for_ingress>
  project:githubRepo: <your_GitHub_repository>

  vpc:regions: "007" # <- This is selected in order to have the option of using NodePools with GPU acceleration
  vpc:loadBalancer: false # Not configured end-to-end
  vpc.autoNEG: false # Working but not totally configured with the Networking

  gke:privateNodes: true # If not set it will default to `false`
  gke:managementAutoRepair: true # If not set it will default to `false`
  gke:managementAutoUpgrade: true # If not set it will default to `false`

  # OPTIONAL GKE values -- default
  gke:name: default
  gke:cidr: 10.0.0.0/16
  gke:nodePoolMachineType: e2-standard-4
  gke:nodePoolDiskSizeGb: 100
  gke:nodePoolDiskType: pd-standard
  gke:nodePoolMaxNodeCount: 5
  gke:nodePoolPreemptible: false
```

Upon reading the [Docs](https://github.com/ClementineM12/MLOps_in_GKE_/blob/main/docs/docs.md) and have configured what is necessary proceed with building your Infrastructure:
```sh
cd iaac
pulumi up
cd ..
```
This will provision all necessary GCP resources.

Once, everything is up and running, connect to the cluster:
```sh
gcloud container clusters get-credentials CLUSTER_NAME --region REGION --project PROJECT_ID
```

## Deploy the MLOps tool of your choise

Depending on the tool to deploy we set `project:target` configuration field in the IaaC. Specific guidelines are provided in `helm` root path for each tool.

**Versions**

* Kubeflow `v1.9.1`
* MLRun `v1.7.2`
* Flyte  [flyte-core] `v1.5.0`

## Shut Down Resources

To clean up all deployed resources:

```sh
pulumi destroy
```
This will remove all Pulumi-managed resources from Google Cloud. 
