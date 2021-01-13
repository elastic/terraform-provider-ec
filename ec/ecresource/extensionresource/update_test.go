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

package extensionresource

import (
	"context"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_updateResource(t *testing.T) {
	tc200withoutFilePath := util.NewResourceData(t, util.ResDataParams{
		ID:     "12345678",
		State:  newExtension(),
		Schema: newSchema(),
	})

	wantTC200statewithoutFilePath := newExtension()
	wantTC200statewithoutFilePath["name"] = "updated_extension"
	wantTC200withoutFilePath := util.NewResourceData(t, util.ResDataParams{
		ID:     "12345678",
		State:  wantTC200statewithoutFilePath,
		Schema: newSchema(),
	})

	tc200withFilePath := util.NewResourceData(t, util.ResDataParams{
		ID:     "12345678",
		State:  newExtensionWithFilePath(),
		Schema: newSchema(),
	})
	wantTC200statewithFilePath := newExtensionWithFilePath()
	wantTC200statewithFilePath["name"] = "updated_extension"
	wantTC200withFilePath := util.NewResourceData(t, util.ResDataParams{
		ID:     "12345678",
		State:  wantTC200statewithFilePath,
		Schema: newSchema(),
	})

	tc500Err := util.NewResourceData(t, util.ResDataParams{
		ID:     "12345678",
		State:  newExtension(),
		Schema: newSchema(),
	})
	wantTC500 := util.NewResourceData(t, util.ResDataParams{
		ID:     "12345678",
		State:  newExtension(),
		Schema: newSchema(),
	})

	lastModified, _ := strfmt.ParseDateTime("2021-01-07T22:13:42.999Z")
	type args struct {
		ctx  context.Context
		d    *schema.ResourceData
		meta interface{}
	}
	tests := []struct {
		name   string
		args   args
		want   diag.Diagnostics
		wantRD *schema.ResourceData
	}{
		{
			name: "returns nil when it receives a 200 without file_path",
			args: args{
				d: tc200withoutFilePath,
				meta: api.NewMock(
					mock.New200StructResponse(models.Extension{ // update request response
						Name:          ec.String("updated_extension"),
						ExtensionType: ec.String("bundle"),
						Description:   "my description",
						Version:       ec.String("*"),
						DownloadURL:   "https://example.com",
						URL:           ec.String("repo://1234"),
						FileMetadata: &models.ExtensionFileMetadata{
							LastModifiedDate: lastModified,
							Size:             1000,
						},
					}),
					mock.New200StructResponse(models.Extension{ // read request response
						Name:          ec.String("updated_extension"),
						ExtensionType: ec.String("bundle"),
						Description:   "my description",
						Version:       ec.String("*"),
						DownloadURL:   "https://example.com",
						URL:           ec.String("repo://1234"),
						FileMetadata: &models.ExtensionFileMetadata{
							LastModifiedDate: lastModified,
							Size:             1000,
						},
					}),
				),
			},
			want:   nil,
			wantRD: wantTC200withoutFilePath,
		},
		{
			name: "returns nil when it receives a 200 with file_path",
			args: args{
				d: tc200withFilePath,
				meta: api.NewMock(
					mock.New200StructResponse(models.Extension{ // update request response
						Name:          ec.String("updated_extension"),
						ExtensionType: ec.String("bundle"),
						Description:   "my description",
						Version:       ec.String("*"),
						DownloadURL:   "https://example.com",
						URL:           ec.String("repo://1234"),
						FileMetadata: &models.ExtensionFileMetadata{
							LastModifiedDate: lastModified,
							Size:             1000,
						},
					}),
					mock.New200StructResponse(nil), // upload request response
					mock.New200StructResponse(models.Extension{ // read request response
						Name:          ec.String("updated_extension"),
						ExtensionType: ec.String("bundle"),
						Description:   "my description",
						Version:       ec.String("*"),
						DownloadURL:   "https://example.com",
						URL:           ec.String("repo://1234"),
						FileMetadata: &models.ExtensionFileMetadata{
							LastModifiedDate: lastModified,
							Size:             1000,
						},
					}),
				),
			},
			want:   nil,
			wantRD: wantTC200withFilePath,
		},
		{
			name: "returns an error when it receives a 500",
			args: args{
				d: tc500Err,
				meta: api.NewMock(mock.NewErrorResponse(500, mock.APIError{
					Code: "some", Message: "message",
				})),
			},
			want: diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "api error: 1 error occurred:\n\t* some: message\n\n",
				},
			},
			wantRD: wantTC500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateResource(tt.args.ctx, tt.args.d, tt.args.meta)
			assert.Equal(t, tt.want, got)
			var want interface{}
			if tt.wantRD != nil {
				if s := tt.wantRD.State(); s != nil {
					want = s.Attributes
				}
			}

			var gotState interface{}
			if s := tt.args.d.State(); s != nil {
				gotState = s.Attributes
			}

			assert.Equal(t, want, gotState)
		})
	}
}
