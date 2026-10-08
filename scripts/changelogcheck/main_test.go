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
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func nulStatus(parts ...string) string {
	return strings.Join(parts, "\x00") + "\x00"
}

func TestRun(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "42.txt")
	require.NoError(t, os.WriteFile(entry, []byte("```release-note:none\n```\n"), 0o644))

	t.Run("ok", func(t *testing.T) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		code := run(
			[]string{"-pr", "42", "-entry", entry},
			strings.NewReader(nulStatus("A", ".changelog/42.txt")),
			stdout,
			stderr,
		)
		assert.Equal(t, 0, code)
		assert.Contains(t, stdout.String(), "ok")
		assert.Empty(t, stderr.String())
	})

	t.Run("missing entry fails closed", func(t *testing.T) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		code := run(
			[]string{"-pr", "99", "-entry", filepath.Join(dir, "missing.txt")},
			strings.NewReader(""),
			stdout,
			stderr,
		)
		assert.Equal(t, 1, code)
		assert.Contains(t, stderr.String(), "::error::missing .changelog/99.txt")
	})

	t.Run("rename of none to expected ok", func(t *testing.T) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		code := run(
			[]string{"-pr", "42", "-entry", entry},
			strings.NewReader(nulStatus("R100", ".changelog/1052.txt", ".changelog/42.txt")),
			stdout,
			stderr,
		)
		assert.Equal(t, 0, code)
		assert.Contains(t, stdout.String(), "ok")
	})

	t.Run("rename of bug content to expected fails", func(t *testing.T) {
		bugEntry := filepath.Join(dir, "bug.txt")
		require.NoError(t, os.WriteFile(bugEntry, []byte("```release-note:bug\nstolen\n```\n"), 0o644))
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		code := run(
			[]string{"-pr", "42", "-entry", bugEntry},
			strings.NewReader(nulStatus("R100", ".changelog/1057.txt", ".changelog/42.txt")),
			stdout,
			stderr,
		)
		assert.Equal(t, 1, code)
		assert.Contains(t, stderr.String(), "not a rename/copy")
	})

	t.Run("path with tab under .changelog is checked", func(t *testing.T) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		code := run(
			[]string{"-pr", "42", "-entry", entry},
			strings.NewReader(nulStatus("A", ".changelog/42.txt", "A", ".changelog/extra\tfile")),
			stdout,
			stderr,
		)
		assert.Equal(t, 1, code)
		assert.Contains(t, stderr.String(), "unexpected path added under .changelog/: .changelog/extra\tfile")
	})

	t.Run("bad pr flag", func(t *testing.T) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		code := run([]string{"-pr", "0"}, strings.NewReader(""), stdout, stderr)
		assert.Equal(t, 2, code)
		assert.Contains(t, stderr.String(), "-pr must be")
	})
}

func TestReadNameStatus(t *testing.T) {
	changes, err := readNameStatus(strings.NewReader(
		nulStatus("A", ".changelog/1.txt", "R100", "old.txt", ".changelog/2.txt", "D", ".changelog/3.txt"),
	))
	require.NoError(t, err)
	require.Len(t, changes, 3)
	assert.Equal(t, Change{Status: 'A', Path: ".changelog/1.txt"}, changes[0])
	assert.Equal(t, Change{Status: 'R', From: "old.txt", Path: ".changelog/2.txt"}, changes[1])
	assert.Equal(t, Change{Status: 'D', Path: ".changelog/3.txt"}, changes[2])
}
