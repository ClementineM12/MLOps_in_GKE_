# MLRun Guide

This guide provides step-by-step instructions for setting up **MLRun** using **Helm** and **Kubernetes**.

## 1. Add and Update Helm Repositories
Add the **V3IO stable** Helm repository and update the chart list:
```sh
helm repo add mlrun-ce https://mlrun.github.io/ce && helm repo update
```

---

## 2. Create a Docker Registry Secret
Create a Kubernetes **secret** for authenticating with your container registry:
```sh
kubectl --namespace mlrun create secret docker-registry registry-credentials \
    --docker-server=<your-registry-server> \ ( if docker hub: `https://registry.hub.docker.com/` )
    --docker-username=<your-username> \
    --docker-password=<your-password> \
    --docker-email=<your-email>
```
Replace placeholders (`<your-registry-server>`, `<your-username>`, etc.) with actual values.

---

## 3. Deploy Ingress Controller
Add the **Ingress Nginx** Helm repository and install the ingress controller:
```sh
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm install nginx-ingress ingress-nginx/ingress-nginx --namespace ingress-nginx --create-namespace
```
Get extrernal IP:
```
kubectl get svc -n ingress-nginx nginx-ingress-ingress-nginx-controller
```

Verify that the **Ingress controller** is running:
```sh
kubectl get svc -n ingress-nginx
```

---

## 4. Install Cert-Manager (For TLS Support)
To enable **SSL/TLS certificates**, install Cert-Manager:
```sh
helm repo add jetstack https://charts.jetstack.io
helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --set crds.enabled=true
```

---

## 5. Deploy MLRun Kit
Install MLRun using Helm with the necessary configurations:
```sh
helm --namespace mlrun \
    install mlrun-kit \
    --wait \
    --timeout 960s \
    --set global.registry.url=index.docker.io/<username> \
    --set global.registry.secretName=registry-credentials \
    --set global.externalHostAddress=<external_ip> \ 
    --set nuclio.dashboard.externalIPAddresses=<list of IP addresses> \
    v3io-stable/mlrun-kit
```

---

## 6. Verify MLRun Deployment
Once the installation is complete, verify that MLRun is deployed successfully:
```sh
kubectl get pods -n mlrun
```

You should see running pods for MLRun, Nuclio, and other services.

---

## 7. Access MLRun Dashboard
To access the MLRun UI, follow these steps:
1. Get the **Ingress IP Address**:
   ```sh
   kubectl get svc -n ingress-nginx
   ```
2. Open your browser and navigate to:
   ```
   http://<external-ip>
   ```
   Replace `<external-ip>` with the external IP address of your ingress controller.

---

## GCP IAP

```
gcloud compute backend-services update BACKEND_SERVICE_NAME \
    --global \
    --iap=enabled,oauth2-client-id=CLIENT_ID,oauth2-client-secret=CLIENT_SECRET
```
