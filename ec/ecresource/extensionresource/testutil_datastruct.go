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

func newExtension() map[string]interface{} {
	return map[string]interface{}{
		"name":           "my_extension",
		"extension_type": "bundle",
		"description":    "my description",
		"version":        "*",
		"download_url":   "https://example.com",
		"url":            "repo://1234",
		"last_modified":  "2021-01-07T22:13:42.999Z",
		"size":           1000,
	}
}

func newExtensionWithFilePath() map[string]interface{} {
	return map[string]interface{}{
		"name":           "my_extension",
		"extension_type": "bundle",
		"description":    "my description",
		"version":        "*",
		"download_url":   "https://example.com",
		"url":            "repo://1234",
		"last_modified":  "2021-01-07T22:13:42.999Z",
		"size":           1000,

		"file_path": "testdata/test_extension_bundle.json",
		"file_hash": "abcd",
	}
}
