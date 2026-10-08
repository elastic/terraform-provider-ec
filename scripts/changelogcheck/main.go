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

// Command changelogcheck validates a pull request's .changelog/{PR}.txt fragment
// against the cloud provider's release-note contract.
//
// Usage:
//
//	changelogcheck -pr <number> [-entry <path>]
//
// Stdin is `git diff --name-status -z` output (NUL-delimited status/path
// fields). Exit status is 1 when validation fails.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("changelogcheck", flag.ContinueOnError)
	fs.SetOutput(stderr)
	pr := fs.Int("pr", 0, "pull request number")
	entry := fs.String("entry", "", "path to .changelog/{PR}.txt (default: .changelog/<pr>.txt)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *pr <= 0 {
		_, _ = fmt.Fprintln(stderr, "changelogcheck: -pr must be a positive pull request number")
		return 2
	}

	entryPath := *entry
	if entryPath == "" {
		entryPath = fmt.Sprintf(".changelog/%d.txt", *pr)
	}

	changes, err := readNameStatus(stdin)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "changelogcheck: read name-status: %v\n", err)
		return 2
	}

	body, hasEntry, err := readEntry(entryPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "changelogcheck: read %s: %v\n", entryPath, err)
		return 2
	}

	errs := Validate(Options{
		PRNumber:  *pr,
		EntryBody: body,
		HasEntry:  hasEntry,
		Changes:   changes,
	})
	if len(errs) == 0 {
		_, _ = fmt.Fprintf(stdout, "changelogcheck: %s ok\n", entryPath)
		return 0
	}
	for _, e := range errs {
		// ::error:: surfaces the failure in the GitHub Actions UI annotation list.
		_, _ = fmt.Fprintf(stderr, "::error::%s\n", e)
	}
	return 1
}

func readEntry(path string) (string, bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return string(b), true, nil
}

// readNameStatus parses `git diff --name-status -z` NUL-delimited fields.
// Records are: status\0path\0 or, for renames/copies, status\0old\0new\0.
func readNameStatus(r io.Reader) ([]Change, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	raw := bytes.Split(data, []byte{0})
	fields := make([]string, 0, len(raw))
	for _, p := range raw {
		if len(p) == 0 {
			continue // trailing NUL produces a final empty field
		}
		fields = append(fields, string(p))
	}

	var changes []Change
	for i := 0; i < len(fields); {
		statusField := fields[i]
		i++
		status := statusField[0]
		switch status {
		case 'R', 'C':
			if i+1 >= len(fields) {
				return nil, fmt.Errorf("invalid rename/copy name-status near %q", statusField)
			}
			changes = append(changes, Change{Status: status, From: fields[i], Path: fields[i+1]})
			i += 2
		default:
			if i >= len(fields) {
				return nil, fmt.Errorf("invalid name-status near %q: missing path", statusField)
			}
			changes = append(changes, Change{Status: status, Path: fields[i]})
			i++
		}
	}
	return changes, nil
}
