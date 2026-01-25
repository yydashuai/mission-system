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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type MissionRef struct {
	Name string `json:"name,omitempty"`
}

type MissionStagePhase string

const (
	MissionStagePhasePending   MissionStagePhase = "Pending"
	MissionStagePhaseRunning   MissionStagePhase = "Running"
	MissionStagePhaseSucceeded MissionStagePhase = "Succeeded"
	MissionStagePhaseFailed    MissionStagePhase = "Failed"
)

type FlightTaskPhase string

const (
	FlightTaskPhasePending   FlightTaskPhase = "Pending"
	FlightTaskPhaseScheduled FlightTaskPhase = "Scheduled"
	FlightTaskPhaseRunning   FlightTaskPhase = "Running"
	FlightTaskPhaseSucceeded FlightTaskPhase = "Succeeded"
	FlightTaskPhaseFailed    FlightTaskPhase = "Failed"
)

type WeaponLoadoutItem struct {
	Weapon      string   `json:"weapon,omitempty"`
	Quantity    int32    `json:"quantity,omitempty"`
	MountPoints []string `json:"mountPoints,omitempty"`
}

type MissionStageFlightTaskTemplate struct {
	Name          string                `json:"name,omitempty"`
	Aircraft      string                `json:"aircraft,omitempty"`
	Role          string                `json:"role,omitempty"`
	Priority      MissionPriority       `json:"priority,omitempty"`
	WeaponLoadout []WeaponLoadoutItem   `json:"weaponLoadout,omitempty"`
	TaskParams    map[string]string     `json:"taskParams,omitempty"`
	PodTemplate   *runtime.RawExtension `json:"podTemplate,omitempty"`
}

type MissionStageSynchronization struct {
	WaitForAll bool   `json:"waitForAll,omitempty"`
	Checkpoint string `json:"checkpoint,omitempty"`
}

type MissionStageDependencyCondition struct {
	Type string `json:"type,omitempty"`
}

type MissionStageDependencies struct {
	Conditions []MissionStageDependencyCondition `json:"conditions,omitempty"`
}

type MissionStageConfig struct {
	Synchronization *MissionStageSynchronization `json:"synchronization,omitempty"`
	Timeout         *metav1.Duration             `json:"timeout,omitempty"`
	Dependencies    *MissionStageDependencies    `json:"dependencies,omitempty"`
}

// MissionStageSpec defines the desired state of MissionStage
type MissionStageSpec struct {
	MissionRef MissionRef `json:"missionRef,omitempty"`

	StageName  string `json:"stageName,omitempty"`
	StageIndex int32  `json:"stageIndex,omitempty"`

	// +kubebuilder:validation:Enum=sequential;parallel;mixed
	StageType StageExecutionType `json:"stageType,omitempty"`

	DependsOn []string `json:"dependsOn,omitempty"`

	FlightTasks []MissionStageFlightTaskTemplate `json:"flightTasks,omitempty"`

	Config *MissionStageConfig `json:"config,omitempty"`
}

type MissionStageFlightTaskStatus struct {
	Name         string          `json:"name,omitempty"`
	Phase        FlightTaskPhase `json:"phase,omitempty"`
	AircraftNode string          `json:"aircraftNode,omitempty"`
	PodName      string          `json:"podName,omitempty"`
	Message      string          `json:"message,omitempty"`
}

// MissionStageStatus defines the observed state of MissionStage
type MissionStageStatus struct {
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed
	Phase MissionStagePhase `json:"phase,omitempty"`

	FlightTasksStatus []MissionStageFlightTaskStatus `json:"flightTasksStatus,omitempty"`

	StartTime      *metav1.Time `json:"startTime,omitempty"`
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	Message        string       `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=ms
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=".status.phase",description="Status phase"

// MissionStage is the Schema for the missionstages API
type MissionStage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MissionStageSpec   `json:"spec,omitempty"`
	Status MissionStageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MissionStageList contains a list of MissionStage
type MissionStageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MissionStage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MissionStage{}, &MissionStageList{})
}
