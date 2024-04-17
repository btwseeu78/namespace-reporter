/*
Copyright 2024.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// +kubebuilder:validation:Enum=Health;Disruption;Checks
type reportingComponents string

const (
	health     reportingComponents = "Health"
	disruption reportingComponents = "Disruption"
	checks     reportingComponents = "Checks"
)

// AuditSpec defines the desired state of Audit
type AuditSpec struct {
	// +optional
	RetentionsPeriod    *int                  `json:"retentionsPeriod,omitempty"`
	Selector            map[string]string     `json:"selector"`
	ReportingComponents []reportingComponents `json:"reportingComponents"`
	Chunks              *int32                `json:"chunks,omitempty"`
	// +optional
	WebhookUrl *string `json:"webhookUrl,omitempty"`
}

// AuditStatus defines the observed state of Audit
type AuditStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Audit is the Schema for the audits API
type Audit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuditSpec   `json:"spec,omitempty"`
	Status AuditStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AuditList contains a list of Audit
type AuditList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Audit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Audit{}, &AuditList{})
}
