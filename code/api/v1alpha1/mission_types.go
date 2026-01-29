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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type MissionType string

const (
	MissionTypeISR    MissionType = "isr"
	MissionTypeStrike MissionType = "strike"
	MissionTypePatrol MissionType = "patrol"
	MissionTypeEscort MissionType = "escort"
)

type MissionPriority string

const (
	MissionPriorityLow      MissionPriority = "low"
	MissionPriorityMedium   MissionPriority = "medium"
	MissionPriorityHigh     MissionPriority = "high"
	MissionPriorityCritical MissionPriority = "critical"
)

type MissionPhase string

const (
	MissionPhasePending   MissionPhase = "待执行"
	MissionPhaseRunning   MissionPhase = "运行中"
	MissionPhaseSucceeded MissionPhase = "已完成"
	MissionPhaseFailed    MissionPhase = "失败"
	MissionPhaseCancelled MissionPhase = "已取消"
)

type StageExecutionType string

const (
	StageExecutionTypeSequential StageExecutionType = "串行"
	StageExecutionTypeParallel   StageExecutionType = "并行"
	StageExecutionTypeMixed      StageExecutionType = "混合"
)

type RetryStrategy string

const (
	RetryStrategyImmediate   RetryStrategy = "immediate"
	RetryStrategyExponential RetryStrategy = "exponential"
	RetryStrategyCustom      RetryStrategy = "custom"
)

type StageFailureAction string

const (
	StageFailureActionAbort    StageFailureAction = "中止"
	StageFailureActionContinue StageFailureAction = "继续"
	StageFailureActionRetry    StageFailureAction = "重试"
)

type MissionObjective struct {
	TargetArea        string            `json:"targetArea,omitempty"`
	TargetCoordinates *GeoCoordinates   `json:"targetCoordinates,omitempty"`
	TargetDescription string            `json:"targetDescription,omitempty"`
	Extra             map[string]string `json:"extra,omitempty"`
}

type GeoCoordinates struct {
	Latitude  string `json:"latitude,omitempty"`
	Longitude string `json:"longitude,omitempty"`
}

type MissionStageTemplate struct {
	Name        string             `json:"name,omitempty"`
	DisplayName string             `json:"displayName,omitempty"`
	Type        StageExecutionType `json:"type,omitempty"`
	DependsOn   []string           `json:"dependsOn,omitempty"`
	Timeout     *metav1.Duration   `json:"timeout,omitempty"`

	FlightTasks []MissionStageFlightTaskTemplate `json:"flightTasks,omitempty"`
}

type MissionConfig struct {
	FailurePolicy      *FailurePolicy      `json:"failurePolicy,omitempty"`
	CancellationPolicy *CancellationPolicy `json:"cancellationPolicy,omitempty"`
	Coordination       *Coordination       `json:"coordination,omitempty"`
}

type FailurePolicy struct {
	MaxRetries         int32              `json:"maxRetries,omitempty"`
	RetryStrategy      RetryStrategy      `json:"retryStrategy,omitempty"`
	StageFailureAction StageFailureAction `json:"stageFailureAction,omitempty"`
}

type CancellationPolicy struct {
	GracePeriod *metav1.Duration `json:"gracePeriod,omitempty"`
	Cleanup     bool             `json:"cleanup,omitempty"`
}

type Coordination struct {
	DataLinkProtocol   string `json:"dataLinkProtocol,omitempty"`
	CommandFrequency   string `json:"commandFrequency,omitempty"`
	EmergencyFrequency string `json:"emergencyFrequency,omitempty"`
}

// MissionSpec defines the desired state of Mission
type MissionSpec struct {
	MissionName string `json:"missionName,omitempty"`

	// +kubebuilder:validation:Enum=isr;strike;patrol;escort
	MissionType MissionType `json:"missionType,omitempty"`

	// +kubebuilder:validation:Enum=low;medium;high;critical
	Priority MissionPriority `json:"priority,omitempty"`

	Objective *MissionObjective `json:"objective,omitempty"`

	Stages []MissionStageTemplate `json:"stages,omitempty"`

	Config *MissionConfig `json:"config,omitempty"`
}

type MissionStageSummary struct {
	Name           string       `json:"name,omitempty"`
	Phase          MissionPhase `json:"phase,omitempty"`
	StartTime      *metav1.Time `json:"startTime,omitempty"`
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

type MissionStatistics struct {
	TotalFlightTasks int32 `json:"totalFlightTasks,omitempty"`
	SucceededTasks   int32 `json:"succeededTasks,omitempty"`
	FailedTasks      int32 `json:"failedTasks,omitempty"`
	RunningTasks     int32 `json:"runningTasks,omitempty"`
	PendingTasks     int32 `json:"pendingTasks,omitempty"`
}

// MissionStatus defines the observed state of Mission
type MissionStatus struct {
	// +kubebuilder:validation:Enum=待执行;运行中;已完成;失败;已取消
	Phase MissionPhase `json:"phase,omitempty"`

	StagesSummary []MissionStageSummary `json:"stagesSummary,omitempty"`
	Statistics    *MissionStatistics    `json:"statistics,omitempty"`

	Message        string       `json:"message,omitempty"`
	StartTime      *metav1.Time `json:"startTime,omitempty"`
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=".status.phase",description="Status phase"

// Mission is the Schema for the missions API
type Mission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MissionSpec   `json:"spec,omitempty"`
	Status MissionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MissionList contains a list of Mission
type MissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mission `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mission{}, &MissionList{})
}
