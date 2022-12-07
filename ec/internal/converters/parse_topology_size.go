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

package converters

import (
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deploymentsize"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ParseTopologySizeTF(size, sizeResource types.String) (*models.TopologySize, error) {
	return ParseTopologySize(&size.Value, &sizeResource.Value)
}

func ParseTopologySize(size, sizeResource *string) (*models.TopologySize, error) {
	if size == nil || *size == "" {
		return nil, nil
	}

	val, err := deploymentsize.ParseGb(*size)
	if err != nil {
		return nil, err
	}

	resource := "memory"

	if sizeResource != nil && *sizeResource != "" {
		resource = *sizeResource
	}

	return &models.TopologySize{
		Value:    ec.Int32(val),
		Resource: ec.String(resource),
	}, nil
}
