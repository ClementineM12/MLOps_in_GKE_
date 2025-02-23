# MLOps Tool Comparison

## 🔹 Prerequisites

Ensure the following tools are installed before running the project:

- [Helm](https://helm.sh/)
- [Golang](https://go.dev/)
- [Pulumi](https://www.pulumi.com/)

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

First, create and deploy your **IaC resources**:

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

## Shut Down Resources

To clean up all deployed resources:

```sh
pulumi destroy
```
This will remove all Pulumi-managed resources from Google Cloud. 
