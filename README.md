# LineQ-Operator
This repository has implemented a Kubernetes operator for the <a href="https://github.com/hamedetemaad/Lineq">LineQ</a> project
## Installation

### 0 - Install HAproxy ingress
```
touch haproxy-auxiliary.cfg
kubectl create ns haproxy-controller
kubectl create configmap haproxy-auxiliary-configmap \
  --from-file haproxy-auxiliary.cfg \
  --namespace haproxy-controller
helm install haproxy-kubernetes-ingress haproxytech/kubernetes-ingress \
  -f cfg/haproxy-values.yaml \
  --namespace haproxy-controller
```
### 1 - Install LineQ
```
helm repo add lineq-charts https://hamedetemaad.github.io/helm-charts/
helm repo update
helm install lineq lineq-charts/lineq -n lineq
```

### 2 - Install LineQ-Operator
```
helm install lineq-operator lineq-charts/lineq-operator -n lineq
```

### 3 - Create waiting room CRD

```
apiVersion: lineq.io/v1alpha1
kind: WaitingRoom
metadata:
  name: test
  namespace: test
spec:
  path: "/"
  activeUsers: 20
  schema: "http"
  host: "example.com"
  backendSvcAddr: test-service
  backendSvcPort: 80
```
