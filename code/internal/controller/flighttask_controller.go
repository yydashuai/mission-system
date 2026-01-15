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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	airforcev1alpha1 "github.com/yydashuai/mission-system/api/v1alpha1"
)

// FlightTaskReconciler reconciles a FlightTask object
type FlightTaskReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// APIReader is used for direct apiserver reads (e.g. listing Events with field selectors),
	// because cached clients do not support arbitrary field selectors.
	APIReader client.Reader
}

//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=flighttasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=flighttasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=flighttasks/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch
//+kubebuilder:rbac:groups=events.k8s.io,resources=events,verbs=get;list;watch
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=weapons,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FlightTask object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *FlightTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var task airforcev1alpha1.FlightTask
	if err := r.Get(ctx, req.NamespacedName, &task); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if task.Status.Phase == "" {
		patch := client.MergeFrom(task.DeepCopy())
		task.Status.Phase = airforcev1alpha1.FlightTaskPhasePending
		if err := r.Status().Patch(ctx, &task, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	if task.Status.Phase == airforcev1alpha1.FlightTaskPhasePending && isStandaloneFlightTask(&task) && task.Status.PodRef == nil {
		patch := client.MergeFrom(task.DeepCopy())
		task.Status.Phase = airforcev1alpha1.FlightTaskPhaseScheduled
		if task.Status.SchedulingInfo == nil {
			task.Status.SchedulingInfo = &airforcev1alpha1.SchedulingInfo{}
		}
		if task.Status.SchedulingInfo.SchedulingAttempts == 0 {
			task.Status.SchedulingInfo.SchedulingAttempts = 1
		}
		if err := r.Status().Patch(ctx, &task, patch); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	podName := fmt.Sprintf("%s-pod", task.Name)
	ensurePod := task.Status.PodRef != nil ||
		task.Status.Phase == airforcev1alpha1.FlightTaskPhaseScheduled ||
		task.Status.Phase == airforcev1alpha1.FlightTaskPhaseRunning
	if ensurePod {
		var pod corev1.Pod
		err := r.Get(ctx, client.ObjectKey{Namespace: task.Namespace, Name: podName}, &pod)
		if err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		if apierrors.IsNotFound(err) {
			desiredPod, err := r.buildPodForTask(ctx, &task, podName)
			if err != nil {
				logger.Error(err, "failed to build pod for FlightTask", "flightTask", task.Name)
				patch := client.MergeFrom(task.DeepCopy())
				task.Status.Phase = airforcev1alpha1.FlightTaskPhaseFailed
				meta := metav1.Condition{
					Type:               "PodCreated",
					Status:             metav1.ConditionFalse,
					Reason:             "InvalidSpec",
					Message:            err.Error(),
					ObservedGeneration: task.Generation,
				}
				apimeta.SetStatusCondition(&task.Status.Conditions, meta)
				_ = r.Status().Patch(ctx, &task, patch)
				return ctrl.Result{}, nil
			}
			if err := controllerutil.SetControllerReference(&task, desiredPod, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.Create(ctx, desiredPod); err != nil {
				return ctrl.Result{}, err
			}
			// Refetch to ensure UID is populated before writing PodRef.
			if err := r.Get(ctx, client.ObjectKey{Namespace: desiredPod.Namespace, Name: desiredPod.Name}, desiredPod); err != nil {
				if apierrors.IsNotFound(err) {
					return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
				}
				return ctrl.Result{}, err
			}

			podRef := podReference(desiredPod)
			if podRef == nil || podRef.UID == "" {
				return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
			}

			patch := client.MergeFrom(task.DeepCopy())
			task.Status.PodRef = podRef
			task.Status.Phase = airforcev1alpha1.FlightTaskPhasePending
			if task.Status.SchedulingInfo == nil {
				task.Status.SchedulingInfo = &airforcev1alpha1.SchedulingInfo{}
			}
			task.Status.SchedulingInfo.SchedulingAttempts = 1
			task.Status.SchedulingInfo.AssignedNode = ""
			task.Status.SchedulingInfo.AssignedTime = nil
			meta := metav1.Condition{
				Type:               "PodCreated",
				Status:             metav1.ConditionTrue,
				Reason:             "Created",
				Message:            "Pod created for FlightTask",
				ObservedGeneration: task.Generation,
			}
			apimeta.SetStatusCondition(&task.Status.Conditions, meta)
			if err := r.Status().Patch(ctx, &task, patch); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}

		original := task.DeepCopy()

		samePod := task.Status.PodRef != nil && task.Status.PodRef.UID != "" && string(task.Status.PodRef.UID) == string(pod.UID)
		if !samePod {
			task.Status.SchedulingInfo = &airforcev1alpha1.SchedulingInfo{SchedulingAttempts: 1}
		}

		desiredPhase := task.Status.Phase
		podScheduled := pod.Spec.NodeName != "" || isPodScheduled(&pod)
		pullReason, pullMessage, pullFailed := imagePullFailure(&pod)
		switch pod.Status.Phase {
		case corev1.PodRunning:
			desiredPhase = airforcev1alpha1.FlightTaskPhaseRunning
		case corev1.PodSucceeded:
			desiredPhase = airforcev1alpha1.FlightTaskPhaseSucceeded
		case corev1.PodFailed:
			desiredPhase = airforcev1alpha1.FlightTaskPhaseFailed
		case corev1.PodPending:
			if podScheduled {
				desiredPhase = airforcev1alpha1.FlightTaskPhaseScheduled
			} else {
				desiredPhase = airforcev1alpha1.FlightTaskPhasePending
			}
		default:
			// keep
		}
		if pullFailed && desiredPhase != airforcev1alpha1.FlightTaskPhaseSucceeded && desiredPhase != airforcev1alpha1.FlightTaskPhaseFailed {
			// 拉镜像失败时维持 Scheduled（或 Pending）以便人工处理/重试，而不是误判为运行中
			desiredPhase = airforcev1alpha1.FlightTaskPhaseScheduled
		}

		summary, summaryErr := r.summarizeFailedScheduling(ctx, &task, &pod)
		if summaryErr != nil {
			logger.V(1).Info("failed to summarize FailedScheduling events", "error", summaryErr)
		}
		desiredAssignedNode := ""
		var desiredAssignedTime *metav1.Time
		if pod.Spec.NodeName != "" {
			desiredAssignedNode = pod.Spec.NodeName
			desiredAssignedTime = podScheduledTime(&pod)
			if desiredAssignedTime == nil && pod.Status.StartTime != nil {
				desiredAssignedTime = pod.Status.StartTime.DeepCopy()
			}
		}

		podScheduledConditionChanged := syncPodScheduledCondition(&task, &pod)
		failedSchedulingConditionChanged := syncFailedSchedulingCondition(&task, &pod, summary)
		podCreatedConditionChanged := ensurePodCreatedCondition(&task, &pod)
		imagePullConditionChanged := syncImagePullFailedCondition(&task, pullFailed, pullReason, pullMessage)

		desiredAttempts := int32(1)
		if samePod && task.Status.SchedulingInfo != nil && task.Status.SchedulingInfo.SchedulingAttempts > desiredAttempts {
			desiredAttempts = task.Status.SchedulingInfo.SchedulingAttempts
		}
		if summary != nil && summary.Attempts > desiredAttempts {
			desiredAttempts = summary.Attempts
		}

		needsPatch := task.Status.PodRef == nil ||
			task.Status.PodRef.Name != pod.Name ||
			task.Status.PodRef.UID == "" ||
			(task.Status.PodRef.UID != "" && task.Status.PodRef.UID != pod.UID) ||
			task.Status.Phase != desiredPhase ||
			podScheduledConditionChanged ||
			failedSchedulingConditionChanged ||
			podCreatedConditionChanged ||
			imagePullConditionChanged
		if task.Status.SchedulingInfo == nil ||
			task.Status.SchedulingInfo.SchedulingAttempts != desiredAttempts ||
			task.Status.SchedulingInfo.AssignedNode != desiredAssignedNode ||
			!timesEqual(task.Status.SchedulingInfo.AssignedTime, desiredAssignedTime) {
			needsPatch = true
		}
		if needsPatch {
			podRef := podReference(&pod)
			if podRef == nil || podRef.UID == "" {
				return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
			}

			patch := client.MergeFrom(original)
			task.Status.Phase = desiredPhase
			task.Status.PodRef = podRef
			if task.Status.SchedulingInfo == nil {
				task.Status.SchedulingInfo = &airforcev1alpha1.SchedulingInfo{}
			}
			task.Status.SchedulingInfo.SchedulingAttempts = desiredAttempts
			task.Status.SchedulingInfo.AssignedNode = desiredAssignedNode
			task.Status.SchedulingInfo.AssignedTime = desiredAssignedTime
			if err := r.Status().Patch(ctx, &task, patch); err != nil {
				return ctrl.Result{}, err
			}
		}

		if pod.Status.Phase == corev1.PodPending && pod.Spec.NodeName == "" {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
	}

	if task.Status.PodRef != nil &&
		task.Status.Phase != airforcev1alpha1.FlightTaskPhaseSucceeded &&
		task.Status.Phase != airforcev1alpha1.FlightTaskPhaseFailed {
		// Periodically resync while the task is still active to catch Pod status transitions
		// even if a watch event is missed.
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

func isStandaloneFlightTask(task *airforcev1alpha1.FlightTask) bool {
	if task.Spec.StageRef.Name != "" {
		return false
	}
	for _, owner := range task.OwnerReferences {
		if owner.Kind == "MissionStage" && strings.HasPrefix(owner.APIVersion, "airforce.airforce.mil/") {
			return false
		}
	}
	return true
}

func podReference(pod *corev1.Pod) *corev1.ObjectReference {
	if pod == nil {
		return nil
	}
	return &corev1.ObjectReference{
		APIVersion: "v1",
		Kind:       "Pod",
		Namespace:  pod.Namespace,
		Name:       pod.Name,
		UID:        pod.UID,
	}
}

func ensurePodCreatedCondition(task *airforcev1alpha1.FlightTask, pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}
	existing := apimeta.FindStatusCondition(task.Status.Conditions, "PodCreated")
	if existing != nil {
		return false
	}
	cond := metav1.Condition{
		Type:               "PodCreated",
		Status:             metav1.ConditionTrue,
		Reason:             "Created",
		Message:            "Pod created for FlightTask",
		ObservedGeneration: task.Generation,
	}
	return setConditionWithTime(&task.Status.Conditions, cond, metav1.Now())
}

func (r *FlightTaskReconciler) buildPodForTask(ctx context.Context, task *airforcev1alpha1.FlightTask, podName string) (*corev1.Pod, error) {
	labels := map[string]string{
		"flighttask": task.Name,
	}
	if task.Labels != nil {
		if v := task.Labels["mission"]; v != "" {
			labels["mission"] = v
		}
		if v := task.Labels["stage"]; v != "" {
			labels["stage"] = v
		}
	}
	if task.Spec.StageRef.Name != "" {
		labels["stageRef"] = task.Spec.StageRef.Name
	}
	if task.Spec.Role != "" {
		labels["role"] = task.Spec.Role
	}
	if task.Spec.AircraftRequirement.Type != "" {
		labels["aircraft"] = task.Spec.AircraftRequirement.Type
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: task.Namespace,
			Name:      podName,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "task",
					Image:   "busybox:1.36",
					Command: []string{"sh", "-c", "echo flighttask-start && sleep 3600"},
				},
			},
		},
	}

	if task.Spec.PodTemplate != nil && len(task.Spec.PodTemplate.Raw) != 0 {
		var tmpl corev1.PodTemplateSpec
		if err := json.Unmarshal(task.Spec.PodTemplate.Raw, &tmpl); err != nil {
			return nil, fmt.Errorf("invalid spec.podTemplate: %w", err)
		}
		if len(tmpl.Spec.Containers) == 0 {
			return nil, fmt.Errorf("invalid spec.podTemplate: spec.containers is required")
		}

		pod.Spec = tmpl.Spec
		pod.Spec.RestartPolicy = corev1.RestartPolicyNever

		if pod.Labels == nil {
			pod.Labels = map[string]string{}
		}
		for k, v := range tmpl.Labels {
			pod.Labels[k] = v
		}
		if pod.Annotations == nil {
			pod.Annotations = map[string]string{}
		}
		for k, v := range tmpl.Annotations {
			pod.Annotations[k] = v
		}
	}

	applyAircraftSchedulingConstraints(pod, task.Spec.AircraftRequirement)

	if err := r.injectWeaponSidecars(ctx, pod, task); err != nil {
		return nil, err
	}

	return pod, nil
}

func (r *FlightTaskReconciler) reader() client.Reader {
	if r.APIReader != nil {
		return r.APIReader
	}
	return r.Client
}

type failedSchedulingSummary struct {
	Attempts      int32
	LastReason    string
	LastMessage   string
	LastEventTime *metav1.Time
}

func (r *FlightTaskReconciler) summarizeFailedScheduling(ctx context.Context, task *airforcev1alpha1.FlightTask, pod *corev1.Pod) (*failedSchedulingSummary, error) {
	summary := &failedSchedulingSummary{}
	seen := map[string]int32{}
	podUID := string(pod.UID)
	if podUID == "" && task.Status.PodRef != nil {
		podUID = string(task.Status.PodRef.UID)
	}
	if podUID == "" {
		return summary, nil
	}

	record := func(uid string, count int32, reason, message string, ts metav1.Time, created metav1.Time) {
		if isSkipDeletingPodMessage(message) {
			return
		}
		if count <= 0 {
			count = 1
		}
		if uid != "" {
			if prev, ok := seen[uid]; ok {
				if count > prev {
					summary.Attempts += count - prev
					seen[uid] = count
				}
			} else {
				seen[uid] = count
				summary.Attempts += count
			}
		} else {
			summary.Attempts += count
		}
		if summary.LastEventTime == nil || ts.After(summary.LastEventTime.Time) {
			summary.LastEventTime = &ts
			summary.LastReason = reason
			summary.LastMessage = message
		}
	}

	var firstErr error

	{
		var list corev1.EventList
		selector := fields.AndSelectors(fields.OneTermEqualSelector("involvedObject.name", pod.Name), fields.OneTermEqualSelector("reason", "FailedScheduling"))
		if err := r.reader().List(ctx, &list, &client.ListOptions{
			Namespace:     pod.Namespace,
			FieldSelector: selector,
		}); err != nil {
			firstErr = err
		} else {
			for i := range list.Items {
				ev := list.Items[i]
				if string(ev.InvolvedObject.UID) != podUID {
					continue
				}
				count := int32(ev.Count)
				ts := eventTimeForCoreEvent(&ev)
				record(string(ev.UID), count, ev.Reason, ev.Message, ts, ev.CreationTimestamp)
			}
		}
	}

	{
		var list eventsv1.EventList
		selector := fields.AndSelectors(fields.OneTermEqualSelector("regarding.name", pod.Name), fields.OneTermEqualSelector("reason", "FailedScheduling"))
		if err := r.reader().List(ctx, &list, &client.ListOptions{
			Namespace:     pod.Namespace,
			FieldSelector: selector,
		}); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		} else {
			for i := range list.Items {
				ev := list.Items[i]
				if string(ev.Regarding.UID) != podUID {
					continue
				}
				count := int32(1) // default to a single occurrence
				if ev.Series != nil && ev.Series.Count > 0 {
					count = ev.Series.Count
				} else if ev.DeprecatedCount > 0 {
					count = ev.DeprecatedCount
				}
				ts := eventTimeForEventsV1(&ev)
				record(string(ev.UID), count, ev.Reason, eventsV1Message(&ev), ts, ev.CreationTimestamp)
			}
		}
	}

	if summary.Attempts == 0 && firstErr != nil {
		return summary, firstErr
	}
	return summary, nil
}

func eventsV1Message(e *eventsv1.Event) string {
	if e == nil {
		return ""
	}
	if e.Note != "" {
		return e.Note
	}
	return ""
}

func withinWindow(eventTime, since metav1.Time) bool {
	if since.IsZero() {
		return true
	}
	if eventTime.IsZero() {
		return true
	}
	return !eventTime.Time.Before(since.Time)
}

func isSkipDeletingPodMessage(msg string) bool {
	msg = strings.TrimSpace(msg)
	return strings.HasPrefix(msg, "skip schedule deleting pod:")
}

func eventTimeForCoreEvent(e *corev1.Event) metav1.Time {
	if e == nil {
		return metav1.Time{}
	}
	if !e.CreationTimestamp.IsZero() {
		return e.CreationTimestamp
	}
	return metav1.Time{}
}

func eventTimeForEventsV1(e *eventsv1.Event) metav1.Time {
	if e == nil {
		return metav1.Time{}
	}
	if !e.CreationTimestamp.IsZero() {
		return e.CreationTimestamp
	}
	return metav1.Time{}
}

func syncPodScheduledCondition(task *airforcev1alpha1.FlightTask, pod *corev1.Pod) bool {
	podSched := findPodCondition(pod.Status.Conditions, corev1.PodScheduled)
	if podSched == nil {
		return false
	}

	cond := metav1.Condition{
		Type:               "PodScheduled",
		ObservedGeneration: task.Generation,
		Reason:             podSched.Reason,
		Message:            podSched.Message,
	}
	if cond.Reason == "" {
		cond.Reason = "PodScheduled"
	}
	switch podSched.Status {
	case corev1.ConditionTrue:
		cond.Status = metav1.ConditionTrue
	case corev1.ConditionFalse:
		cond.Status = metav1.ConditionFalse
	default:
		cond.Status = metav1.ConditionUnknown
	}
	return setConditionWithTime(&task.Status.Conditions, cond, podSched.LastTransitionTime)
}

func syncFailedSchedulingCondition(task *airforcev1alpha1.FlightTask, pod *corev1.Pod, summary *failedSchedulingSummary) bool {
	if summary == nil {
		summary = &failedSchedulingSummary{}
	}

	// Fallback: if we couldn't pull events but PodScheduled is False, surface that as a FailedScheduling condition.
	if summary.Attempts == 0 && pod != nil {
		if podSched := findPodCondition(pod.Status.Conditions, corev1.PodScheduled); podSched != nil && podSched.Status == corev1.ConditionFalse {
			summary.Attempts = 1
			if summary.LastReason == "" {
				summary.LastReason = podSched.Reason
			}
			if summary.LastMessage == "" {
				summary.LastMessage = podSched.Message
			}
			summary.LastEventTime = &podSched.LastTransitionTime
		}
	}

	cond := metav1.Condition{
		Type:               "FailedScheduling",
		ObservedGeneration: task.Generation,
	}
	if summary.Attempts > 0 {
		cond.Status = metav1.ConditionTrue
		cond.Reason = summary.LastReason
		if cond.Reason == "" {
			cond.Reason = "FailedScheduling"
		}
		cond.Message = summary.LastMessage
		ts := metav1.Now()
		if summary.LastEventTime != nil {
			ts = *summary.LastEventTime
		}
		return setConditionWithTime(&task.Status.Conditions, cond, ts)
	}

	cond.Status = metav1.ConditionFalse
	cond.Reason = "NoFailedSchedulingEvents"
	cond.Message = "No FailedScheduling events observed"
	return setConditionWithTime(&task.Status.Conditions, cond, metav1.Now())
}

func findPodCondition(conditions []corev1.PodCondition, t corev1.PodConditionType) *corev1.PodCondition {
	for i := range conditions {
		if conditions[i].Type == t {
			return &conditions[i]
		}
	}
	return nil
}

func isPodScheduled(pod *corev1.Pod) bool {
	c := findPodCondition(pod.Status.Conditions, corev1.PodScheduled)
	return c != nil && c.Status == corev1.ConditionTrue
}

func imagePullFailure(pod *corev1.Pod) (string, string, bool) {
	check := func(statuses []corev1.ContainerStatus) (string, string, bool) {
		for i := range statuses {
			cs := statuses[i]
			if cs.State.Waiting == nil {
				continue
			}
			reason := cs.State.Waiting.Reason
			switch reason {
			case "ErrImagePull", "ImagePullBackOff", "RegistryUnavailable", "InvalidImageName":
				msg := cs.State.Waiting.Message
				if msg == "" {
					msg = "pod image pull failed"
				}
				if reason == "" {
					reason = "ImagePullFailed"
				}
				return reason, msg, true
			}
		}
		return "", "", false
	}

	if reason, msg, failed := check(pod.Status.ContainerStatuses); failed {
		return reason, msg, true
	}
	if reason, msg, failed := check(pod.Status.InitContainerStatuses); failed {
		return reason, msg, true
	}
	return "", "", false
}

func syncImagePullFailedCondition(task *airforcev1alpha1.FlightTask, failed bool, reason, message string) bool {
	existing := apimeta.FindStatusCondition(task.Status.Conditions, "ImagePullFailed")
	cond := metav1.Condition{
		Type:               "ImagePullFailed",
		ObservedGeneration: task.Generation,
	}
	var transitionTime metav1.Time
	if existing != nil {
		transitionTime = existing.LastTransitionTime
	}
	if failed {
		cond.Status = metav1.ConditionTrue
		cond.Reason = reason
		if cond.Reason == "" {
			cond.Reason = "ImagePullFailed"
		}
		if message != "" {
			cond.Message = message
		} else {
			cond.Message = "pod image pull failed"
		}
		if existing == nil || existing.Status != cond.Status || existing.Reason != cond.Reason || existing.Message != cond.Message {
			transitionTime = metav1.Now()
		}
		return setConditionWithTime(&task.Status.Conditions, cond, transitionTime)
	}

	cond.Status = metav1.ConditionFalse
	cond.Reason = "NoImagePullError"
	cond.Message = "no image pull errors observed"
	if existing == nil || existing.Status != cond.Status || existing.Reason != cond.Reason || existing.Message != cond.Message {
		transitionTime = metav1.Now()
	}
	return setConditionWithTime(&task.Status.Conditions, cond, transitionTime)
}

func podScheduledTime(pod *corev1.Pod) *metav1.Time {
	c := findPodCondition(pod.Status.Conditions, corev1.PodScheduled)
	if c == nil {
		return nil
	}
	if c.LastTransitionTime.IsZero() {
		return nil
	}
	return c.LastTransitionTime.DeepCopy()
}

func timesEqual(a, b *metav1.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Time.Equal(b.Time)
}

func setConditionWithTime(conditions *[]metav1.Condition, cond metav1.Condition, transitionTime metav1.Time) bool {
	cond.LastTransitionTime = transitionTime
	for i := range *conditions {
		existing := &(*conditions)[i]
		if existing.Type != cond.Type {
			continue
		}
		(*conditions)[i] = cond
		return !(existing.Status == cond.Status &&
			existing.Reason == cond.Reason &&
			existing.Message == cond.Message &&
			existing.ObservedGeneration == cond.ObservedGeneration &&
			existing.LastTransitionTime.Time.Equal(cond.LastTransitionTime.Time))
	}
	*conditions = append(*conditions, cond)
	return true
}

func applyAircraftSchedulingConstraints(pod *corev1.Pod, req airforcev1alpha1.AircraftRequirement) {
	aircraftType := strings.TrimSpace(req.Type)
	if aircraftType != "" {
		if pod.Spec.NodeSelector == nil {
			pod.Spec.NodeSelector = map[string]string{}
		}
		if pod.Spec.NodeSelector["aircraft.mil/type"] == "" && pod.Spec.NodeSelector["aircraft.type"] == "" {
			pod.Spec.NodeSelector["aircraft.mil/type"] = aircraftType
			if pod.Spec.NodeSelector["aircraft.mil/status"] == "" && pod.Spec.NodeSelector["aircraft.status"] == "" {
				pod.Spec.NodeSelector["aircraft.mil/status"] = "ready"
			}
		}
	}

	var required []corev1.NodeSelectorRequirement
	if req.MinFuelLevel > 0 {
		minFuel := int(req.MinFuelLevel) - 1
		if minFuel < 0 {
			minFuel = 0
		}
		required = append(required, corev1.NodeSelectorRequirement{
			Key:      "aircraft.mil/fuel.level",
			Operator: corev1.NodeSelectorOpGt,
			Values:   []string{strconv.Itoa(minFuel)},
		})
	}
	if req.RequiredHardpoints > 0 {
		minHardpoints := int(req.RequiredHardpoints) - 1
		if minHardpoints < 0 {
			minHardpoints = 0
		}
		required = append(required, corev1.NodeSelectorRequirement{
			Key:      "aircraft.mil/hardpoint.available",
			Operator: corev1.NodeSelectorOpGt,
			Values:   []string{strconv.Itoa(minHardpoints)},
		})
	}
	for _, capName := range req.Capabilities {
		capName = strings.TrimSpace(capName)
		if capName == "" {
			continue
		}
		required = append(required, corev1.NodeSelectorRequirement{
			Key:      "aircraft.mil/capability." + sanitizeLabelKeySuffix(capName),
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{"true"},
		})
	}

	preferredLocation := strings.TrimSpace(req.PreferredLocation)
	var preferred []corev1.PreferredSchedulingTerm
	if preferredLocation != "" {
		preferred = append(preferred, corev1.PreferredSchedulingTerm{
			Weight: 50,
			Preference: corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      "aircraft.mil/location.zone",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{preferredLocation},
					},
				},
			},
		})
	}

	if len(required) == 0 && len(preferred) == 0 {
		return
	}

	if pod.Spec.Affinity != nil && pod.Spec.Affinity.NodeAffinity != nil {
		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil ||
			len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 0 {
			return
		}
	}

	if pod.Spec.Affinity == nil {
		pod.Spec.Affinity = &corev1.Affinity{}
	}
	if pod.Spec.Affinity.NodeAffinity == nil {
		pod.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}
	if len(required) != 0 && pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{MatchExpressions: required},
			},
		}
	}
	if len(preferred) != 0 && len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
		pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = preferred
	}
}

func (r *FlightTaskReconciler) injectWeaponSidecars(ctx context.Context, pod *corev1.Pod, task *airforcev1alpha1.FlightTask) error {
	if len(task.Spec.WeaponLoadout) == 0 {
		return nil
	}

	aircraftType := strings.TrimSpace(task.Spec.AircraftRequirement.Type)
	for i := range task.Spec.WeaponLoadout {
		item := task.Spec.WeaponLoadout[i]
		weaponName := strings.TrimSpace(item.WeaponRef.Name)
		if weaponName == "" {
			return fmt.Errorf("invalid spec.weaponLoadout[%d]: weaponRef.name is required", i)
		}

		var weapon airforcev1alpha1.Weapon
		if err := r.Get(ctx, client.ObjectKey{Namespace: task.Namespace, Name: weaponName}, &weapon); err != nil {
			return fmt.Errorf("weapon %q not found: %w", weaponName, err)
		}

		image := ""
		if weapon.Spec.Image != nil {
			repo := strings.TrimSpace(weapon.Spec.Image.Repository)
			tag := strings.TrimSpace(weapon.Spec.Image.Tag)
			if repo != "" && tag != "" && !strings.Contains(repo, ":") {
				image = repo + ":" + tag
			} else {
				image = repo
			}
		}
		if image == "" {
			return fmt.Errorf("weapon %q missing spec.image.repository", weaponName)
		}

		if weapon.Spec.Compatibility != nil && len(weapon.Spec.Compatibility.AircraftTypes) != 0 && aircraftType != "" {
			if !containsString(weapon.Spec.Compatibility.AircraftTypes, aircraftType) {
				return fmt.Errorf("weapon %q is not compatible with aircraft type %q", weaponName, aircraftType)
			}
		}

		if weapon.Spec.Compatibility != nil && len(weapon.Spec.Compatibility.HardpointTypes) != 0 {
			for j, mp := range item.MountPoints {
				mp = strings.TrimSpace(mp)
				if mp == "" {
					continue
				}
				if !containsString(weapon.Spec.Compatibility.HardpointTypes, mp) {
					return fmt.Errorf("weapon %q is not compatible with mountPoint %q (index %d)", weaponName, mp, j)
				}
			}
		}

		baseName := sanitizeDNSLabel("weapon-" + weaponName)
		containerName := uniqueContainerName(pod.Spec.Containers, baseName)

		env := []corev1.EnvVar{
			{Name: "WEAPON_NAME", Value: weaponName},
			{Name: "WEAPON_TYPE", Value: weapon.Spec.WeaponType},
			{Name: "QUANTITY", Value: strconv.FormatInt(int64(item.Quantity), 10)},
			{Name: "MOUNT_POINTS", Value: strings.Join(item.MountPoints, ",")},
			{Name: "AIRCRAFT_TYPE", Value: aircraftType},
		}

		volumeMounts := []corev1.VolumeMount{{Name: "weapon-interface", MountPath: "/interface"}}
		if weapon.Spec.Container != nil {
			env = append(env, weapon.Spec.Container.Env...)
			volumeMounts = append(volumeMounts, weapon.Spec.Container.VolumeMounts...)
		}

		sidecar := corev1.Container{
			Name:         containerName,
			Image:        image,
			Env:          env,
			VolumeMounts: volumeMounts,
		}
		if weapon.Spec.Image != nil && weapon.Spec.Image.PullPolicy != "" {
			sidecar.ImagePullPolicy = weapon.Spec.Image.PullPolicy
		}
		if weapon.Spec.Container != nil {
			if len(weapon.Spec.Container.Ports) != 0 {
				sidecar.Ports = append(sidecar.Ports, weapon.Spec.Container.Ports...)
			}
			if weapon.Spec.Container.LivenessProbe != nil {
				sidecar.LivenessProbe = weapon.Spec.Container.LivenessProbe.DeepCopy()
			}
		}

		pod.Spec.Containers = append(pod.Spec.Containers, sidecar)
	}

	ensureEmptyDirVolume(&pod.Spec, "weapon-interface")
	ensureVolumeMountAllContainers(&pod.Spec, corev1.VolumeMount{Name: "weapon-interface", MountPath: "/interface"})
	return nil
}

func containsString(values []string, needle string) bool {
	for _, v := range values {
		if strings.EqualFold(strings.TrimSpace(v), needle) {
			return true
		}
	}
	return false
}

func sanitizeDNSLabel(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "x"
	}
	var b strings.Builder
	b.Grow(len(s))
	prevDash := false
	for _, r := range s {
		isAlnum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlnum {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "x"
	}
	if len(out) > 63 {
		out = out[:63]
		out = strings.TrimRight(out, "-")
		if out == "" {
			out = "x"
		}
	}
	return out
}

func sanitizeLabelKeySuffix(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "x"
	}
	var b strings.Builder
	b.Grow(len(s))
	lastWasSep := false
	for _, r := range s {
		isAllowed := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.'
		if isAllowed {
			b.WriteRune(r)
			lastWasSep = false
			continue
		}
		if !lastWasSep {
			b.WriteByte('-')
			lastWasSep = true
		}
	}
	out := strings.Trim(b.String(), "-_.")
	if out == "" {
		out = "x"
	}
	if len(out) > 63 {
		out = out[:63]
		out = strings.TrimRight(out, "-_.")
		if out == "" {
			out = "x"
		}
	}
	return out
}

func uniqueContainerName(existing []corev1.Container, base string) string {
	name := base
	for i := 1; ; i++ {
		if !containerNameExists(existing, name) {
			return name
		}
		suffix := "-" + strconv.Itoa(i)
		trunc := base
		if len(trunc)+len(suffix) > 63 {
			trunc = trunc[:63-len(suffix)]
			trunc = strings.TrimRight(trunc, "-")
		}
		name = trunc + suffix
	}
}

func containerNameExists(existing []corev1.Container, name string) bool {
	for _, c := range existing {
		if c.Name == name {
			return true
		}
	}
	return false
}

func ensureEmptyDirVolume(podSpec *corev1.PodSpec, name string) {
	for i := range podSpec.Volumes {
		if podSpec.Volumes[i].Name == name {
			return
		}
	}
	podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})
}

func ensureVolumeMountAllContainers(podSpec *corev1.PodSpec, mount corev1.VolumeMount) {
	for ci := range podSpec.Containers {
		if hasVolumeMount(podSpec.Containers[ci].VolumeMounts, mount.Name, mount.MountPath) {
			continue
		}
		podSpec.Containers[ci].VolumeMounts = append(podSpec.Containers[ci].VolumeMounts, mount)
	}
}

func hasVolumeMount(mounts []corev1.VolumeMount, name, mountPath string) bool {
	for _, m := range mounts {
		if m.Name == name && m.MountPath == mountPath {
			return true
		}
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlightTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&airforcev1alpha1.FlightTask{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
