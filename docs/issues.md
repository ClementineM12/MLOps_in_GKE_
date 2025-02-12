## GCP 

1. Error while updating **instance groups**:

By default, GKE instance groups use the service account with the format PROJECT_ID@cloudservices.gserviceaccount.com. Confirm that this default service account is assigned the Editor role, which should typically include the necessary permissions.

[Reference](https://cloud.google.com/knowledge/kb/permission-error-when-making-changes-to-instance-groups-000004593)

```
roles/compute.instanceAdmin.v1
roles/compute.networkUser
roles/compute.imageUser
roles/iam.serviceAccountUser
```

Resolve:
```bash
gcloud projects add-iam-policy-binding PROJECT \
  --member="serviceAccount:PROJECT_ID@cloudservices.gserviceaccount.com" \
  --role="roles/compute.instanceAdmin.v1"

gcloud projects add-iam-policy-binding PROJECT \
  --member="serviceAccount:PROJECT_ID@cloudservices.gserviceaccount.com" \
  --role="roles/compute.networkUser"

gcloud projects add-iam-policy-binding PROJECT \
  --member="serviceAccount:PROJECT_ID@cloudservices.gserviceaccount.com" \
  --role="roles/compute.imageUser"
```

[Extra Docs](https://mouliveera.medium.com/permissions-error-required-compute-instancegroups-update-permission-for-project-8a7f759c30c2)

Resolved: https://github.com/pulumi/pulumi/discussions/15902

# Argo cd proper removal of CRDs
```sh
kubectl delete application --all -n argocd 
kubectl delete namespace argocd --wait
kubectl delete crd $(kubectl get crds | grep argoproj.io | awk '{print $1}')
```

## Istio

```sh
  GET / HTTP/1.1" 401 - jwt_authn_access_denied{Jwks_doesn't_have_key_to_match_kid_or_alg_from_Jwt}
```

Run: 
