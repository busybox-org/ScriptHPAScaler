# ScriptHPAScaler
# 概述
`ScriptHPAScaler`是一个脚本式方案的`kubernetes`水平`pod` `autoscaler`控制器. 您可以使用`ScriptHPAScaler`在 kubernetes 中定义的支持`scale`资源(例如`Deployment`和`StatefulSet`）的任何种类的对象。

# 安装
1. 安装CRD
```bash
kubectl apply -f config/crd/
```
2. 安装`RBAC`
```bash
kubectl apply -f config/rbac/
```
3. 部署`Manager`
```bash
kubectl apply -f config/manager/
```
4. 验证安装
```bash
kubectl get deploy scripthpascaler-controller -n kube-system -o wide 

➜ kubectl get deploy scripthpascaler-controller -n kube-system
NAME                            DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
scripthpascaler-controller         1         1         1            1           49s
```

# 例子
请试用[示例文件夹中的示例](https://github.com/xmapst/ScriptHPAScaler/tree/main/config/samples)。