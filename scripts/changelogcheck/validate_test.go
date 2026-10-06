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
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	// Inline sample matching a real historical fragment shape. Do not read
	// .changelog/*.txt from disk: release prep deletes those files and would
	// break make unit on the release-prep PR and on master until a new fragment lands.
	goldenBug := "```release-note:bug\n" +
		"resource/ec_deployment: Use the deployment template Kibana size and zone count when `kibana` is set without those attributes, instead of clamping to 1g and 1 zone.\n" +
		"```\n"

	tests := []struct {
		name    string
		opts    Options
		wantErr []string // substrings that must each appear in some error
		wantOK  bool
	}{
		{
			name: "missing entry",
			opts: Options{
				PRNumber: 42,
				HasEntry: false,
			},
			wantErr: []string{"missing .changelog/42.txt"},
		},
		{
			name: "invalid pr number",
			opts: Options{
				PRNumber:  0,
				HasEntry:  true,
				EntryBody: "```release-note:bug\nfix\n```",
			},
			wantErr: []string{"pull request number must be positive"},
		},
		{
			name: "no release-note block",
			opts: Options{
				PRNumber:  42,
				HasEntry:  true,
				EntryBody: "just some text\n",
			},
			wantErr: []string{"no ```release-note:<type> fenced block found"},
		},
		{
			name: "empty type",
			opts: Options{
				PRNumber:  42,
				HasEntry:  true,
				EntryBody: "```release-note:\nbody\n```\n",
			},
			wantErr: []string{"missing a type"},
		},
		{
			name: "untyped release-note fence",
			opts: Options{
				PRNumber:  42,
				HasEntry:  true,
				EntryBody: "```release-note\nbody\n```\n",
			},
			wantErr: []string{"missing a type"},
		},
		{
			name: "unhyphenated releasenote rejected",
			opts: Options{
				PRNumber:  42,
				HasEntry:  true,
				EntryBody: "```releasenote:bug\nfix it\n```\n",
			},
			wantErr: []string{"hyphenated"},
		},
		{
			name: "prose mentioning releasenote is ok",
			opts: Options{
				PRNumber: 42,
				HasEntry: true,
				EntryBody: "Do not use the unhyphenated releasenote spelling.\n" +
					"```release-note:bug\nfix it\n```\n",
			},
			wantOK: true,
		},
		{
			name: "bug with empty body",
			opts: Options{
				PRNumber:  42,
				HasEntry:  true,
				EntryBody: "```release-note:bug\n```\n",
			},
			wantErr: []string{"empty body"},
		},
		{
			name: "none with empty body ok",
			opts: Options{
				PRNumber:   42,
				HasEntry:   true,
				EntryBody:  "```release-note:none\n```\n",
				AddedPaths: []string{".changelog/42.txt"},
			},
			wantOK: true,
		},
		{
			name: "none with whitespace body ok",
			opts: Options{
				PRNumber:  42,
				HasEntry:  true,
				EntryBody: "```release-note:none\n\n```\n",
			},
			wantOK: true,
		},
		{
			name: "realistic bug fragment",
			opts: Options{
				PRNumber:   1057,
				HasEntry:   true,
				EntryBody:  goldenBug,
				AddedPaths: []string{".changelog/1057.txt"},
			},
			wantOK: true,
		},
		{
			name: "crlf fragment",
			opts: Options{
				PRNumber:  42,
				HasEntry:  true,
				EntryBody: "```release-note:enhancement\r\nresource/ec_deployment: Add thing\r\n```\r\n",
			},
			wantOK: true,
		},
		{
			name: "multiple valid blocks",
			opts: Options{
				PRNumber: 42,
				HasEntry: true,
				EntryBody: "```release-note:bug\nfix a\n```\n" +
					"```release-note:enhancement\nadd b\n```\n",
			},
			wantOK: true,
		},
		{
			name: "extra added path any extension",
			opts: Options{
				PRNumber:   42,
				HasEntry:   true,
				EntryBody:  "```release-note:none\n```\n",
				AddedPaths: []string{".changelog/42.txt", ".changelog/notes.md"},
			},
			wantErr: []string{"unexpected path added under .changelog/: .changelog/notes.md"},
		},
		{
			name: "wrong filename added",
			opts: Options{
				PRNumber:   42,
				HasEntry:   true,
				EntryBody:  "```release-note:none\n```\n",
				AddedPaths: []string{".changelog/999.txt"},
			},
			wantErr: []string{"unexpected path added under .changelog/: .changelog/999.txt"},
		},
		{
			name: "non-changelog added paths ignored",
			opts: Options{
				PRNumber:   42,
				HasEntry:   true,
				EntryBody:  "```release-note:none\n```\n",
				AddedPaths: []string{"README.md", "scripts/changelogcheck/main.go"},
			},
			wantOK: true,
		},
		{
			name: "missing entry still reports unexpected paths",
			opts: Options{
				PRNumber:   42,
				HasEntry:   false,
				AddedPaths: []string{".changelog/notes.md"},
			},
			wantErr: []string{
				"unexpected path added under .changelog/: .changelog/notes.md",
				"missing .changelog/42.txt",
			},
		},
	}

	droppedTypes := []string{
		"improvement",
		"new-datasource",
		"deprecation",
		"new-ephemeral",
		"new-function",
		"new-action",
	}
	for _, typ := range droppedTypes {
		typ := typ
		tests = append(tests, struct {
			name    string
			opts    Options
			wantErr []string
			wantOK  bool
		}{
			name: "library type dropped by template: " + typ,
			opts: Options{
				PRNumber:  42,
				HasEntry:  true,
				EntryBody: "```release-note:" + typ + "\nbody\n```\n",
			},
			wantErr: []string{`unknown release-note type "` + typ + `"`},
		})
	}

	for _, typ := range templateTypes(t) {
		typ := typ
		tests = append(tests, struct {
			name    string
			opts    Options
			wantErr []string
			wantOK  bool
		}{
			name: "allowed type: " + typ,
			opts: Options{
				PRNumber:  42,
				HasEntry:  true,
				EntryBody: "```release-note:" + typ + "\nsome note body\n```\n",
			},
			wantOK: true,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := Validate(tt.opts)
			if tt.wantOK {
				assert.Empty(t, errs)
				return
			}
			require.NotEmpty(t, errs)
			joined := strings.Join(errs, "\n")
			for _, want := range tt.wantErr {
				assert.Contains(t, joined, want)
			}
		})
	}
}

func TestAllowedTypesMatchChangelogTmpl(t *testing.T) {
	fromTmpl := map[string]struct{}{}
	for _, typ := range templateTypes(t) {
		fromTmpl[typ] = struct{}{}
	}
	for typ := range fromTmpl {
		_, ok := allowedTypes[typ]
		assert.Truef(t, ok, "scripts/changelog.tmpl references %q but allowedTypes does not", typ)
	}
	for typ := range allowedTypes {
		if typ == "none" {
			continue
		}
		_, ok := fromTmpl[typ]
		assert.Truef(t, ok, "allowedTypes has %q but scripts/changelog.tmpl does not reference it", typ)
	}
}

var (
	tmplIndexType = regexp.MustCompile(`index\s+\.NotesByType\s+"([^"]+)"`)
	tmplDotType   = regexp.MustCompile(`\.NotesByType\.([a-z0-9-]+)`)
)

func templateTypes(t *testing.T) []string {
	t.Helper()
	body := readRepoFile(t, filepath.Join("..", "changelog.tmpl"))
	seen := map[string]struct{}{}
	for _, re := range []*regexp.Regexp{tmplIndexType, tmplDotType} {
		for _, m := range re.FindAllStringSubmatch(body, -1) {
			seen[m[1]] = struct{}{}
		}
	}
	require.NotEmpty(t, seen, "failed to extract types from scripts/changelog.tmpl")
	out := make([]string, 0, len(seen))
	for typ := range seen {
		out = append(out, typ)
	}
	return out
}

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	path := filepath.Join(filepath.Dir(thisFile), rel)
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(b)
}
