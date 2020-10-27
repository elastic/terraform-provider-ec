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

package deploymentresource

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// imports a deployment providing the ability to add arbitrary keys and values
// in the "key=value" format.
func importFunc(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	properties, err := splitID(d.Id())

	if id, ok := properties["id"]; ok {
		d.SetId(id)
		delete(properties, "id")
	}

	if err != nil {
		return nil, err
	}

	merr := multierror.NewPrefixed("failed setting import keys")
	for k, v := range properties {
		if err := d.Set(k, v); err != nil {
			merr = merr.Append(
				fmt.Errorf(`failed setting "%s" as "%s": %w`, k, v, err),
			)
		}
	}

	if err := merr.ErrorOrNil(); err != nil {
		return nil, merr.ErrorOrNil()
	}

	return []*schema.ResourceData{d}, nil
}

func splitID(id string) (map[string]string, error) {
	result := make(map[string]string)
	properties := strings.Split(id, ":")
	result["id"] = properties[0]

	if len(properties) == 1 {
		return result, nil
	}

	merr := multierror.NewPrefixed("invalid import id")
	for _, prop := range properties[1:] {
		if p := strings.Split(prop, "="); len(p) == 2 {
			result[p[0]] = p[1]
			continue
		}
		merr = merr.Append(fmt.Errorf(`"%s" not in <key>=<value> format`, prop))
	}

	if err := merr.ErrorOrNil(); err != nil {
		return nil, merr.ErrorOrNil()
	}

	return result, nil
}
