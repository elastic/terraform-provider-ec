#!/usr/bin/env bash
# Self-check for extract-release-notes.sh. Does not tag, push, or publish.

set -euo pipefail

dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(dirname "$dir")"
extract="$dir/extract-release-notes.sh"
fixture="$dir/testdata/changelog-extract.md"
fail=0

assert_eq() {
	local name="$1" got="$2" want="$3"
	if [[ "$got" != "$want" ]]; then
		echo "FAIL: $name" >&2
		printf 'got:\n%s\nwant:\n%s\n' "$got" "$want" >&2
		fail=1
	else
		echo "ok: $name"
	fi
}

assert_fails() {
	local name="$1"
	shift
	if "$@" >/dev/null 2>&1; then
		echo "FAIL: $name (expected non-zero exit)" >&2
		fail=1
	else
		echo "ok: $name"
	fi
}

got="$("$extract" 0.13.1 "$fixture")"
want=$'FEATURES:\n\n* first line of 0.13.1\n\nENHANCEMENTS:\n\n* second line of 0.13.1'
assert_eq "fixture 0.13.1 body" "$got" "$want"

got="$("$extract" v0.13.0 "$fixture")"
want=$'FEATURES:\n\n* only 0.13.0'
assert_eq "fixture 0.13.0 strips v prefix" "$got" "$want"

got="$("$extract" 0.13.10 "$fixture")"
want=$'FEATURES:\n\n* should not leak into 0.13.1'
assert_eq "fixture 0.13.10 is not a prefix of 0.13.1" "$got" "$want"

assert_fails "missing version" "$extract" 9.9.9 "$fixture"
assert_fails "empty section" "$extract" 0.0.1 "$fixture"
assert_fails "usage: no args" "$extract"
assert_fails "changelog missing" "$extract" 0.13.1 "$dir/testdata/does-not-exist.md"

# Live CHANGELOG.md: same shape as GitHub notes for already-shipped versions.
assert_live() {
	local version="$1" must_have="$2" must_not="$3"
	local got
	got="$("$extract" "$version" "$root/CHANGELOG.md")"
	if printf '%s\n' "$got" | grep -q '^# '; then
		echo "FAIL: live $version included a markdown heading" >&2
		printf 'got:\n%s\n' "$got" >&2
		fail=1
		return
	fi
	if [[ "$got" != FEATURES:* && "$got" != ENHANCEMENTS:* && "$got" != "BUG FIXES:"* && "$got" != NOTES:* && "$got" != "BREAKING CHANGES:"* ]]; then
		echo "FAIL: live $version did not start at a changelog category" >&2
		printf 'got:\n%s\n' "$got" >&2
		fail=1
		return
	fi
	if [[ "$got" != *"$must_have"* ]]; then
		echo "FAIL: live $version missing expected text: $must_have" >&2
		printf 'got:\n%s\n' "$got" >&2
		fail=1
		return
	fi
	if [[ "$got" == *"$must_not"* ]]; then
		echo "FAIL: live $version leaked: $must_not" >&2
		printf 'got:\n%s\n' "$got" >&2
		fail=1
		return
	fi
	echo "ok: live CHANGELOG $version"
}

assert_live 0.13.0 "encryption_key_path" "linked"
assert_live 0.13.1 "linked" "encryption_key_path"

if [[ "$fail" -ne 0 ]]; then
	exit 1
fi
echo "extract-release-notes_test.sh: all ok"
