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
  - [x] 开发CRD定义
    - [x] Mission CRD
    - [x] MissionStage CRD
    - [x] FlightTask CRD
    - [x] Weapon CRD
  - [x] 开发核心Controller
    - [x] Mission Controller
    - [x] MissionStage Controller
    - [x] FlightTask Controller
    - [x] Weapon Controller（最小闭环）
  - [x] FlightTask 基础调度约束（NodeSelector/NodeAffinity）
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

**最后更新**: 2026-01-13

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

#### Weapon Controller 最小闭环（E）
- ✅ Weapon Controller：初始化 `Weapon.status.phase=Available`（若为空）
- ✅ Weapon Controller：初始化 `Weapon.status.usage`（若为空）
- ✅ Weapon Controller：根据 `spec.compatibility.aircraftTypes` 回写 `status.compatibilityChecks`（best-effort，可观测）

---

### 2026-01-13

#### FlightTask 调度约束与可观测性（F）
- ✅ FlightTask Controller：根据 `spec.aircraftRequirement` 自动补齐 Pod 调度约束
  - `aircraft.mil/type`、`aircraft.mil/status=ready`
  - `aircraft.mil/fuel.level`、`aircraft.mil/hardpoint.available`、`aircraft.mil/capability.*` 的 NodeAffinity（required）
  - `aircraft.mil/location.zone` 的 NodeAffinity（preferred）
- ✅ FlightTask 状态推进语义修正：Pod 创建后保持 `Scheduled`，由 Pod 实际状态推进到 `Running/Succeeded/Failed`
- ✅ MissionStage.status.flightTasksStatus：补齐 `podName` 与 `aircraftNode` 字段（基于 FlightTask.status）

#### Pod/FlightTask 状态对齐增强（G）
- ✅ FlightTask Controller：观测 `PodScheduled` 条件并回写到 `FlightTask.status.conditions`（便于直接看到 Unschedulable 原因）
- ✅ FlightTask Controller：基于 `FailedScheduling` 事件累计 `status.schedulingInfo.schedulingAttempts`（并在 Pod Pending 且未分配 Node 时周期性 Requeue 以刷新）
- ✅ FlightTask Controller：在 Pod 真正被分配 Node 时，使用 Pod 的 `PodScheduled.lastTransitionTime`（或 `pod.status.startTime`）回写 `status.schedulingInfo.assignedTime`
- ✅ FailedScheduling 计数增强：同时按 `involvedObject.uid` 与 `involvedObject.name` 聚合（避免漏计），并合并到同一次 status patch（避免条件滞后）
- ✅ 兼容两套 Event API：根据podUID同时统计 `core/v1 Event` 与 `events.k8s.io/v1 Event` 的 FailedScheduling（避免不同组件写入不同资源导致漏计）
- ✅ 状态一致性：PodScheduled condition 每次都会同步（不管内容是否变化），FailedScheduling 不再按时间窗口过滤，避免漏计

### 2026-01-15

#### FlightTask 状态对齐与镜像拉取失败处理
- ✅ FlightTask Controller：为已有关联 Pod 且未终态的任务增加周期性 Requeue（默认 5s），避免漏掉事件导致 FlightTask 停在 Pending/Scheduled
- ✅ 创建 Pod 后的短暂 NotFound 改为重排队重查，提升容错
- ✅ 新增 `ImagePullFailed` 条件：检测 ErrImagePull/ImagePullBackOff 等 Waiting 原因，记录 reason/message，但不再将 FlightTask 置为 Failed（保持 Scheduled/Pending 便于人工修复）
- ✅ 示例 `example/flighttask-with-weapon.yaml` 增加 `podTemplate`，使用不存在的镜像触发 ImagePullBackOff，可用于验证上述逻辑

#### CRD 易用性
- ✅ 为 MissionStage/FlightTask 增加短名：`ms`、`ft`（kubebuilder 注解与 CRD bases 均已更新）

### 2026-01-16

#### 示例验证清单
- ✅ 更新示例任务：`example/mission.yaml` 扩展为 3 个阶段、每阶段 3 个 FlightTask

### 2026-01-19

#### FlightTask 失败兜底与状态展示
- ✅ FlightTask：Pod 创建失败（Invalid）时，写入 `PodCreated=False` 条件并将任务置为 Failed，避免卡在 Scheduled
- ✅ CRD：为 Mission/MissionStage/FlightTask/Weapon 增加 `Phase` 列（`kubectl get` 可直接看到 status.phase）
- ✅ Mission failurePolicy：新增 `stageFailureAction=continue` 分支，阶段失败后允许依赖阶段继续推进且任务不中断
#### 验证基本功能
- ✅ 执行 `kubectl get pod -w` 观察：stage1 全部完成后 stage2 串行推进，stage2 完成后 stage3 并行启动
- ✅ 所有阶段 Pod 均创建并完成（Completed），链路推进正常
- ✅ 阶段类型调整为并行 → 串行 → 并行，并配置依赖关系（stage2 依赖 stage1，stage3 依赖 stage2）
- ✅ 每个任务补齐 `podTemplate`，使用短时 `sleep 5` 便于快速验证链路推进

### 2026-01-23

#### 前端初始搭建（主节点展示）
- ✅ 创建前端目录：`/project/frontend_master`
- ✅ 初始化 Vite + Vue3 项目
- ✅ 完成基础视觉系统与页面骨架（Header/Sidebar/Main/Detail）
- ✅ 建立 Dashboard 初版页面与静态占位数据
- ✅ 引入 Vue Router + Pinia，拆分布局组件
  - TopBar / SideBar / DetailPanel 组件化
  - 路由结构与页面占位（Missions/Stages/FlightTasks/Weapons/Cluster/Settings）

#### 前端交互与状态联动
- ✅ 右侧详情面板改为随页面选中项联动，不再固定显示
- ✅ 新增 Pinia 的 focus 状态管理，Missions/Stages/FlightTasks/Weapons/Cluster 页面点击条目会更新详情
- ✅ Cluster 节点卡片支持选中态样式

#### 前端数据层与 K8s 直连调试
- ✅ 新增 data store + seed + normalize，支持从 K8s 原生列表 items 映射到页面字段
- ✅ API 请求支持 K8s 模式路径（`/apis` 与 `/api/v1`），新增 API_MODE/namespace 配置
- ✅ 增加鉴权配置（AUTH_HEADER/AUTH_SCHEME/AUTH_TOKEN），请求与 healthz 检测可携带 header
- ✅ 使用 Vite 代理解决 CORS，API_BASE 为空走同源 `/apis`/`/api` 请求
- ✅ Sync 按钮支持立即 API 检测 + 数据刷新 + UTC 时间同步
- ✅ Sidebar System Health 改为动态统计 nodes/events
- ✅ 更换浏览器标签页图标与左上角品牌图标为 `空战图标.png`
- ✅ MissionStage CRD 增加 dependsOn 字段，controller 同步写入并在前端展示依赖信息

### 2026-01-25

#### 前端 metrics server 对接
- ✅ 接入 `metrics.k8s.io/v1beta1` 节点指标，计算 CPU/内存使用百分比
- ✅ Data store 合并 Node 列表与 metrics，Cluster 页面实时展示 CPU/Memory
- ✅ 统计 Running Pods 数量并写回 Node `pods` 展示

#### 前端刷新体验优化
- ✅ 轮询刷新时保留旧数据，仅在首次加载显示 Loading，减少列表闪烁
- ✅ API 状态轮询改为静默更新，避免 Sidebar 在检查时闪烁
