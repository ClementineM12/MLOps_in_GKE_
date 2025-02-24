# MLRun Guide

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
