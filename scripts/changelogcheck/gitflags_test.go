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
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gitDiffFlags must stay in sync with .github/workflows/pr-changelog-check.yml.
var gitDiffFlags = []string{"diff", "--name-status", "-z", "-M", "-C100%", "--find-copies-harder", "HEAD^", "HEAD"}

func TestGitDiffFlags_verbatimCopyIsC100(t *testing.T) {
	dir := initChangelogRepo(t)
	writeFrag(t, dir, "1.txt", "```release-note:bug\nstolen note\n```\n")
	commitAll(t, dir, "base")

	require.NoError(t, copyFile(
		filepath.Join(dir, ".changelog", "1.txt"),
		filepath.Join(dir, ".changelog", "2.txt"),
	))
	commitAll(t, dir, "verbatim copy")

	changes, err := readNameStatus(bytes.NewReader(gitDiff(t, dir)))
	require.NoError(t, err)
	require.Len(t, changes, 1)
	assert.Equal(t, byte('C'), changes[0].Status)
	assert.Equal(t, ".changelog/1.txt", changes[0].From)
	assert.Equal(t, ".changelog/2.txt", changes[0].Path)
}

func TestGitDiffFlags_similarNewNoteIsA(t *testing.T) {
	dir := initChangelogRepo(t)
	writeFrag(t, dir, "1.txt", "```release-note:enhancement\nresource/ec_deployment: Add foo support\n```\n")
	commitAll(t, dir, "base")

	// Similar fences/wording but not byte-identical — must stay A under -C100%.
	writeFrag(t, dir, "2.txt", "```release-note:enhancement\nresource/ec_deployment: Add bar support for users\n```\n")
	commitAll(t, dir, "similar new note")

	changes, err := readNameStatus(bytes.NewReader(gitDiff(t, dir)))
	require.NoError(t, err)
	require.Len(t, changes, 1)
	assert.Equal(t, byte('A'), changes[0].Status)
	assert.Equal(t, ".changelog/2.txt", changes[0].Path)
}

func initChangelogRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "test")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".changelog"), 0o755))
	return dir
}

func writeFrag(t *testing.T, dir, name, body string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".changelog", name), []byte(body), 0o644))
}

func commitAll(t *testing.T, dir, msg string) {
	t.Helper()
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", msg)
}

func gitDiff(t *testing.T, dir string) []byte {
	t.Helper()
	cmd := exec.Command("git", gitDiffFlags...)
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err, "git %v", gitDiffFlags)
	return out
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}
