# Controller实现逻辑

## 一、Controller架构概览

### 1.1 Controller职责划分

```
┌────────────────────────────────────────────────────────────┐
│                  Mission Controller                        │
│  职责：管理任务生命周期，创建和协调MissionStage             │
└────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────┐
│                MissionStage Controller                     │
│  职责：管理阶段执行，创建FlightTask，处理阶段依赖          │
└────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────┐
│                 FlightTask Controller                      │
│  职责：管理飞行任务，创建Pod，处理任务调度和重试           │
└────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────┐
│                 Aircraft Controller                        │
│  职责：管理飞机节点状态，更新Node标签和资源                │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────┐
│                  Weapon Controller                         │
│  职责：管理武器资源，处理武器更新和兼容性检查              │
└────────────────────────────────────────────────────────────┘
```

### 1.2 Controller通信方式

- **通过K8s API Server**：所有Controller通过Watch机制监听资源变化
- **通过资源状态**：上层Controller通过更新资源Status通知下层Controller
- **通过OwnerReference**：建立资源间的父子关系，实现级联删除

## 二、Mission Controller

### 2.1 核心职责

1. 监听Mission CRD的创建、更新、删除
2. 根据Mission.Spec.Stages创建MissionStage资源
3. 管理阶段间的依赖关系和执行顺序
4. 汇总各阶段状态，更新Mission整体状态
5. 处理任务的取消和清理

### 2.2 Reconcile逻辑

```go
package controllers

import (
    "context"
    "time"

    airforcev1alpha1 "airforce.mil/api/v1alpha1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"
)

type MissionReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func (r *MissionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    logger.Info("开始调和Mission", "name", req.Name)

    // 1. 获取Mission对象
    mission := &airforcev1alpha1.Mission{}
    if err := r.Get(ctx, req.NamespacedName, mission); err != nil {
        if errors.IsNotFound(err) {
            // Mission已被删除，无需处理（级联删除会自动清理子资源）
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    // 2. 检查Mission是否被标记为删除
    if !mission.ObjectMeta.DeletionTimestamp.IsZero() {
        return r.handleMissionDeletion(ctx, mission)
    }

    // 3. 初始化Mission状态
    if mission.Status.Phase == "" {
        mission.Status.Phase = "Pending"
        mission.Status.StartTime = metav1.Now()
        if err := r.Status().Update(ctx, mission); err != nil {
            return ctrl.Result{}, err
        }
    }

    // 4. 创建或更新MissionStage资源
    if err := r.reconcileStages(ctx, mission); err != nil {
        return ctrl.Result{}, err
    }

    // 5. 检查所有Stage的状态
    allStages, err := r.listMissionStages(ctx, mission)
    if err != nil {
        return ctrl.Result{}, err
    }

    // 6. 根据Stage状态判断是否可以启动下一个Stage
    if err := r.progressMissionStages(ctx, mission, allStages); err != nil {
        return ctrl.Result{}, err
    }

    // 7. 汇总Stage状态，更新Mission状态
    if err := r.updateMissionStatus(ctx, mission, allStages); err != nil {
        return ctrl.Result{}, err
    }

    // 8. 检查任务是否完成
    if r.isMissionComplete(mission, allStages) {
        mission.Status.Phase = "Succeeded"
        mission.Status.CompletionTime = &metav1.Time{Time: time.Now()}
        if err := r.Status().Update(ctx, mission); err != nil {
            return ctrl.Result{}, err
        }
        logger.Info("任务完成", "mission", mission.Name)
        return ctrl.Result{}, nil
    }

    // 9. 检查任务是否失败
    if r.isMissionFailed(mission, allStages) {
        if mission.Status.RetryCount < mission.Spec.Config.FailurePolicy.MaxRetries {
            // 重试任务
            logger.Info("任务失败，准备重试", "mission", mission.Name, "retryCount", mission.Status.RetryCount)
            return r.retryMission(ctx, mission)
        } else {
            // 任务彻底失败
            mission.Status.Phase = "Failed"
            mission.Status.Message = "任务失败，已达到最大重试次数"
            if err := r.Status().Update(ctx, mission); err != nil {
                return ctrl.Result{}, err
            }
            logger.Error(nil, "任务失败", "mission", mission.Name)
            return ctrl.Result{}, nil
        }
    }

    // 10. 定期重新调和（检查状态变化）
    return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

// reconcileStages 创建或更新MissionStage资源
func (r *MissionReconciler) reconcileStages(ctx context.Context, mission *airforcev1alpha1.Mission) error {
    logger := log.FromContext(ctx)

    for i, stageSpec := range mission.Spec.Stages {
        // 检查Stage是否已存在
        stage := &airforcev1alpha1.MissionStage{}
        stageName := mission.Name + "-" + stageSpec.Name
        err := r.Get(ctx, client.ObjectKey{Name: stageName, Namespace: mission.Namespace}, stage)

        if err != nil && errors.IsNotFound(err) {
            // Stage不存在，创建新的Stage
            newStage := &airforcev1alpha1.MissionStage{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      stageName,
                    Namespace: mission.Namespace,
                    Labels: map[string]string{
                        "mission":     mission.Name,
                        "stage-index": fmt.Sprintf("%d", i),
                    },
                    // 设置OwnerReference，实现级联删除
                    OwnerReferences: []metav1.OwnerReference{
                        *metav1.NewControllerRef(mission, airforcev1alpha1.GroupVersion.WithKind("Mission")),
                    },
                },
                Spec: airforcev1alpha1.MissionStageSpec{
                    MissionRef:  corev1.LocalObjectReference{Name: mission.Name},
                    StageName:   stageSpec.DisplayName,
                    StageIndex:  i,
                    StageType:   stageSpec.Type,
                    FlightTasks: []airforcev1alpha1.FlightTaskSpec{}, // 从任务定义中解析
                    Config: airforcev1alpha1.StageConfig{
                        Timeout:    stageSpec.Timeout,
                        DependsOn:  stageSpec.DependsOn,
                    },
                },
            }

            // 解析FlightTask定义（从Mission配置或外部配置中读取）
            newStage.Spec.FlightTasks = r.parseFlightTasksForStage(stageSpec)

            if err := r.Create(ctx, newStage); err != nil {
                logger.Error(err, "创建MissionStage失败", "stage", stageName)
                return err
            }
            logger.Info("创建MissionStage成功", "stage", stageName)
        } else if err != nil {
            return err
        } else {
            // Stage已存在，检查是否需要更新
            if r.needsStageUpdate(stage, stageSpec) {
                stage.Spec = airforcev1alpha1.MissionStageSpec{
                    MissionRef:  corev1.LocalObjectReference{Name: mission.Name},
                    StageName:   stageSpec.DisplayName,
                    StageIndex:  i,
                    StageType:   stageSpec.Type,
                    FlightTasks: r.parseFlightTasksForStage(stageSpec),
                    Config: airforcev1alpha1.StageConfig{
                        Timeout:   stageSpec.Timeout,
                        DependsOn: stageSpec.DependsOn,
                    },
                }
                if err := r.Update(ctx, stage); err != nil {
                    logger.Error(err, "更新MissionStage失败", "stage", stageName)
                    return err
                }
                logger.Info("更新MissionStage成功", "stage", stageName)
            }
        }
    }

    return nil
}

// progressMissionStages 推进任务阶段（检查依赖，启动新阶段）
func (r *MissionReconciler) progressMissionStages(ctx context.Context, mission *airforcev1alpha1.Mission, allStages []airforcev1alpha1.MissionStage) error {
    logger := log.FromContext(ctx)

    for i := range allStages {
        stage := &allStages[i]

        // 如果Stage已经在运行或已完成，跳过
        if stage.Status.Phase == "Running" || stage.Status.Phase == "Succeeded" {
            continue
        }

        // 如果Stage是Pending，检查依赖是否满足
        if stage.Status.Phase == "Pending" || stage.Status.Phase == "" {
            dependenciesMet := r.checkStageDependencies(stage, allStages)
            if dependenciesMet {
                // 启动Stage
                stage.Status.Phase = "Running"
                stage.Status.StartTime = metav1.Now()
                if err := r.Status().Update(ctx, stage); err != nil {
                    return err
                }
                logger.Info("启动MissionStage", "stage", stage.Name)
            }
        }
    }

    return nil
}

// checkStageDependencies 检查Stage的依赖是否满足
func (r *MissionReconciler) checkStageDependencies(stage *airforcev1alpha1.MissionStage, allStages []airforcev1alpha1.MissionStage) bool {
    // 如果没有依赖，直接返回true
    if len(stage.Spec.Config.DependsOn) == 0 {
        return true
    }

    // 检查所有依赖的Stage是否都已成功完成
    for _, depName := range stage.Spec.Config.DependsOn {
        depStage := r.findStageByName(depName, allStages)
        if depStage == nil || depStage.Status.Phase != "Succeeded" {
            return false
        }
    }

    return true
}

// updateMissionStatus 汇总Stage状态，更新Mission状态
func (r *MissionReconciler) updateMissionStatus(ctx context.Context, mission *airforcev1alpha1.Mission, allStages []airforcev1alpha1.MissionStage) error {
    // 统计各状态的Stage数量
    var pending, running, succeeded, failed int
    stagesSummary := []airforcev1alpha1.StageSummary{}

    for _, stage := range allStages {
        switch stage.Status.Phase {
        case "Pending", "":
            pending++
        case "Running":
            running++
        case "Succeeded":
            succeeded++
        case "Failed":
            failed++
        }

        stagesSummary = append(stagesSummary, airforcev1alpha1.StageSummary{
            Name:           stage.Name,
            Phase:          stage.Status.Phase,
            StartTime:      stage.Status.StartTime,
            CompletionTime: stage.Status.CompletionTime,
        })
    }

    // 统计FlightTask数量（从所有Stage中汇总）
    totalTasks, succeededTasks, failedTasks, runningTasks, pendingTasks := r.countFlightTasks(ctx, mission)

    // 更新Mission状态
    mission.Status.StagesSummary = stagesSummary
    mission.Status.Statistics = airforcev1alpha1.MissionStatistics{
        TotalFlightTasks: totalTasks,
        SucceededTasks:   succeededTasks,
        FailedTasks:      failedTasks,
        RunningTasks:     runningTasks,
        PendingTasks:     pendingTasks,
    }
    mission.Status.LastUpdateTime = metav1.Now()

    // 更新Phase
    if running > 0 {
        mission.Status.Phase = "Running"
    } else if pending > 0 && succeeded > 0 {
        mission.Status.Phase = "Running"  // 有些Stage完成了，有些还在等待
    }

    mission.Status.Message = fmt.Sprintf("%d阶段运行中，%d阶段已完成，%d阶段待执行", running, succeeded, pending)

    return r.Status().Update(ctx, mission)
}

// isMissionComplete 检查任务是否完成
func (r *MissionReconciler) isMissionComplete(mission *airforcev1alpha1.Mission, allStages []airforcev1alpha1.MissionStage) bool {
    if len(allStages) == 0 {
        return false
    }

    for _, stage := range allStages {
        if stage.Status.Phase != "Succeeded" {
            return false
        }
    }

    return true
}

// isMissionFailed 检查任务是否失败
func (r *MissionReconciler) isMissionFailed(mission *airforcev1alpha1.Mission, allStages []airforcev1alpha1.MissionStage) bool {
    for _, stage := range allStages {
        if stage.Status.Phase == "Failed" {
            // 检查失败策略
            if mission.Spec.Config.FailurePolicy.StageFailureAction == "abort" {
                return true
            }
        }
    }
    return false
}

// SetupWithManager 设置Controller
func (r *MissionReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&airforcev1alpha1.Mission{}).
        Owns(&airforcev1alpha1.MissionStage{}).  // 监听子资源变化
        Complete(r)
}
```

## 三、MissionStage Controller

### 3.1 核心职责

1. 监听MissionStage CRD的创建、更新、删除
2. 根据Stage.Spec.FlightTasks创建FlightTask资源
3. 协调阶段内多个FlightTask的执行（串行、并行或混合）
4. 处理阶段超时
5. 汇总FlightTask状态，更新Stage状态

### 3.2 Reconcile逻辑

```go
type MissionStageReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func (r *MissionStageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // 1. 获取MissionStage对象
    stage := &airforcev1alpha1.MissionStage{}
    if err := r.Get(ctx, req.NamespacedName, stage); err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    // 2. 检查Stage是否应该运行（由Mission Controller设置为Running）
    if stage.Status.Phase != "Running" {
        return ctrl.Result{}, nil
    }

    // 3. 创建或更新FlightTask资源
    if err := r.reconcileFlightTasks(ctx, stage); err != nil {
        return ctrl.Result{}, err
    }

    // 4. 获取所有FlightTask的状态
    allTasks, err := r.listFlightTasks(ctx, stage)
    if err != nil {
        return ctrl.Result{}, err
    }

    // 5. 根据StageType协调FlightTask的执行
    switch stage.Spec.StageType {
    case "sequential":
        if err := r.progressSequentialTasks(ctx, stage, allTasks); err != nil {
            return ctrl.Result{}, err
        }
    case "parallel":
        if err := r.progressParallelTasks(ctx, stage, allTasks); err != nil {
            return ctrl.Result{}, err
        }
    case "mixed":
        if err := r.progressMixedTasks(ctx, stage, allTasks); err != nil {
            return ctrl.Result{}, err
        }
    }

    // 6. 更新Stage状态
    if err := r.updateStageStatus(ctx, stage, allTasks); err != nil {
        return ctrl.Result{}, err
    }

    // 7. 检查阶段是否完成
    if r.isStageComplete(stage, allTasks) {
        stage.Status.Phase = "Succeeded"
        stage.Status.CompletionTime = &metav1.Time{Time: time.Now()}
        if err := r.Status().Update(ctx, stage); err != nil {
            return ctrl.Result{}, err
        }
        logger.Info("阶段完成", "stage", stage.Name)
        return ctrl.Result{}, nil
    }

    // 8. 检查阶段是否失败
    if r.isStageFailed(stage, allTasks) {
        stage.Status.Phase = "Failed"
        stage.Status.Message = "阶段内有任务失败"
        if err := r.Status().Update(ctx, stage); err != nil {
            return ctrl.Result{}, err
        }
        return ctrl.Result{}, nil
    }

    // 9. 检查超时
    if r.isStageTimeout(stage) {
        stage.Status.Phase = "Failed"
        stage.Status.Message = "阶段执行超时"
        if err := r.Status().Update(ctx, stage); err != nil {
            return ctrl.Result{}, err
        }
        return ctrl.Result{}, nil
    }

    return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// progressSequentialTasks 串行执行任务
func (r *MissionStageReconciler) progressSequentialTasks(ctx context.Context, stage *airforcev1alpha1.MissionStage, allTasks []airforcev1alpha1.FlightTask) error {
    logger := log.FromContext(ctx)

    // 按顺序执行任务，一次只运行一个
    for i := range allTasks {
        task := &allTasks[i]

        // 如果任务已完成，继续下一个
        if task.Status.Phase == "Succeeded" {
            continue
        }

        // 如果任务正在运行，等待完成
        if task.Status.Phase == "Running" {
            return nil
        }

        // 如果任务失败，停止执行
        if task.Status.Phase == "Failed" {
            return nil
        }

        // 启动下一个任务
        if task.Status.Phase == "Pending" || task.Status.Phase == "" {
            // 检查前面的任务是否都已完成
            if i > 0 && allTasks[i-1].Status.Phase != "Succeeded" {
                // 前一个任务未完成，等待
                return nil
            }

            // 启动任务
            task.Status.Phase = "Scheduled"
            task.Status.SchedulingInfo.SchedulingAttempts = 1
            if err := r.Status().Update(ctx, task); err != nil {
                return err
            }
            logger.Info("启动FlightTask", "task", task.Name)
            return nil  // 一次只启动一个
        }
    }

    return nil
}

// progressParallelTasks 并行执行任务
func (r *MissionStageReconciler) progressParallelTasks(ctx context.Context, stage *airforcev1alpha1.MissionStage, allTasks []airforcev1alpha1.FlightTask) error {
    logger := log.FromContext(ctx)

    // 同时启动所有任务
    for i := range allTasks {
        task := &allTasks[i]

        if task.Status.Phase == "Pending" || task.Status.Phase == "" {
            task.Status.Phase = "Scheduled"
            task.Status.SchedulingInfo.SchedulingAttempts = 1
            if err := r.Status().Update(ctx, task); err != nil {
                return err
            }
            logger.Info("启动FlightTask", "task", task.Name)
        }
    }

    return nil
}

// SetupWithManager 设置Controller
func (r *MissionStageReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&airforcev1alpha1.MissionStage{}).
        Owns(&airforcev1alpha1.FlightTask{}).
        Complete(r)
}
```

## 四、FlightTask Controller

### 4.1 核心职责

1. 监听FlightTask CRD的创建、更新、删除
2. 根据FlightTask.Spec创建Pod
3. 注入武器Sidecar容器
4. 监听Pod状态，处理任务失败和重试
5. 更新飞机节点的资源占用情况

### 4.2 Reconcile逻辑

```go
type FlightTaskReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func (r *FlightTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // 1. 获取FlightTask对象
    task := &airforcev1alpha1.FlightTask{}
    if err := r.Get(ctx, req.NamespacedName, task); err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    // 2. 检查任务是否应该调度（由Stage Controller设置为Scheduled）
    if task.Status.Phase != "Scheduled" && task.Status.Phase != "Pending" {
        // 如果任务已经在运行或完成，监听Pod状态
        if task.Status.PodRef.Name != "" {
            return r.syncPodStatus(ctx, task)
        }
        return ctrl.Result{}, nil
    }

    // 3. 检查是否已有Pod存在
    if task.Status.PodRef.Name != "" {
        return r.syncPodStatus(ctx, task)
    }

    // 4. 创建Pod
    pod, err := r.createPodForFlightTask(ctx, task)
    if err != nil {
        logger.Error(err, "创建Pod失败", "task", task.Name)
        task.Status.Phase = "Failed"
        task.Status.Message = fmt.Sprintf("创建Pod失败: %v", err)
        r.Status().Update(ctx, task)
        return ctrl.Result{}, err
    }

    // 5. 更新FlightTask状态
    task.Status.Phase = "Running"
    task.Status.PodRef = corev1.LocalObjectReference{Name: pod.Name}
    task.Status.SchedulingInfo.AssignedNode = pod.Spec.NodeName
    task.Status.SchedulingInfo.AssignedTime = metav1.Now()
    task.Status.StartTime = metav1.Now()
    if err := r.Status().Update(ctx, task); err != nil {
        return ctrl.Result{}, err
    }

    logger.Info("FlightTask已调度", "task", task.Name, "pod", pod.Name, "node", pod.Spec.NodeName)

    return ctrl.Result{}, nil
}

// createPodForFlightTask 为FlightTask创建Pod
func (r *FlightTaskReconciler) createPodForFlightTask(ctx context.Context, task *airforcev1alpha1.FlightTask) (*corev1.Pod, error) {
    logger := log.FromContext(ctx)

    // 1. 基于PodTemplate创建Pod
    pod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("%s-pod-%s", task.Name, randString(6)),
            Namespace: task.Namespace,
            Labels:    task.Spec.PodTemplate.Labels,
            OwnerReferences: []metav1.OwnerReference{
                *metav1.NewControllerRef(task, airforcev1alpha1.GroupVersion.WithKind("FlightTask")),
            },
        },
        Spec: task.Spec.PodTemplate.Spec,
    }

    // 2. 注入武器Sidecar容器
    if err := r.injectWeaponSidecars(ctx, pod, task); err != nil {
        return nil, err
    }

    // 3. 创建Pod
    if err := r.Create(ctx, pod); err != nil {
        return nil, err
    }

    logger.Info("创建Pod成功", "pod", pod.Name)
    return pod, nil
}

// injectWeaponSidecars 注入武器Sidecar容器
func (r *FlightTaskReconciler) injectWeaponSidecars(ctx context.Context, pod *corev1.Pod, task *airforcev1alpha1.FlightTask) error {
    for _, weaponLoadout := range task.Spec.WeaponLoadout {
        // 获取Weapon CRD
        weapon := &airforcev1alpha1.Weapon{}
        if err := r.Get(ctx, client.ObjectKey{Name: weaponLoadout.WeaponRef.Name, Namespace: "weapons"}, weapon); err != nil {
            return err
        }

        // 创建Sidecar容器
        sidecar := corev1.Container{
            Name:  fmt.Sprintf("weapon-%s", weapon.Name),
            Image: fmt.Sprintf("%s:%s", weapon.Spec.Image.Repository, weapon.Spec.Image.Tag),
            Env:   weapon.Spec.Container.Env,
            VolumeMounts: weapon.Spec.Container.VolumeMounts,
            Ports:  weapon.Spec.Container.Ports,
            LivenessProbe: weapon.Spec.Container.LivenessProbe,
        }

        // 添加武器特定的环境变量
        sidecar.Env = append(sidecar.Env, corev1.EnvVar{
            Name:  "QUANTITY",
            Value: fmt.Sprintf("%d", weaponLoadout.Quantity),
        })
        sidecar.Env = append(sidecar.Env, corev1.EnvVar{
            Name:  "MOUNT_POINTS",
            Value: strings.Join(weaponLoadout.MountPoints, ","),
        })

        // 添加到Pod
        pod.Spec.Containers = append(pod.Spec.Containers, sidecar)
    }

    return nil
}

// syncPodStatus 同步Pod状态到FlightTask
func (r *FlightTaskReconciler) syncPodStatus(ctx context.Context, task *airforcev1alpha1.FlightTask) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // 获取Pod
    pod := &corev1.Pod{}
    if err := r.Get(ctx, client.ObjectKey{Name: task.Status.PodRef.Name, Namespace: task.Namespace}, pod); err != nil {
        if errors.IsNotFound(err) {
            // Pod不存在，可能被删除了
            logger.Info("Pod不存在", "pod", task.Status.PodRef.Name)
            return r.handlePodLost(ctx, task)
        }
        return ctrl.Result{}, err
    }

    // 同步Pod状态
    switch pod.Status.Phase {
    case corev1.PodSucceeded:
        // 任务成功完成
        task.Status.Phase = "Succeeded"
        task.Status.CompletionTime = &metav1.Time{Time: time.Now()}
        task.Status.Message = "任务成功完成"
        if err := r.Status().Update(ctx, task); err != nil {
            return ctrl.Result{}, err
        }
        logger.Info("FlightTask完成", "task", task.Name)

        // 释放飞机节点资源
        r.releaseAircraftResources(ctx, task, pod)

    case corev1.PodFailed:
        // 任务失败，尝试重试
        logger.Info("Pod失败", "pod", pod.Name, "task", task.Name)
        return r.handlePodFailure(ctx, task, pod)

    case corev1.PodRunning:
        // 任务运行中，更新执行状态
        task.Status.Phase = "Running"
        // 从Pod注解或容器日志中读取执行状态（位置、油量等）
        r.updateExecutionStatus(ctx, task, pod)
        if err := r.Status().Update(ctx, task); err != nil {
            return ctrl.Result{}, err
        }
    }

    return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// handlePodFailure 处理Pod失败
func (r *FlightTaskReconciler) handlePodFailure(ctx context.Context, task *airforcev1alpha1.FlightTask, pod *corev1.Pod) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // 检查重试次数
    if task.Status.RetryCount >= 3 {
        // 达到最大重试次数，标记任务失败
        task.Status.Phase = "Failed"
        task.Status.Message = fmt.Sprintf("任务失败，已达到最大重试次数(3次)")
        r.Status().Update(ctx, task)
        logger.Error(nil, "FlightTask失败", "task", task.Name)
        return ctrl.Result{}, nil
    }

    // 重试任务
    task.Status.RetryCount++
    task.Status.LastRetryTime = &metav1.Time{Time: time.Now()}
    task.Status.Phase = "Retrying"
    task.Status.Message = fmt.Sprintf("任务失败，准备第%d次重试", task.Status.RetryCount)

    // 标记失败的Node为不可用（添加Taint）
    if pod.Spec.NodeName != "" {
        r.taintFailedNode(ctx, pod.Spec.NodeName, task.Name)
    }

    // 删除失败的Pod
    r.Delete(ctx, pod)

    // 重置Pod引用，触发重新调度
    task.Status.PodRef.Name = ""
    task.Status.SchedulingInfo.AssignedNode = ""

    if err := r.Status().Update(ctx, task); err != nil {
        return ctrl.Result{}, err
    }

    logger.Info("FlightTask重试", "task", task.Name, "retryCount", task.Status.RetryCount)

    // 延迟重试（指数退避）
    retryDelay := time.Duration(1<<uint(task.Status.RetryCount)) * time.Second
    return ctrl.Result{RequeueAfter: retryDelay}, nil
}

// taintFailedNode 给失败的Node添加Taint
func (r *FlightTaskReconciler) taintFailedNode(ctx context.Context, nodeName string, taskName string) error {
    node := &corev1.Node{}
    if err := r.Get(ctx, client.ObjectKey{Name: nodeName}, node); err != nil {
        return err
    }

    // 添加Taint，防止任务再次调度到这个Node
    taint := corev1.Taint{
        Key:    "task-failed",
        Value:  taskName,
        Effect: corev1.TaintEffectNoSchedule,
    }

    node.Spec.Taints = append(node.Spec.Taints, taint)
    return r.Update(ctx, node)
}

// SetupWithManager 设置Controller
func (r *FlightTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&airforcev1alpha1.FlightTask{}).
        Owns(&corev1.Pod{}).
        Complete(r)
}
```

## 五、Aircraft Controller

### 5.1 核心职责

1. 定期更新飞机节点的状态标签（油量、挂载点、位置等）
2. 监听Pod调度事件，更新节点资源占用
3. 检测节点健康状态
4. 处理节点维护和上下线

### 5.2 实现逻辑

```go
type AircraftController struct {
    client.Client
    Scheme *runtime.Scheme
}

func (r *AircraftController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. 获取Node
    node := &corev1.Node{}
    if err := r.Get(ctx, req.NamespacedName, node); err != nil {
        return ctrl.Result{}, err
    }

    // 2. 检查是否是飞机节点
    if _, ok := node.Labels["aircraft.mil/type"]; !ok {
        return ctrl.Result{}, nil
    }

    // 3. 更新节点状态标签
    if err := r.updateAircraftStatus(ctx, node); err != nil {
        return ctrl.Result{}, err
    }

    // 4. 检查节点上的Pod状态
    if err := r.syncPodsOnNode(ctx, node); err != nil {
        return ctrl.Result{}, err
    }

    // 5. 更新资源分配情况
    if err := r.updateResourceAllocation(ctx, node); err != nil {
        return ctrl.Result{}, err
    }

    return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// updateAircraftStatus 更新飞机状态（模拟从飞机系统获取数据）
func (r *AircraftController) updateAircraftStatus(ctx context.Context, node *corev1.Node) error {
    // 模拟从飞机系统获取状态数据
    // 实际应该通过Agent或API获取
    fuelLevel := r.getAircraftFuelLevel(node.Name)  // 模拟函数
    location := r.getAircraftLocation(node.Name)    // 模拟函数

    // 更新标签
    if node.Labels == nil {
        node.Labels = make(map[string]string)
    }

    node.Labels["aircraft.mil/fuel.level"] = fmt.Sprintf("%d", fuelLevel)
    if fuelLevel > 70 {
        node.Labels["aircraft.mil/fuel.status"] = "high"
    } else if fuelLevel > 30 {
        node.Labels["aircraft.mil/fuel.status"] = "medium"
    } else {
        node.Labels["aircraft.mil/fuel.status"] = "low"
    }

    node.Labels["aircraft.mil/location.zone"] = location

    return r.Update(ctx, node)
}
```

## 六、Controller协作流程

### 6.1 任务创建流程

```
用户创建Mission
       ↓
Mission Controller监听到创建事件
       ↓
创建MissionStage资源（设置Phase=Pending）
       ↓
检查依赖，将Stage1设置为Phase=Running
       ↓
Stage Controller监听到Stage1变为Running
       ↓
创建FlightTask资源（设置Phase=Scheduled）
       ↓
FlightTask Controller监听到创建事件
       ↓
创建Pod，注入武器Sidecar
       ↓
K8s Scheduler调度Pod到合适的飞机Node
       ↓
Pod运行，FlightTask Controller同步状态
       ↓
Pod完成，FlightTask标记为Succeeded
       ↓
Stage Controller检测所有FlightTask完成，标记Stage为Succeeded
       ↓
Mission Controller检测Stage1完成，启动Stage2
       ↓
...循环直到所有Stage完成
       ↓
Mission标记为Succeeded
```

### 6.2 任务失败重试流程

```
Pod执行失败
       ↓
FlightTask Controller检测到失败
       ↓
检查重试次数 < 3
       ↓
给失败的Node添加Taint
       ↓
删除失败的Pod
       ↓
FlightTask状态设为Retrying
       ↓
重新创建Pod
       ↓
Scheduler避开被Taint的Node，选择新的Node
       ↓
新Pod运行
```

## 七、下一步

下一章将详细讲解**武器管理方案**，包括：
- 武器容器化设计
- 武器挂载和注入机制
- 武器版本管理和升级
- 武器池预加载优化
