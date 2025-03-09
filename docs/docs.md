## Prerequisites

* GCP Account
* Pulumi Account

> [!NOTE]
> A nice thing that Google Cloud provides is a [cost calculator](https://cloud.google.com/products/calculator?hl=en), so before beginning building anything on cloud proceed with checking the costs of it.

## Installation

1. Follow the proper [instruction](https://cloud.google.com/sdk/docs/install) to install google cloud cli for your working environment.
2. Install [pulumi](https://www.pulumi.com/docs/iac/get-started/gcp/begin/) and the proper language version ( go ).
3. Pulumi Kubernetes Plugin
    ```bash
      pulumi plugin install resource kubernetes
    ```

> [!IMPORTANT]
> If you are using the Google Cloud Free Program or the 12-month trial period with $300 credit, note that the free tier does not offer enough resources for default full Kubeflow installation. You need to upgrade to a paid account. For more info, check [here](https://googlecloudplatform.github.io/kubeflow-gke-docs/dev/docs/deploy/project-setup/#setting-up-a-project).

## GCP Project

Create your GCP project ( e.g. mlops-project-01 ) on the console.

It is important to activate the **Cloud Resource Manager API**

## Service account

In your Google Cloud project, create a service account dedicated for pulumi. You can name it however you see fit in order to showcase its usage ( .e.g. pulumi-deployer ). The following roles must be set upon your new service account: 

1. **GKE Cluster Roles**

To create and manage a GKE cluster, the service account or user must have the following roles:

  - roles/container.admin: Full control over GKE clusters.
  - roles/compute.networkAdmin: To create or configure networks and firewall rules for the cluster.
  - roles/iam.serviceAccountUser: To allow the cluster to use the Kubernetes service account.

2. **VPC Roles**

For network setup, the roles required are:

  - roles/compute.networkAdmin: For creating or managing VPCs, subnets, and routes.
  - roles/compute.securityAdmin: For managing firewall rules.
  - roles/compute.viewer: For viewing existing network configurations.

3. **IAM Roles for Kubeflow and Storage**

Kubeflow requires permissions for Kubernetes resources and Google Cloud services (like GCS and IAM):

**Storage Roles**:
  - roles/storage.admin: For managing buckets and objects in GCS.
  - roles/storage.objectAdmin: For accessing objects in GCS.

**IAM and Kubernetes Roles**:
  - roles/iam.serviceAccountAdmin: To manage service accounts used by Kubeflow.
  - `Kubernetes Engine Admin`/`roles/container.clusterAdmin`: To administer Kubernetes clusters.
  - roles/container.developer: To manage Kubernetes applications.

4. **Load Balancer Roles**

For the load balancer to work with GKE and expose Kubeflow services:

  - roles/compute.loadBalancerAdmin: To manage external/internal load balancers.
  - roles/dns.admin: If using managed DNS for domain names.

> [!NOTE]
> The reason we are using service-account is for security and following best [practises](https://cloud.google.com/sdk/docs/authorizing).

**Activation**

After you have created the service account in your Google Cloud console procced ith the creation of a service account key. Upon its creation it will be also downloaded. Save the key to your desired directory.
To activate the usage of the service-account first you must export in your working environment the following variable:
```bash
export GCLOUD_KEYFILE_JSON=~<directory><key_name>.json
```
, and run the following command:
```bash
gcloud auth activate-service-account --key-file=$GCLOUD_KEYFILE_JSON
```

> [!NOTE]
> The set up of Cloud Identity-Aware Proxy ( Cloud IAP ) is recommended for production deployments or deployments with access to sensitive data, thus this step was not configured.

# GO Pulumi-GCP Package

Manages a [Google Kubernetes Engine (GKE)](https://pkg.go.dev/github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/container) cluster.

To get more information about GKE clusters, see:

[The API reference](https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters)
How-to guides
[GKE overview](https://cloud.google.com/kubernetes-engine/docs/concepts/kubernetes-engine-overview)
[About cluster configuration choices](https://cloud.google.com/kubernetes-engine/docs/concepts/types-of-clusters)

> On version 5.0.0+ of the provider, you must explicitly set `deletionProtection = false` and run `pulumi up` to write the field to state in order to destroy a cluster.

> All arguments and attributes (including certificate outputs) will be stored in the raw state as plaintext. [Read more about secrets in state](https://www.pulumi.com/docs/intro/concepts/programming-model/#secrets).

# Github Actions

Retrieve value for **WORKLOAD_IDENTITY_PROVIDER** to set as secret by running: 
```sh
  gcloud iam workload-identity-pools providers list \
    --workload-identity-pool=MY_POOL_ID \
    --location=global \
    --project=MY_PROJECT_ID
```