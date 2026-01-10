# CRD资源定义详解

## 一、资源层次结构

```
Mission (任务)
├── MissionStage (任务阶段)
│   └── FlightTask (飞行任务)
│       └── Pod (任务执行实例)
│
AircraftNode (飞机节点) - 对应K8s Node
Weapon (武器装备)
WeaponLoadout (武器挂载配置)
```

## 二、核心CRD定义

### 2.1 Mission（任务）

**用途**：描述完整的作战任务，包含多个阶段及其依赖关系

```yaml
apiVersion: airforce.mil/v1alpha1
kind: Mission
metadata:
  name: far-sea-strike-mission-001
  namespace: operations
spec:
  # 任务基本信息
  missionName: "远海方向重要目标综合侦察-压制-打击示范行动"
  missionType: strike  # 任务类型：isr(侦察), strike(打击), patrol(巡逻), escort(护航)
  priority: high       # 优先级：low, medium, high, critical

  # 任务目标
  objective:
    targetArea: "东海某海域"
    targetCoordinates:
      latitude: 28.5
      longitude: 122.3
    targetDescription: "海上高价值机动平台"

  # 任务阶段定义（按顺序执行）
  stages:
  - name: stage1-isr
    displayName: "侦察与态势建立"
    type: sequential    # 阶段内任务执行方式：sequential(串行), parallel(并行), mixed(混合)
    dependsOn: []       # 依赖的前置阶段
    timeout: 30m        # 阶段超时时间

  - name: stage2-suppression
    displayName: "制空掩护与电子压制"
    type: parallel
    dependsOn: ["stage1-isr"]  # 依赖阶段1完成
    timeout: 45m

  - name: stage3-strike
    displayName: "远程打击与评估"
    type: mixed
    dependsOn: ["stage2-suppression"]
    timeout: 60m

  # 任务配置
  config:
    # 失败策略
    failurePolicy:
      maxRetries: 3                    # 最大重试次数
      retryStrategy: exponential       # 重试策略：immediate, exponential, custom
      stageFailureAction: abort        # 阶段失败处理：abort(终止), continue(继续), retry(重试)

    # 取消策略
    cancellationPolicy:
      gracePeriod: 300s                # 取消任务的宽限期
      cleanup: true                    # 是否清理资源

    # 协同配置
    coordination:
      dataLinkProtocol: "Link-16"      # 数据链协议
      commandFrequency: "UHF-243MHz"   # 指挥频率
      emergencyFrequency: "121.5MHz"   # 应急频率

status:
  # 任务状态
  phase: Running  # Pending, Running, Succeeded, Failed, Cancelled

  # 阶段状态摘要
  stagesSummary:
  - name: stage1-isr
    phase: Succeeded
    startTime: "2024-03-15T10:00:00Z"
    completionTime: "2024-03-15T10:28:00Z"
  - name: stage2-suppression
    phase: Running
    startTime: "2024-03-15T10:28:00Z"
  - name: stage3-strike
    phase: Pending

  # 统计信息
  statistics:
    totalFlightTasks: 12
    succeededTasks: 5
    failedTasks: 0
    runningTasks: 4
    pendingTasks: 3

  # 当前状态
  message: "阶段2正在执行，4架飞机正在执行任务"
  startTime: "2024-03-15T10:00:00Z"
  lastUpdateTime: "2024-03-15T10:30:00Z"
```

### 2.2 MissionStage（任务阶段）

**用途**：任务的执行阶段，包含多个协同的飞行任务

```yaml
apiVersion: airforce.mil/v1alpha1
kind: MissionStage
metadata:
  name: stage1-isr
  namespace: operations
  labels:
    mission: far-sea-strike-mission-001
    stage-index: "1"
  ownerReferences:
  - apiVersion: airforce.mil/v1alpha1
    kind: Mission
    name: far-sea-strike-mission-001
    uid: xxx
    controller: true
spec:
  # 关联的任务
  missionRef:
    name: far-sea-strike-mission-001

  # 阶段信息
  stageName: "侦察与态势建立"
  stageIndex: 1
  stageType: sequential

  # 飞行任务列表
  flightTasks:
  - name: ft-kj500-001
    aircraft: kj500        # 飞机型号
    role: awacs-control    # 任务角色
    priority: critical     # 该任务的优先级

    # 所需武器/装备
    weaponLoadout:
    - weapon: aesa-radar
      quantity: 1
    - weapon: datalink-pod
      quantity: 1

    # 任务参数
    taskParams:
      altitude: 8000m
      speed: 450km/h
      loiterTime: 30m
      controlRange: 400km

  - name: ft-wz10-001
    aircraft: wz10
    role: reconnaissance
    priority: high

    weaponLoadout:
    - weapon: sar-radar-pod
      quantity: 1
    - weapon: eo-ir-sensor
      quantity: 1

    taskParams:
      altitude: 12000m
      speed: 600km/h
      scanArea: "target-area-alpha"

  - name: ft-j20-001
    aircraft: j20
    role: escort
    priority: high

    weaponLoadout:
    - weapon: pl10-missile
      quantity: 2

    taskParams:
      altitude: 10000m
      speed: 800km/h
      patrolRoute: "escort-route-1"

  # 阶段配置
  config:
    # 同步策略（任务间如何协同）
    synchronization:
      waitForAll: false      # 是否等待所有任务就绪后再开始
      checkpoint: "target-in-sight"  # 同步检查点

    # 超时配置
    timeout: 30m

    # 依赖条件
    dependencies:
      conditions:
      - type: WeatherClear    # 天气晴朗
      - type: DataLinkReady   # 数据链就绪

status:
  phase: Running  # Pending, Running, Succeeded, Failed

  # 飞行任务状态
  flightTasksStatus:
  - name: ft-kj500-001
    phase: Running
    aircraftNode: kj500-01
    podName: ft-kj500-001-pod-abc123
  - name: ft-wz10-001
    phase: Running
    aircraftNode: wz10-03
    podName: ft-wz10-001-pod-def456
  - name: ft-j20-001
    phase: Pending
    message: "等待合适的飞机节点"

  startTime: "2024-03-15T10:00:00Z"
  completionTime: null
  message: "2/3任务正在执行"
```

### 2.3 FlightTask（飞行任务）

**用途**：单架飞机的具体任务，最终会创建Pod执行

```yaml
apiVersion: airforce.mil/v1alpha1
kind: FlightTask
metadata:
  name: ft-j20-001
  namespace: operations
  labels:
    mission: far-sea-strike-mission-001
    stage: stage2-suppression
    aircraft-type: j20
    role: air-superiority
spec:
  # 关联的任务阶段
  stageRef:
    name: stage2-suppression

  # 飞机要求
  aircraftRequirement:
    type: j20                    # 飞机型号
    minFuelLevel: 70             # 最低油量百分比
    capabilities:                # 必需的能力
    - stealth
    - bvr-combat
    requiredHardpoints: 4        # 需要的挂载点数量
    preferredLocation: "east-sea" # 优选位置

  # 任务角色和参数
  role: air-superiority
  taskParams:
    altitude: 11000m
    speed: 900km/h
    missionDuration: 45m
    operationArea:
      center: {lat: 28.5, lon: 122.3}
      radius: 100km

    # 任务阶段
    phases:
    - name: ingress
      duration: 15m
      waypoints: ["wp1", "wp2", "wp3"]
    - name: combat
      duration: 20m
      tactics: "sweep-and-clear"
    - name: egress
      duration: 10m
      waypoints: ["wp4", "wp5"]

  # 武器挂载配置
  weaponLoadout:
  - weaponRef:
      name: pl15-missile
    quantity: 4
    mountPoints: ["hp1", "hp2", "hp3", "hp4"]

  - weaponRef:
      name: pl10-missile
    quantity: 2
    mountPoints: ["hp5", "hp6"]

  # Pod模板（用于创建实际执行的Pod）
  podTemplate:
    metadata:
      labels:
        app: flight-task
        task: ft-j20-001

    spec:
      # 节点选择器
      nodeSelector:
        aircraft.type: j20
        aircraft.status: ready

      # 亲和性配置
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: aircraft.fuel
                operator: Gt
                values: ["70"]
              - key: aircraft.hardpoint.available
                operator: Gt
                values: ["4"]

          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            preference:
              matchExpressions:
              - key: aircraft.fuel
                operator: Gt
                values: ["90"]

      # 容器定义
      containers:
      # 主容器：飞机控制系统
      - name: aircraft-control
        image: registry.mil/aircraft/j20-control:v3.2.1
        env:
        - name: MISSION_ID
          value: "far-sea-strike-mission-001"
        - name: TASK_ROLE
          value: "air-superiority"
        - name: ALTITUDE
          value: "11000"
        resources:
          requests:
            cpu: "2"
            memory: "4Gi"
            aircraft.mil/fuel: "70"  # 自定义资源：燃料
          limits:
            cpu: "4"
            memory: "8Gi"
        volumeMounts:
        - name: mission-data
          mountPath: /mission
        - name: weapon-interface
          mountPath: /weapons

      # Sidecar容器：武器系统 PL-15
      - name: weapon-pl15
        image: registry.mil/weapons/pl15-missile:v2.3.0
        env:
        - name: QUANTITY
          value: "4"
        - name: MOUNT_POINTS
          value: "hp1,hp2,hp3,hp4"
        volumeMounts:
        - name: weapon-interface
          mountPath: /interface

      # Sidecar容器：武器系统 PL-10
      - name: weapon-pl10
        image: registry.mil/weapons/pl10-missile:v1.8.0
        env:
        - name: QUANTITY
          value: "2"
        - name: MOUNT_POINTS
          value: "hp5,hp6"
        volumeMounts:
        - name: weapon-interface
          mountPath: /interface

      # Init容器：任务预检
      initContainers:
      - name: preflight-check
        image: registry.mil/utils/preflight-checker:v1.0
        command: ["sh", "-c"]
        args:
        - |
          echo "执行飞行前检查..."
          # 检查武器系统就绪
          # 检查燃料充足
          # 检查通信链路
          echo "检查完成"

      volumes:
      - name: mission-data
        configMap:
          name: mission-data-config
      - name: weapon-interface
        emptyDir: {}

      restartPolicy: OnFailure

status:
  phase: Running  # Pending, Scheduled, Running, Succeeded, Failed

  # 调度信息
  schedulingInfo:
    assignedNode: "j20-02"
    assignedTime: "2024-03-15T10:28:30Z"
    schedulingAttempts: 1

  # Pod信息
  podRef:
    name: ft-j20-001-pod-xyz789
    namespace: operations

  # 执行状态
  executionStatus:
    currentPhase: "combat"
    location: {lat: 28.3, lon: 122.1}
    altitude: 11000m
    speed: 900km/h
    fuelRemaining: 65
    weaponsRemaining:
      pl15: 4
      pl10: 2

  # 重试信息
  retryCount: 0
  lastRetryTime: null

  startTime: "2024-03-15T10:28:35Z"
  completionTime: null
  message: "任务进行中，已完成ingress阶段"
```

### 2.4 AircraftNode（飞机节点）

**用途**：描述飞机的详细信息和能力，作为K8s Node的扩展

**注意**：这不是完全的CRD，而是在Node基础上添加的标签和注解

```yaml
apiVersion: v1
kind: Node
metadata:
  name: j20-02
  labels:
    # 基本信息
    aircraft.mil/type: j20
    aircraft.mil/model: j20a
    aircraft.mil/serial: "2024-J20-002"

    # 状态标签
    aircraft.mil/status: ready       # ready, busy, maintenance, offline
    aircraft.mil/health: healthy     # healthy, degraded, critical

    # 位置信息
    aircraft.mil/location.zone: east-sea
    aircraft.mil/location.base: "东海基地"

    # 能力标签
    aircraft.mil/capability.stealth: "true"
    aircraft.mil/capability.supercruise: "true"
    aircraft.mil/capability.bvr: "true"
    aircraft.mil/capability.air-refuel: "true"

    # 资源状态
    aircraft.mil/fuel.level: "85"          # 当前油量百分比
    aircraft.mil/fuel.status: high         # low(<30), medium(30-70), high(>70)
    aircraft.mil/hardpoint.total: "10"     # 总挂载点数
    aircraft.mil/hardpoint.available: "10" # 可用挂载点数

    # 任务信息
    aircraft.mil/mission.current: ""       # 当前执行的任务（空表示无任务）
    aircraft.mil/mission.count: "127"      # 累计执行任务数
    aircraft.mil/flight.hours: "450.5"     # 飞行小时数

  annotations:
    # 详细配置（JSON格式）
    aircraft.mil/spec: |
      {
        "manufacturer": "成都飞机工业集团",
        "firstFlight": "2020-03-15",
        "maxSpeed": 2100,
        "combatRadius": 2000,
        "serviceceiling": 20000,
        "maxTakeoffWeight": 37000,
        "engine": "涡扇-15",
        "avionics": "第四代综合航电系统",
        "radar": "有源相控阵雷达"
      }

    # 武器兼容性（JSON数组）
    aircraft.mil/compatible-weapons: |
      ["pl10", "pl12", "pl15", "pl21", "cm400akg", "ls6"]

    # 维护信息
    aircraft.mil/maintenance: |
      {
        "lastMaintenance": "2024-03-10T08:00:00Z",
        "nextMaintenance": "2024-04-10T08:00:00Z",
        "maintenanceLevel": "一级维护",
        "technician": "张工程师"
      }

    # 通信配置
    aircraft.mil/communication: |
      {
        "datalink": "Link-16",
        "radio": ["UHF", "VHF", "HF"],
        "satcom": true,
        "encryption": "军用加密"
      }

spec:
  # K8s Node原生字段
  podCIDR: 10.244.2.0/24

  # 自定义资源容量
  capacity:
    cpu: "8"
    memory: "32Gi"
    aircraft.mil/fuel: "100"          # 燃料资源（百分比）
    aircraft.mil/hardpoints: "10"     # 挂载点资源
    aircraft.mil/weapon-weight: "8000" # 最大武器载重（kg）

  allocatable:
    cpu: "8"
    memory: "32Gi"
    aircraft.mil/fuel: "85"
    aircraft.mil/hardpoints: "10"
    aircraft.mil/weapon-weight: "8000"

status:
  conditions:
  - type: Ready
    status: "True"
    lastHeartbeatTime: "2024-03-15T10:30:00Z"
    lastTransitionTime: "2024-03-15T08:00:00Z"
    reason: AircraftReady
    message: "飞机状态正常，系统就绪"

  - type: FuelSufficient
    status: "True"
    lastHeartbeatTime: "2024-03-15T10:30:00Z"
    lastTransitionTime: "2024-03-15T10:25:00Z"
    reason: FuelLevelHigh
    message: "燃料充足：85%"

  - type: WeaponSystemReady
    status: "True"
    lastHeartbeatTime: "2024-03-15T10:30:00Z"
    lastTransitionTime: "2024-03-15T08:00:00Z"
    reason: AllSystemsOperational
    message: "武器系统就绪"

  addresses:
  - type: InternalIP
    address: "10.100.2.5"
  - type: Hostname
    address: "j20-02"

  nodeInfo:
    operatingSystem: "AircraftOS"
    architecture: "arm64"
    kernelVersion: "5.15.0-aircraft"
```

### 2.5 Weapon（武器装备）

**用途**：定义武器的规格、镜像和兼容性

```yaml
apiVersion: airforce.mil/v1alpha1
kind: Weapon
metadata:
  name: pl15-missile
  namespace: weapons
spec:
  # 武器基本信息
  weaponName: "霹雳-15中远程空空导弹"
  weaponType: air-to-air-missile
  category: beyond-visual-range  # 类别：bvr, wvr, air-to-ground, etc.

  # 武器规格
  specifications:
    manufacturer: "中国航天科工集团"
    weight: 200kg
    length: 4.2m
    diameter: 0.2m
    range: 200km
    speed: 4.0  # 马赫数
    guidance: "主动雷达制导"
    warhead: "破片杀伤"

  # 容器镜像
  image:
    repository: registry.mil/weapons/pl15-missile
    tag: v2.3.0
    pullPolicy: IfNotPresent

  # 资源需求
  resources:
    hardpoints: 1           # 占用挂载点数
    weight: 200             # 重量（kg）
    power: 0                # 功耗（watts）
    cooling: low            # 散热需求：low, medium, high

  # 兼容性
  compatibility:
    # 兼容的飞机型号
    aircraftTypes:
    - j20
    - j16
    - j10c

    # 兼容的挂载点类型
    hardpointTypes:
    - internal-bay          # 内置弹舱
    - under-wing           # 翼下挂架

  # 容器配置
  container:
    # 环境变量
    env:
    - name: WEAPON_TYPE
      value: "pl15"
    - name: GUIDANCE_MODE
      value: "active-radar"

    # 卷挂载
    volumeMounts:
    - name: weapon-interface
      mountPath: /interface
    - name: weapon-config
      mountPath: /config

    # 端口（用于与主容器通信）
    ports:
    - name: control
      containerPort: 9001
      protocol: TCP

    # 健康检查
    livenessProbe:
      httpGet:
        path: /health
        port: 9001
      initialDelaySeconds: 5
      periodSeconds: 10

  # 版本管理
  version:
    current: v2.3.0
    changelog: |
      v2.3.0:
      - 改进目标识别算法
      - 增强抗干扰能力
      - 修复已知bug
    releaseDate: "2024-01-15"

status:
  # 部署状态
  phase: Available  # Available, Updating, Deprecated

  # 使用统计
  usage:
    totalDeployed: 48       # 当前部署数量
    totalFired: 127         # 累计发射次数
    successRate: 0.95       # 成功率

  # 兼容性检查结果
  compatibilityChecks:
  - aircraftType: j20
    compatible: true
    lastChecked: "2024-03-15T08:00:00Z"
  - aircraftType: j16
    compatible: true
    lastChecked: "2024-03-15T08:00:00Z"
```

### 2.6 WeaponLoadout（武器挂载配置）

**用途**：定义标准的武器挂载方案，可复用

```yaml
apiVersion: airforce.mil/v1alpha1
kind: WeaponLoadout
metadata:
  name: j20-air-superiority-loadout
  namespace: weapons
spec:
  # 适用的飞机型号
  aircraftType: j20

  # 挂载方案名称
  loadoutName: "制空作战标准挂载"
  loadoutType: air-superiority

  # 武器配置
  weapons:
  # 中远程空空导弹
  - weaponRef:
      name: pl15-missile
    quantity: 4
    mountPoints: ["internal-bay-1", "internal-bay-2", "internal-bay-3", "internal-bay-4"]
    priority: primary

  # 近距格斗导弹
  - weaponRef:
      name: pl10-missile
    quantity: 2
    mountPoints: ["internal-bay-5", "internal-bay-6"]
    priority: secondary

  # 总重量和资源统计
  totalWeight: 1000kg
  totalHardpoints: 6

  # 适用场景
  scenarios:
  - air-superiority
  - escort
  - patrol

  # 推荐配置
  recommended: true

  # 备注
  description: |
    标准制空作战挂载，适用于夺取制空权任务。
    配备4枚PL-15中远程导弹用于超视距作战，
    2枚PL-10格斗导弹用于近距离空战。
    保持隐身性能，所有武器内置。

status:
  # 使用次数
  usageCount: 45
  lastUsed: "2024-03-15T10:00:00Z"

  # 有效性
  validated: true
  validationTime: "2024-03-01T08:00:00Z"
```

## 三、CRD间的关系

### 3.1 所有权关系（OwnerReference）

```
Mission (owns)
  └── MissionStage (owns)
        └── FlightTask (owns)
              └── Pod
```

当Mission被删除时，自动级联删除所有MissionStage、FlightTask和Pod。

### 3.2 引用关系

```
FlightTask
  ├── references → MissionStage
  ├── references → Weapon (via weaponRef)
  ├── references → WeaponLoadout (optional)
  └── scheduled on → AircraftNode
```

### 3.3 标签关联

所有资源使用统一的标签体系进行关联：

```yaml
labels:
  mission: far-sea-strike-mission-001  # 任务ID
  stage: stage2-suppression            # 阶段ID
  aircraft-type: j20                   # 飞机型号
  role: air-superiority                # 任务角色
```

## 四、自定义资源（Custom Resources）

K8s允许定义自定义资源类型，我们定义以下资源：

### 4.1 aircraft.mil/fuel

燃料资源，表示飞机的燃油量（百分比）

```yaml
resources:
  requests:
    aircraft.mil/fuel: "70"  # 任务需要至少70%的燃料
  limits:
    aircraft.mil/fuel: "100"
```

### 4.2 aircraft.mil/hardpoints

挂载点资源，表示飞机的武器挂载点数量

```yaml
resources:
  requests:
    aircraft.mil/hardpoints: "6"  # 任务需要6个挂载点
```

### 4.3 aircraft.mil/weapon-weight

武器载重资源，表示飞机可承载的武器重量（kg）

```yaml
resources:
  requests:
    aircraft.mil/weapon-weight: "1000"  # 任务需要1000kg的武器载重
```

## 五、CRD安装

### 5.1 生成CRD定义文件

使用kubebuilder生成CRD定义：

```bash
# 生成所有CRD
kubebuilder create api --group airforce --version v1alpha1 --kind Mission
kubebuilder create api --group airforce --version v1alpha1 --kind MissionStage
kubebuilder create api --group airforce --version v1alpha1 --kind FlightTask
kubebuilder create api --group airforce --version v1alpha1 --kind Weapon
kubebuilder create api --group airforce --version v1alpha1 --kind WeaponLoadout
```

### 5.2 安装CRD到集群

```bash
# 应用所有CRD定义
kubectl apply -f config/crd/bases/

# 验证CRD安装
kubectl get crd | grep airforce.mil

# 输出示例：
# missions.airforce.mil
# missionstages.airforce.mil
# flighttasks.airforce.mil
# weapons.airforce.mil
# weaponloadouts.airforce.mil
```

## 六、使用示例

### 6.1 创建完整的任务

```bash
# 1. 创建武器资源
kubectl apply -f weapons/pl15-missile.yaml
kubectl apply -f weapons/pl10-missile.yaml

# 2. 创建任务
kubectl apply -f missions/far-sea-strike.yaml

# 3. Mission Controller自动创建MissionStage
# 4. Stage Controller自动创建FlightTask
# 5. FlightTask Controller自动创建Pod并调度到合适的飞机节点

# 6. 查看任务状态
kubectl get mission far-sea-strike-mission-001 -o yaml

# 7. 查看阶段状态
kubectl get missionstage -l mission=far-sea-strike-mission-001

# 8. 查看飞行任务状态
kubectl get flighttask -l mission=far-sea-strike-mission-001

# 9. 查看实际运行的Pod
kubectl get pods -l mission=far-sea-strike-mission-001
```

### 6.2 动态更新任务

```bash
# 修改Mission，添加新的阶段
kubectl edit mission far-sea-strike-mission-001

# Mission Controller会自动创建新的MissionStage
```

### 6.3 查看飞机节点状态

```bash
# 查看所有飞机节点
kubectl get nodes -l aircraft.mil/type

# 查看具体飞机的详细信息
kubectl describe node j20-02

# 查看飞机的资源分配情况
kubectl top node j20-02
```

## 七、下一步

下一章将详细讲解**Controller的实现逻辑**，包括：
- Mission Controller如何管理任务生命周期
- Stage Controller如何协调阶段执行
- FlightTask Controller如何调度和管理飞行任务
- 各Controller之间如何协作
