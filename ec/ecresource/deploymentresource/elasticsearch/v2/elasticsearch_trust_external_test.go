package v2

import (
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/stretchr/testify/require"
)

func Test_readElasticsearchTrustExternals(t *testing.T) {
	tests := []struct {
		name             string
		settings         *models.ElasticsearchClusterSettings
		expectedAccounts ElasticsearchTrustExternals
	}{
		{
			name:             "should return an empty list when the settings are nil",
			expectedAccounts: ElasticsearchTrustExternals{},
		},
		{
			name:             "should return an empty list when the trust settings are nil",
			settings:         &models.ElasticsearchClusterSettings{},
			expectedAccounts: ElasticsearchTrustExternals{},
		},
		{
			name: "should return an empty list when the trust settings are nil",
			settings: &models.ElasticsearchClusterSettings{
				Trust: &models.ElasticsearchClusterTrustSettings{},
			},
			expectedAccounts: ElasticsearchTrustExternals{},
		},
		{
			name: "should return an empty list when the trust settings are empty",
			settings: &models.ElasticsearchClusterSettings{
				Trust: &models.ElasticsearchClusterTrustSettings{
					External: []*models.ExternalTrustRelationship{},
				},
			},
			expectedAccounts: ElasticsearchTrustExternals{},
		},
		{
			name: "should return an empty list when the trust settings are empty",
			settings: &models.ElasticsearchClusterSettings{
				Trust: &models.ElasticsearchClusterTrustSettings{
					External: []*models.ExternalTrustRelationship{},
				},
			},
			expectedAccounts: ElasticsearchTrustExternals{},
		},
		{
			name: "should return a list of the included trusted accounts",
			settings: &models.ElasticsearchClusterSettings{
				Trust: &models.ElasticsearchClusterTrustSettings{
					External: []*models.ExternalTrustRelationship{
						{
							TrustRelationshipID: ptr("complicated"),
							TrustAll:            ptr(false),
							TrustAllowlist:      []string{"abc123", "def456"},
						},
						{
							TrustRelationshipID: ptr("blessed"),
							TrustAll:            ptr(true),
						},
						{
							TrustAllowlist: []string{"abc123", "def456"},
						},
						nil,
					},
				},
			},
			expectedAccounts: ElasticsearchTrustExternals{
				{
					RelationshipId: ptr("complicated"),
					TrustAll:       ptr(false),
					TrustAllowlist: []string{"abc123", "def456"},
				},
				{
					RelationshipId: ptr("blessed"),
					TrustAll:       ptr(true),
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
			accounts, err := readElasticsearchTrustExternals(tt.settings)
			require.NoError(t, err)
			require.Equal(t, tt.expectedAccounts, accounts)
		})
	}
}
