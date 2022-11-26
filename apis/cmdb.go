/*
Copyright 2020 The Crossplane Authors.

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

// Package apis contains Kubernetes API for the CMDB provider.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	idenreconv1alpha1 "github.com/crossplane/provider-cmdb/apis/idenrecon/v1alpha1"
	tablev1alpha1 "github.com/crossplane/provider-cmdb/apis/table/v1alpha1"
	cmdbv1alpha1 "github.com/crossplane/provider-cmdb/apis/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		cmdbv1alpha1.SchemeBuilder.AddToScheme,
		idenreconv1alpha1.SchemeBuilder.AddToScheme,
		tablev1alpha1.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
