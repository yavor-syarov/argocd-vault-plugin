### HashiCorp Vault
We support AppRole, Token, Github, Kubernetes and Userpass Auth Method for getting secrets from Vault.

We currently support retrieving secrets from KV-V1 and KV-V2 backends.

**Note**: For KV-V2 backends, the path needs to be specified as `${vault-kvv2-backend-path}/data/{path-to-secret}` where `vault-kvv2-backend-path` is the path to the KV-V2 backend (usually just `secret`) and `path-to-secret` is the path to the secret in Vault.

##### AppRole Authentication
For AppRole Authentication, these are the required parameters:
```
VAULT_ADDR: Your HashiCorp Vault Address
AVP_TYPE: vault
AVP_AUTH_TYPE: approle
AVP_ROLE_ID: Your AppRole Role ID
AVP_SECRET_ID: Your AppRole Secret ID
```

##### Vault Token Authentication
For Vault Token Authentication, these are the required parameters:
```
VAULT_ADDR: Your HashiCorp Vault Address
VAULT_TOKEN: Your Vault token
AVP_TYPE: vault
AVP_AUTH_TYPE: token
```

This option may be the easiest to test with locally, depending on your Vault setup.

##### Github Authentication
For Github Authentication, these are the required parameters:
```
VAULT_ADDR: Your HashiCorp Vault Address
AVP_TYPE: vault
AVP_AUTH_TYPE: github
AVP_GITHUB_TOKEN: Your Github Personal Access Token
```

##### Kubernetes Authentication
In order to use Kubernetes Authentication a couple of things are required.

###### 1. Configuring Argo CD
You can either use your own Service Account or the default Argo CD service account. To use the default Argo CD service account all you need to do is set `automountServiceAccountToken` to true in the `argocd-repo-server`.

```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  name: argocd-repo-server
spec:
  template:
    spec:
      automountServiceAccountToken: true
```

This will put the Service Account token in the default path of `/var/run/secrets/kubernetes.io/serviceaccount/token`.

If you want to use your own Service Account, you would first create the Service Account.
`kubectl create serviceaccount your-service-account`.

<b>*Note*</b>: The service account that you use must have access to the Kubernetes TokenReview API. You can find the Vault documentation on configuring Kubernetes [here](https://www.vaultproject.io/docs/auth/kubernetes#configuring-kubernetes).

And then you will update the `argocd-repo-server` to use that service account.

```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  name: argocd-repo-server
spec:
  template:
    spec:
      serviceAccount: your-service-account
      automountServiceAccountToken: true
```

###### 2. Configuring Kubernetes
Use the /config endpoint to configure Vault to talk to Kubernetes. Use `kubectl cluster-info` to validate the Kubernetes host address and TCP port. For the list of available configuration options, please see the [API documentation](https://www.vaultproject.io/api/auth/kubernetes).

```
$ vault write auth/kubernetes/config \
    token_reviewer_jwt="<your service account JWT>" \
    kubernetes_host=https://192.168.99.100:<your TCP port or blank for 443> \
    kubernetes_ca_cert=@ca.crt
```

And then create a named role:
```
vault write auth/kubernetes/role/argocd \
    bound_service_account_names=your-service-account \
    bound_service_account_namespaces=argocd \
    policies=argocd \
    ttl=1h
```
This role authorizes the "vault-auth" service account in the default namespace and it gives it the default policy.

You can find the full documentation on configuring Kubernetes Authentication [here](https://www.vaultproject.io/docs/auth/kubernetes#configuration).

--- 

Once Argo CD and Kubernetes are configured, you can then set the required environment variables for the plugin:
```
VAULT_ADDR: Your HashiCorp Vault Address
AVP_TYPE: vault
AVP_AUTH_TYPE: k8s
AVP_K8S_MOUNT_PATH: Mount Path of your kubernetes Auth (optional)
AVP_K8S_ROLE: Your Kuberetes Auth Role
AVP_K8S_TOKEN_PATH: Path to JWT (optional)
```

##### Userpass Authentication
For Userpass Authentication, these are the required parameters:
```
VAULT_ADDR: Your HashiCorp Vault Address
AVP_TYPE: vault
AVP_AUTH_TYPE: userpass
AVP_USERNAME: Your Username
AVP_PASSWORD: Your Password
```

##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: vault-example
  annotations:
    avp.kubernetes.io/path: "secret/data/database"
type: Opaque
data:
  username: <username>
  password: <password>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: vault-example
type: Opaque
data:
  username: <path:secret/data/database#username>
  password: <path:secret/data/database#password>
```

###### Versioned secrets

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: vault-example
  annotations:
    avp.kubernetes.io/path: "secret/data/database"
    avp.kubernetes.io/secret-version: "2" # 2 is the latest revision in this example
type: Opaque
data:
  username: <username>
  password: <password>
  username-current: <path:secret/data/database#username#2> # same as <username>
  password-current: <path:secret/data/database#password#2> # same as <password>
  username-old: <path:secret/data/database#username#1>
  password-old: <path:secret/data/database#password#1>
```

**Note**: Only Vault KV-V2 backends support versioning. Versions specified with a KV-V1 Vault will be ignored and the latest version will be retrieved.

### IBM Cloud Secrets Manager

The path for IBM Cloud Secret Manager secrets can be specified in two ways:
1. `ibmcloud/<SECRET_TYPE>/secrets/groups/<GROUP>#<SECRET_NAME>`, or
2. `ibmcloud/<SECRET_TYPE>/secrets/groups/<GROUP>/<SECRET_NAME>#<SECRET_KEY>`

Where:
* `<SECRET_TYPE>` can be one of the following: `arbitrary`, `iam_credentials`, `imported_cert`, `kv`, `private_cert`, `public_cert`, or `username_password`.
* `<GROUP>` can be a secret group ID or name.
* `<SECRET_NAME>` is the name of the secret.
* `<SECRET_KEY>` is the key name within the secret. Specifically, the following keys are available for extraction:
  * `api_key` for the `iam_credentials` secret type
  * `username` and `password` for the `username_password` secret type
  * `certificate`, `private_key`, `intermediate` for the `imported_cert` or `public_cert` secret types
  * `certificate`, `private_key`, `issuing_ca`, `ca_chain` for the `private_cert` secret type
  * any key of the `kv` secret type
  `<SECRET_KEY>` is not supported for the `arbitrary` secret type.

When using the first path syntax, secrets that are JSON data (i.e, non `arbitrary` secrets or an `arbitrary` secret with JSON `payload`) can have select keys (listed under `<SECRET_KEY>` above) interpolated with the [jsonPath](./howitworks.md#jsonPath) modifier. With the second path syntax, the interpolation with the `jsonPath` modifier is not necessary. 

##### Authentication

IAM authentication is only supported at this time. The following parameters are required for IAM authentication:
```
AVP_IBM_INSTANCE_URL or VAULT_ADDR: Your IBM Cloud Secret Manager Endpoint
AVP_TYPE: ibmsecretsmanager
AVP_IBM_API_KEY: Your IBM Cloud API Key
```

##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: ibm-example
  annotations:
    avp.kubernetes.io/path: "ibmcloud/arbitrary/secrets/groups/123" # 123 represents your Secret Group ID
type: Opaque
data:
  username: <username>
  password: <password>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: ibm-example
type: Opaque
data:
  username: <path:ibmcloud/arbitrary/secrets/groups/123#username>
  password: <path:ibmcloud/arbitrary/secrets/groups/123#password>
```

###### Non-arbitrary secret

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: ibm-example
  annotations:
    avp.kubernetes.io/path: "ibmcloud/imported_cert/secrets/groups/123" # 123 represents your Secret Group ID
type: Opaque
stringData:
  PUBLIC_CRT: |
    <my-cert-secret | jsonPath {.certificate}>
  PUBLIC_CRT_PREVIOUS: |
    <path:ibmcloud/imported_cert/secrets/groups/123#my-cert-secret#previous | jsonPath {.certificate}>
  PRIVATE_KEY: |
    <my-cert-secret | jsonPath {.private_key}>
```

###### Non-arbitrary secrets (alternative path syntax)

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: ibm-example
  annotations:
    avp.kubernetes.io/path: "ibmcloud/imported_cert/secrets/groups/myGroup/my-cert-secret"
type: Opaque
stringData:
  PUBLIC_CRT: |
    <certificate>
  PRIVATE_KEY: |
    <private_key>
  USERNAME: <path:ibmcloud/username_password/secrets/groups/myGroup/basic-auth#username>
  PASSWORD: <path:ibmcloud/username_password/secrets/groups/myGroup/basic-auth#password>
```

### AWS Secrets Manager

##### AWS Authentication
Refer to the [AWS SDK for Go V2
documentation](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials) for
supplying AWS credentials. Supported credentials and the order in which they are loaded are
described [here](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials).

**Note About Region**
If you provide the full AWS ARN as the secret path, ex. `arn:aws:secretsmanager:us-east-1:123123123:secret:some-secret`, 
the region from the ARN (us-east-1) in this example, will take precedents over the AWS_REGION environment variable listed below. 

These are the parameters for AWS:
```
AVP_TYPE: awssecretsmanager
AWS_REGION: Your AWS Region (Optional: defaults to us-east-2)
```

##### Examples

###### Path Annotation

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-example
  annotations:
    avp.kubernetes.io/path: "test-aws-secret" # The name of your AWS Secret
stringData:
  sample-secret: <test-secret>
type: Opaque
```

###### Inline Path

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-example
stringData:
  sample-secret: <path:test-aws-secret#test-secret>
type: Opaque
```

###### Versioned secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-example
  annotations:
    avp.kubernetes.io/path: "some-path/secret"
    avp.kubernetes.io/secret-version: "AWSCURRENT"
stringData:
  sample-secret: <test-secret>
  sample-secret-again: <path:some-path/secret#test-secret#AWSCURRENT>
  sample-secret-old: <path:some-path/secret#test-secret#AWSPREVIOUS>
type: Opaque
```

###### Secret in the same account

The 'friendly' name of the secret can be used in this case.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-example
stringData:
  sample-secret: <path:test-aws-secret#test-secret>
type: Opaque
```

###### Secret in a different account

The arn of the secret needs to be used in this case:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-example
stringData:
  sample-secret: <path:arn:aws:secretsmanager:<REGION>:<ACCOUNT_NUMBER>:<SECRET_ID>#<key>>
type: Opaque
```

###### Retrieving of binary data

Since there is no way to set a key for binary type in AWS Secret Manager, set the `<key>` part to `SecretBinary` to retrieve binary data:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-example
stringData:
  sample-secret: <path:arn:aws:secretsmanager:<REGION>:<ACCOUNT_NUMBER>:<SECRET_ID>#SecretBinary>
type: Opaque
```

**NOTE**
For cross account access there is the need to configure the correct permissions between accounts, please check:
https://aws.amazon.com/premiumsupport/knowledge-center/secrets-manager-share-between-accounts
https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access_examples_cross.html

### GCP Secret Manager

##### GCP Authentication
Refer to the [Authentication Overview](https://cloud.google.com/docs/authentication) for Google Cloud APIs.

These are the parameters for GCP:
```
AVP_TYPE: gcpsecretmanager
```

##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: projects/12345678987/secrets/test-secret
type: Opaque
data:
  password: <test-secret>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:projects/12345678987/secrets/test-secret#test-secret>
```

###### Versioned secrets

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "projects/12345678987/secrets/test-secret"
    avp.kubernetes.io/secret-version: "latest"
type: Opaque
data:
  current-password: <password>
  current-password-again: <path:projects/12345678987/secrets/test-secret#password#latest>
  password-old: <path:projects/12345678987/secrets/test-secret#password#another-version-id>
```

### AZURE Key Vault

##### Azure Authentication
Refer to the [Use environment-based authentication](https://docs.microsoft.com/en-us/azure/developer/go/azure-sdk-authorization#use-environment-based-authentication) in the Azure SDK for Go.

Any secrets that are disabled in the key vault will be skipped if found.

For Azure, `path` is the unique name of your key vault.

**Note**: Versioning is only supported for inline paths.

**Note**: Due to the way the Azure backend works, templates that use _inline-path placeholders are more efficient_
(fewer HTTP calls and therefore lower chance of hitting rate limit) than generic placeholders.

These are the parameters for Azure:
```
AVP_TYPE: azurekeyvault
```

##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "keyvault"
type: Opaque
data:
  password: <test-secret>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:keyvault#test-secret>
```

###### Versioned secrets

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  current-password: <path:keyvault#password>
  current-password-again: <path:keyvault#password#8f8da2e06c8240808ee439ff093803b5>
  password-old: <path:keyvault#password#33740fc26214497f8904d93f20f7db6d>
```

### SOPS
##### SOPS Authentication
Refer to the [SOPS project page](https://github.com/mozilla/sops) for authentication options/environment variables.

For SOPS, `path` is file path to a JSON or YAML file encrypted using SOPS  and `key` is a top level key in the document, `jsonpath` can be used to fetch subkeys.

**Note**: Versioning is not supported.

These are the parameters for SOPS:
```
AVP_TYPE: sops
```

##### Examples
Given a file encrypted with SOPS named `example.yaml` and containing the following data:
```yaml
test-secret: test-data
parent:
  child: value
```

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "example.yaml"
type: Opaque
data:
  password: <test-secret>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:example.yaml#test-secret>
```

###### Sub key

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "example.yaml"
type: Opaque
stringData:
  password: <parent | jsonPath {.child}>
```

### Yandex Cloud Lockbox
##### YCL Authentication
Refer to the [IAM overview](https://cloud.yandex.com/en/docs/iam/concepts/) for yandex cloud APIs authorization.

These are the parameters for YCL:
```
AVP_TYPE: yandexcloudlockbox
AVP_YCL_SERVICE_ACCOUNT_ID: Service account ID
AVP_YCL_KEY_ID: Service account authorized Key ID
AVP_YCL_PRIVATE_KEY: Service account authorized private key
```
##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "secret-id"
type: Opaque
data:
  password: <key>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:secret-id#key>
```

###### Versioned secrets

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
    avp.kubernetes.io/path: "secret-id"
    avp.kubernetes.io/secret-version: "version-id"
type: Opaque
data:
  current-password: <password>
  current-password-again: <path:secret-id#password#version-id>
  password-old: <path:secret-id#password#old-version-id>
```

### 1Password Connect

**Note**: The 1Password Connect backend does not support versioning, so specifying a version will be ignored.

##### 1Password Connect Authentication

Refer to the [1Password Secrets Automation overview](https://support.1password.com/secrets-automation/) for 1Password Connect usage.

These are the parameters for 1Password Connect:

```
AVP_TYPE: 1passwordconnect
OP_CONNECT_TOKEN: Your 1Password Connect access token
OP_CONNECT_HOST: The hostname of your 1Password Connect server
```

##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "vaults/vault-uuid-or-title/items/item-uuid-or-title"
type: Opaque
data:
  password: <key>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:vaults/vault-uuid-or-title/items/item-uuid-or-title#key>
```

### Keeper Secrets Manager

**Note**: The Keeper Secrets Manager backend does not support versioning, or annotations. It does not support injecting attached files.

#### Keeper Authentication

These are the parameters for Keeper:
```
AVP_TYPE: keepersecretsmanager
AVP_KEEPER_CONFIG_PATH: the path to the keeper configuration file on disk.
```

##### Examples
Examples assume that the secrets are not saved base64 encoded in the Secret Server.

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "secret-id"
type: Opaque
stringData:
  password: <key>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:secret-id#key | base64encode>
```

### Delinea Secret Server

**Note**: The Delinea Secret Server backend does not support versioning.

##### Delinea Authentication
Refer to the [REST API documentation on your Delinea Secret Server](https://your-delinea-server/SecretServer/Documents/restapi/) for API authentication.

These are the parameters for Delinea:
```
AVP_TYPE: delineasecretserver
AVP_DELINEA_URL: The URL of the Dilenea Secret Server
AVP_DELINEA_USER: The account for authentication
AVP_DELINEA_PASSWORD: The password for authentication

Optional:
AVP_DELINEA_DOMAIN: The authentication domain (e.g. the Active Directory domain)
```
##### Examples
Examples assume that the secrets are not saved base64 encoded in the Secret Server.

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "secret-id"
type: Opaque
stringData:
  password: <key>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:secret-id#key | base64encode>
```

### Kubernetes Secret

Inject values from any kubernetes secret

**Note**: The Kubernetes Secret backend does not support versioning

##### Kubernetes Secret Authentication

Backend inherits same in-cluster service-account as the plugin itself

These are the parameters for Kubernetes Secret:

```
AVP_TYPE: kubernetessecret
```

##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "my-secret"
type: Opaque
data:
  password: <key>
```

###### Path Annotation With Namespace

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "prod:my-secret"
type: Opaque
data:
  password: <key>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:my-secret#key>
```

###### Inline Path With Namespace

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:prod:my-secret#key>
```

### CyberArk Secrets Manager

This guide explains how to integrate ArgoCD Vault Plugin with CyberArk Secrets Manager for secure secret retrieval in your Kubernetes clusters. CyberArk Secrets Manager provides two authentication methods, each suited for different deployment scenarios.

#### Prerequisites

Before setting up CyberArk Secrets Manager integration, ensure you have:

- **Helm 3.x** installed and configured
- **kubectl** installed and configured to access your cluster with appropriate **RBAC permissions** to create namespaces, ConfigMaps, and modify deployments
- **ArgoCD** deployed in your Kubernetes cluster (typically in the `argocd` namespace)
- CyberArk Secrets Manager **certificate file**

#### Authentication Methods
CyberArk Secrets Manager supports two authentication methods for Kubernetes:
1. **Certificate-based Kubernetes Authentication** - Uses mutual TLS (mTLS) for authentication
2. **JWT (Token-based) Kubernetes Authentication** - Uses Kubernetes service account JWT tokens

**Before you begin, collect the following information:**

| Parameter             | Description                                                                                  |
|-----------------------|----------------------------------------------------------------------------------------------|
| **SM_ACCOUNT**        | The account name used when deploying Secrets Manager. **Example:** `conjur`                  |
| **SM_URL**            | The Secrets Manager service URL.                                                             |
| **SM_CERT_FILE_PATH** | The path to the Secrets Manager certificate file.                                            |
| **SM_SERVICE_ID**     | The service ID that will be used for Kubernetes Authenticator.<br>**Example:** `dev-cluster` |



#### Certificate-based Kubernetes Authentication

Enables Kubernetes workloads to securely authenticate with CyberArk Secrets Manager using mutual TLS (mTLS).

> **Note:** Certificate-based authentication is **not supported** by CyberArk Secrets Manager SaaS.

##### 1. Initial Configuration

**Kubernetes cluster admin:**

In this step, you create the following Kubernetes resources for the Kubernetes Authenticator inside your Kubernetes cluster:

- A namespace, called `cyberark-conjur`
- A service account for the Kubernetes Authenticator, `authn-k8s-sa`
- A cluster role with the necessary RBAC permissions, `conjur-clusterrole`
- A Golden ConfigMap, `conjur-configmap`, containing connection and configuration information

```bash
helm repo add cyberark https://cyberark.github.io/helm-charts
  
helm install cluster-prep cyberark/conjur-config-cluster-prep \
--namespace cyberark-conjur \
--create-namespace \
--set conjur.account="$SM_ACCOUNT" \
--set conjur.applianceUrl="$SM_URL" \
--set conjur.certificateBase64="$(cat $SM_CERT_FILE_PATH | base64 -w 0)" \
--set authnK8s.authenticatorID="$SM_SERVICE_ID" \
--set authnK8s.serviceAccount.name="authn-k8s-sa" \
--set authnK8s.clusterRole.name="conjur-authn-role"
```

> **Note:** Use the goldenConfigMap `name` and `namespace` created here in the subsequent steps.

##### 2. Define and Load a Kubernetes Authenticator Policy

Refer to the [CyberArk documentation](https://docs.cyberark.com/conjur-enterprise/latest/en/content/integrations/k8s-ocp/k8s-k8s-authn.htm?tocpath=Integrations%7COpenShift%252FKubernetes%7CAuthenticate%20OpenShift%252FKubernetes%7C_____3) for detailed policy configuration.

##### 3. Create a Workload Identity for Kubernetes

Refer to the [CyberArk documentation](https://docs.cyberark.com/secrets-manager-sh/latest/en/content/integrations/k8s-ocp/k8s-app-identity.htm) for detailed workload identity setup.

##### 4. Configure Kubernetes Authenticator Client

###### 4.1. Create ConfigMap for the Kubernetes Authenticator Client

```bash
helm install namespace-prep cyberark/conjur-config-namespace-prep \
--namespace argocd \
--set authnK8s.goldenConfigMap="conjur-configmap" \
--set authnK8s.namespace="cyberark-conjur"
```

###### 4.2. Create ConfigMap for the AVP Plugin

```bash
kubectl create configmap vault-configuration \
--namespace argocd \
--from-literal=AVP_SECRETS_MANAGER_ACCOUNT="$SM_ACCOUNT" \
--from-literal=AVP_SECRETS_MANAGER_URL="$SM_URL" \
--from-literal=AVP_TYPE=cyberarksecretsmanager \
--from-literal=AVP_SECRETS_MANAGER_TOKEN_FILE=/run/conjur/access-token \
--from-file=AVP_SECRETS_MANAGER_SSL_CERT="$SM_CERT_FILE_PATH"
```

###### 4.3. Define CMP Plugin Configuration

Create a ConfigMap named `cmp-plugin` in the `argocd` namespace with the following content:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cmp-plugin
  namespace: argocd
data:
  avp.yaml: |
    apiVersion: argoproj.io/v1alpha1
    kind: ConfigManagementPlugin
    metadata:
      name: argocd-vault-plugin
    spec:
      allowConcurrency: true
      discover:
        find:
          command:
            - sh
            - "-c"
            - "find . -name '*.yaml' | xargs -I {} grep \"<path\\|avp\\.kubernetes\\.io\" {} | grep ."
      generate:
        command: ["argocd-vault-plugin", "generate", "."]
      lockRepo: false
```

Apply the ConfigMap:

```bash
kubectl apply -f cmp-plugin.yaml
```
> **Note:** The example above assumes Kubernetes yaml manifests. Adjust the `command` accordingly if using other formats (e.g., Helm, Kustomize).

###### 4.4. Patch argocd-repo-server

Patch the argocd-repo-server deployment to integrate with CyberArk Secrets Manager using the Kubernetes Authenticator Client and ArgoCD Vault Plugin as sidecars.

**Before applying, populate these values:**
- `CONJUR_AUTHN_LOGIN`: The host identity for authentication (e.g., `host/argocd-repo-server`)
- avp container `image`: An argocd-vault-plugin image with CyberArk Secrets Manager support

```yaml
spec:
  template:
    spec:
      containers:
        - name: authenticator
          image: cyberark/conjur-authn-k8s-client
          imagePullPolicy: Always
          env:
            - name: CONJUR_AUTHN_LOGIN
              value: <the host identity, e.g., host/argocd-repo-server>
            - name: MY_POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: MY_POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          envFrom:
            - configMapRef:
                name: conjur-connect
          volumeMounts:
            - name: conjur-access-token
              mountPath: /run/conjur
        - name: avp
          image: <image>
          command:
            - /var/run/argocd/argocd-cmp-server
          envFrom:
            - configMapRef:
                name: vault-configuration
          securityContext:
            runAsNonRoot: true
            runAsUser: 999
          volumeMounts:
            - name: var-files
              mountPath: /var/run/argocd
            - name: plugins
              mountPath: /home/argocd/cmp-server/plugins
            - name: tmp
              mountPath: /tmp
            - name: cmp-plugin
              mountPath: /home/argocd/cmp-server/config/plugin.yaml
              subPath: avp.yaml
            - name: conjur-access-token
              mountPath: /run/conjur
        - name: argocd-repo-server
          volumeMounts:
            - name: argocd-repo-server-secret
              mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      volumes:
        - name: conjur-access-token
          emptyDir:
            medium: Memory
        - name: cmp-plugin
          configMap:
            name: cmp-plugin
            defaultMode: 420
        - name: argocd-repo-server-secret
          secret:
            secretName: argocd-repo-server-secret
            defaultMode: 511
            optional: true
```

Apply the patch:

```bash
kubectl patch deployment argocd-repo-server -n argocd --patch-file=patch.yaml
```

##### 5. Proxy configuration (if needed)
If your CyberArk Secrets Manager is behind a reverse proxy that terminates the TLS connection, configure the proxy to forward the client certificate using the `X-SSL-Client-Certificate` header.

For example, with NGINX as a reverse proxy, add the following directive inside the `server` block (ensure SSL client certificate authentication is enabled):

```nginx
proxy_set_header X-SSL-Client-Certificate $ssl_client_escaped_cert;
```

#### JWT Authentication

The JWT authentication method uses short-lived JWT tokens issued by the Kubernetes API server, mounted as a projected volume. The default TTL is 1 hour (3600 seconds), configurable via `expirationSeconds`. Kubernetes automatically rotates tokens before expiration to ensure continuous authentication.

> **Note:** For more details, see the [CyberArk JWT authentication documentation](https://docs.cyberark.com/conjur-enterprise/latest/en/content/integrations/k8s-ocp/k8s-jwt-authn.htm?tocpath=Integrations%7COpenShift%252FKubernetes%7CAuthenticate%20OpenShift%252FKubernetes%7C_____2).

##### 1. Initial Configuration

**Kubernetes cluster admin:**

In this step, you create the following Kubernetes resources for the Kubernetes Authenticator inside your Kubernetes cluster:

- A namespace, called `cyberark-conjur-jwt`
- A Golden ConfigMap, `conjur-configmap`, containing Conjur connection and configuration information

```bash
helm repo add cyberark https://cyberark.github.io/helm-charts

helm install cluster-prep cyberark/conjur-config-cluster-prep \
--namespace "cyberark-conjur-jwt" \
--create-namespace \
--set conjur.account="$SM_ACCOUNT" \
--set conjur.applianceUrl="$SM_URL" \
--set conjur.certificateBase64="$(cat $SM_CERT_FILE_PATH | base64 -w 0)" \
--set authnK8s.authenticatorID="$SM_SERVICE_ID" \
--set authnK8s.clusterRole.create=false \
--set authnK8s.serviceAccount.create=false
```

> **Note:** Use the goldenConfigMap `name` and `namespace` created here in the subsequent steps.

##### 2. Define and Load a Kubernetes Authenticator Policy

Refer to the [CyberArk documentation](https://docs.cyberark.com/secrets-manager-sh/latest/en/content/integrations/k8s-ocp/k8s-jwt-authn.htm) for detailed policy configuration.

##### 3. Create a Workload Identity for Kubernetes

Refer to the [CyberArk documentation](https://docs.cyberark.com/secrets-manager-sh/latest/en/content/integrations/k8s-ocp/cjr-k8s-authn-client-authjwt.htm) for detailed workload identity setup.

##### 4. Configure Kubernetes Authenticator Client

###### 4.1. Create ConfigMap for Kubernetes Authenticator Client

```bash
helm install namespace-prep cyberark/conjur-config-namespace-prep \
--namespace argocd \
--set conjurConfigMap.authnMethod="authn-jwt" \
--set authnK8s.goldenConfigMap="conjur-configmap" \
--set authnK8s.namespace="cyberark-conjur-jwt" \
--set authnRoleBinding.create="false"
```

###### 4.2. Create ConfigMap for AVP Plugin

```bash
kubectl create configmap vault-configuration \
--namespace argocd \
--from-literal=AVP_SECRETS_MANAGER_ACCOUNT="$SM_ACCOUNT" \
--from-literal=AVP_SECRETS_MANAGER_URL="$SM_URL" \
--from-literal=AVP_TYPE=cyberarksecretsmanager \
--from-literal=AVP_SECRETS_MANAGER_TOKEN_FILE=/run/conjur/access-token \
--from-file=AVP_SECRETS_MANAGER_SSL_CERT="$SM_CERT_FILE_PATH"
```

###### 4.3. Define CMP Plugin Configuration

Create a ConfigMap named `cmp-plugin` in the `argocd` namespace with the following content:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cmp-plugin
  namespace: argocd
data:
  avp.yaml: |
    apiVersion: argoproj.io/v1alpha1
    kind: ConfigManagementPlugin
    metadata:
      name: argocd-vault-plugin
    spec:
      allowConcurrency: true
      discover:
        find:
          command:
            - sh
            - "-c"
            - "find . -name '*.yaml' | xargs -I {} grep \"<path\\|avp\\.kubernetes\\.io\" {} | grep ."
      generate:
        command: ["argocd-vault-plugin", "generate", "."]
      lockRepo: false
```

Apply the ConfigMap:

```bash
kubectl apply -f cmp-plugin.yaml
```
> **Note:** The example above assumes Kubernetes yaml manifests. Adjust the `command` accordingly if using other formats (e.g., Helm, Kustomize).

###### 4.4. Patch argocd-repo-server

Patch the argocd-repo-server Deployment to enable CyberArk Secrets Manager integration using the Kubernetes Authenticator Client as a sidecar container for authentication. The authenticator uses a JWT token (from the projected service account token) to authenticate with CyberArk Secrets Manager and writes the access token to a shared in-memory volume.

The AVP (ArgoCD Vault Plugin) container reads the access token from the shared memory volume (`/run/conjur`) and uses the `vault-configuration` ConfigMap to fetch secrets.

**Before applying, populate these values:**
- avp container `image`: An argocd-vault-plugin image with CyberArk Secrets Manager support
- serviceAccountToken `audience`: The audience value as defined in your CyberArk Secrets Manager authn policy

```yaml
# patch_repo_server.yaml
spec:
  template:
    spec:
      containers:
        - name: authenticator
          image: cyberark/conjur-authn-k8s-client
          imagePullPolicy: Always
          env:
            - name: JWT_TOKEN_PATH
              value: /var/run/secrets/tokens/token
          envFrom:
            - configMapRef:
                name: conjur-connect
          volumeMounts:
            - name: conjur-access-token
              mountPath: /run/conjur
            - name: argocd-repo-server-secret
              mountPath: /var/run/secrets/tokens
              readOnly: true
        - name: avp
          image: <image>
          imagePullPolicy: Always
          command:
            - /var/run/argocd/argocd-cmp-server
          envFrom:
            - configMapRef:
                name: vault-configuration
          securityContext:
            runAsNonRoot: true
            runAsUser: 999
          volumeMounts:
            - name: var-files
              mountPath: /var/run/argocd
            - name: plugins
              mountPath: /home/argocd/cmp-server/plugins
            - name: tmp
              mountPath: /tmp
            - name: cmp-plugin
              mountPath: /home/argocd/cmp-server/config/plugin.yaml
              subPath: avp.yaml
            - name: conjur-access-token
              mountPath: /run/conjur
        - name: argocd-repo-server
          volumeMounts:
            - name: argocd-repo-server-secret
              mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      volumes:
        - name: conjur-access-token
          emptyDir:
            medium: Memory
        - name: argocd-repo-server-secret
          projected:
            sources:
              - serviceAccountToken:
                  path: token
                  expirationSeconds: 6000
                  audience: <audience>
        - name: cmp-plugin
          configMap:
            name: cmp-plugin
            defaultMode: 420
```

Apply the patch:

```bash
kubectl patch deployment argocd-repo-server -n argocd --patch-file=patch.yaml
```

---

#### Secret Retrieval

Once authentication is set up, secrets can be retrieved from CyberArk Secrets Manager using either path annotations or inline paths in your Kubernetes Secret manifests.

For detailed information on secret retrieval patterns and usage, refer to the [official Argo CD Vault Plugin documentation](https://argocd-vault-plugin.readthedocs.io/en/stable/howitworks/).

---

#### Troubleshooting
1. **Verify ConfigMaps are created:**
   ```bash
   kubectl get configmap -n argocd conjur-connect vault-configuration cmp-plugin -o yaml
   ```
2. **Check argocd-repo-server pods:**
   ```bash
   kubectl get pods -n argocd -l app.kubernetes.io/name=argocd-repo-server
   ```
3. **Inspect authenticator logs:**
   ```bash
   kubectl logs -n argocd <pod-name> -c authenticator --tail=50
   ```
4. **Verify access token is being written:**
   ```bash
   kubectl exec -n argocd <pod-name> -c avp -- ls -la /run/conjur/
   ```

##### Additional Resources

- [CyberArk Kubernetes Authentication Documentation](https://docs.cyberark.com/secrets-manager-sh/latest/en/content/integrations/k8s-ocp/k8s-admin-lp.htm?tocpath=Integrations%7COpenShift%252FKubernetes%7CAuthenticate%20OpenShift%252FKubernetes%7C_____0)
- [ArgoCD Vault Plugin Documentation](https://argocd-vault-plugin.readthedocs.io/)
- [CyberArk Community Forums](https://community.cyberark.com/s/)