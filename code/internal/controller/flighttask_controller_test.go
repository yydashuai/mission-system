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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	airforcev1alpha1 "github.com/yydashuai/mission-system/api/v1alpha1"
)

var _ = Describe("FlightTask Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		flighttask := &airforcev1alpha1.FlightTask{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind FlightTask")
			err := k8sClient.Get(ctx, typeNamespacedName, flighttask)
			if err != nil && errors.IsNotFound(err) {
				resource := &airforcev1alpha1.FlightTask{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
						Labels: map[string]string{
							"mission": "m1",
							"stage":   "s1",
						},
					},
					Spec: airforcev1alpha1.FlightTaskSpec{
						StageRef: airforcev1alpha1.MissionStageRef{Name: "s1"},
						AircraftRequirement: airforcev1alpha1.AircraftRequirement{
							Type: "j20",
						},
						Role: "air-superiority",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
				patch := client.MergeFrom(resource.DeepCopy())
				resource.Status.Phase = airforcev1alpha1.FlightTaskPhaseScheduled
				Expect(k8sClient.Status().Patch(ctx, resource, patch)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &airforcev1alpha1.FlightTask{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			pod := &corev1.Pod{}
			_ = k8sClient.Get(ctx, types.NamespacedName{Name: resourceName + "-pod", Namespace: "default"}, pod)
			_ = k8sClient.Delete(ctx, pod)

			By("Cleanup the specific resource instance FlightTask")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &FlightTaskReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			var pod corev1.Pod
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: resourceName + "-pod", Namespace: "default"}, &pod)).To(Succeed())

			var updated airforcev1alpha1.FlightTask
			Expect(k8sClient.Get(ctx, typeNamespacedName, &updated)).To(Succeed())
			Expect(updated.Status.Phase).To(Equal(airforcev1alpha1.FlightTaskPhaseRunning))
			Expect(updated.Status.PodRef).NotTo(BeNil())
			Expect(updated.Status.PodRef.Name).To(Equal(pod.Name))
		})
	})
})
