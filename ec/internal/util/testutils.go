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

package util

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// Check conversion to attr.Value
// it should catch cases when e.g. the func under test returns types.List{}
func CheckConverionToAttrValue(t *testing.T, dt datasource.DataSource, attributeName string, attributeValue types.List) {
	resp := datasource.SchemaResponse{}
	dt.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Nil(t, resp.Diagnostics)

	attrType := resp.Schema.Attributes[attributeName].GetType()
	assert.NotNil(t, attrType, fmt.Sprintf("Type of attribute '%s' cannot be nil", attributeName))
	var target types.List
	diags := tfsdk.ValueFrom(context.Background(), attributeValue, attrType, &target)
	assert.Nil(t, diags)
}

func StringListAsType(t *testing.T, in []string) types.List {
	res, diags := types.ListValueFrom(context.Background(), types.StringType, in)
	assert.Nil(t, diags)
	return res
}

func StringMapAsType(t *testing.T, in map[string]string) types.Map {
	res, diags := types.MapValueFrom(context.Background(), types.StringType, in)
	assert.Nil(t, diags)
	return res
}
