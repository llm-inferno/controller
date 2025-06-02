# Steps to run demo

## Create kind cluster

```bash
kind create cluster --name mycluster
```

Optionally, load locally built images, and list loaded images, if needed.
Otherwise, they will be pulled from `quay.io/atantawi`.

```bash
kind load docker-image inferno inferno-controller -n mycluster
docker exec -it mycluster-control-plane crictl images
```

## Deploy REST API inferno optimizer

`$INFERNO_REPO` is the path to the cloned [inferno optimizer repository](https://github.com/llm-inferno/inferno).
This will create a namespace `inferno`, a deployment `inferno-optimizer`, and a service `inferno-optimizer`.

```bash
kubectl apply -f $INFERNO_REPO/manifests/yamls/deploy-optimizer.yaml
```

Inspect the logs of the running optimizer pod.

```bash
POD_OPT=$(kubectl get pod -l app=inferno-optimizer -n inferno -o jsonpath="{.items[0].metadata.name}")
kubectl logs -f $POD_OPT -n inferno 
```

## Deploy inferno controller

`$INFERNO_CTRL_REPO` is the path to the cloned [inferno controller repository](https://github.com/llm-inferno/controller).
This will create namespace `controller-system`, install the CRDs, a deployment `controller-controller-manager`, as well as required RBACs and services.

```bash
kubectl apply -f $INFERNO_CTRL_REPO/manifests/yamls/deploy.yaml
```

Inspect the logs of the running controller pod.

```bash
POD_CTRL=$(kubectl get pod -l app.kubernetes.io/name=controller -n controller-system -o jsonpath="{.items[0].metadata.name}")
kubectl logs -f $POD_CTRL -n controller-system 
```

## Deploy (sample data) custom resources

```bash
kubectl apply -k $INFERNO_CTRL_REPO/manifests/yamls/
```

## Patch optimizer CR to invoke optimizer

Watch changes to the optimizer custom resource.

```bash
watch kubectl get optimizer inferno -o yaml
```

Invoke the optimizer by setting the spec `optimize` to `true`.

```bash
kubectl patch --type merge --patch '{"spec":{"optimize":true}}' optimizer inferno
```

## Inspect optimization decisions

Get the server custom resources, then selectively `kubectl describe` them, checking `spec.desiredAlloc`.

```bash
kubectl get server
```

Or, forward the inferno optimizer service

```bash
kubectl port-forward service/inferno-optimizer -n inferno 8080:80
```

and watch changes to the servers.

```bash
watch -n 20 $(curl http://localhost:8080/getServers | grep -A +12 desiredAlloc)
```

## Cleanup

```bash
kubectl delete -k $INFERNO_CTRL_REPO/manifests/yamls/
kubectl delete -f $INFERNO_CTRL_REPO/manifests/yamls/deploy.yaml

kubectl delete -f $INFERNO_REPO/manifests/yamls/deploy-optimizer.yaml

kind delete cluster --name mycluster
```