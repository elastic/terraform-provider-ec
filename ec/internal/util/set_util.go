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

import "sort"

// StringToItems takes in a slice of strings and returns a []interface{}.
func StringToItems(elems ...string) (result []interface{}) {
	for _, e := range elems {
		result = append(result, e)
	}

	return result
}

// ItemsToString takes in an []interface{} and returns a slice of strings.
func ItemsToString(elems []interface{}) (result []string) {
	for _, e := range elems {
		result = append(result, e.(string))
	}
	sort.Strings(result)

	return result
}
