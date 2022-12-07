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
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//go:embed regionPrivateLinkMap.json
var privateLinkDataJson string

type provider struct {
	name             string
	populateResource func(map[string]interface{}, *schema.ResourceData) error
}

var (
	errUnknownRegion   = errors.New("could not find a privatelink endpoint for region")
	errUnknownProvider = errors.New("could not find a privatelink endpoint map for provider")
	errMissingKey      = errors.New("expected region data key not available")
	errWrongType       = errors.New("unexapected type in region data key")
)

func readContextFor(p provider) func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
	return func(ctx context.Context, rd *schema.ResourceData, i interface{}) diag.Diagnostics {
		regionName, ok := rd.Get("region").(string)
		if !ok {
			return diag.Errorf("a region is required to lookup a privatelink endpoint")
		}

		if rd.Id() == "" {
			rd.SetId(strconv.Itoa(schema.HashString(fmt.Sprintf("%s:%s", p.name, regionName))))
		}

		regionData, err := getRegionData(p.name, regionName)
		if err != nil {
			return diag.FromErr(err)
		}

		return diag.FromErr(p.populateResource(regionData, rd))
	}
}

type configMap = map[string]interface{}
type regionToConfigMap = map[string]configMap
type providerToRegionMap = map[string]regionToConfigMap

func getRegionData(providerName string, regionName string) (map[string]interface{}, error) {
	var providerMap providerToRegionMap
	if err := json.Unmarshal([]byte(privateLinkDataJson), &providerMap); err != nil {
		return nil, err
	}

	providerData, ok := providerMap[providerName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", errUnknownProvider, providerName)
	}

	regionData, ok := providerData[regionName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", errUnknownRegion, regionName)
	}

	return regionData, nil
}

func copyToStateAs[T any](key string, from map[string]interface{}, rd *schema.ResourceData) error {
	value, ok := from[key]
	if !ok {
		return fmt.Errorf("%w: %s", errMissingKey, key)
	}

	castValue, ok := value.(T)
	if !ok {
		return fmt.Errorf("%w: %s", errWrongType, key)
	}

	return rd.Set(key, castValue)
}
