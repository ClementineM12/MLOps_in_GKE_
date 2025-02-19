# 🚀 MLOps Project

## 🔹 Prerequisites

Ensure the following tools are installed before running the project:

- [Helm](https://helm.sh/)
- [Golang](https://go.dev/)
- [Pulumi](https://www.pulumi.com/)

---

## 📌 Getting Started

This repository contains **Infrastructure-as-Code (IaC)** definitions using [Pulumi](https://www.pulumi.com/) to manage resources on **Google Cloud Platform (GCP)**.

The infrastructure includes:
- **GKE Clusters**
- **Istio Service Mesh**
- **VPC Networks**
- **IAM Roles**
- **FluxCD for GitOps**
- **Kubeflow for ML Pipelines**

---

## 🚀 Deploy Infrastructure

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

## ⚙️ Deploy Kubeflow with Helm

```sh
# Navigate to the Helm chart directory
cd helm/kubeflow-flux

# Install Kubeflow
helm install kubeflow --namespace flux-system
```

## 🌐 Access the Kubeflow Central Dashboard

Forward the Istio Ingress Gateway:

```sh
kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80
```
Now, access Kubeflow Dashboard at http://localhost:8080.

## 🛑 Shut Down Resources

To clean up all deployed resources:

```sh
pulumi destroy
```
This will remove all Pulumi-managed resources from Google Cloud.

## 📌 Running a Kubeflow Pipeline

Install Kubeflow Pipelines SDK (v2.11.0)
```sh
pip install kfp
```
Now, you can create and run Kubeflow Pipelines!

