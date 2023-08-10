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
