// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package deploymenttemplates

import "github.com/hashicorp/terraform-plugin-framework/types"

type deploymentTemplatesDataSourceModel struct {
	Id             types.String              `tfsdk:"id"`
	Region         types.String              `tfsdk:"region"`
	StackVersion   types.String              `tfsdk:"stack_version"`
	ShowDeprecated types.Bool                `tfsdk:"show_deprecated"`
	Templates      []deploymentTemplateModel `tfsdk:"templates"` //< deploymentTemplateModel
}

type deploymentTemplateModel struct {
	ID                 string              `tfsdk:"id"`
	Name               string              `tfsdk:"name"`
	Description        string              `tfsdk:"description"`
	MinStackVersion    string              `tfsdk:"min_stack_version"`
	Deprecated         bool                `tfsdk:"deprecated"`
	Elasticsearch      *elasticsearchModel `tfsdk:"elasticsearch"`
	Kibana             *statelessModel     `tfsdk:"kibana"`
	Apm                *statelessModel     `tfsdk:"apm"`
	EnterpriseSearch   *statelessModel     `tfsdk:"enterprise_search"`
	IntegrationsServer *statelessModel     `tfsdk:"integrations_server"`
}
