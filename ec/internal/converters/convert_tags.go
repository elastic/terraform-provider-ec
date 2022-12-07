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

package converters

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

// flattenTags takes in Deployment Metadata resource models and returns its
// Tags in flattened form.
func TagsToTypeMap(metadataItems []*models.MetadataItem) types.Map {
	var tags = make(map[string]attr.Value)
	for _, res := range metadataItems {
		if res.Key != nil {
			tags[*res.Key] = types.String{Value: *res.Value}
		}
	}
	return types.Map{ElemType: types.StringType, Elems: tags}
}

// flattenTags takes in Deployment Metadata resource models and returns its
// Tags as Go map
func TagsToMap(metadataItems []*models.MetadataItem) map[string]string {
	if len(metadataItems) == 0 {
		return nil
	}
	res := make(map[string]string)
	for _, item := range metadataItems {
		if item.Key != nil {
			res[*item.Key] = *item.Value
		}
	}
	return res
}

func MapToTags(raw map[string]string) []*models.MetadataItem {
	result := make([]*models.MetadataItem, 0, len(raw))
	for k, v := range raw {
		result = append(result, &models.MetadataItem{
			Key:   ec.String(k),
			Value: ec.String(v),
		})
	}

	// Sort by key
	sort.SliceStable(result, func(i, j int) bool {
		return *result[i].Key < *result[j].Key
	})

	return result
}

func TFmapToTags(ctx context.Context, raw types.Map) ([]*models.MetadataItem, diag.Diagnostics) {
	result := make([]*models.MetadataItem, 0, len(raw.Elems))
	for k, v := range raw.Elems {
		var tag string
		if diags := tfsdk.ValueAs(ctx, v, &tag); diags.HasError() {
			return nil, diags
		}
		result = append(result, &models.MetadataItem{
			Key:   ec.String(k),
			Value: ec.String(tag),
		})
	}

	// Sort by key
	sort.SliceStable(result, func(i, j int) bool {
		return *result[i].Key < *result[j].Key
	})

	return result, nil
}
