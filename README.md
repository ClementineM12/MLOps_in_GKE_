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
- **ArgoCD for GitOps**
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

## 🎯 Deploy and Access ArgoCD

Get `helm-chart` and deploy:
```sh
helm repo add argo https://argoproj.github.io/argo-helm && helm repo update
helm install argocd argo/argo-cd --namespace argocd --create-namespace --values helm/argocd/values.yaml 
```

Retrieve ArgoCD Admin Credentials:

```sh
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```
* **Username**: `admin`

Default Port Forwarding:
```sh
kubectl port-forward svc/argocd-server -n argocd 8080:443
```
Now, access ArgoCD UI at https://localhost:8080.

## ⚙️ Deploy Kubeflow with Helm

```sh
# Navigate to the Helm chart directory
cd helm/kubeflow

# Uninstall existing Kubeflow release (if any)
helm uninstall kubeflow --namespace argocd
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

## 🎉 You're All Set!

Now that your MLOps infrastructure is up and running, you can begin managing ML workloads using Kubeflow and ArgoCD.

For further improvements, consider:

* ✅ Automating deployment with CI/CD
* ✅ Securing Kubeflow & ArgoCD
* ✅ Optimizing GCP resources for cost efficiency
* 🚀 Happy MLOps!

### **📌 Key Enhancements:**

* ✅ **Markdown-friendly** formatting for `README.md`  
* ✅ **Proper code blocks (`sh`)** for commands  
* ✅ **Emojis for better readability** (optional but engaging)  
* ✅ **Section breaks (`---`)** for better structure  

