/*
Copyright 2023.

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

// GameSpec defines the desired state of Game
type GameSpec struct {

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	GameName string `json:"gameName"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern:=`^https?:\/\/(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$`
	Url string `json:"url"`

	// +optional
	// +kubebuilder:default:=false
	// +kubebuilder:validation:Type=boolean
	ForceRedeploy bool `json:"forceRedeploy,omitempty"`

	// +optional
	// +kubebuilder:default=8080
	// +kubebuilder:validation:Type=integer
	Port int `json:"port,omitempty"`
}

// GameStatus defines the observed state of Game
type GameStatus struct {
	Ready bool `json:"ready,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Game is the Schema for the games API
// +kubebuilder:printcolumn:name="Game",type=string,JSONPath=`.spec.gameName`
// +kubebuilder:printcolumn:name="Url",type=string,JSONPath=`.spec.Url`
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
type Game struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GameSpec   `json:"spec,omitempty"`
	Status GameStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GameList contains a list of Game
type GameList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Game `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Game{}, &GameList{})
}
