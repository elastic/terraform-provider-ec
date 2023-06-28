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

package v2

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type IntegrationsServerConfig v1.IntegrationsServerConfig

func readIntegrationsServerConfigs(in *models.IntegrationsServerConfiguration) (*IntegrationsServerConfig, error) {
	var cfg IntegrationsServerConfig

	if in.UserSettingsYaml != "" {
		cfg.UserSettingsYaml = &in.UserSettingsYaml
	}

	if in.UserSettingsOverrideYaml != "" {
		cfg.UserSettingsOverrideYaml = &in.UserSettingsOverrideYaml
	}

	if o := in.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			cfg.UserSettingsJson = ec.String(string(b))
		}
	}

	if o := in.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			cfg.UserSettingsOverrideJson = ec.String(string(b))
		}
	}

	if in.DockerImage != "" {
		cfg.DockerImage = &in.DockerImage
	}

	if in.SystemSettings != nil {
		if in.SystemSettings.DebugEnabled != nil {
			cfg.DebugEnabled = in.SystemSettings.DebugEnabled
		}
	}

	if cfg == (IntegrationsServerConfig{}) {
		return nil, nil
	}

	return &cfg, nil
}

func integrationsServerConfigPayload(ctx context.Context, cfgObj attr.Value, res *models.IntegrationsServerConfiguration) diag.Diagnostics {
	var diags diag.Diagnostics

	if cfgObj.IsNull() || cfgObj.IsUnknown() {
		return nil
	}

	var cfg *v1.IntegrationsServerConfigTF

	if diags = tfsdk.ValueAs(ctx, cfgObj, &cfg); diags.HasError() {
		return nil
	}

	if cfg == nil {
		return nil
	}

	if !cfg.DebugEnabled.IsNull() {
		if res.SystemSettings == nil {
			res.SystemSettings = &models.IntegrationsServerSystemSettings{}
		}
		res.SystemSettings.DebugEnabled = ec.Bool(cfg.DebugEnabled.ValueBool())
	}

	if cfg.UserSettingsJson.ValueString() != "" {
		if err := json.Unmarshal([]byte(cfg.UserSettingsJson.ValueString()), &res.UserSettingsJSON); err != nil {
			diags.AddError("failed expanding IntegrationsServer user_settings_json", err.Error())
		}
	}

	if cfg.UserSettingsOverrideJson.ValueString() != "" {
		if err := json.Unmarshal([]byte(cfg.UserSettingsOverrideJson.ValueString()), &res.UserSettingsOverrideJSON); err != nil {
			diags.AddError("failed expanding IntegrationsServer user_settings_override_json", err.Error())
		}
	}

	if !cfg.UserSettingsYaml.IsNull() {
		res.UserSettingsYaml = cfg.UserSettingsYaml.ValueString()
	}

	if !cfg.UserSettingsOverrideYaml.IsNull() {
		res.UserSettingsOverrideYaml = cfg.UserSettingsOverrideYaml.ValueString()
	}

	if !cfg.DockerImage.IsNull() {
		res.DockerImage = cfg.DockerImage.ValueString()
	}

	return diags
}
