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
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type ApmConfig = v1.ApmConfig

func readApmConfigs(in *models.ApmConfiguration) (v1.ApmConfigs, error) {
	var cfg ApmConfig

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

	if cfg == (ApmConfig{}) {
		return nil, nil
	}

	return v1.ApmConfigs{cfg}, nil
}

func apmConfigPayload(ctx context.Context, cfg v1.ApmConfigTF, model *models.ApmConfiguration) diag.Diagnostics {
	if !cfg.DebugEnabled.IsNull() {
		if model.SystemSettings == nil {
			model.SystemSettings = &models.ApmSystemSettings{}
		}
		model.SystemSettings.DebugEnabled = &cfg.DebugEnabled.Value
	}

	var diags diag.Diagnostics
	if cfg.UserSettingsJson.Value != "" {
		if err := json.Unmarshal([]byte(cfg.UserSettingsJson.Value), &model.UserSettingsJSON); err != nil {
			diags.AddError("failed expanding apm user_settings_json", err.Error())
			return diags
		}
	}

	if cfg.UserSettingsOverrideJson.Value != "" {
		if err := json.Unmarshal([]byte(cfg.UserSettingsOverrideJson.Value), &model.UserSettingsOverrideJSON); err != nil {
			diags.AddError("failed expanding apm user_settings_override_json", err.Error())
			return diags
		}
	}

	if !cfg.UserSettingsYaml.IsNull() {
		model.UserSettingsYaml = cfg.UserSettingsYaml.Value
	}

	if !cfg.UserSettingsOverrideYaml.IsNull() {
		model.UserSettingsOverrideYaml = cfg.UserSettingsOverrideYaml.Value
	}

	if !cfg.DockerImage.IsNull() {
		model.DockerImage = cfg.DockerImage.Value
	}

	return nil
}
