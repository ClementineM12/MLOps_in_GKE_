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