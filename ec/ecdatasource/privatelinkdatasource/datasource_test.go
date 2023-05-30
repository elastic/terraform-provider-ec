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

package privatelinkdatasource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrivateLinkDataSource_ReadRegionData(t *testing.T) {
	t.Run("should error out when accessing an unknown region", func(t *testing.T) {
		source := privateLinkDataSource[v0AwsModel]{csp: "aws"}
		model := v0AwsModel{
			RegionField: "us-north-7",
		}

		_, err := source.readRegionData(model)
		require.ErrorIs(t, err, errUnknownRegion)
	})
	t.Run("should error out when accessing an unknown provider", func(t *testing.T) {
		source := privateLinkDataSource[v0AwsModel]{csp: "ibm"}
		model := v0AwsModel{
			RegionField: "us-north-7",
		}

		_, err := source.readRegionData(model)
		require.ErrorIs(t, err, errUnknownProvider)
	})
	t.Run("should return a populate state model when accessing a valid region", func(t *testing.T) {
		region := "us-east-1"
		source := privateLinkDataSource[v0AwsModel]{csp: "aws"}
		model := v0AwsModel{
			RegionField: region,
		}

		state, err := source.readRegionData(model)
		require.NoError(t, err)
		require.NotEmpty(t, state.DomainName)
		require.NotEmpty(t, state.VpcServiceName)
		require.NotEmpty(t, state.ZoneIDs)
		require.Equal(t, region, state.RegionField)
	})
}
