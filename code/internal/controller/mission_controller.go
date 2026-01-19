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
	"strings"
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

// MissionReconciler reconciles a Mission object
type MissionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=missions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=missions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=missions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Mission object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *MissionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var mission airforcev1alpha1.Mission
	if err := r.Get(ctx, req.NamespacedName, &mission); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if mission.Status.Phase == "" {
		patch := client.MergeFrom(mission.DeepCopy())
		mission.Status.Phase = airforcev1alpha1.MissionPhasePending
		now := metav1.Now()
		mission.Status.StartTime = &now
		mission.Status.LastUpdateTime = &now
		if err := r.Status().Patch(ctx, &mission, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 1) Ensure MissionStage resources exist (and update their Spec if needed).
	for index, stage := range mission.Spec.Stages {
		if stage.Name == "" {
			continue
		}

		stageObjName := fmt.Sprintf("%s-%s", mission.Name, stage.Name)
		var missionStage airforcev1alpha1.MissionStage
		err := r.Get(ctx, client.ObjectKey{Namespace: mission.Namespace, Name: stageObjName}, &missionStage)
		if err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		if apierrors.IsNotFound(err) {
			flightTasks := normalizeStageFlightTasks(stage.FlightTasks)
			missionStage = airforcev1alpha1.MissionStage{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: mission.Namespace,
					Name:      stageObjName,
					Labels: map[string]string{
						"mission":     mission.Name,
						"stage-name":  stage.Name,
						"stage-index": fmt.Sprintf("%d", index+1),
					},
				},
				Spec: airforcev1alpha1.MissionStageSpec{
					MissionRef:  airforcev1alpha1.MissionRef{Name: mission.Name},
					StageName:   stage.DisplayName,
					StageIndex:  int32(index + 1),
					StageType:   stage.Type,
					FlightTasks: flightTasks,
					Config: &airforcev1alpha1.MissionStageConfig{
						Timeout: stage.Timeout,
					},
				},
			}
			if missionStage.Spec.StageName == "" {
				missionStage.Spec.StageName = stage.Name
			}
			if err := controllerutil.SetControllerReference(&mission, &missionStage, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.Create(ctx, &missionStage); err != nil {
				return ctrl.Result{}, err
			}
			continue
		}

		patch := client.MergeFrom(missionStage.DeepCopy())
		changed := false

		if missionStage.Labels == nil {
			missionStage.Labels = map[string]string{}
		}
		if missionStage.Labels["mission"] != mission.Name {
			missionStage.Labels["mission"] = mission.Name
			changed = true
		}
		if missionStage.Labels["stage-name"] != stage.Name {
			missionStage.Labels["stage-name"] = stage.Name
			changed = true
		}
		if missionStage.Labels["stage-index"] != fmt.Sprintf("%d", index+1) {
			missionStage.Labels["stage-index"] = fmt.Sprintf("%d", index+1)
			changed = true
		}

		if missionStage.Spec.MissionRef.Name != mission.Name {
			missionStage.Spec.MissionRef.Name = mission.Name
			changed = true
		}
		stageName := stage.DisplayName
		if stageName == "" {
			stageName = stage.Name
		}
		if missionStage.Spec.StageName != stageName {
			missionStage.Spec.StageName = stageName
			changed = true
		}
		if missionStage.Spec.StageIndex != int32(index+1) {
			missionStage.Spec.StageIndex = int32(index + 1)
			changed = true
		}
		if missionStage.Spec.StageType != stage.Type {
			missionStage.Spec.StageType = stage.Type
			changed = true
		}
		desiredFlightTasks := normalizeStageFlightTasks(stage.FlightTasks)
		if !missionStageFlightTasksEqual(missionStage.Spec.FlightTasks, desiredFlightTasks) {
			missionStage.Spec.FlightTasks = desiredFlightTasks
			changed = true
		}
		if missionStage.Spec.Config == nil {
			missionStage.Spec.Config = &airforcev1alpha1.MissionStageConfig{}
			changed = true
		}
		if missionStage.Spec.Config.Timeout == nil && stage.Timeout != nil {
			missionStage.Spec.Config.Timeout = stage.Timeout
			changed = true
		}
		if missionStage.Spec.Config.Timeout != nil && stage.Timeout == nil {
			missionStage.Spec.Config.Timeout = nil
			changed = true
		}
		if missionStage.Spec.Config.Timeout != nil && stage.Timeout != nil && missionStage.Spec.Config.Timeout.Duration != stage.Timeout.Duration {
			missionStage.Spec.Config.Timeout = stage.Timeout
			changed = true
		}

		if changed {
			if err := r.Patch(ctx, &missionStage, patch); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// 2) Delete MissionStage resources that are no longer referenced by Mission.spec.stages.
	var existingMissionStages airforcev1alpha1.MissionStageList
	if err := r.List(ctx, &existingMissionStages, &client.ListOptions{
		Namespace:     mission.Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{"mission": mission.Name}),
	}); err != nil {
		return ctrl.Result{}, err
	}
	desiredNames := make(map[string]struct{}, len(mission.Spec.Stages))
	for _, stage := range mission.Spec.Stages {
		if stage.Name == "" {
			continue
		}
		desiredNames[fmt.Sprintf("%s-%s", mission.Name, stage.Name)] = struct{}{}
	}
	for _, ms := range existingMissionStages.Items {
		if _, ok := desiredNames[ms.Name]; ok {
			continue
		}
		_ = r.Delete(ctx, &ms)
	}

	// 3) Progress stage phases based on Mission.spec.stages dependency graph.
	failureAction := stageFailureAction(&mission)
	stagesByName := make(map[string]*airforcev1alpha1.MissionStage, len(existingMissionStages.Items))
	for i := range existingMissionStages.Items {
		stage := &existingMissionStages.Items[i]
		stagesByName[stage.Name] = stage
	}
	for _, stageTemplate := range mission.Spec.Stages {
		if stageTemplate.Name == "" {
			continue
		}
		stageObjName := fmt.Sprintf("%s-%s", mission.Name, stageTemplate.Name)
		stage, ok := stagesByName[stageObjName]
		if !ok {
			continue
		}

		if stage.Status.Phase == "" {
			patch := client.MergeFrom(stage.DeepCopy())
			stage.Status.Phase = airforcev1alpha1.MissionStagePhasePending
			if err := r.Status().Patch(ctx, stage, patch); err != nil {
				return ctrl.Result{}, err
			}
		}

		if stage.Status.Phase != airforcev1alpha1.MissionStagePhasePending {
			continue
		}

		depsMet := true
		for _, dep := range stageTemplate.DependsOn {
			if dep == "" {
				continue
			}
			depObjName := fmt.Sprintf("%s-%s", mission.Name, dep)
			depStage, ok := stagesByName[depObjName]
			if !ok {
				depsMet = false
				break
			}
			if depStage.Status.Phase != airforcev1alpha1.MissionStagePhaseSucceeded {
				if failureAction == airforcev1alpha1.StageFailureActionContinue &&
					depStage.Status.Phase == airforcev1alpha1.MissionStagePhaseFailed {
					continue
				}
				depsMet = false
				break
			}
		}
		if !depsMet {
			continue
		}

		patch := client.MergeFrom(stage.DeepCopy())
		stage.Status.Phase = airforcev1alpha1.MissionStagePhaseRunning
		now := metav1.Now()
		if stage.Status.StartTime == nil {
			stage.Status.StartTime = &now
		}
		stage.Status.Message = "Stage started by Mission controller"
		if err := r.Status().Patch(ctx, stage, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 3) Summarize stage phases back into Mission.status.
	if err := r.List(ctx, &existingMissionStages, &client.ListOptions{
		Namespace:     mission.Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{"mission": mission.Name}),
	}); err != nil {
		return ctrl.Result{}, err
	}
	missionStageByName := make(map[string]airforcev1alpha1.MissionStage, len(existingMissionStages.Items))
	for _, ms := range existingMissionStages.Items {
		missionStageByName[ms.Name] = ms
	}

	patch := client.MergeFrom(mission.DeepCopy())
	now := metav1.Now()
	mission.Status.LastUpdateTime = &now

	summaries := make([]airforcev1alpha1.MissionStageSummary, 0, len(mission.Spec.Stages))
	stagePhases := make([]airforcev1alpha1.MissionPhase, 0, len(mission.Spec.Stages))
	var failedStages, runningStages, pendingStages, succeededStages int
	for _, stage := range mission.Spec.Stages {
		if stage.Name == "" {
			continue
		}
		stageObjName := fmt.Sprintf("%s-%s", mission.Name, stage.Name)
		ms, ok := missionStageByName[stageObjName]
		if !ok {
			summaries = append(summaries, airforcev1alpha1.MissionStageSummary{
				Name:  stage.Name,
				Phase: airforcev1alpha1.MissionPhasePending,
			})
			stagePhases = append(stagePhases, airforcev1alpha1.MissionPhasePending)
			continue
		}

		phase := airforcev1alpha1.MissionPhasePending
		switch ms.Status.Phase {
		case airforcev1alpha1.MissionStagePhaseRunning:
			phase = airforcev1alpha1.MissionPhaseRunning
		case airforcev1alpha1.MissionStagePhaseSucceeded:
			phase = airforcev1alpha1.MissionPhaseSucceeded
		case airforcev1alpha1.MissionStagePhaseFailed:
			phase = airforcev1alpha1.MissionPhaseFailed
		}

		summaries = append(summaries, airforcev1alpha1.MissionStageSummary{
			Name:           stage.Name,
			Phase:          phase,
			StartTime:      ms.Status.StartTime,
			CompletionTime: ms.Status.CompletionTime,
		})
		stagePhases = append(stagePhases, phase)
		switch phase {
		case airforcev1alpha1.MissionPhaseFailed:
			failedStages++
		case airforcev1alpha1.MissionPhaseRunning:
			runningStages++
		case airforcev1alpha1.MissionPhaseSucceeded:
			succeededStages++
		default:
			pendingStages++
		}
	}
	mission.Status.StagesSummary = summaries

	desiredMissionPhase := airforcev1alpha1.MissionPhasePending
	if len(stagePhases) == 0 {
		desiredMissionPhase = airforcev1alpha1.MissionPhasePending
	} else {
		if failureAction == airforcev1alpha1.StageFailureActionContinue {
			if runningStages > 0 {
				desiredMissionPhase = airforcev1alpha1.MissionPhaseRunning
			} else if pendingStages > 0 {
				desiredMissionPhase = airforcev1alpha1.MissionPhasePending
			} else {
				desiredMissionPhase = airforcev1alpha1.MissionPhaseSucceeded
			}
		} else {
			if failedStages > 0 {
				desiredMissionPhase = airforcev1alpha1.MissionPhaseFailed
			} else if runningStages > 0 {
				desiredMissionPhase = airforcev1alpha1.MissionPhaseRunning
			} else if pendingStages > 0 {
				desiredMissionPhase = airforcev1alpha1.MissionPhasePending
			} else {
				desiredMissionPhase = airforcev1alpha1.MissionPhaseSucceeded
			}
		}
	}
	mission.Status.Phase = desiredMissionPhase

	// 4) Summarize FlightTask statistics (best effort).
	var taskList airforcev1alpha1.FlightTaskList
	if err := r.List(ctx, &taskList, &client.ListOptions{
		Namespace:     mission.Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{"mission": mission.Name}),
	}); err != nil {
		logger.Error(err, "failed to list FlightTasks for mission statistics")
	} else {
		stats := &airforcev1alpha1.MissionStatistics{}
		stats.TotalFlightTasks = int32(len(taskList.Items))
		for _, task := range taskList.Items {
			switch task.Status.Phase {
			case airforcev1alpha1.FlightTaskPhaseSucceeded:
				stats.SucceededTasks++
			case airforcev1alpha1.FlightTaskPhaseFailed:
				stats.FailedTasks++
			case airforcev1alpha1.FlightTaskPhaseRunning:
				stats.RunningTasks++
			case airforcev1alpha1.FlightTaskPhaseScheduled, airforcev1alpha1.FlightTaskPhasePending, "":
				stats.PendingTasks++
			default:
				stats.PendingTasks++
			}
		}
		mission.Status.Statistics = stats
	}

	if err := r.Status().Patch(ctx, &mission, patch); err != nil {
		logger.Error(err, "failed to update Mission status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func normalizeStageFlightTasks(tasks []airforcev1alpha1.MissionStageFlightTaskTemplate) []airforcev1alpha1.MissionStageFlightTaskTemplate {
	if len(tasks) == 0 {
		return nil
	}
	out := make([]airforcev1alpha1.MissionStageFlightTaskTemplate, 0, len(tasks))
	for i, task := range tasks {
		if strings.TrimSpace(task.Name) == "" {
			name := strings.TrimSpace(task.Aircraft)
			if name == "" {
				name = "task"
			}
			task.Name = fmt.Sprintf("%s-%02d", name, i+1)
		}
		out = append(out, task)
	}
	return out
}

func missionStageFlightTasksEqual(a, b []airforcev1alpha1.MissionStageFlightTaskTemplate) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name {
			return false
		}
		if a[i].Aircraft != b[i].Aircraft {
			return false
		}
		if a[i].Role != b[i].Role {
			return false
		}
		if a[i].Priority != b[i].Priority {
			return false
		}
		if len(a[i].WeaponLoadout) != len(b[i].WeaponLoadout) {
			return false
		}
		for j := range a[i].WeaponLoadout {
			if a[i].WeaponLoadout[j].Weapon != b[i].WeaponLoadout[j].Weapon {
				return false
			}
			if a[i].WeaponLoadout[j].Quantity != b[i].WeaponLoadout[j].Quantity {
				return false
			}
		}
		if len(a[i].TaskParams) != len(b[i].TaskParams) {
			return false
		}
		for k, v := range a[i].TaskParams {
			if b[i].TaskParams[k] != v {
				return false
			}
		}
	}
	return true
}

func stageFailureAction(mission *airforcev1alpha1.Mission) airforcev1alpha1.StageFailureAction {
	if mission == nil || mission.Spec.Config == nil || mission.Spec.Config.FailurePolicy == nil {
		return airforcev1alpha1.StageFailureActionAbort
	}
	action := mission.Spec.Config.FailurePolicy.StageFailureAction
	switch action {
	case airforcev1alpha1.StageFailureActionAbort,
		airforcev1alpha1.StageFailureActionContinue,
		airforcev1alpha1.StageFailureActionRetry:
		return action
	default:
		return airforcev1alpha1.StageFailureActionAbort
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MissionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&airforcev1alpha1.Mission{}).
		Owns(&airforcev1alpha1.MissionStage{}).
		Complete(r)
}
