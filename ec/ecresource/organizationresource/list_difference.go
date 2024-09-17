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

package organizationresource

// Returns all the elements from array a that are not in array b
func difference[T interface{}](a, b []*T, getKey func(T) string) []*T {
	var diff []*T
	m := make(map[string]T)
	for _, item := range b {
		if item == nil {
			continue
		}
		key := getKey(*item)
		m[key] = *item
	}

	for _, item := range a {
		if item == nil {
			continue
		}
		key := getKey(*item)
		if _, ok := m[key]; !ok {
			diff = append(diff, item)
		}
	}

	return diff
}
