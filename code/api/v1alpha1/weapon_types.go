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

type WeaponCoolingLevel string

const (
	WeaponCoolingLevelLow    WeaponCoolingLevel = "low"
	WeaponCoolingLevelMedium WeaponCoolingLevel = "medium"
	WeaponCoolingLevelHigh   WeaponCoolingLevel = "high"
)

type WeaponSpecImage struct {
	Repository string            `json:"repository,omitempty"`
	Tag        string            `json:"tag,omitempty"`
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

type WeaponSpecifications struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	Weight       string `json:"weight,omitempty"`
	Length       string `json:"length,omitempty"`
	Diameter     string `json:"diameter,omitempty"`
	Range        string `json:"range,omitempty"`
	Speed        string `json:"speed,omitempty"`
	Guidance     string `json:"guidance,omitempty"`
	Warhead      string `json:"warhead,omitempty"`
}

type WeaponResources struct {
	Hardpoints int32              `json:"hardpoints,omitempty"`
	Weight     int32              `json:"weight,omitempty"`
	Power      int32              `json:"power,omitempty"`
	Cooling    WeaponCoolingLevel `json:"cooling,omitempty"`
}

type WeaponCompatibility struct {
	AircraftTypes  []string `json:"aircraftTypes,omitempty"`
	HardpointTypes []string `json:"hardpointTypes,omitempty"`
}

type WeaponContainerSpec struct {
	Env           []corev1.EnvVar        `json:"env,omitempty"`
	VolumeMounts  []corev1.VolumeMount   `json:"volumeMounts,omitempty"`
	Ports         []corev1.ContainerPort `json:"ports,omitempty"`
	LivenessProbe *corev1.Probe          `json:"livenessProbe,omitempty"`
}

type WeaponVersion struct {
	Current     string       `json:"current,omitempty"`
	Changelog   string       `json:"changelog,omitempty"`
	ReleaseDate *metav1.Time `json:"releaseDate,omitempty"`
}

// WeaponSpec defines the desired state of Weapon
type WeaponSpec struct {
	WeaponName string `json:"weaponName,omitempty"`
	WeaponType string `json:"weaponType,omitempty"`
	Category   string `json:"category,omitempty"`

	Specifications *WeaponSpecifications `json:"specifications,omitempty"`
	Image          *WeaponSpecImage      `json:"image,omitempty"`
	Resources      *WeaponResources      `json:"resources,omitempty"`
	Compatibility  *WeaponCompatibility  `json:"compatibility,omitempty"`
	Container      *WeaponContainerSpec  `json:"container,omitempty"`
	Version        *WeaponVersion        `json:"version,omitempty"`
}

type WeaponPhase string

const (
	WeaponPhaseAvailable  WeaponPhase = "Available"
	WeaponPhaseUpdating   WeaponPhase = "Updating"
	WeaponPhaseDeprecated WeaponPhase = "Deprecated"
)

type WeaponUsage struct {
	TotalDeployed int32  `json:"totalDeployed,omitempty"`
	TotalFired    int32  `json:"totalFired,omitempty"`
	SuccessRate   string `json:"successRate,omitempty"`
}

type WeaponCompatibilityCheck struct {
	AircraftType string       `json:"aircraftType,omitempty"`
	Compatible   bool         `json:"compatible,omitempty"`
	LastChecked  *metav1.Time `json:"lastChecked,omitempty"`
}

// WeaponStatus defines the observed state of Weapon
type WeaponStatus struct {
	// +kubebuilder:validation:Enum=Available;Updating;Deprecated
	Phase WeaponPhase `json:"phase,omitempty"`

	Usage               *WeaponUsage               `json:"usage,omitempty"`
	CompatibilityChecks []WeaponCompatibilityCheck `json:"compatibilityChecks,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=".status.phase",description="Status phase"

// Weapon is the Schema for the weapons API
type Weapon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WeaponSpec   `json:"spec,omitempty"`
	Status WeaponStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WeaponList contains a list of Weapon
type WeaponList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Weapon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Weapon{}, &WeaponList{})
}
