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

package v1

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ApmConfigTF struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	DebugEnabled             types.Bool   `tfsdk:"debug_enabled"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type ApmConfig struct {
	DockerImage              *string `tfsdk:"docker_image"`
	DebugEnabled             *bool   `tfsdk:"debug_enabled"`
	UserSettingsJson         *string `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson *string `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         *string `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml *string `tfsdk:"user_settings_override_yaml"`
}

type ApmConfigs []ApmConfig
