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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
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

	if task.Status.Phase == airforcev1alpha1.FlightTaskPhasePending && isStandaloneFlightTask(&task) {
		patch := client.MergeFrom(task.DeepCopy())
		task.Status.Phase = airforcev1alpha1.FlightTaskPhaseScheduled
		if err := r.Status().Patch(ctx, &task, patch); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if task.Status.Phase == airforcev1alpha1.FlightTaskPhaseScheduled || task.Status.Phase == airforcev1alpha1.FlightTaskPhaseRunning {
		podName := fmt.Sprintf("%s-pod", task.Name)

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

	if err := r.injectWeaponSidecars(ctx, pod, task); err != nil {
		return nil, err
	}

	return pod, nil
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
