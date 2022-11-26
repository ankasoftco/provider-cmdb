/*
Copyright 2022 The Crossplane Authors.

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
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// ItemValue for observation
/*type ItemValue struct {
	Values
}*/

// CIParameters are the configurable fields of Identification and Reconciliation API.
type CIParameters struct {
	SysParamDataSource string            `json:"sysParamDataSource"`
	ClassName          string            `json:"className"`
	Name               string            `json:"name"`
	Values             map[string]string `json:"values,omitempty"`
}

// CIObservation are the observable fields of Identification and Reconciliation API.
type CIObservation struct {
}

// CISpec defines the desired state of Identification and Reconciliation API.
type CISpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CIParameters `json:"forProvider"`
}

// CIStatus represents the observed state of Identification and Reconciliation API.
type CIStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CIObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CI is Identification and Reconciliation API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cmdb}
type CI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CISpec   `json:"spec"`
	Status CIStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CIList contains a list of Identification and Reconciliation API
type CIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CI `json:"items"`
}

// CI type metadata.
var (
	CIKind             = reflect.TypeOf(CI{}).Name()
	CIGroupKind        = schema.GroupKind{Group: Group, Kind: CIKind}.String()
	CIKindAPIVersion   = CIKind + "." + SchemeGroupVersion.String()
	CIGroupVersionKind = SchemeGroupVersion.WithKind(CIKind)
)

func init() {
	SchemeBuilder.Register(&CI{}, &CIList{})
}
