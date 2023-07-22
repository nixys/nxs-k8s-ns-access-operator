# DEPRECATED

This repository is considered deprecated and will be archived. For the new version of this tool please go to [nxs-rbac-operator](https://github.com/nixys/nxs-rbac-operator) project.

# nxs-k8s-ns-access-operator

It is the Kubernetes operator to the automatically create the role binding to specific user for created namespaces with the names satisfied to the flollowing format: `$APP_NAME-msvc-$MICROSEVICE_NAME-$BRANCH_NAME` (using regexp: `^(.*)-msvc-.*$`).

The operator supports following optional env variables:
* `KUBECONFIG`: contains the path to kubeconfig file. Use this env variable in case when the operator execute outside the Kubernetes cluster.
* `CLUSTER_ROLE_NAME`: defines the cluster role name to be bindings in created namespace. If this env variable not set, cluster role `edit` will be used.

## Deploy

To deploy the operator into the Kubernetes cluster use the deployment `deployment.yml` from the `docs` directory:

```
kubectl apply -f docs
```

## Create new project user

To create by the operator role binding in created namespace you need to create the user with `$APP_NAME` name. Do the following steps:

* First you need to create certificate for new user. This action must be performed on Kubernetes master node:

```
openssl genrsa -out user.key 2048
openssl req -new -key user.key -out user.csr -subj "/CN=$APP_NAME/O=$APP_NAME"
openssl x509 -req -in user.csr -CA /etc/kubernetes/ssl/ca.pem -CAkey /etc/kubernetes/ssl/ca-key.pem -CAcreateserial -out user.crt -days 365
```

* Grant permission new user for create namespaces:

```
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ns-creator-$APP_NAME
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ns-creator
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: $APP_NAME
EOF
```

* On client node add new context to kubeconfig (placeholders: `CLUSTER_NAME`):

```
kubectl config set-credentials $APP_NAME --client-certificate=/path/to/user.crt  --client-key=/path/to/user.key
kubectl config set-context $APP_NAME-context --cluster=CLUSTER_NAME --user=$APP_NAME
```

The method above describes how to setup your kubeconfig to use certificate and key files. If you need to store the data of this files insteed, do the following:

* Transform certificate and key to base64-lines:

```
cat user.crt | base64 -w0
cat user.key | base64 -w0
```

* Replace the options `client-certificate` and `client-key` manually in appropriate block within `~/.kube/config` file to `client-certificate-data` and `client-key-data` and replace the path to certificate and key files with the data obtained above.
