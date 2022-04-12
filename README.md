# SupersetScalers
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

# 概述
 `SupersetScalers`是一个插件式方案的`kubernetes`水平`pod` `autoscaler`控制器. 您可以使用`HPAScaler`在 kubernetes 中定义的支持`scale`资源(例如`Deployment`和`StatefulSet`以及`ReplicaSet`）的任何种类的对象。

 # 安装
1. 安装CRD
```bash
kubectl apply -f config/crd/k8s.q1.com_hpascalers.yaml
```
2. 安装`RBAC`
```bash
# create ClusterRole 
kubectl apply -f config/rbac/role.yaml

# create ServiceAccount
kubectl apply -f config/rbac/service_account.yaml

# create ClusterRolebinding 
kubectl apply -f config/rbac/role_binding.yaml
```
3. 部署`supersetscaler-controller`
```bash
kubectl apply -f config/deploy/deploy.yaml
```
4. 验证安装
```bash
kubectl get deploy supersetscaler-controller -n kube-system -o wide 

➜  supersetscaler-controller git:(master) ✗ kubectl get deploy supersetscaler-controller -n kube-system
NAME                            DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
supersetscaler-controller         1         1         1            1           49s
```

# 例子
请试用[示例文件夹中的示例](https://github.com/xmapst/SupersetScalers/tree/main/example)。

# Plugin
- [x] rabbitmq
- [ ] rocketmq
- [ ] kafka
- [ ] redis
- [x] http
- [ ] shell
- [ ] python
- [ ] lua

# Makefile help
```text
Usage:
  make <target>

General
  help             Display this help.

Development
  manifests        Generate ClusterRole and CustomResourceDefinition objects.
  generate         Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
  fmt              Run go fmt against code.
  vet              Run go vet against code.
  test             Run tests.

Build
  build            Build manager binary.
  run              Run a controller from your host.
  docker-build     Build docker image with the manager.
  docker-push      Push docker image with the manager.

Deployment
  install          Install CRDs into the K8s cluster specified in ~/.kube/config.
  uninstall        Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
  deploy           Deploy controller to the K8s cluster specified in ~/.kube/config.
  undeploy         Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
  controller-gen   Download controller-gen locally if necessary.
  envtest          Download envtest-setup locally if necessary.
```
