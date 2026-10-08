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

package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/go-changelog"
)

// unhyphenatedFence matches a go-changelog fence that uses "releasenote"
// instead of the hyphenated "release-note" this repo requires.
var unhyphenatedFence = regexp.MustCompile("(?m)^```releasenote")

// allowedTypes are the release-note categories scripts/changelog.tmpl renders,
// plus "none" for non-user-facing PRs. This is intentionally not
// changelog.TypeValues: that list rejects new-data-source / new-guide and
// accepts types the template silently drops.
var allowedTypes = map[string]struct{}{
	"breaking-change": {},
	"note":            {},
	"feature":         {},
	"new-resource":    {},
	"new-data-source": {},
	"new-guide":       {},
	"enhancement":     {},
	"bug":             {},
	"none":            {},
}

// Change is one path change from `git diff --name-status`.
// For renames/copies, Path is the destination and From is the source.
type Change struct {
	Status byte // A, C, D, M, R, …
	Path   string
	From   string
}

// Options configures Validate.
type Options struct {
	PRNumber  int
	EntryBody string
	HasEntry  bool
	Changes   []Change
}

// Validate checks the PR changelog fragment and .changelog/ path changes.
// It returns human-readable error messages; an empty slice means pass.
func Validate(opts Options) []string {
	var errs []string

	if opts.PRNumber <= 0 {
		return []string{"pull request number must be positive"}
	}

	expected := fmt.Sprintf(".changelog/%d.txt", opts.PRNumber)

	// Git's default rename detection pairs a deleted blob with an identical
	// added blob as R100. Release-prep deletes pending release-note:none
	// fragments and adds a new none for the prep PR — that looks like a
	// rename. Allow R/C onto the expected path only when the new fragment is
	// solely release-note:none. Renaming a user-facing note onto {PR}.txt
	// (stealing another PR's release note) still fails.
	onlyNone := opts.HasEntry && entryIsOnlyNone(opts.EntryBody)

	expectedIntroduced := false
	var unexpected []string
	for _, ch := range opts.Changes {
		path := normalizePath(ch.Path)
		if path == "" || !strings.HasPrefix(path, ".changelog/") {
			continue
		}
		switch ch.Status {
		case 'A':
			if path == expected {
				expectedIntroduced = true
				continue
			}
			unexpected = append(unexpected, path)
		case 'R', 'C':
			if path == expected {
				if onlyNone {
					expectedIntroduced = true
					continue
				}
				from := normalizePath(ch.From)
				errs = append(errs, fmt.Sprintf("%s must be a new file (git status A), not a rename/copy from %s", expected, from))
				continue
			}
			unexpected = append(unexpected, path)
		}
	}
	sort.Strings(unexpected)
	for _, p := range unexpected {
		errs = append(errs, fmt.Sprintf("unexpected path added under .changelog/: %s (only %s is allowed)", p, expected))
	}

	if !opts.HasEntry {
		errs = append(errs, fmt.Sprintf("missing %s — add it in a follow-up commit after the PR number is known (see CONTRIBUTING.md)", expected))
		return errs
	}
	if !expectedIntroduced {
		errs = append(errs, fmt.Sprintf("%s must be newly added in this PR (git status A), or a rename/copy of a release-note:none fragment during release prep", expected))
		return errs
	}

	if unhyphenatedFence.MatchString(opts.EntryBody) {
		errs = append(errs, fmt.Sprintf("%s: use the hyphenated ```release-note:… fence; ```releasenote is not accepted", expected))
	}

	notes := changelog.NotesFromEntry(changelog.Entry{Body: opts.EntryBody})
	if len(notes) == 0 {
		errs = append(errs, fmt.Sprintf("%s: no ```release-note:<type> fenced block found", expected))
		return errs
	}

	for _, note := range notes {
		typ := strings.TrimSpace(note.Type)
		if typ == "" {
			errs = append(errs, fmt.Sprintf("%s: release-note block is missing a type (use ```release-note:<type>)", expected))
			continue
		}
		if _, ok := allowedTypes[typ]; !ok {
			errs = append(errs, fmt.Sprintf("%s: unknown release-note type %q (allowed: %s)", expected, typ, allowedTypeList()))
			continue
		}
		if typ != "none" && strings.TrimSpace(note.Body) == "" {
			errs = append(errs, fmt.Sprintf("%s: release-note:%s block has an empty body", expected, typ))
		}
	}

	return errs
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "./")
	return p
}

// entryIsOnlyNone reports whether every parsed release-note block is type none
// (the release-prep / non-user-facing marker).
func entryIsOnlyNone(body string) bool {
	notes := changelog.NotesFromEntry(changelog.Entry{Body: body})
	if len(notes) == 0 {
		return false
	}
	for _, note := range notes {
		if strings.TrimSpace(note.Type) != "none" {
			return false
		}
	}
	return true
}

func allowedTypeList() string {
	types := make([]string, 0, len(allowedTypes))
	for t := range allowedTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	return strings.Join(types, ", ")
}
