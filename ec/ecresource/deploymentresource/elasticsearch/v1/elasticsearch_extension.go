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

type ElasticsearchExtensionTF struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Version types.String `tfsdk:"version"`
	Url     types.String `tfsdk:"url"`
}

type ElasticsearchExtensionsTF types.Set

type ElasticsearchExtension struct {
	Name    string `tfsdk:"name"`
	Type    string `tfsdk:"type"`
	Version string `tfsdk:"version"`
	Url     string `tfsdk:"url"`
}

type ElasticsearchExtensions []ElasticsearchExtension
