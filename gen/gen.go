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

// This program generates ec/version.go. It can be invoked by running
// make generate
//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"log"
	"os"
	"text/template"
)

var templateFormat = `
package ec

// Version contains the current terraform provider version.
const Version = "{{ .Version }}"
`[1:]

type format struct {
	Version string
}

func main() {
	version := os.Getenv("VERSION")
	t := template.Must(template.New("").Parse(templateFormat))

	var buf = new(bytes.Buffer)
	if err := t.Execute(buf, format{Version: version}); err != nil {
		log.Fatalln(err)
	}

	os.WriteFile("./ec/version.go", buf.Bytes(), 0666)
}
