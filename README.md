# Kubernetes Cluster API Provider Metalsoft

Kubernetes-native declarative infrastructure for Metalsoft.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/cluster-api-provider-metalsoft:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/cluster-api-provider-metalsoft:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```

## Contributing

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## Dev notes

### Prerequisites
-   git
-   make
-   [Go](https://go.dev/dl/)
-   [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder/releases)
-   [Docker](https://docs.docker.com/get-docker/)
-   [Tilt](https://docs.tilt.dev/install.html)
-   [Kubectl](https://kubernetes.io/docs/tasks/tools/)
-   [Kustomize](https://github.com/kubernetes-sigs/kustomize)
-   [kind](https://kind.sigs.k8s.io/)
-   [clusterctl](https://github.com/kubernetes-sigs/cluster-api/releases)

### Getting started

- Clone this repo into ~/go/src/github.com/metalsoft-io/cluster-api-provider-metalsoft
 ```
mkdir -p ~/go/src/github.com/metalsoft-io
git clone https://github.com/metalsoft-io/cluster-api-provider-metalsoft.git ~/go/src/github.com/metalsoft-io/cluster-api-provider-metalsoft
 ```
- Clone cluster-api repo into ~/go/src/sigs.k8s.io/cluster-api (optionally, fork it and clone the forked repo)
```
mkdir -p ~/go/src/sigs.k8s.io
git clone https://github.com/kubernetes-sigs/cluster-api.git ~/go/src/sigs.k8s.io/cluster-api
```
- Create tilt-settings.json in the cluster-api directory
```
cat > ~/go/src/sigs.k8s.io/cluster-api/tilt-settings.json <<EOF
{
    "default_registry": "registry.metalsoft.dev/cluster-api",
    "provider_repos": ["../../github.com/metalsoft-io/cluster-api-provider-metalsoft"],
    "enable_providers": ["metalsoft-capi-provider", "kubeadm-bootstrap", "kubeadm-control-plane"],
    "kustomize_substitutions": {
        "EXP_MACHINE_POOL": "true",
        "EXP_CLUSTER_RESOURCE_SET": "true"
    },
    "extra_args": {
        "metalsoft-capi-provider": ["-zap-log-level=debug"]
    },
    "debug": {
        "metalsoft-capi-provider": {
            "continue": true,
            "port": 31000
        }
    }
}
EOF
```
- Start Docker Desktop on your workstation
- Create Kind cluster config in the cluster-api directory
```
cat  > ~/go/src/sigs.k8s.io/cluster-api/kind-cluster-with-extramounts.yaml <<EOF  
kind: Cluster  
apiVersion: kind.x-k8s.io/v1alpha4  
name: capi-test  
nodes:  
- role: control-plane 
  image: kindest/node:v1.29.2@sha256:51a1434a5397193442f0be2a297b488b6c919ce8a3931be0ce822606ea5ca245 
  extraMounts:  
  - hostPath: /var/run/docker.sock  
    containerPath: /var/run/docker.sock  
EOF
```
- Create Kind cluster
```
kind create cluster --config ~/go/src/sigs.k8s.io/cluster-api/kind-cluster-with-extramounts.yaml
```
- Bring Tilt up
```
cd ~/go/src/sigs.k8s.io/cluster-api
tilt up
```
- Press spacebar to open Tilt UI in browser

### References
- [Cluster API Book](https://cluster-api.sigs.k8s.io)
- [Kubebuilder Book](https://book.kubebuilder.io)
- [Tutorial - implementation part](https://capi-samples.github.io/kubecon-na-2022-tutorial/docs/cluster-implementation)
- [Tutorial Talk Video](https://www.youtube.com/watch?v=5-X6haLVO5A&t=3875s)





