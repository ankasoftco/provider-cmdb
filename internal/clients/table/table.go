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

package table

import (
	"github.com/anka-software/cmdb-sdk/pkg/client/table"

	"github.com/crossplane/provider-cmdb/internal/clients"
)

// NewTableClient returns a new Table service
func NewTableClient(cfg clients.Config) table.ClientService {
	cmdbConfig := clients.NewClient(cfg)

	return cmdbConfig.Table
}

// GenerateGetTableItemsOptions get items.
func GenerateGetTableItemsOptions(tableName string, ciName string) *table.GetTableItemsParams {
	var query = "name=" + ciName

	var params = table.NewGetTableItemParams().WithTableName(
		tableName).WithQuery(
		&query)

	return params
}
