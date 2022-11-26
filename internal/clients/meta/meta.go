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

package meta

import (
	"github.com/anka-software/cmdb-sdk/pkg/client/cmdb_meta"

	"github.com/crossplane/provider-cmdb/internal/clients"
)

// NewMetaClient returns a new Table service
func NewMetaClient(cfg clients.Config) cmdb_meta.ClientService {
	cmdbConfig := clients.NewClient(cfg)

	return cmdbConfig.CmdbMeta
}

// GenerateGetMetaOptions get items.
func GenerateGetMetaOptions(className string) *cmdb_meta.GetCmdbMetaParams {

	var params = cmdb_meta.NewGetCmdbMetaParams().WithClassName(
		className)

	return params
}
