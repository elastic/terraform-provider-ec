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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase[T regioner] struct {
	validRegion      string
	makeModel        func(region string) T
	dataSource       privateLinkDataSource[T]
	verifyValidState func(*testing.T, T)
}

func testDataSource[T regioner](t *testing.T, testCase testCase[T]) {
	t.Run(fmt.Sprintf("should error out when accessing an unknown region for %s", testCase.dataSource.csp), func(t *testing.T) {
		model := testCase.makeModel("antarctic-7")

		_, err := testCase.dataSource.readRegionData(model)
		require.ErrorIs(t, err, errUnknownRegion)
	})
	t.Run(fmt.Sprintf("should return a populate state model when accessing a valid region for %s", testCase.dataSource.csp), func(t *testing.T) {
		model := testCase.makeModel(testCase.validRegion)

		state, err := testCase.dataSource.readRegionData(model)
		require.NoError(t, err)
		testCase.verifyValidState(t, state)
	})
}

func TestPrivateLinkDataSource_ReadRegionData(t *testing.T) {
	testDataSource[v0AwsModel](t, testCase[v0AwsModel]{
		validRegion: "us-east-1",
		makeModel:   func(region string) v0AwsModel { return v0AwsModel{RegionField: region} },
		dataSource:  privateLinkDataSource[v0AwsModel]{csp: "aws"},
		verifyValidState: func(t *testing.T, state v0AwsModel) {
			require.NotEmpty(t, state.DomainName)
			require.NotEmpty(t, state.VpcServiceName)
			require.NotEmpty(t, state.ZoneIDs)
			require.Equal(t, "us-east-1", state.RegionField)
		},
	})
	testDataSource[v0GcpModel](t, testCase[v0GcpModel]{
		validRegion: "us-central1",
		makeModel:   func(region string) v0GcpModel { return v0GcpModel{RegionField: region} },
		dataSource:  privateLinkDataSource[v0GcpModel]{csp: "gcp"},
		verifyValidState: func(t *testing.T, state v0GcpModel) {
			require.NotEmpty(t, state.DomainName)
			require.NotEmpty(t, state.ServiceAttachmentUri)
			require.Equal(t, "us-central1", state.RegionField)
		},
	})
	testDataSource[v0AzureModel](t, testCase[v0AzureModel]{
		validRegion: "australiaeast",
		makeModel:   func(region string) v0AzureModel { return v0AzureModel{RegionField: region} },
		dataSource:  privateLinkDataSource[v0AzureModel]{csp: "azure"},
		verifyValidState: func(t *testing.T, state v0AzureModel) {
			require.NotEmpty(t, state.DomainName)
			require.NotEmpty(t, state.ServiceAlias)
			require.Equal(t, "australiaeast", state.RegionField)
		},
	})

	t.Run("should error out when accessing an unknown provider", func(t *testing.T) {
		source := privateLinkDataSource[v0AwsModel]{csp: "ibm"}
		model := v0AwsModel{
			RegionField: "us-north-7",
		}

		_, err := source.readRegionData(model)
		require.ErrorIs(t, err, errUnknownProvider)
	})
}
