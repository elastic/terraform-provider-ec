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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestVersionValidator_ValidateString(t *testing.T) {
	tests := []struct {
		name    string
		isValid bool
		version types.String
	}{
		{
			name:    "should treat null values as valid",
			isValid: true,
			version: types.StringNull(),
		},
		{
			name:    "should treat unknown values as valid",
			isValid: true,
			version: types.StringUnknown(),
		},
		{
			name:    "should treat valid version strings as valid",
			isValid: true,
			version: types.StringValue("7.9.0"),
		},
		{
			name:    "should treat valid, tagged version strings as valid",
			isValid: true,
			version: types.StringValue("7.9.0-foo"),
		},
		{
			name:    "should treat invalid version strings as invalid",
			isValid: false,
			version: types.StringValue("not a real version"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := isVersion{}
			resp := validator.StringResponse{}
			v.ValidateString(context.Background(), validator.StringRequest{
				ConfigValue: tt.version,
			}, &resp)

			require.Equal(t, tt.isValid, !resp.Diagnostics.HasError())
		})
	}
}
