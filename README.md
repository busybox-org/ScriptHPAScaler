# ScriptHPAScaler
## 概述
`ScriptHPAScaler`是一个脚本式方案的`kubernetes`水平`pod` `autoscaler`控制器. 您可以使用`ScriptHPAScaler`在 kubernetes 中定义的支持`scale`资源(例如`Deployment`和`StatefulSet`）的任何种类的对象。

## 安装
```bash
kubectl apply -f dist/install.yaml
```
## 验证安装

```bash
kubectl get deploy scripthpascaler-controller -n kube-system -o wide 

➜ kubectl get deploy scripthpascaler-controller -n kube-system
NAME                            DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
scripthpascaler-controller         1         1         1            1           49s
```

## 例子
请试用[示例文件夹中的示例](https://github.com/xmapst/ScriptHPAScaler/tree/main/config/samples)。