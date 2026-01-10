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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type MissionStageRef struct {
	Name string `json:"name,omitempty"`
}

type AircraftRequirement struct {
	Type               string   `json:"type,omitempty"`
	MinFuelLevel       int32    `json:"minFuelLevel,omitempty"`
	Capabilities       []string `json:"capabilities,omitempty"`
	RequiredHardpoints int32    `json:"requiredHardpoints,omitempty"`
	PreferredLocation  string   `json:"preferredLocation,omitempty"`
}

type OperationArea struct {
	Center GeoCoordinates `json:"center,omitempty"`
	Radius string         `json:"radius,omitempty"`
}

type TaskPhase struct {
	Name      string           `json:"name,omitempty"`
	Duration  *metav1.Duration `json:"duration,omitempty"`
	Waypoints []string         `json:"waypoints,omitempty"`
	Tactics   string           `json:"tactics,omitempty"`
}

type FlightTaskParams struct {
	Altitude        string            `json:"altitude,omitempty"`
	Speed           string            `json:"speed,omitempty"`
	MissionDuration *metav1.Duration  `json:"missionDuration,omitempty"`
	OperationArea   *OperationArea    `json:"operationArea,omitempty"`
	Phases          []TaskPhase       `json:"phases,omitempty"`
	Extra           map[string]string `json:"extra,omitempty"`
}

type WeaponRef struct {
	Name string `json:"name,omitempty"`
}

type FlightTaskWeaponLoadoutItem struct {
	WeaponRef   WeaponRef `json:"weaponRef,omitempty"`
	Quantity    int32     `json:"quantity,omitempty"`
	MountPoints []string  `json:"mountPoints,omitempty"`
}

// FlightTaskSpec defines the desired state of FlightTask
type FlightTaskSpec struct {
	StageRef MissionStageRef `json:"stageRef,omitempty"`

	AircraftRequirement AircraftRequirement `json:"aircraftRequirement,omitempty"`

	Role       string            `json:"role,omitempty"`
	TaskParams *FlightTaskParams `json:"taskParams,omitempty"`

	WeaponLoadout []FlightTaskWeaponLoadoutItem `json:"weaponLoadout,omitempty"`

	PodTemplate *corev1.PodTemplateSpec `json:"podTemplate,omitempty"`
}

type SchedulingInfo struct {
	AssignedNode       string       `json:"assignedNode,omitempty"`
	AssignedTime       *metav1.Time `json:"assignedTime,omitempty"`
	SchedulingAttempts int32        `json:"schedulingAttempts,omitempty"`
}

type ExecutionStatus struct {
	CurrentPhase     string            `json:"currentPhase,omitempty"`
	Location         *GeoCoordinates   `json:"location,omitempty"`
	Altitude         string            `json:"altitude,omitempty"`
	Speed            string            `json:"speed,omitempty"`
	FuelRemaining    int32             `json:"fuelRemaining,omitempty"`
	WeaponsRemaining map[string]int32  `json:"weaponsRemaining,omitempty"`
	Extra            map[string]string `json:"extra,omitempty"`
}

// FlightTaskStatus defines the observed state of FlightTask
type FlightTaskStatus struct {
	// +kubebuilder:validation:Enum=Pending;Scheduled;Running;Succeeded;Failed
	Phase FlightTaskPhase `json:"phase,omitempty"`

	SchedulingInfo  *SchedulingInfo         `json:"schedulingInfo,omitempty"`
	PodRef          *corev1.ObjectReference `json:"podRef,omitempty"`
	ExecutionStatus *ExecutionStatus        `json:"executionStatus,omitempty"`

	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FlightTask is the Schema for the flighttasks API
type FlightTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlightTaskSpec   `json:"spec,omitempty"`
	Status FlightTaskStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FlightTaskList contains a list of FlightTask
type FlightTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlightTask `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FlightTask{}, &FlightTaskList{})
}
