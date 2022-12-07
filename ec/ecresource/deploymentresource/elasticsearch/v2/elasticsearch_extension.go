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
	"context"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v1"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchExtensions v1.ElasticsearchExtensions

func ReadElasticsearchExtensions(in *models.ElasticsearchConfiguration) (ElasticsearchExtensions, error) {
	if len(in.UserBundles) == 0 && len(in.UserPlugins) == 0 {
		return nil, nil
	}

	extensions := make(ElasticsearchExtensions, 0, len(in.UserBundles)+len(in.UserPlugins))

	for _, model := range in.UserBundles {
		extension, err := ReadFromUserBundle(model)
		if err != nil {
			return nil, err
		}

		extensions = append(extensions, *extension)
	}

	for _, model := range in.UserPlugins {
		extension, err := ReadFromUserPlugin(model)
		if err != nil {
			return nil, err
		}

		extensions = append(extensions, *extension)
	}

	return extensions, nil
}

func elasticsearchExtensionPayload(ctx context.Context, extensions types.Set, es *models.ElasticsearchConfiguration) diag.Diagnostics {
	for _, elem := range extensions.Elems {
		var extension v1.ElasticsearchExtensionTF

		if diags := tfsdk.ValueAs(ctx, elem, &extension); diags.HasError() {
			return diags
		}

		version := extension.Version.Value
		url := extension.Url.Value
		name := extension.Name.Value

		if extension.Type.Value == "bundle" {
			es.UserBundles = append(es.UserBundles, &models.ElasticsearchUserBundle{
				Name:                 &name,
				ElasticsearchVersion: &version,
				URL:                  &url,
			})
		}

		if extension.Type.Value == "plugin" {
			es.UserPlugins = append(es.UserPlugins, &models.ElasticsearchUserPlugin{
				Name:                 &name,
				ElasticsearchVersion: &version,
				URL:                  &url,
			})
		}
	}
	return nil
}

func ReadFromUserBundle(in *models.ElasticsearchUserBundle) (*v1.ElasticsearchExtension, error) {
	var ext v1.ElasticsearchExtension

	ext.Type = "bundle"

	if in.ElasticsearchVersion == nil {
		return nil, utils.MissingField("ElasticsearchUserBundle.ElasticsearchVersion")
	}
	ext.Version = *in.ElasticsearchVersion

	if in.URL == nil {
		return nil, utils.MissingField("ElasticsearchUserBundle.URL")
	}
	ext.Url = *in.URL

	if in.Name == nil {
		return nil, utils.MissingField("ElasticsearchUserBundle.Name")
	}
	ext.Name = *in.Name

	return &ext, nil
}

func ReadFromUserPlugin(in *models.ElasticsearchUserPlugin) (*v1.ElasticsearchExtension, error) {
	var ext v1.ElasticsearchExtension

	ext.Type = "plugin"

	if in.ElasticsearchVersion == nil {
		return nil, utils.MissingField("ElasticsearchUserPlugin.ElasticsearchVersion")
	}
	ext.Version = *in.ElasticsearchVersion

	if in.URL == nil {
		return nil, utils.MissingField("ElasticsearchUserPlugin.URL")
	}
	ext.Url = *in.URL

	if in.Name == nil {
		return nil, utils.MissingField("ElasticsearchUserPlugin.Name")
	}
	ext.Name = *in.Name

	return &ext, nil
}
