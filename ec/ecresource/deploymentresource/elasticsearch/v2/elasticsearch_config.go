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
	"reflect"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type ElasticsearchConfig v1.ElasticsearchConfig

func (c *ElasticsearchConfig) IsEmpty() bool {
	return c == nil || reflect.ValueOf(*c).IsZero()
}

func ReadElasticsearchConfig(in *models.ElasticsearchConfiguration) (*ElasticsearchConfig, error) {
	var config ElasticsearchConfig

	if in == nil {
		return &ElasticsearchConfig{}, nil
	}

	if len(in.EnabledBuiltInPlugins) > 0 {
		config.Plugins = append(config.Plugins, in.EnabledBuiltInPlugins...)
	}

	if in.UserSettingsYaml != "" {
		config.UserSettingsYaml = &in.UserSettingsYaml
	}

	if in.UserSettingsOverrideYaml != "" {
		config.UserSettingsOverrideYaml = &in.UserSettingsOverrideYaml
	}

	if o := in.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			config.UserSettingsJson = ec.String(string(b))
		}
	}

	if o := in.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			config.UserSettingsOverrideJson = ec.String(string(b))
		}
	}

	if in.DockerImage != "" {
		config.DockerImage = ec.String(in.DockerImage)
	}

	return &config, nil
}

func ElasticsearchConfigPayload(ctx context.Context, cfgObj attr.Value, model *models.ElasticsearchConfiguration) (*models.ElasticsearchConfiguration, diag.Diagnostics) {
	if cfgObj.IsNull() || cfgObj.IsUnknown() {
		return model, nil
	}

	var cfg v1.ElasticsearchConfigTF

	diags := tfsdk.ValueAs(ctx, cfgObj, &cfg)

	if diags.HasError() {
		return nil, diags
	}

	if cfg.UserSettingsJson.Value != "" {
		if err := json.Unmarshal([]byte(cfg.UserSettingsJson.Value), &model.UserSettingsJSON); err != nil {
			diags.AddError("failed expanding elasticsearch user_settings_json", err.Error())
		}
	}

	if cfg.UserSettingsOverrideJson.Value != "" {
		if err := json.Unmarshal([]byte(cfg.UserSettingsOverrideJson.Value), &model.UserSettingsOverrideJSON); err != nil {
			diags.AddError("failed expanding elasticsearch user_settings_override_json", err.Error())
		}
	}

	if !cfg.UserSettingsYaml.IsNull() {
		model.UserSettingsYaml = cfg.UserSettingsYaml.Value
	}

	if !cfg.UserSettingsOverrideYaml.IsNull() {
		model.UserSettingsOverrideYaml = cfg.UserSettingsOverrideYaml.Value
	}

	ds := cfg.Plugins.ElementsAs(ctx, &model.EnabledBuiltInPlugins, true)

	diags = append(diags, ds...)

	if !cfg.DockerImage.IsNull() {
		model.DockerImage = cfg.DockerImage.Value
	}

	return model, diags
}
