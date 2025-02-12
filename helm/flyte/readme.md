1. Obtain the IP address for your Ingress resource:

```bash
kubectl get ingress -n flyte
```

Example output:

```bash
NAME              CLASS    HOSTS                     ADDRESS         PORTS     AGE
flyte-core        <none>   flyteontf.uniondemo.run   35.237.42.230   80, 443   3m1s
flyte-core-grpc   <none>   flyteontf.uniondemo.run   35.237.42.230   80, 443   3m1s
```
2. Create a DNS `A` record in a zone you own, pointing to the Ingress IP.

>NOTE: it may take a while before `cert-manager` can issue a certificate for your deployment, especially because for that process to work,
the FQDN of your deployment needs to be resolvable and DNS propagation takes time. 

3. Update your `$HOME\.flyte\config,yaml` and make `endpoint` your DNS name:
```yaml
...
#Example
endpoint: dns:///flyteontf.uniondemo.run 
insecure: false #it means, the connection uses SSL, even if it's a temporary cert-manager cert.

#Uncomment only if you want to test CLI commands and the certificate is not generated yet.
# You can confirm the cert by either going to the UI (a valid certificate should be used) or
#from your terminal: kubectl get challenges.acme.cert-manager.io -n flyte (there should not be any pending challenge). With this flag enabled, SSL is still used but the client doesn't verify the certificate chain.

#insecureSkipVerify: true 
```

> NOTE: this is only needed for CLI access (`flytectl` or `pyflyte`)

### Testing your deployment

5. In your browser, go to `https://<your-DNS-record>/console`

> WARNING: At this point, Flyte's UI would be exposed to the Internet. We stronly encourage you to add authentication to your deployment by following [the documentation](https://docs.flyte.org/en/latest/deployment/configuration/auth_setup.html)

### How to connect to Artifact Registry?

>NOTE: Read more about authentication to Artifact Registry using Access Tokens [here](https://cloud.google.com/artifact-registry/docs/docker/authentication#token)
1. Create a key for the Google Service Account you'll be impersonating in order to push Images to Artifact Registry:

>NOTE: in this example we're using `flyte` as the value for `local.application` and `gcp` for `local.environment` . Replace to match what you indicated in the `locals.tf` file

```bash
gcloud iam service-accounts keys create gcp-artifact-writer.key --iam-account=flyte-gcp-registrywriter@<YOUR-GCP-PROJECT_ID>.iam.gserviceaccount.com
```

2. Activate the Service Account in your gcloud session:

```bash
gcloud auth activate-service-account flyte-gcp-registrywriter@<YOUR-GPC-PROJECT_ID>.iam.gserviceaccount.com --key-file=gcp-artifact-writer.key
```
3. Generate a token and authenticate to Docker:

```bash
gcloud auth print-access-token | docker login -u oauth2accesstoken --password-stdin https://<YOUR-GCP-REGION>-docker.pkg.dev
```
At this point you can use ImageSpec or Dockerfiles to build and push custom images to your Artifact Registry repo.

#### Giving permissions to Flyte Pods to pull Images

As part of the development process, Task Pods will pull the custom Image you defined in the workflow registration phase. For this to work, the `default` Service Account in each `project-domain` namespace will need to mount an `imagePullSecret`.

1. Generate a key for the Google Service Account that's been created with the permissions to read Images from Artifact Registry

```bash
gcloud iam service-accounts keys create gcp-artifact-reader.key --iam-account=flyte-gcp-flyteworkers@<YOUR-GCP-PROJECT_ID>.iam.gserviceaccount.com
```

2. Create a Kubernetes secret in the `project-domain` namespace where you'll run your first workflow:

```bash
kubectl create secret docker-registry artifact-registry --docker-server=https://<YOUR-GCP-REGION>-docker.pkg.dev --docker-email=flyte-gcp-flyteworkers@<YOUR-GCP-PROJECT>.iam.gserviceaccount.com --docker-username=_json_key --docker-password="$(cat gcp-artifact-reader.key)" --namespace flytesnacks-development
```

3. Edit your `default` Service Account:

```bash
kubectl edit sa default -n flytesnacks-development
```

4. Add the `imagePullSecret`:

```yaml
imagePullSecrets:
- name: artifact-registry
```
4. Run [the example](https://docs.flyte.org/projects/cookbook/en/latest/auto_examples/customizing_dependencies/image_spec.html#image-spec-example) in the docs to confirm.