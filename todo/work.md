# 空军作战任务K8s平台 - 工作日志

> 开始时间: 2026-01-10
> 项目目标: 基于 Kubernetes 实现空军作战任务编排系统

---

## 工作记录

### 2026-01-10

#### 项目初始化
- ✅ 阅读项目规划文档（todo/plan 目录下的6个核心文档）
- ✅ 了解集群配置信息（k8s-cluster-info.md）
  - K8s v1.29.2，3节点集群（1 master + 2 worker）
  - 网络插件: Calico
  - 容器运行时: Docker 29.1.3
- ✅ 创建工作日志文件 work.md

#### Kubebuilder 代码推进
- ✅ 确认资源层级：Mission（第一层）→ MissionStage（第二层）→ FlightTask（第三层）→ Pod（执行层），并补充 Weapon CRD
- ✅ 生成并补全 CRD Go 类型（api/v1alpha1）
  - Mission：补充 missionName/missionType/priority/objective/stages/config 及 status phase/summary/statistics 等字段
  - MissionStage：补充 missionRef/stageName/stageIndex/stageType/flightTasks/config 及 status phase/flightTasksStatus 等字段
  - FlightTask：生成骨架并补充 stageRef/aircraftRequirement/role/taskParams/weaponLoadout/podTemplate 及 status phase/scheduling/podRef 等字段
  - Weapon：生成骨架并补充 weaponName/specifications/image/resources/compatibility/container/version 及 status phase/usage 等字段
- ✅ 更新示例清单（config/samples）为可参考的示例字段（Mission/MissionStage/FlightTask/Weapon）
- ✅ 修复 controller-gen 生成阻塞：将 CRD 中 float 字段改为 string（避免跨语言兼容问题），并成功生成 CRD bases 与 deepcopy
  - 生成输出：code/config/crd/bases/airforce.airforce.mil_{missions,missionstages,flighttasks,weapons}.yaml
  - 生成输出：code/api/v1alpha1/zz_generated.deepcopy.go

#### 测试与网络依赖
- ⚠️ `make test` 在下载 envtest 工具（setup-envtest）时访问 `proxy.golang.org` 失败（connection refused）
- ✅ 调整 Makefile：保持 `ENVTEST_VERSION=latest`（setup-envtest 是独立 module，不能用 controller-runtime 的 v0.17.0 作为版本号）
- 运行建议（需要能访问 Go module 代理）：
  - `GOPROXY=https://goproxy.cn,direct GOSUMDB=off make test`
  - 若仅做编译自检：`make build`

#### 集群安装问题修复
- ✅ 修复 `make install` 失败：FlightTask CRD 过大导致 `metadata.annotations` 超过 256Ki（kubectl client-side apply 的 last-applied 注解会膨胀）
  - 将 `FlightTask.spec.podTemplate` 从 `corev1.PodTemplateSpec` 改为 `runtime.RawExtension`（schemaless，保留未知字段），显著缩小 CRD schema
  - `make install` 改为 `kubectl apply --server-side`，避免写入 last-applied 注解
- ✅ 已成功在集群安装 4 个 CRD：missions/missionstages/flighttasks/weapons

#### 实施阶段
当前阶段: **阶段一 - 基础平台搭建（1-2个月）**

根据实施路线图 (todo/plan/06-系统总结与实施建议.md):
- 阶段一任务清单:
  - [x] 搭建K8s集群（集群已经搭建完成）
  - [ ] 开发CRD定义
    - [ ] Mission CRD
    - [ ] MissionStage CRD
    - [ ] FlightTask CRD
    - [ ] Weapon CRD
  - [ ] 开发核心Controller
    - [ ] Mission Controller
    - [ ] MissionStage Controller
    - [ ] FlightTask Controller
  - [ ] 验证基本功能

---

## 下一步计划

1. 使用 kubebuilder 初始化项目
2. 创建 CRD 定义
3. 实现基础 Controller 逻辑
4. 编写单元测试

---

## 技术笔记

### 核心架构设计
- 三层任务抽象: Mission → MissionStage → FlightTask → Pod
- 武器采用 Sidecar 模式而非独立 Deployment
- 使用自定义调度器实现智能任务分配
- 多层次容错机制

### 关键技术栈
- Kubernetes: v1.29.2
- Kubebuilder: v3.x (待安装)
- 编程语言: Go 1.21+
- 容器运行时: Containerd (计划从Docker迁移)

---

## 问题与解决

（暂无）

---

**最后更新**: 2026-01-10

---

### 2026-01-11

#### Controller 开发推进
- ✅ Mission Controller：补齐阶段推进逻辑
  - 基于 `Mission.spec.stages[].dependsOn` 推进 `MissionStage.status.phase`：依赖满足则从 Pending → Running，并写入 `startTime`
  - 汇总 `MissionStage.status.phase` 回写 `Mission.status.stagesSummary` 与 `Mission.status.phase`
  - 统计 `FlightTask` 数量与各 phase，回写 `Mission.status.statistics`（按 label `mission=<mission.name>` 聚合）
- ✅ MissionStage Controller：实现基础编排
  - 根据 `MissionStage.spec.flightTasks` 创建/更新/清理 `FlightTask`（OwnerReference 指向 MissionStage）
  - 任务推进：Stage Running 时将任务置为 `Scheduled`（sequential 只推进一个；parallel/mixed 推进全部）
  - 汇总 `FlightTask.status.phase` 回写 `MissionStage.status.flightTasksStatus`、`message`，并在全部成功/有失败时置 `MissionStage` 为 Succeeded/Failed
  - 支持 `MissionStage.spec.config.timeout` 超时失败

#### 测试与环境说明
- ⚠️ 本地 `go test` 需要网络/代理拉取 Go modules；另外默认 `GOCACHE=/root/.cache/go-build` 在当前环境会遇到权限问题
- 建议你本地运行时设置：
  - `GOCACHE=$PWD/.gocache`（避免写 /root/.cache）
  - `GOPROXY` 指向可访问的代理（例如 `https://goproxy.cn,direct`）

#### FlightTask 最小闭环（A）
- ✅ FlightTask Controller：当 `FlightTask.status.phase=Scheduled` 时创建执行 Pod（默认 busybox，或使用 `spec.podTemplate`）
- ✅ 监听 Pod 状态并回写 `FlightTask.status.phase`：Running/Succeeded/Failed，同时写入 `status.podRef`
- ✅ envtest：补充 `flighttask_controller_test.go` 断言（Scheduled→创建 Pod→FlightTask 进入 Running）

#### Mission → MissionStage → FlightTask → Pod 串联
- ✅ 扩展 `Mission.spec.stages[].flightTasks`（对齐 `todo/plan/05-远海打击任务完整示例.md` 的完整任务定义）
- ✅ Mission Controller 将 `stages[].flightTasks` 下发到 `MissionStage.spec.flightTasks`（自动补齐缺失的 task name）
- ✅ 更新 `Mission` CRD schema：支持 `spec.stages[].flightTasks` 字段
- ✅ 更新示例：`code/config/samples/airforce_v1alpha1_mission.yaml` 增加 `flightTasks`，用于一键触发整条链路

---

### 2026-01-12

#### 开发环境与生成器修复
- ✅ 安装 envtest 工具：`make envtest`（自动下载 Go toolchain 以满足 setup-envtest 的 Go 版本要求）
- ✅ 下载 envtest k8s 资源：`./bin/setup-envtest-latest use 1.29.0 --bin-dir ./bin -p path`
- ✅ 修复 `make manifests` 失败：
  - 原因：`code/.gomodcache/`（以及 `code/.gocache/`）位于 Go module 根目录下，`controller-gen paths=./...` 会扫描到缓存里的依赖源码，导致报 `missing go.sum entry`
  - 处理：删除并从 git 里移除缓存目录，避免误提交；同时在 `.gitignore` 中忽略 `.gomodcache/` 与 `.gocache/`
  - 处理：设置 `GOCACHE=$PWD/.gocache` 避免写入 `/root/.cache/go-build` 引起 permission denied
- ✅ RBAC：确保 `manager-role` 包含 `core/v1 pods` 权限（用于 FlightTask Controller 在真实集群创建 Pod）

#### Weapon Sidecar 注入（B）
- ✅ FlightTask Controller：基于 `spec.weaponLoadout` 拉取 `Weapon` CR 并向执行 Pod 注入武器 sidecar（共享 `EmptyDir` 卷 `weapon-interface`）
- ✅ 兼容性检查：按 `Weapon.spec.compatibility.aircraftTypes/hardpointTypes` 做最小校验，不满足则将 FlightTask 标记为 Failed 并写入 condition
- ✅ 单测更新：`flighttask_controller_test.go` 增加武器注入断言（创建 Weapon + FlightTask，期望 Pod 存在 `weapon-<name>` 容器）

#### 集群验证（C）
- ✅ 部署 controller 镜像：`docker.io/woshiasher/controllers:v0.1.0`，Pod 已正常 Ready
- ✅ 验证 Weapon sidecar 注入：
  - `kubectl apply -f example/weapon.yaml` + `kubectl apply -f example/flighttask-with-weapon.yaml`
  - 将 `FlightTask` status 手动 patch 为 `Scheduled` 后成功创建 Pod：`demo-flighttask-with-weapon-pod`
  - Pod 容器包含：`task` 与 `weapon-pl-15`，并挂载共享卷：`weapon-interface`
  - 备注：示例中 `weapon-pl-15` 使用 `busybox` 会立刻退出（Completed），因此 Pod `Ready` 可能为 False，但注入本身已验证成功

#### FlightTask 自动推进（D）
- ✅ 对“独立 FlightTask”（无 `spec.stageRef.name` 且不归属 `MissionStage`）自动从 `Pending -> Scheduled`，避免需要手动 patch status
- ✅ 保持与 MissionStage 编排不冲突：只对独立任务生效，由 MissionStage 创建/管理的 FlightTask 仍由 MissionStage Controller 推进
