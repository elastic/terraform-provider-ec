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
// Added paths under .changelog/ are read from stdin (one path per line), typically
// from `git diff --name-only --diff-filter=A HEAD^1 HEAD`. Exit status is 1 when
// validation fails.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
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

	added, err := readLines(stdin)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "changelogcheck: read added paths: %v\n", err)
		return 2
	}

	body, hasEntry, err := readEntry(entryPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "changelogcheck: read %s: %v\n", entryPath, err)
		return 2
	}

	errs := Validate(Options{
		PRNumber:   *pr,
		EntryBody:  body,
		HasEntry:   hasEntry,
		AddedPaths: added,
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

func readLines(r io.Reader) ([]string, error) {
	var lines []string
	sc := bufio.NewScanner(r)
	// git path lists are short; 1 MiB per line is plenty.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, sc.Err()
}
