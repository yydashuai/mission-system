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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
}

//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=flighttasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=flighttasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=flighttasks/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete

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

	if task.Status.Phase == airforcev1alpha1.FlightTaskPhaseScheduled || task.Status.Phase == airforcev1alpha1.FlightTaskPhaseRunning {
		podName := fmt.Sprintf("%s-pod", task.Name)

		var pod corev1.Pod
		err := r.Get(ctx, client.ObjectKey{Namespace: task.Namespace, Name: podName}, &pod)
		if err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		if apierrors.IsNotFound(err) {
			desiredPod, err := r.buildPodForTask(&task, podName)
			if err != nil {
				logger.Error(err, "failed to build pod for FlightTask", "flightTask", task.Name)
				return ctrl.Result{}, err
			}
			if err := controllerutil.SetControllerReference(&task, desiredPod, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.Create(ctx, desiredPod); err != nil {
				return ctrl.Result{}, err
			}

			patch := client.MergeFrom(task.DeepCopy())
			task.Status.PodRef = &corev1.ObjectReference{
				APIVersion: "v1",
				Kind:       "Pod",
				Namespace:  task.Namespace,
				Name:       desiredPod.Name,
			}
			task.Status.Phase = airforcev1alpha1.FlightTaskPhaseRunning
			if err := r.Status().Patch(ctx, &task, patch); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}

		desiredPhase := task.Status.Phase
		switch pod.Status.Phase {
		case corev1.PodRunning:
			desiredPhase = airforcev1alpha1.FlightTaskPhaseRunning
		case corev1.PodSucceeded:
			desiredPhase = airforcev1alpha1.FlightTaskPhaseSucceeded
		case corev1.PodFailed:
			desiredPhase = airforcev1alpha1.FlightTaskPhaseFailed
		case corev1.PodPending:
			if desiredPhase == "" || desiredPhase == airforcev1alpha1.FlightTaskPhasePending {
				desiredPhase = airforcev1alpha1.FlightTaskPhaseScheduled
			}
		default:
			// keep
		}

		needsPatch := task.Status.PodRef == nil ||
			task.Status.PodRef.Name != pod.Name ||
			task.Status.Phase != desiredPhase
		if needsPatch {
			patch := client.MergeFrom(task.DeepCopy())
			task.Status.Phase = desiredPhase
			if task.Status.PodRef == nil {
				task.Status.PodRef = &corev1.ObjectReference{
					APIVersion: "v1",
					Kind:       "Pod",
					Namespace:  pod.Namespace,
					Name:       pod.Name,
				}
			}
			if err := r.Status().Patch(ctx, &task, patch); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *FlightTaskReconciler) buildPodForTask(task *airforcev1alpha1.FlightTask, podName string) (*corev1.Pod, error) {
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

	if task.Spec.PodTemplate == nil {
		return pod, nil
	}

	if len(task.Spec.PodTemplate.Raw) == 0 {
		return pod, nil
	}

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

	return pod, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlightTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&airforcev1alpha1.FlightTask{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
