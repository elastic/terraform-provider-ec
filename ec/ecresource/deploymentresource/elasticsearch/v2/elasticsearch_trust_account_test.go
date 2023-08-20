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
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/stretchr/testify/require"
)

func ptr[T any](t T) *T {
	return &t
}

func Test_readElasticsearchTrustAccounts(t *testing.T) {
	tests := []struct {
		name             string
		settings         *models.ElasticsearchClusterSettings
		expectedAccounts ElasticsearchTrustAccounts
	}{
		{
			name:             "should return an empty list when the settings are nil",
			expectedAccounts: ElasticsearchTrustAccounts{},
		},
		{
			name:             "should return an empty list when the trust settings are nil",
			settings:         &models.ElasticsearchClusterSettings{},
			expectedAccounts: ElasticsearchTrustAccounts{},
		},
		{
			name: "should return an empty list when the trust settings are nil",
			settings: &models.ElasticsearchClusterSettings{
				Trust: &models.ElasticsearchClusterTrustSettings{},
			},
			expectedAccounts: ElasticsearchTrustAccounts{},
		},
		{
			name: "should return an empty list when the trust settings are empty",
			settings: &models.ElasticsearchClusterSettings{
				Trust: &models.ElasticsearchClusterTrustSettings{
					Accounts: []*models.AccountTrustRelationship{},
				},
			},
			expectedAccounts: ElasticsearchTrustAccounts{},
		},
		{
			name: "should return an empty list when the trust settings are empty",
			settings: &models.ElasticsearchClusterSettings{
				Trust: &models.ElasticsearchClusterTrustSettings{
					Accounts: []*models.AccountTrustRelationship{},
				},
			},
			expectedAccounts: ElasticsearchTrustAccounts{},
		},
		{
			name: "should return a list of the included trusted accounts",
			settings: &models.ElasticsearchClusterSettings{
				Trust: &models.ElasticsearchClusterTrustSettings{
					Accounts: []*models.AccountTrustRelationship{
						{
							AccountID:      ptr("account-id"),
							TrustAll:       ptr(false),
							TrustAllowlist: []string{"abc123", "def456"},
						},
						{
							AccountID: ptr("account-id"),
							TrustAll:  ptr(true),
						},
						{
							TrustAllowlist: []string{"abc123", "def456"},
						},
						nil,
					},
				},
			},
			expectedAccounts: ElasticsearchTrustAccounts{
				{
					AccountId:      ptr("account-id"),
					TrustAll:       ptr(false),
					TrustAllowlist: []string{"abc123", "def456"},
				},
				{
					AccountId: ptr("account-id"),
					TrustAll:  ptr(true),
				},
				{
					TrustAllowlist: []string{"abc123", "def456"},
				},
				{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, err := readElasticsearchTrustAccounts(tt.settings)
			require.NoError(t, err)
			require.Equal(t, tt.expectedAccounts, accounts)
		})
	}
}
