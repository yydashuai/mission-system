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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	airforcev1alpha1 "github.com/yydashuai/mission-system/api/v1alpha1"
)

// WeaponReconciler reconciles a Weapon object
type WeaponReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=weapons,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=weapons/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=airforce.airforce.mil,resources=weapons/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Weapon object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *WeaponReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var weapon airforcev1alpha1.Weapon
	if err := r.Get(ctx, req.NamespacedName, &weapon); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if weapon.Status.Phase == "" {
		patch := client.MergeFrom(weapon.DeepCopy())
		weapon.Status.Phase = airforcev1alpha1.WeaponPhaseAvailable
		if err := r.Status().Patch(ctx, &weapon, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WeaponReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&airforcev1alpha1.Weapon{}).
		Complete(r)
}
