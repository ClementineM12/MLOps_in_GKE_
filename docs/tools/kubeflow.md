# Kubeflow Guide

### Port-Forward

The default way of accessing Kubeflow is through **port-forwarding**. This allows you to get started quickly without additional networking configurations. Run the following command to forward Istio's Ingress-Gateway to local port `8080`:

```sh
kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80
```

After running the command, access the **Kubeflow Central Dashboard** by following these steps:

1. Open your browser and visit `http://localhost:8080`. You should see the Dex login screen.
2. Login with the default user credentials:
   - **Email:** `user@example.com`
   - **Password:** `12341234`

### NodePort / LoadBalancer / Ingress

To connect to Kubeflow using **NodePort**, **LoadBalancer**, or **Ingress**, you need to set up **HTTPS**. Many Kubeflow applications (e.g., Tensorboard, Jupyter, Katib UI) use [Secure Cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#restrict_access_to_cookies), making HTTP access over non-localhost domains ineffective.

Exposing your Kubeflow cluster with proper **HTTPS** depends on your environment. You may also consider third-party [Kubeflow distributions](https://www.kubeflow.org/docs/started/installing-kubeflow/#install-a-packaged-kubeflow-distribution).

> **Note:** If you **must** expose Kubeflow over HTTP, disable the `Secure Cookies` feature by setting the `APP_SECURE_COOKIES` environment variable to `false` in every relevant web app. This is **not recommended** due to security risks.

---

## Changing Default Credentials

### Change Default Username & Email

For security reasons, avoid using the **default username and email** in security-sensitive environments. Instead, customize them before deploying Kubeflow.

To modify the **default user** credentials:

1. Edit `common/dex/overlays/oauth2-proxy/config-map.yaml` and update the fields:

    ```yaml
    ...
      staticPasswords:
      - email: <REPLACE_WITH_YOUR_EMAIL>
        username: <REPLACE_WITH_PREFERRED_USERNAME>
    ```

### Change Default User Password

Using an **identity provider** (e.g., LDAP, GitHub, Google, Microsoft, OIDC, SAML, GitLab) is recommended instead of static passwords. However, if needed, follow these steps to update the default password securely.

#### Generate a Secure Password Hash

Use `bcrypt` to hash a new password:

```sh
python3 -c 'from passlib.hash import bcrypt; import getpass; print(bcrypt.using(rounds=12, ident="2y").hash(getpass.getpass()))'
```

Example output:

```sh
Password:       # Enter your new password
$2y$12$example_hash_generated  # Copy this hash
```

#### Before Creating the Cluster

1. Edit `common/dex/base/dex-passwords.yaml` and update the `DEX_USER_PASSWORD` field with your new hash:

    ```yaml
    ...
      stringData:
        DEX_USER_PASSWORD: <REPLACE_WITH_HASH>
    ```

#### After Creating the Cluster

1. Delete the existing `dex-passwords` secret:

    ```sh
    kubectl delete secret dex-passwords -n auth
    ```

2. Create a new `dex-passwords` secret with the updated hash:

    ```sh
    kubectl create secret generic dex-passwords --from-literal=DEX_USER_PASSWORD='REPLACE_WITH_HASH' -n auth
    ```

3. Restart the Dex pod to apply changes:

    ```sh
    kubectl delete pods --all -n auth
    ```

4. Login using the new password.

---

## Upgrading and Extending Kubeflow

For modifying or upgrading your Kubeflow deployment, follow these best practices:

- **Never edit manifests directly.** Use **Kustomize overlays** and [components](https://github.com/kubernetes-sigs/kustomize/blob/master/examples/components.md) on top of the [example.yaml](https://github.com/kubeflow/manifests/blob/master/example/kustomization.yaml).
- Upgrade by referencing the new manifests, building with `kustomize`, and applying with `kubectl apply`.
- Adjust **overlays and components** as needed during upgrades.
- **Prune old resources** by adding [labels](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/labels/) to all resources from the start.
- Use `kubectl apply --prune --dry-run` to list removable resources before applying changes.
- Major changes (e.g., the **1.9 release switch to oauth2-proxy**) may require additional configuration.

> With basic Kubernetes knowledge, upgrading Kubeflow should be manageable.

---

This README provides essential guidelines for installing, securing, and upgrading Kubeflow. For further details, refer to the official [Kubeflow documentation](https://www.kubeflow.org/docs/).

### References
* [Kubeflow GKE Docs](https://googlecloudplatform.github.io/kubeflow-gke-docs/dev/docs/deploy/deploy-cli/)
* [Tool Architecture](https://www.kubeflow.org/docs/started/architecture/)
* [Istio in Kubeflow](https://www.kubeflow.org/docs/concepts/multi-tenancy/istio/#why-kubeflow-needs-istio)


