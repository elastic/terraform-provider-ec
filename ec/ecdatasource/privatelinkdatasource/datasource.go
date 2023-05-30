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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/mitchellh/mapstructure"
)

//go:embed regionPrivateLinkMap.json
var privateLinkDataJson string

var (
	errUnknownRegion   = errors.New("could not find a privatelink endpoint for region")
	errUnknownProvider = errors.New("could not find a privatelink endpoint map for provider")
)

type regioner interface {
	Region() string
}

type privateLinkDataSource[T regioner] struct {
	csp             string
	privateLinkName string
}

func (d *privateLinkDataSource[T]) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = fmt.Sprintf("%s_%s_%s_endpoint", request.ProviderTypeName, d.csp, d.privateLinkName)
}

func (d privateLinkDataSource[T]) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state T

	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	state, err := d.readRegionData(state)
	if err != nil {
		response.Diagnostics.AddError("Failed to read private link data for region", err.Error())
	}
	response.State.Set(ctx, state)
}

func (d privateLinkDataSource[T]) readRegionData(state T) (T, error) {
	regionData, err := getRegionData(d.csp, state.Region())
	if err != nil {
		return state, err
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &state,
		TagName: "tfsdk",
	})
	if err != nil {
		return state, err
	}

	err = decoder.Decode(regionData)
	return state, err
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
