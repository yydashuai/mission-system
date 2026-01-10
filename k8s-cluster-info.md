# Kubernetes 集群信息

> 生成时间: 2026-01-10
> 分析工具: kubectl

---

## 集群概览

### 基本信息
- **Kubernetes 版本**: v1.29.2
- **kubectl 版本**: v1.29.2
- **Kustomize 版本**: v5.0.4-0.20230601165947-6ce0bf390ce3
- **集群年龄**: 2天18小时
- **控制平面地址**: https://192.168.40.11:6443

### 集群组件
- **CoreDNS**: 运行正常
- **API Server**: 运行正常
- **Metrics Server**: 已安装并运行正常 (v0.8.0)

---

## 节点信息

集群共有 **3个节点** (1个控制平面节点 + 2个工作节点)

| 节点名称 | 角色 | 状态 | IP地址 | 操作系统 | 内核版本 | 容器运行时 | Kubernetes版本 | 年龄 |
|---------|------|------|--------|----------|----------|------------|---------------|------|
| k8s-master | control-plane | Ready | 192.168.40.11 | Rocky Linux 9.6 (Blue Onyx) | 5.14.0-570.17.1.el9_6.x86_64 | Docker 29.1.3 | v1.29.2 | 2d18h |
| k8s-node01 | worker | Ready | 192.168.40.12 | Rocky Linux 9.6 (Blue Onyx) | 5.14.0-570.17.1.el9_6.x86_64 | Docker 29.1.3 | v1.29.2 | 2d18h |
| k8s-node02 | worker | Ready | 192.168.40.13 | Rocky Linux 9.6 (Blue Onyx) | 5.14.0-570.17.1.el9_6.x86_64 | Docker 29.1.3 | v1.29.2 | 2d18h |

**节点特点**:
- 所有节点运行相同版本的 Kubernetes (v1.29.2)
- 所有节点使用 Docker 作为容器运行时 (v29.1.3)
- 所有节点运行 Rocky Linux 9.6 操作系统
- 所有节点状态为 Ready

---

## 命名空间

集群共有 **4个命名空间**:

| 命名空间 | 状态 | 年龄 | 说明 |
|---------|------|------|------|
| default | Active | 2d18h | 默认命名空间 |
| kube-node-lease | Active | 2d18h | 节点心跳租约 |
| kube-public | Active | 2d18h | 公共资源 |
| kube-system | Active | 2d18h | 系统组件 |

---

## Pod 资源统计

集群共运行 **14个 Pod**，全部位于 kube-system 命名空间。

### Pod 列表详情

| 命名空间 | Pod名称 | 就绪状态 | 状态 | 重启次数 | 年龄 |
|---------|---------|---------|------|---------|------|
| kube-system | calico-kube-controllers-558d465845-6l2kn | 1/1 | Running | 1 | 2d18h |
| kube-system | calico-node-jcljn | 1/1 | Running | 1 | 2d18h |
| kube-system | calico-node-qhvr4 | 1/1 | Running | 1 | 2d18h |
| kube-system | calico-node-zlsd5 | 1/1 | Running | 1 | 2d18h |
| kube-system | calico-typha-5b56944f9b-kxgqp | 1/1 | Running | 1 | 2d18h |
| kube-system | coredns-857d9ff4c9-45ncc | 1/1 | Running | 1 | 2d18h |
| kube-system | coredns-857d9ff4c9-s4lbd | 1/1 | Running | 1 | 2d18h |
| kube-system | etcd-k8s-master | 1/1 | Running | 1 | 2d18h |
| kube-system | kube-apiserver-k8s-master | 1/1 | Running | 1 | 2d18h |
| kube-system | kube-controller-manager-k8s-master | 1/1 | Running | 1 | 2d18h |
| kube-system | kube-proxy-2b56h | 1/1 | Running | 1 | 2d18h |
| kube-system | kube-proxy-lgp95 | 1/1 | Running | 1 | 2d18h |
| kube-system | kube-proxy-vrkbz | 1/1 | Running | 1 | 2d18h |
| kube-system | kube-scheduler-k8s-master | 1/1 | Running | 2 | 2d18h |

**Pod 健康状态**:
- ✅ 所有 Pod 状态正常 (Running)
- ✅ 所有 Pod 容器就绪 (Ready 1/1)
- ℹ️ 大部分 Pod 重启过1次（21小时前），kube-scheduler 重启过2次

---

## Deployment 资源

集群共有 **3个 Deployment**:

| 命名空间 | Deployment名称 | 期望副本 | 当前副本 | 可用副本 | 年龄 |
|---------|---------------|---------|---------|---------|------|
| kube-system | calico-kube-controllers | 1/1 | 1 | 1 | 2d18h |
| kube-system | calico-typha | 1/1 | 1 | 1 | 2d18h |
| kube-system | coredns | 2/2 | 2 | 2 | 2d18h |

**Deployment 状态**: 所有部署均正常，副本数符合期望

---

## Service 资源

集群共有 **3个 Service**:

| 命名空间 | Service名称 | 类型 | Cluster IP | 外部IP | 端口 | 年龄 |
|---------|------------|------|------------|--------|------|------|
| default | kubernetes | ClusterIP | 10.0.0.1 | - | 443/TCP | 2d18h |
| kube-system | calico-typha | ClusterIP | 10.9.161.150 | - | 5473/TCP | 2d18h |
| kube-system | kube-dns | ClusterIP | 10.0.0.10 | - | 53/UDP,53/TCP,9153/TCP | 2d18h |

**Service 说明**:
- **kubernetes**: Kubernetes API Server 服务
- **calico-typha**: Calico Typha 服务（网络策略）
- **kube-dns**: CoreDNS 服务（集群 DNS）

---

## 存储资源

- **PersistentVolume (PV)**: 0个
- **PersistentVolumeClaim (PVC)**: 0个

**说明**: 当前集群未配置持久化存储

---

## ConfigMap 资源

集群共有 **12个 ConfigMap**:

| 命名空间 | ConfigMap名称 | 数据项数 | 年龄 |
|---------|--------------|---------|------|
| default | kube-root-ca.crt | 1 | 2d18h |
| kube-node-lease | kube-root-ca.crt | 1 | 2d18h |
| kube-public | cluster-info | 1 | 2d18h |
| kube-public | kube-root-ca.crt | 1 | 2d18h |
| kube-system | calico-config | 4 | 2d18h |
| kube-system | coredns | 1 | 2d18h |
| kube-system | extension-apiserver-authentication | 6 | 2d18h |
| kube-system | kube-apiserver-legacy-service-account-token-tracking | 1 | 2d18h |
| kube-system | kube-proxy | 2 | 2d18h |
| kube-system | kube-root-ca.crt | 1 | 2d18h |
| kube-system | kubeadm-config | 1 | 2d18h |
| kube-system | kubelet-config | 1 | 2d18h |

---

## 网络配置

### CNI (Container Network Interface)
- **网络插件**: Calico
- **Calico 组件**:
  - calico-kube-controllers (Deployment)
  - calico-typha (Deployment + DaemonSet)
  - calico-node (DaemonSet，每个节点一个实例)

### 服务网络
- **Service CIDR**: 10.0.0.0/16 (推测，基于 ClusterIP 范围)
- **DNS 服务**: CoreDNS (2个副本)
- **DNS Service IP**: 10.0.0.10

---

## 集群特征总结

### 优点
1. ✅ 集群健康状态良好，所有节点和 Pod 运行正常
2. ✅ 使用成熟的网络方案 Calico
3. ✅ 高可用 DNS (CoreDNS 有2个副本)
4. ✅ 节点配置统一，便于管理
5. ✅ 使用较新的 Kubernetes 版本 (v1.29.2)

### 待改进项
1. ⚠️ 未安装 Metrics Server（无法使用 `kubectl top` 查看资源使用情况）
2. ⚠️ 控制平面节点只有1个（单点故障风险）
3. ⚠️ 未配置持久化存储 (PV/PVC)
4. ⚠️ 工作节点较少（仅2个），扩展性有限
5. ⚠️ 使用 Docker 作为容器运行时（Kubernetes 已弃用 dockershim，建议迁移到 containerd 或 CRI-O）

### 建议
1. 安装 Metrics Server 以支持资源监控: `kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml`
2. 配置持久化存储方案（如 NFS、Ceph、或云存储）
3. 考虑添加更多控制平面节点实现高可用（3或5个节点）
4. 迁移容器运行时从 Docker 到 containerd
5. 根据业务需求，考虑安装以下组件：
   - Ingress Controller (nginx-ingress, traefik)
   - 监控系统 (Prometheus + Grafana)
   - 日志收集 (EFK/ELK Stack)
   - Service Mesh (Istio, Linkerd)

---

## 快速命令参考

```bash
# 查看集群信息
kubectl cluster-info

# 查看节点
kubectl get nodes -o wide

# 查看所有 Pod
kubectl get pods --all-namespaces

# 查看特定命名空间的资源
kubectl get all -n kube-system

# 查看集群事件
kubectl get events --all-namespaces --sort-by='.lastTimestamp'

# 查看资源使用情况（需要安装 metrics-server）
kubectl top nodes
kubectl top pods --all-namespaces
```

---

**文档生成**: Claude Code (Anthropic)
**最后更新**: 2026-01-10
