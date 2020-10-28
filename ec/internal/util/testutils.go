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
	"errors"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// ResDataParams holds the raw configuration for NewResourceData to consume
type ResDataParams struct {
	// ID to set for the resource.
	ID string

	// The resource's schema.
	Schema map[string]*schema.Schema

	// The current resource state, to simulate a create or a case where no
	// previous state has been persisted, only State should be specified.
	State map[string]interface{}

	// The desired resource configuration, this is useful to simulate "update"
	// changes on a given resource.
	Change map[string]interface{}
}

// Validate the parameters
func (params ResDataParams) Validate() error {
	merr := multierror.NewPrefixed("invalid NewResourceData parameters")
	if params.ID == "" {
		merr = merr.Append(errors.New("id cannot be empty"))
	}

	if len(params.Schema) == 0 {
		merr = merr.Append(errors.New("schema cannot be empty"))
	}

	if params.State == nil {
		merr = merr.Append(errors.New("state cannot be empty"))
	}

	return merr.ErrorOrNil()
}

// NewResourceData creates a ResourceData from a raw configuration map and schema.
func NewResourceData(t *testing.T, params ResDataParams) *schema.ResourceData {
	t.Helper()
	if err := params.Validate(); err != nil {
		t.Fatal(err)
	}

	return resourceDataRaw(t,
		params.ID, params.Schema, params.State, params.Change,
	)
}

// resourceDataRaw creates a ResourceData from a raw configuration map.
// Setting the ID to the specified value, and using the desired map as diff
// to be applied, if not specified, then the current is used as the desired
// configuration starting off from an empty state.
func resourceDataRaw(t *testing.T, id string, schemaMap map[string]*schema.Schema, current, desired map[string]interface{}) *schema.ResourceData {
	t.Helper()

	result := generateRD(t, schemaMap, current, nil)
	result.SetId(id)
	if len(desired) == 0 {
		return result
	}

	return generateRD(t, schemaMap, desired, result.State())
}

func generateRD(t *testing.T, schemaMap map[string]*schema.Schema, rawAttr map[string]interface{}, state *terraform.InstanceState) *schema.ResourceData {
	resCfg := terraform.NewResourceConfigRaw(rawAttr)
	sm := schema.InternalMap(schemaMap)

	diff, err := sm.Diff(context.Background(), state, resCfg, nil, nil, true)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	result, err := sm.Data(state, diff)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return result
}
