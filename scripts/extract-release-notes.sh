#!/usr/bin/env bash
#
# Print the CHANGELOG.md section for a version (GitHub release body).
# Usage: extract-release-notes.sh X.Y.Z [CHANGELOG.md]
#

set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
	echo "usage: $0 X.Y.Z [CHANGELOG.md]" >&2
	exit 2
fi

version="${1#v}"
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
changelog="${2:-$(dirname "${__dir}")/CHANGELOG.md}"

if [[ ! -f "$changelog" ]]; then
	echo "changelog not found: $changelog" >&2
	exit 1
fi

awk -v ver="$version" '
	$0 ~ "^# " ver "( |$)" { found = 1; next }
	found && /^# / { exit }
	found {
		if (!started) {
			if ($0 ~ /^[[:space:]]*$/) next
			started = 1
		}
		lines[++n] = $0
	}
	END {
		if (!started) {
			print "no changelog section for " ver " in " FILENAME > "/dev/stderr"
			exit 1
		}
		while (n > 0 && lines[n] ~ /^[[:space:]]*$/) n--
		for (i = 1; i <= n; i++) print lines[i]
	}
' "$changelog"
