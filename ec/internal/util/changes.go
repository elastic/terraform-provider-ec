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

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// ObjectRemoved takes in a ResourceData and a key string, returning whether
// or not the object ([]intreface{} type) is being removed in the current
// change.
func ObjectRemoved(d *schema.ResourceData, key string) bool {
	old, new := d.GetChange(key)
	return len(old.([]interface{})) > 0 && len(new.([]interface{})) == 0
}
