/*
Copyright 2026 yydashuai.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	airforcev1alpha1 "github.com/yydashuai/mission-system/api/v1alpha1"
)

// MissionStageReconciler reconciles a MissionStage object
type MissionStageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=missionstages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=missionstages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=missionstages/finalizers,verbs=update
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=flighttasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=flighttasks/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MissionStage object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *MissionStageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var stage airforcev1alpha1.MissionStage
	if err := r.Get(ctx, req.NamespacedName, &stage); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if stage.Status.Phase == "" {
		patch := client.MergeFrom(stage.DeepCopy())
		stage.Status.Phase = airforcev1alpha1.MissionStagePhasePending
		if err := r.Status().Patch(ctx, &stage, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.reconcileFlightTasks(ctx, &stage); err != nil {
		return ctrl.Result{}, err
	}

	tasks, err := r.listFlightTasks(ctx, &stage)
	if err != nil {
		return ctrl.Result{}, err
	}

	if stage.Status.Phase == airforcev1alpha1.MissionStagePhaseRunning && stage.Status.StartTime == nil {
		patch := client.MergeFrom(stage.DeepCopy())
		now := metav1.Now()
		stage.Status.StartTime = &now
		if err := r.Status().Patch(ctx, &stage, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	if stage.Status.Phase == airforcev1alpha1.MissionStagePhaseRunning {
		if err := r.progressTasks(ctx, &stage, tasks); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.updateStageStatus(ctx, &stage, tasks); err != nil {
		logger.Error(err, "failed to update MissionStage status")
		return ctrl.Result{}, err
	}

	if stage.Status.Phase == airforcev1alpha1.MissionStagePhaseRunning && r.isStageTimeout(&stage) {
		patch := client.MergeFrom(stage.DeepCopy())
		stage.Status.Phase = airforcev1alpha1.MissionStagePhaseFailed
		stage.Status.Message = "Stage timed out"
		now := metav1.Now()
		stage.Status.CompletionTime = &now
		if err := r.Status().Patch(ctx, &stage, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MissionStageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&airforcev1alpha1.MissionStage{}).
		Owns(&airforcev1alpha1.FlightTask{}).
		Complete(r)
}

func (r *MissionStageReconciler) listFlightTasks(ctx context.Context, stage *airforcev1alpha1.MissionStage) ([]airforcev1alpha1.FlightTask, error) {
	var taskList airforcev1alpha1.FlightTaskList
	if err := r.List(ctx, &taskList, &client.ListOptions{
		Namespace: stage.Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{
			"mission": stage.Spec.MissionRef.Name,
			"stage":   stage.Name,
		}),
	}); err != nil {
		return nil, err
	}
	return taskList.Items, nil
}

func (r *MissionStageReconciler) reconcileFlightTasks(ctx context.Context, stage *airforcev1alpha1.MissionStage) error {
	desired := make(map[string]airforcev1alpha1.MissionStageFlightTaskTemplate, len(stage.Spec.FlightTasks))
	for _, tmpl := range stage.Spec.FlightTasks {
		if tmpl.Name == "" {
			continue
		}
		desired[tmpl.Name] = tmpl
	}

	for index, tmpl := range stage.Spec.FlightTasks {
		if tmpl.Name == "" {
			continue
		}
		taskObjName := fmt.Sprintf("%s-%s", stage.Name, tmpl.Name)

		var task airforcev1alpha1.FlightTask
		err := r.Get(ctx, client.ObjectKey{Namespace: stage.Namespace, Name: taskObjName}, &task)
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}

		desiredLabels := map[string]string{
			"mission":     stage.Spec.MissionRef.Name,
			"stage":       stage.Name,
			"task-name":   tmpl.Name,
			"task-index":  strconv.Itoa(index + 1),
			"aircraft":    tmpl.Aircraft,
			"task-role":   tmpl.Role,
			"stage-index": strconv.Itoa(int(stage.Spec.StageIndex)),
		}

		desiredSpec := airforcev1alpha1.FlightTaskSpec{
			StageRef: airforcev1alpha1.MissionStageRef{Name: stage.Name},
			AircraftRequirement: airforcev1alpha1.AircraftRequirement{
				Type: tmpl.Aircraft,
			},
			Role: tmpl.Role,
		}
		if len(tmpl.TaskParams) > 0 {
			desiredSpec.TaskParams = &airforcev1alpha1.FlightTaskParams{Extra: tmpl.TaskParams}
		}
		if len(tmpl.WeaponLoadout) > 0 {
			desiredSpec.WeaponLoadout = make([]airforcev1alpha1.FlightTaskWeaponLoadoutItem, 0, len(tmpl.WeaponLoadout))
			for _, weapon := range tmpl.WeaponLoadout {
				if weapon.Weapon == "" {
					continue
				}
				desiredSpec.WeaponLoadout = append(desiredSpec.WeaponLoadout, airforcev1alpha1.FlightTaskWeaponLoadoutItem{
					WeaponRef: airforcev1alpha1.WeaponRef{Name: weapon.Weapon},
					Quantity:  weapon.Quantity,
				})
			}
		}

		if apierrors.IsNotFound(err) {
			task = airforcev1alpha1.FlightTask{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: stage.Namespace,
					Name:      taskObjName,
					Labels:    desiredLabels,
				},
				Spec: desiredSpec,
			}
			if err := controllerutil.SetControllerReference(stage, &task, r.Scheme); err != nil {
				return err
			}
			if err := r.Create(ctx, &task); err != nil {
				return err
			}
			continue
		}

		patch := client.MergeFrom(task.DeepCopy())
		changed := false

		if task.Labels == nil {
			task.Labels = map[string]string{}
		}
		for k, v := range desiredLabels {
			if task.Labels[k] != v {
				task.Labels[k] = v
				changed = true
			}
		}

		if task.Spec.StageRef.Name != desiredSpec.StageRef.Name {
			task.Spec.StageRef = desiredSpec.StageRef
			changed = true
		}
		if task.Spec.AircraftRequirement.Type != desiredSpec.AircraftRequirement.Type {
			task.Spec.AircraftRequirement.Type = desiredSpec.AircraftRequirement.Type
			changed = true
		}
		if task.Spec.Role != desiredSpec.Role {
			task.Spec.Role = desiredSpec.Role
			changed = true
		}
		if desiredSpec.TaskParams == nil && task.Spec.TaskParams != nil {
			task.Spec.TaskParams = nil
			changed = true
		}
		if desiredSpec.TaskParams != nil {
			if task.Spec.TaskParams == nil || !mapsEqual(task.Spec.TaskParams.Extra, desiredSpec.TaskParams.Extra) {
				task.Spec.TaskParams = desiredSpec.TaskParams
				changed = true
			}
		}
		if len(desiredSpec.WeaponLoadout) == 0 && len(task.Spec.WeaponLoadout) != 0 {
			task.Spec.WeaponLoadout = nil
			changed = true
		}
		if len(desiredSpec.WeaponLoadout) != 0 && !weaponLoadoutEqual(task.Spec.WeaponLoadout, desiredSpec.WeaponLoadout) {
			task.Spec.WeaponLoadout = desiredSpec.WeaponLoadout
			changed = true
		}

		if changed {
			if err := r.Patch(ctx, &task, patch); err != nil {
				return err
			}
		}
	}

	// Delete tasks that are no longer referenced in stage.spec.flightTasks.
	tasks, err := r.listFlightTasks(ctx, stage)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		taskName := task.Labels["task-name"]
		if taskName == "" {
			continue
		}
		if _, ok := desired[taskName]; ok {
			continue
		}
		_ = r.Delete(ctx, &task)
	}

	return nil
}

func (r *MissionStageReconciler) progressTasks(ctx context.Context, stage *airforcev1alpha1.MissionStage, tasks []airforcev1alpha1.FlightTask) error {
	if len(tasks) == 0 {
		return nil
	}

	sort.Slice(tasks, func(i, j int) bool {
		li, _ := strconv.Atoi(tasks[i].Labels["task-index"])
		lj, _ := strconv.Atoi(tasks[j].Labels["task-index"])
		if li == 0 || lj == 0 {
			return tasks[i].Name < tasks[j].Name
		}
		return li < lj
	})

	switch stage.Spec.StageType {
	case airforcev1alpha1.StageExecutionTypeSequential:
		for i := range tasks {
			task := &tasks[i]
			if task.Status.Phase == airforcev1alpha1.FlightTaskPhaseSucceeded {
				continue
			}
			if task.Status.Phase == airforcev1alpha1.FlightTaskPhaseFailed {
				return nil
			}
			if task.Status.Phase == airforcev1alpha1.FlightTaskPhaseRunning || task.Status.Phase == airforcev1alpha1.FlightTaskPhaseScheduled {
				return nil
			}

			if task.Status.Phase == airforcev1alpha1.FlightTaskPhasePending || task.Status.Phase == "" {
				patch := client.MergeFrom(task.DeepCopy())
				task.Status.Phase = airforcev1alpha1.FlightTaskPhaseScheduled
				now := metav1.Now()
				if task.Status.SchedulingInfo == nil {
					task.Status.SchedulingInfo = &airforcev1alpha1.SchedulingInfo{}
				}
				if task.Status.SchedulingInfo.SchedulingAttempts == 0 {
					task.Status.SchedulingInfo.SchedulingAttempts = 1
					task.Status.SchedulingInfo.AssignedTime = &now
				}
				if err := r.Status().Patch(ctx, task, patch); err != nil {
					return err
				}
				return nil
			}
		}
		return nil

	case airforcev1alpha1.StageExecutionTypeParallel, airforcev1alpha1.StageExecutionTypeMixed:
		for i := range tasks {
			task := &tasks[i]
			if task.Status.Phase != airforcev1alpha1.FlightTaskPhasePending && task.Status.Phase != "" {
				continue
			}
			patch := client.MergeFrom(task.DeepCopy())
			task.Status.Phase = airforcev1alpha1.FlightTaskPhaseScheduled
			now := metav1.Now()
			if task.Status.SchedulingInfo == nil {
				task.Status.SchedulingInfo = &airforcev1alpha1.SchedulingInfo{}
			}
			if task.Status.SchedulingInfo.SchedulingAttempts == 0 {
				task.Status.SchedulingInfo.SchedulingAttempts = 1
				task.Status.SchedulingInfo.AssignedTime = &now
			}
			if err := r.Status().Patch(ctx, task, patch); err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}

func (r *MissionStageReconciler) updateStageStatus(ctx context.Context, stage *airforcev1alpha1.MissionStage, tasks []airforcev1alpha1.FlightTask) error {
	taskByName := make(map[string]airforcev1alpha1.FlightTask, len(tasks))
	for _, task := range tasks {
		taskByName[task.Labels["task-name"]] = task
	}

	statuses := make([]airforcev1alpha1.MissionStageFlightTaskStatus, 0, len(stage.Spec.FlightTasks))
	var pending, scheduled, running, succeeded, failed int
	for _, tmpl := range stage.Spec.FlightTasks {
		if tmpl.Name == "" {
			continue
		}
		task, ok := taskByName[tmpl.Name]
		phase := airforcev1alpha1.FlightTaskPhasePending
		if ok && task.Status.Phase != "" {
			phase = task.Status.Phase
		}

		switch phase {
		case airforcev1alpha1.FlightTaskPhaseSucceeded:
			succeeded++
		case airforcev1alpha1.FlightTaskPhaseFailed:
			failed++
		case airforcev1alpha1.FlightTaskPhaseRunning:
			running++
		case airforcev1alpha1.FlightTaskPhaseScheduled:
			scheduled++
		default:
			pending++
		}

		statuses = append(statuses, airforcev1alpha1.MissionStageFlightTaskStatus{
			Name:  tmpl.Name,
			Phase: phase,
		})
	}

	patch := client.MergeFrom(stage.DeepCopy())
	stage.Status.FlightTasksStatus = statuses
	stage.Status.Message = fmt.Sprintf("tasks: pending=%d scheduled=%d running=%d succeeded=%d failed=%d", pending, scheduled, running, succeeded, failed)

	if stage.Status.Phase == airforcev1alpha1.MissionStagePhaseRunning {
		if failed > 0 {
			stage.Status.Phase = airforcev1alpha1.MissionStagePhaseFailed
			now := metav1.Now()
			stage.Status.CompletionTime = &now
		} else if len(statuses) == 0 || succeeded == len(statuses) {
			stage.Status.Phase = airforcev1alpha1.MissionStagePhaseSucceeded
			now := metav1.Now()
			stage.Status.CompletionTime = &now
		}
	}

	if stage.Status.Phase == airforcev1alpha1.MissionStagePhaseSucceeded || stage.Status.Phase == airforcev1alpha1.MissionStagePhaseFailed {
		if stage.Status.CompletionTime == nil {
			now := metav1.Now()
			stage.Status.CompletionTime = &now
		}
	}

	if stage.Status.StartTime == nil && stage.Status.Phase == airforcev1alpha1.MissionStagePhaseRunning {
		now := metav1.Now()
		stage.Status.StartTime = &now
	}

	return r.Status().Patch(ctx, stage, patch)
}

func (r *MissionStageReconciler) isStageTimeout(stage *airforcev1alpha1.MissionStage) bool {
	if stage.Spec.Config == nil || stage.Spec.Config.Timeout == nil {
		return false
	}
	if stage.Status.StartTime == nil {
		return false
	}
	timeout := stage.Spec.Config.Timeout.Duration
	return time.Since(stage.Status.StartTime.Time) > timeout
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func weaponLoadoutEqual(a, b []airforcev1alpha1.FlightTaskWeaponLoadoutItem) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].WeaponRef.Name != b[i].WeaponRef.Name {
			return false
		}
		if a[i].Quantity != b[i].Quantity {
			return false
		}
		if len(a[i].MountPoints) != len(b[i].MountPoints) {
			return false
		}
		for j := range a[i].MountPoints {
			if a[i].MountPoints[j] != b[i].MountPoints[j] {
				return false
			}
		}
	}
	return true
}
