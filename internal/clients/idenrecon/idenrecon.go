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

package idenrecon

import (
	"fmt"
	"strings"

	"github.com/anka-software/cmdb-sdk/pkg/client/cmdb"
	"github.com/anka-software/cmdb-sdk/pkg/models"

	"github.com/crossplane/provider-cmdb/apis/idenrecon/v1alpha1"
	"github.com/crossplane/provider-cmdb/internal/clients"
)

// NewIdenReconClient returns a new Identification and Reconciliation service
func NewIdenReconClient(cfg clients.Config) cmdb.ClientService {
	cmdbConfig := clients.NewClient(cfg)

	return cmdbConfig.Cmdb
}

// GenerateCIOptions creates/updates.
func GenerateCIOptions(d *v1alpha1.CIParameters) *cmdb.CreateIdentifyReconcileParams {
	d.Values["name"] = d.Name
	var params = cmdb.NewCreateIdentifyReconcileParams().WithSysParamDataSource(
		&d.SysParamDataSource).WithBody(&models.IdentifyReconcileItemList{
		Items: []*models.IdentifyReconcileItem{{ClassName: d.ClassName, Values: d.Values}},
	})

	return params
}

// IsResourceUpToDate for observation
func IsResourceUpToDate(desired map[string]string, current map[string]interface{}) bool {
	for k, v := range desired {
		str, ok := current[k].(string)
		if ok {
			if str != v {
				fmt.Printf("The field %v is not up to date. Current State:%v Desired State:%v\n", k, str, v)
				return false
			}
		}
	}
	return true
}

// ContainsField for linter
func ContainsField(s []string, e map[string]string) error {
	var isExist bool
	for k := range e {
		isExist = false
		for _, a := range s {
			if a == k {
				isExist = true
				break
			}
		}
		if !isExist {
			similarFields := GetSimilarFields(s, k)
			return fmt.Errorf("The field:%v is not recognized.\nAvailable fields that are similar to %v:\n%v", k, k, similarFields)
		}
	}

	return nil
}

// GetSimilarFields for linter
func GetSimilarFields(s []string, str string) []string {
	var similarFields []string
	for _, v := range s {
		isSimilar := strings.Contains(v, str)
		if isSimilar {
			similarFields = append(similarFields, v)
		}
	}
	return similarFields
}
