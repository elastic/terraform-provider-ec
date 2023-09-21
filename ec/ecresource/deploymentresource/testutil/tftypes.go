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

package testutil

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func attrValueFromGoTypeValue(t *testing.T, goValue any, attributeType attr.Type) attr.Value {
	var attrValue attr.Value
	diags := tfsdk.ValueFrom(context.Background(), goValue, attributeType, &attrValue)
	assert.Nil(t, diags)
	return attrValue
}

func TfTypesValueFromGoTypeValue(t *testing.T, goValue any, attributeType attr.Type) tftypes.Value {
	attrValue := attrValueFromGoTypeValue(t, goValue, attributeType)
	tftypesValue, err := attrValue.ToTerraformValue(context.Background())
	assert.Nil(t, err)
	return tftypesValue
}
