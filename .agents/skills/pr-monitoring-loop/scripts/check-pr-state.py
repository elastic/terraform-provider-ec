#!/usr/bin/env python3
"""Collect deterministic GitHub PR state for watcher subagents.

The script is intentionally read-only against GitHub. It fetches PR metadata,
checks (pinned to the current head SHA), reviews, comments, review threads, and
issue events, then computes a `summary` block that watcher subagents use to
decide whether the PR is actionable.

Key concepts:

- `--state-file` persists across watcher restarts. It records seen comment/
  review IDs and timestamps so "new comment since last poll" survives the
  short-lived nature of fresh subagents.
- New-vs-old discrimination is purely ID/timestamp based. There is NO author
  filtering: bots (verify-openspec, macroscope, github-actions) emit
  first-class signal, and the PR author is frequently the human reviewer
  because the agent commits on their behalf.
- `summary.reviews.verifyOpenspec.runState` derives the verify-openspec
  workflow state. The workflow REMOVES its own label as soon as it picks the
  PR up, so label absence on `pr.labels` is not a signal to re-request.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Callable, Iterable, Optional


PR_VIEW_FIELDS = [
    "author",
    "baseRefName",
    "baseRefOid",
    "headRefName",
    "headRefOid",
    "isDraft",
    "labels",
    "mergeStateStatus",
    "mergeable",
    "number",
    "reviewDecision",
    "state",
    "statusCheckRollup",
    "title",
    "updatedAt",
    "url",
]

# Fields `gh pr checks --json` actually accepts. `conclusion` and `detailsUrl`
# are rejected, which makes the whole fallback call fail.
CHECK_FIELDS = [
    "bucket",
    "completedAt",
    "description",
    "event",
    "link",
    "name",
    "startedAt",
    "state",
    "workflow",
]

VERIFY_OPENSPEC_LABEL = "verify-openspec"
# The verify-openspec workflow runs under GitHub Actions and posts as
# `github-actions[bot]`. Author identity alone is not enough (that bot posts
# many other things), so we ALSO require the body to contain the workflow's
# distinctive "OpenSpec verify" / "Verification Report" signature.
VERIFY_OPENSPEC_REVIEW_AUTHORS = {"github-actions[bot]", "github-actions"}
VERIFY_OPENSPEC_REVIEW_BODY_MARKERS = (
    "OpenSpec verify",
    "Verification Report",
)

# Cloud-provider CI split. Auto-fixable names are the GitHub Actions jobs a
# watcher may patch. Out-of-band names are Buildkite commit statuses: report
# them and escalate failures. Never auto-fix and never re-trigger acceptance
# (each run creates paid Elastic Cloud deployments).
AUTO_FIXABLE_CHECK_NAMES = {"Unit", "Validate OpenSpecs"}
# Branch protection on master requires these in addition to Buildkite acceptance.
REQUIRED_PASSED_CHECK_NAMES = ("Unit", "CLA")
ACCEPTANCE_CHECK_NAME = "buildkite/terraform-provider-ec-acceptance"
OUT_OF_BAND_CHECK_NAMES = {
    ACCEPTANCE_CHECK_NAME,
    "buildkite/terraform-provider-ec-release",
}

# cp-hosted-team#4386 lands the verify workflow. Match any filename so a later
# .md, .yml, or compiled .lock.yml still counts.
VERIFY_WORKFLOW_GLOB = ".github/workflows/*openspec-verify*"

EXIT_OK = 0
EXIT_TRANSIENT = 2
EXIT_TIMEOUT = 124

DEFAULT_INTERVAL_SECONDS = 60
DEFAULT_MAX_DURATION_SECONDS = 1800
# Acceptance runs about 120 minutes, plus queue time. Poll slowly and stop
# after 3 hours so a run that never posts still returns to a human.
ACCEPTANCE_INTERVAL_SECONDS = 300
ACCEPTANCE_MAX_DURATION_SECONDS = 3 * 60 * 60
SEEN_ID_CAP = 1000

# Maximum age of lastPolledAt before we treat it as stale and reset
# timestamp-based discrimination. This prevents a long-dead watcher from
# missing updates because its cached lastPolledAt predates the actual
# current state of the PR.
STATE_LAST_POLLED_TTL_SECONDS = 30 * 60


# ---------------------------------------------------------------------------
# Subprocess wrappers (mockable seam for tests)
# ---------------------------------------------------------------------------


class TransientGhError(RuntimeError):
    """Raised when a `gh` invocation fails after retries."""


def run_gh(
    args: list[str],
    *,
    allow_failure: bool = False,
    retries: int = 1,
) -> str:
    """Run a `gh` command, optionally retrying once on transient failures."""

    attempt = 0
    last_stderr = ""
    while True:
        result = subprocess.run(
            ["gh", *args],
            check=False,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        if result.returncode == 0:
            return result.stdout
        last_stderr = result.stderr.strip()
        if attempt >= retries:
            break
        attempt += 1
        time.sleep(1.0 * attempt)

    if allow_failure:
        return ""

    raise TransientGhError(
        json.dumps(
            {
                "error": "gh command failed",
                "command": ["gh", *args],
                "stderr": last_stderr,
                "returncode": result.returncode,
            }
        )
    )


def run_git(
    args: list[str], *, allow_failure: bool = False
) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["git", *args],
        check=not allow_failure,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )


def loads_gh_output(output: str) -> Any:
    """Parse `gh` stdout, including one JSON value per `--paginate` page.

    A single page, including an array that `gh --paginate` already merged, is
    returned as that value. Object pages such as check-runs are concatenated
    and come back as a list of pages.
    """

    decoder = json.JSONDecoder()
    pages: list[Any] = []
    idx = 0
    length = len(output)
    while idx < length:
        while idx < length and output[idx].isspace():
            idx += 1
        if idx >= length:
            break
        value, end = decoder.raw_decode(output, idx)
        pages.append(value)
        idx = end
    if len(pages) == 1:
        return pages[0]
    if not pages:
        raise json.JSONDecodeError("no JSON value", output, 0)
    return pages


def flatten_item_pages(data: Any) -> list[Any]:
    """Flatten a list endpoint that may be one page or many.

    One page is a list of items. A list of page-lists is flattened.
    """

    if not isinstance(data, list):
        return []
    if data and all(isinstance(page, list) for page in data):
        flat: list[Any] = []
        for page in data:
            flat.extend(page)
        return flat
    return data


def check_runs_from_payload(data: Any) -> list[dict[str, Any]]:
    """Collect `check_runs` from one status object or a list of page objects."""

    pages = [data] if isinstance(data, dict) else data if isinstance(data, list) else []
    runs: list[dict[str, Any]] = []
    for page in pages:
        if isinstance(page, dict):
            runs.extend(page.get("check_runs") or [])
    return runs


def combined_status_from_payload(data: Any) -> dict[str, Any]:
    """Merge `statuses` from one combined-status object or a list of pages."""

    pages = [data] if isinstance(data, dict) else data if isinstance(data, list) else []
    statuses: list[dict[str, Any]] = []
    combined: dict[str, Any] = {}
    for page in pages:
        if not isinstance(page, dict):
            continue
        if not combined:
            combined = dict(page)
        statuses.extend(page.get("statuses") or [])
    if not combined:
        return {}
    combined["statuses"] = statuses
    return combined


def gh_json(
    args: list[str],
    *,
    default: Any,
    allow_failure: bool = False,
    retries: int = 1,
) -> Any:
    output = run_gh(args, allow_failure=allow_failure, retries=retries)
    if not output.strip():
        return default
    try:
        return loads_gh_output(output)
    except json.JSONDecodeError:
        if allow_failure:
            return default
        raise TransientGhError(
            json.dumps(
                {
                    "error": "gh command returned invalid JSON",
                    "command": ["gh", *args],
                    "stdout": output[:2000],
                }
            )
        )


# ---------------------------------------------------------------------------
# Data fetchers
# ---------------------------------------------------------------------------


def repo_info() -> dict[str, str]:
    data = gh_json(["repo", "view", "--json", "owner,name"], default={})
    owner = data.get("owner", {})
    return {"owner": owner.get("login", ""), "name": data.get("name", "")}


def pr_view(pr: str) -> dict[str, Any]:
    return gh_json(
        ["pr", "view", pr, "--json", ",".join(PR_VIEW_FIELDS)],
        default={},
    )


def pr_checks(pr: str) -> list[dict[str, Any]]:
    return gh_json(
        ["pr", "checks", pr, "--json", ",".join(CHECK_FIELDS)],
        default=[],
        allow_failure=True,
    )


def commit_check_runs(owner: str, repo: str, sha: str) -> list[dict[str, Any]]:
    """Check-runs pinned to a specific commit SHA (canonical for the head)."""

    if not (owner and repo and sha):
        return []
    data = gh_json(
        [
            "api",
            f"repos/{owner}/{repo}/commits/{sha}/check-runs",
            "--paginate",
        ],
        default={},
    )
    return check_runs_from_payload(data)


def commit_combined_status(owner: str, repo: str, sha: str) -> dict[str, Any]:
    if not (owner and repo and sha):
        return {}
    data = gh_json(
        ["api", f"repos/{owner}/{repo}/commits/{sha}/status", "--paginate"],
        default={},
    )
    return combined_status_from_payload(data)


def issue_comments(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    data = gh_json(
        ["api", f"repos/{owner}/{repo}/issues/{number}/comments", "--paginate"],
        default=[],
    )
    return flatten_item_pages(data)


def review_comments(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    data = gh_json(
        ["api", f"repos/{owner}/{repo}/pulls/{number}/comments", "--paginate"],
        default=[],
    )
    return flatten_item_pages(data)


def reviews(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    data = gh_json(
        ["api", f"repos/{owner}/{repo}/pulls/{number}/reviews", "--paginate"],
        default=[],
    )
    return flatten_item_pages(data)


def issue_events(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    data = gh_json(
        ["api", f"repos/{owner}/{repo}/issues/{number}/events", "--paginate"],
        default=[],
    )
    return flatten_item_pages(data)


THREAD_COMMENT_FIELDS = """
              id
              databaseId
              author { login }
              body
              createdAt
              updatedAt
              diffHunk
              path
              line
              originalLine
              outdated
              url
"""


def review_threads(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    query = f"""
query($owner: String!, $repo: String!, $number: Int!, $cursor: String) {{
  repository(owner: $owner, name: $repo) {{
    pullRequest(number: $number) {{
      reviewThreads(first: 100, after: $cursor) {{
        pageInfo {{ hasNextPage endCursor }}
        nodes {{
          id
          isResolved
          isOutdated
          path
          line
          originalLine
          comments(first: 100) {{
            pageInfo {{ hasNextPage endCursor }}
            nodes {{
{THREAD_COMMENT_FIELDS}
            }}
          }}
        }}
      }}
    }}
  }}
}}
"""
    nodes: list[dict[str, Any]] = []
    cursor = ""
    while True:
        api_args = [
            "api",
            "graphql",
            "-f",
            f"owner={owner}",
            "-f",
            f"repo={repo}",
            "-F",
            f"number={number}",
            "-f",
            f"query={query}",
        ]
        if cursor:
            api_args.extend(["-f", f"cursor={cursor}"])
        data = gh_json(api_args, default={})
        threads = (
            data.get("data", {})
            .get("repository", {})
            .get("pullRequest", {})
            .get("reviewThreads", {})
        )
        nodes.extend(threads.get("nodes") or [])
        page_info = threads.get("pageInfo") or {}
        if not page_info.get("hasNextPage"):
            break
        cursor = page_info.get("endCursor") or ""
        if not cursor:
            break
    for thread in nodes:
        extend_thread_comments(thread)
    return nodes


def extend_thread_comments(thread: dict[str, Any]) -> None:
    """Fetch comment pages after the first 100 on one review thread."""

    comments = thread.setdefault("comments", {})
    nodes = comments.setdefault("nodes", [])
    page = comments.get("pageInfo") or {}
    cursor = page.get("endCursor") or ""
    thread_id = thread.get("id")
    seen_cursors: set[str] = set()
    query = f"""
query($id: ID!, $cursor: String!) {{
  node(id: $id) {{
    ... on PullRequestReviewThread {{
      comments(first: 100, after: $cursor) {{
        pageInfo {{ hasNextPage endCursor }}
        nodes {{
{THREAD_COMMENT_FIELDS}
        }}
      }}
    }}
  }}
}}
"""
    while page.get("hasNextPage") and cursor and thread_id and cursor not in seen_cursors:
        seen_cursors.add(cursor)
        data = gh_json(
            [
                "api",
                "graphql",
                "-f",
                f"id={thread_id}",
                "-f",
                f"cursor={cursor}",
                "-f",
                f"query={query}",
            ],
            default={},
        )
        more = ((data.get("data") or {}).get("node") or {}).get("comments") or {}
        nodes.extend(more.get("nodes") or [])
        page = more.get("pageInfo") or {}
        cursor = page.get("endCursor") or ""
    comments["pageInfo"] = page


def fetch_pr_merge_refs(pr: dict[str, Any]) -> dict[str, str]:
    number = pr["number"]
    base_ref = pr["baseRefName"]
    local_base_ref = f"refs/remotes/origin/{base_ref}"
    local_pr_ref = f"refs/remotes/origin/pr-{number}-head"
    fetched = run_git(
        [
            "fetch",
            "--quiet",
            "origin",
            f"+refs/heads/{base_ref}:{local_base_ref}",
            f"+refs/pull/{number}/head:{local_pr_ref}",
        ],
        allow_failure=True,
    )
    if fetched.returncode != 0:
        raise RuntimeError(fetched.stderr.strip() or "git fetch failed")
    base_sha = run_git(["rev-parse", local_base_ref], allow_failure=True).stdout.strip()
    head_sha = run_git(["rev-parse", local_pr_ref], allow_failure=True).stdout.strip()
    return {"base": base_sha, "head": head_sha}


def parse_merge_tree_conflicts(output: str) -> list[str]:
    files: set[str] = set()
    for line in output.splitlines():
        if "\t" in line:
            metadata, path = line.split("\t", 1)
            parts = metadata.split()
            # `merge-tree --write-tree` emits `<mode> <object> <stage>`.
            if len(parts) in {3, 4} and parts[2] in {"1", "2", "3"}:
                files.add(path)
                continue
        marker = " Merge conflict in "
        if marker in line:
            files.add(line.split(marker, 1)[1])
    return sorted(files)


def merge_conflicts(pr: dict[str, Any]) -> dict[str, Any]:
    fallback_conflict = (
        pr.get("mergeable") == "CONFLICTING" or pr.get("mergeStateStatus") == "DIRTY"
    )
    try:
        refs = fetch_pr_merge_refs(pr)
        if not refs.get("base") or not refs.get("head"):
            raise RuntimeError("could not resolve PR merge refs")
        result = run_git(
            ["merge-tree", "--write-tree", refs["base"], refs["head"]],
            allow_failure=True,
        )
    except (
        KeyError,
        subprocess.CalledProcessError,
        FileNotFoundError,
        RuntimeError,
    ) as exc:
        return {
            "analysisAvailable": False,
            "hasConflicts": fallback_conflict,
            "files": [],
            "source": "github-mergeability-fallback",
            "error": str(exc),
        }
    output = "\n".join(part for part in [result.stdout, result.stderr] if part)
    files = parse_merge_tree_conflicts(output)
    return {
        "analysisAvailable": result.returncode in {0, 1},
        "hasConflicts": bool(files) or fallback_conflict,
        "files": files,
        "source": "git-merge-tree",
        "base": refs["base"],
        "head": refs["head"],
        "mergeTreeExitCode": result.returncode,
        "details": output.strip(),
    }


# ---------------------------------------------------------------------------
# State file
# ---------------------------------------------------------------------------


def default_state_file_path(pr_number: int) -> str:
    """Default to a state directory alongside this script.

    State sits in this working tree, next to the script. Each git worktree
    therefore has its own file. That is intentional: ``.git`` inside a
    worktree is a file pointing at a worktree-private git dir, and writing
    there is easy to get wrong. Watchers of the same PR must use the same
    checkout (or pass the same ``--state-file``).
    """

    state_dir = Path(__file__).resolve().parent / "state"
    return str(state_dir / f".pr-monitor-{pr_number}.json")


def load_state(path: str | None) -> dict[str, Any]:
    if not path:
        return {}
    try:
        with open(path, "r", encoding="utf-8") as fh:
            data = json.load(fh)
            if isinstance(data, dict):
                return data
    except (FileNotFoundError, json.JSONDecodeError, OSError):
        pass
    return {}


def save_state(path: str | None, state: dict[str, Any]) -> None:
    if not path:
        return
    try:
        target = Path(path)
        target.parent.mkdir(parents=True, exist_ok=True)
        with open(target, "w", encoding="utf-8") as fh:
            json.dump(state, fh, indent=2, sort_keys=True)
    except OSError:
        # State persistence is best-effort; we don't fail the poll over it.
        pass


def cap_seen_ids(ids: Iterable[Any], cap: int = SEEN_ID_CAP) -> list[Any]:
    deduped: list[Any] = []
    seen: set[Any] = set()
    for item in ids:
        if item in seen:
            continue
        seen.add(item)
        deduped.append(item)
    if len(deduped) > cap:
        deduped = deduped[-cap:]
    return deduped


# ---------------------------------------------------------------------------
# Pure helpers (used by compute_payload and unit tests)
# ---------------------------------------------------------------------------


def parse_iso(ts: str | None) -> Optional[datetime]:
    if not ts:
        return None
    try:
        if ts.endswith("Z"):
            ts = ts[:-1] + "+00:00"
        return datetime.fromisoformat(ts)
    except (TypeError, ValueError):
        return None


def now_iso() -> str:
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def normalize_check(check: dict[str, Any]) -> dict[str, Any]:
    state = (check.get("state") or check.get("status") or "").upper()
    conclusion = (check.get("conclusion") or "").upper()
    bucket = (check.get("bucket") or "").lower()

    # `gh pr checks` reports skipped jobs (the always-skipped Renovate Approve
    # job, for example) as bucket "skipping" / conclusion SKIPPED. That is not
    # a failure.
    skipped = bucket in {"skip", "skipping"} or conclusion == "SKIPPED"
    failed = (not skipped) and (
        bucket in {"fail", "cancel"}
        or conclusion
        in {
            "ACTION_REQUIRED",
            "CANCELLED",
            "FAILURE",
            "STALE",
            "TIMED_OUT",
        }
    )
    pending = (not skipped) and (
        bucket in {"pending", "running", "unknown"}
        or state
        in {
            "EXPECTED",
            "IN_PROGRESS",
            "PENDING",
            "QUEUED",
            "REQUESTED",
            "WAITING",
        }
    )
    passed = (not skipped) and (bucket == "pass" or conclusion == "SUCCESS")

    return {
        **check,
        "derived": {
            "failed": failed,
            "pending": pending and not failed,
            "passed": passed and not failed,
        },
    }


def normalize_check_run(run: dict[str, Any]) -> dict[str, Any]:
    """Normalize a REST `check-runs` entry into the same `derived` shape."""

    status = (run.get("status") or "").upper()
    conclusion = (run.get("conclusion") or "").upper()
    failed = conclusion in {
        "ACTION_REQUIRED",
        "CANCELLED",
        "FAILURE",
        "TIMED_OUT",
        "STALE",
    }
    pending = status in {"QUEUED", "IN_PROGRESS", "PENDING", "REQUESTED", "WAITING"}
    passed = conclusion == "SUCCESS"
    return {
        **run,
        "derived": {
            "failed": failed,
            "pending": pending and not failed,
            "passed": passed and not failed,
        },
    }


def normalize_combined_status_entry(entry: dict[str, Any]) -> dict[str, Any]:
    """Normalize a combined-status entry.

    The REST status API uses `context` and `target_url`. Check runs and
    `gh pr checks` use `name` and `html_url`. Copy the status fields onto
    those keys so dedup and classification see Buildkite and CLA.
    """

    state = (entry.get("state") or "").upper()
    failed = state in {"FAILURE", "ERROR"}
    pending = state == "PENDING"
    passed = state == "SUCCESS"
    return {
        **entry,
        "name": entry.get("name") or entry.get("context"),
        "html_url": entry.get("html_url") or entry.get("target_url") or "",
        "derived": {
            "failed": failed,
            "pending": pending and not failed,
            "passed": passed and not failed,
        },
    }


def check_name(check: dict[str, Any]) -> Optional[str]:
    name = check.get("name") or check.get("context")
    if name is None:
        return None
    return str(name)


def classify_check_name(name: Optional[str]) -> str:
    if name in AUTO_FIXABLE_CHECK_NAMES:
        return "auto-fixable"
    if name in OUT_OF_BAND_CHECK_NAMES:
        return "out-of-band"
    return "unknown"


def check_url(check: dict[str, Any]) -> str:
    return (
        check.get("html_url")
        or check.get("detailsUrl")
        or check.get("target_url")
        or check.get("link")
        or ""
    )


def out_of_band_state(check: dict[str, Any]) -> str:
    derived = check["derived"]
    if derived["failed"]:
        return "failed"
    if derived["pending"]:
        return "pending"
    if derived["passed"]:
        return "passed"
    return "skipped"


def out_of_band_settled(out_of_band: list[dict[str, Any]]) -> bool:
    """True once acceptance has a non-pending state and no out-of-band check is pending.

    A missing acceptance status is not settled: Buildkite may not have posted it
    yet. Release is waited on only when its status is already present.
    """

    saw_acceptance = False
    for item in out_of_band:
        if item.get("state") == "pending":
            return False
        if item.get("name") == ACCEPTANCE_CHECK_NAME:
            saw_acceptance = True
    return saw_acceptance


def check_timestamp(check: dict[str, Any]) -> Optional[datetime]:
    for key in ("started_at", "startedAt", "updated_at", "created_at"):
        ts = parse_iso(check.get(key))
        if ts is not None:
            return ts
    return None


def detect_verify_workflow(root: Optional[str] = None) -> bool:
    """True when a verify-openspec workflow file exists in this repo.

    `CHECK_PR_STATE_VERIFY_WORKFLOW=1` or `0` overrides the filesystem probe.
    The workflow lands in cp-hosted-team#4386; until then this stays false.
    """

    override = os.environ.get("CHECK_PR_STATE_VERIFY_WORKFLOW")
    if override == "1":
        return True
    if override == "0":
        return False
    if root is None:
        completed = run_git(["rev-parse", "--show-toplevel"], allow_failure=True)
        if completed.returncode != 0:
            return False
        root = completed.stdout.strip()
    if not root:
        return False
    return any(Path(root).glob(VERIFY_WORKFLOW_GLOB))


def derive_verify_openspec(
    review_data: list[dict[str, Any]],
    event_data: list[dict[str, Any]],
    head_sha: str,
    openspec_change: Optional[str] = None,
) -> dict[str, Any]:
    """Compute verify-openspec workflow state.

    The verify-openspec workflow removes its own label as soon as it starts
    processing, so label absence is NOT a signal to re-request. We derive
    runState from the labeled/unlabeled timeline plus the bot's reviews.

    Approvals are permanent — once a verify-openspec APPROVED review exists,
    it never goes stale. The only way to re-trigger is to re-apply the label.
    """

    def is_verify_review(review: dict[str, Any]) -> bool:
        login = ((review.get("user") or {}).get("login") or "").lower()
        if login not in {a.lower() for a in VERIFY_OPENSPEC_REVIEW_AUTHORS}:
            return False
        body = review.get("body") or ""
        return any(marker in body for marker in VERIFY_OPENSPEC_REVIEW_BODY_MARKERS)

    label_applied_at: Optional[str] = None
    label_removed_at: Optional[str] = None
    for event in event_data:
        ev_type = event.get("event")
        label = (event.get("label") or {}).get("name")
        if label != VERIFY_OPENSPEC_LABEL:
            continue
        created = event.get("created_at")
        if ev_type == "labeled":
            if not label_applied_at or (
                parse_iso(created) or datetime.min.replace(tzinfo=timezone.utc)
            ) > (
                parse_iso(label_applied_at) or datetime.min.replace(tzinfo=timezone.utc)
            ):
                label_applied_at = created
        elif ev_type == "unlabeled":
            if not label_removed_at or (
                parse_iso(created) or datetime.min.replace(tzinfo=timezone.utc)
            ) > (
                parse_iso(label_removed_at) or datetime.min.replace(tzinfo=timezone.utc)
            ):
                label_removed_at = created

    all_verify_reviews = sorted(
        [r for r in review_data if is_verify_review(r)],
        key=lambda r: (
            parse_iso(r.get("submitted_at"))
            or datetime.min.replace(tzinfo=timezone.utc)
        ),
    )
    change_id = (openspec_change or "").strip()
    verify_reviews = [
        review
        for review in all_verify_reviews
        if not change_id or f"`{change_id}`" in (review.get("body") or "")
    ]

    last_verify_review = verify_reviews[-1] if verify_reviews else None
    last_verify_review_at = (
        parse_iso(last_verify_review.get("submitted_at"))
        if last_verify_review
        else None
    )

    last_approval = None
    for r in reversed(verify_reviews):
        if (r.get("state") or "").upper() == "APPROVED":
            last_approval = r
            break

    last_approval_at = (last_approval or {}).get("submitted_at")
    last_approval_head = (last_approval or {}).get("commit_id")

    label_applied_dt = parse_iso(label_applied_at)
    label_removed_dt = parse_iso(label_removed_at)
    label_active = label_applied_dt is not None and (
        label_removed_dt is None or label_removed_dt < label_applied_dt
    )
    label_consumed = (
        label_applied_dt is not None
        and label_removed_dt is not None
        and label_removed_dt >= label_applied_dt
    )

    # A finished label cycle whose report names a different change does not
    # keep this change in-progress.
    if (
        change_id
        and label_consumed
        and label_applied_dt is not None
        and not verify_reviews
        and any(
            (
                parse_iso(review.get("submitted_at"))
                or datetime.min.replace(tzinfo=timezone.utc)
            )
            >= label_applied_dt
            for review in all_verify_reviews
        )
    ):
        label_applied_at = None
        label_removed_at = None
        label_applied_dt = None
        label_removed_dt = None
        label_active = False
        label_consumed = False

    # If a label was applied after the most recent verify review, the
    # workflow has been re-requested.
    if (
        label_applied_dt is not None
        and last_verify_review_at is not None
        and label_applied_dt > last_verify_review_at
    ):
        run_state = "pending-pickup" if label_active else "in-progress"
    elif verify_reviews:
        last_state = (last_verify_review.get("state") or "").upper()
        if last_state == "APPROVED":
            run_state = "approved"
        elif last_state == "CHANGES_REQUESTED":
            run_state = "changes-requested"
        else:
            run_state = "none"
    elif label_active:
        run_state = "pending-pickup"
    elif label_consumed:
        run_state = "in-progress"
    else:
        run_state = "none"

    return {
        "lastLabelAppliedAt": label_applied_at,
        "lastLabelRemovedAt": label_removed_at,
        "lastApprovalAt": last_approval_at,
        "lastApprovalHeadSha": last_approval_head,
        "runState": run_state,
    }


def derive_latest_by_reviewer(
    review_data: list[dict[str, Any]],
) -> dict[str, dict[str, Any]]:
    latest: dict[str, dict[str, Any]] = {}
    for review in review_data:
        login = ((review.get("user") or {}).get("login") or "").lower()
        if not login:
            continue
        state = (review.get("state") or "").upper()
        # Ignore COMMENTED-only reviews for "decision" purposes.
        if state not in {"APPROVED", "CHANGES_REQUESTED", "DISMISSED"}:
            continue
        existing = latest.get(login)
        existing_dt = parse_iso((existing or {}).get("submittedAt"))
        candidate_dt = parse_iso(review.get("submitted_at"))
        if existing is None or (
            candidate_dt is not None
            and (existing_dt is None or candidate_dt > existing_dt)
        ):
            latest[login] = {
                "state": state,
                "submittedAt": review.get("submitted_at"),
                "id": review.get("id"),
                "commitId": review.get("commit_id"),
            }
    return latest


def derive_effective_decision(latest_by_reviewer: dict[str, dict[str, Any]]) -> str:
    """Return APPROVED, CHANGES_REQUESTED, or REVIEW_REQUIRED."""

    if not latest_by_reviewer:
        return "REVIEW_REQUIRED"
    states = {entry["state"] for entry in latest_by_reviewer.values()}
    if "CHANGES_REQUESTED" in states:
        return "CHANGES_REQUESTED"
    if "APPROVED" in states:
        return "APPROVED"
    return "REVIEW_REQUIRED"


def select_new(
    items: list[dict[str, Any]],
    *,
    id_key: str,
    seen_ids: set[Any],
    timestamp_keys: list[str],
    since_dt: Optional[datetime],
) -> list[dict[str, Any]]:
    """Return items not in seen_ids and (when applicable) newer than `since`."""

    out: list[dict[str, Any]] = []
    for item in items:
        item_id = item.get(id_key)
        if item_id is not None and item_id in seen_ids:
            continue
        if since_dt is not None:
            ts: Optional[datetime] = None
            for key in timestamp_keys:
                ts = parse_iso(item.get(key))
                if ts is not None:
                    break
            if ts is not None and ts <= since_dt:
                # Older than `since` AND not previously seen — treat as new on
                # first run (no state file). Honour `since` only when a since
                # was explicitly provided.
                continue
        out.append(item)
    return out


def thread_root_comment_id(thread: dict[str, Any]) -> Optional[Any]:
    nodes = (thread.get("comments") or {}).get("nodes") or []
    if not nodes:
        return None
    return nodes[0].get("id")


def thread_last_updated(thread: dict[str, Any]) -> Optional[datetime]:
    latest: Optional[datetime] = None
    for comment in (thread.get("comments") or {}).get("nodes") or []:
        for key in ("updatedAt", "createdAt"):
            ts = parse_iso(comment.get(key))
            if ts is not None and (latest is None or ts > latest):
                latest = ts
                break
    return latest


# ---------------------------------------------------------------------------
# Payload computation (pure: no I/O, easy to test)
# ---------------------------------------------------------------------------


def compute_payload(
    *,
    repo: dict[str, str],
    pr: dict[str, Any],
    pr_check_data: list[dict[str, Any]],
    commit_check_run_data: list[dict[str, Any]],
    commit_status_data: dict[str, Any],
    review_data: list[dict[str, Any]],
    issue_comment_data: list[dict[str, Any]],
    review_comment_data: list[dict[str, Any]],
    thread_data: list[dict[str, Any]],
    event_data: list[dict[str, Any]],
    merge_conflict_data: dict[str, Any],
    state: dict[str, Any],
    head_sha_override: Optional[str] = None,
    since_override: Optional[str] = None,
    openspec_change: Optional[str] = None,
    verify_workflow_present: bool = False,
    retain_stale_since: bool = False,
) -> tuple[dict[str, Any], dict[str, Any]]:
    """Build the JSON payload and the next-state-file content."""

    head_sha = head_sha_override or pr.get("headRefOid") or ""

    # --- stale-state invalidation -----------------------------------------
    # If the PR head has changed since the last poll, the old lastPolledAt
    # is no longer a valid baseline for "new since" discrimination.
    # Reset it (but keep seen IDs so exact duplicates are still filtered).
    last_head_sha_in_state = state.get("lastHeadSha")
    if last_head_sha_in_state is not None and last_head_sha_in_state != head_sha:
        state = {**state, "lastPolledAt": None}

    # Also reset lastPolledAt if it is older than the TTL. This handles the
    # case where a watcher was killed and restarted hours later — old
    # timestamp discrimination would hide updates that happened in between.
    # --watch-acceptance freezes lastPolledAt on purpose for the whole run
    # (up to 3 hours), so that reset must not apply there.
    since_str = since_override or state.get("lastPolledAt")
    since_dt = parse_iso(since_str)
    if since_dt is not None and not retain_stale_since:
        age_seconds = (datetime.now(timezone.utc) - since_dt).total_seconds()
        if age_seconds > STATE_LAST_POLLED_TTL_SECONDS:
            state = {**state, "lastPolledAt": None}
            since_dt = None

    # --- normalize checks -------------------------------------------------
    raw_pr_checks = [normalize_check(c) for c in pr_check_data]
    pinned_check_runs = [normalize_check_run(r) for r in commit_check_run_data]
    pinned_statuses = [
        normalize_combined_status_entry(s)
        for s in (commit_status_data.get("statuses") or [])
    ]

    # Prefer commit-pinned data. Fall back to gh pr checks only when the
    # caller did not pin a SHA. An override with no statuses must stay empty
    # rather than borrow the current head's rollup.
    pinned_combined = pinned_check_runs + pinned_statuses
    pinned_to_commit = bool(pinned_combined or head_sha_override)
    canonical_checks = pinned_combined if pinned_to_commit else raw_pr_checks

    # Deduplicate check runs by name, preferring the most recent one.
    # This handles re-runs for the same commit where old runs remain in the API.
    def deduplicate_checks(checks):
        by_name = {}
        for c in checks:
            name = check_name(c)
            if name is None:
                continue
            existing = by_name.get(name)
            existing_ts = check_timestamp(existing) if existing else None
            candidate_ts = check_timestamp(c)
            if existing is None or (
                candidate_ts is not None
                and (existing_ts is None or candidate_ts > existing_ts)
            ):
                by_name[name] = c
        return list(by_name.values())

    canonical_checks = deduplicate_checks(canonical_checks)

    # Fallback: if deduplication left us with nothing (e.g. every check
    # entry lacked a name), fall back to the raw pr-checks list so we
    # never report null/empty counts when check data is available.
    if not canonical_checks and not pinned_to_commit:
        canonical_checks = raw_pr_checks
    canonical_checks = [
        {**c, "class": classify_check_name(check_name(c))} for c in canonical_checks
    ]
    # Out-of-band Buildkite checks are reported separately. They do not
    # contribute to failed/pending/passed, so a pending acceptance run does
    # not block verify-openspec and a red acceptance run is not auto-fixable.
    out_of_band_checks = [c for c in canonical_checks if c["class"] == "out-of-band"]
    counted_checks = [c for c in canonical_checks if c["class"] != "out-of-band"]
    failed_checks = [c for c in counted_checks if c["derived"]["failed"]]
    pending_checks = [c for c in counted_checks if c["derived"]["pending"]]
    passed_checks = [c for c in counted_checks if c["derived"]["passed"]]
    all_failed_checks = [c for c in canonical_checks if c["derived"]["failed"]]

    # --- new-since discrimination ----------------------------------------
    seen_issue_ids: set[Any] = set(state.get("seenIssueCommentIds") or [])
    seen_review_comment_ids: set[Any] = set(state.get("seenReviewCommentIds") or [])
    seen_review_ids: set[Any] = set(state.get("seenReviewIds") or [])
    seen_thread_ids: set[Any] = set(state.get("seenReviewThreadIds") or [])

    head_pushed_recently = (
        last_head_sha_in_state is not None and last_head_sha_in_state != head_sha
    )

    # The state watermark must not hide an id this poll has never recorded.
    # `--since` is the only timestamp cutoff for comments and reviews.
    explicit_since = parse_iso(since_override) if since_override else None
    new_issue_comments = select_new(
        issue_comment_data,
        id_key="id",
        seen_ids=seen_issue_ids,
        timestamp_keys=["created_at", "updated_at"],
        since_dt=explicit_since,
    )
    new_review_comments = select_new(
        review_comment_data,
        id_key="id",
        seen_ids=seen_review_comment_ids,
        timestamp_keys=["created_at", "updated_at"],
        since_dt=explicit_since,
    )
    new_reviews = select_new(
        review_data,
        id_key="id",
        seen_ids=seen_review_ids,
        timestamp_keys=["submitted_at"],
        since_dt=explicit_since,
    )

    # --- threads ----------------------------------------------------------
    # An unresolved thread stays actionable until it is resolved or outdated,
    # including after a push that clears lastPolledAt.
    unresolved_threads = [
        t for t in thread_data if not t.get("isResolved") and not t.get("isOutdated")
    ]
    seen_thread_ids -= {t.get("id") for t in unresolved_threads}
    unresolved_new = []
    unresolved_updated_since_head = []
    for thread in unresolved_threads:
        thread_id = thread.get("id")
        if thread_id and thread_id not in seen_thread_ids:
            unresolved_new.append(thread)
        last_updated = thread_last_updated(thread)
        if (
            last_updated is not None
            and since_dt is not None
            and last_updated > since_dt
        ):
            if thread not in unresolved_new:
                unresolved_updated_since_head.append(thread)

    # A comment on a review thread is reported with that thread. Counting it
    # again as a standalone review comment leaves an empty actionable signal
    # after the thread is resolved.
    thread_comment_ids: set[Any] = set()
    for thread in thread_data:
        for comment in (thread.get("comments") or {}).get("nodes") or []:
            comment_id = comment.get("databaseId")
            if comment_id is not None:
                thread_comment_ids.add(comment_id)
    if thread_comment_ids:
        new_review_comments = [
            comment
            for comment in new_review_comments
            if comment.get("id") not in thread_comment_ids
        ]

    # --- reviews ---------------------------------------------------------
    latest_by_reviewer = derive_latest_by_reviewer(review_data)
    effective_decision = derive_effective_decision(latest_by_reviewer)
    verify_openspec = derive_verify_openspec(
        review_data, event_data, head_sha, openspec_change=openspec_change
    )

    # --- merge state -----------------------------------------------------
    merge_state = pr.get("mergeStateStatus")
    mergeable = pr.get("mergeable")
    has_merge_conflicts = bool(merge_conflict_data.get("hasConflicts"))
    merge_blocked = has_merge_conflicts or merge_state in {
        "BEHIND",
        "BLOCKED",
        "UNKNOWN",
        "UNSTABLE",
    }
    last_merge_state = state.get("lastMergeStateStatus")

    # --- actionable list (rebased on new* / effectiveDecision) -----------
    actionable: list[str] = []
    if failed_checks:
        actionable.append("failed_checks")
    if any(c["derived"]["failed"] for c in out_of_band_checks):
        actionable.append("acceptance_failed")
    if new_issue_comments:
        actionable.append("issue_comments")
    if new_review_comments:
        actionable.append("review_comments")
    if unresolved_new or unresolved_updated_since_head:
        actionable.append("unresolved_review_threads")
    if effective_decision == "CHANGES_REQUESTED":
        actionable.append("changes_requested")
    if any(
        (review.get("state") or "").upper() == "COMMENTED"
        and (review.get("body") or "").strip()
        for review in new_reviews
    ):
        actionable.append("commented_reviews")
    if has_merge_conflicts:
        actionable.append("merge_conflicts")
    elif merge_state in {"BEHIND", "UNKNOWN", "UNSTABLE"}:
        actionable.append("merge_or_branch_state")
    elif (
        merge_state == "BLOCKED"
        and last_merge_state is not None
        and last_merge_state != "BLOCKED"
    ):
        actionable.append("merge_or_branch_state")

    summary = {
        "actionable": actionable,
        "hasActionable": bool(actionable),
        "pr": {
            "number": pr.get("number"),
            "url": pr.get("url"),
            "title": pr.get("title"),
            "state": pr.get("state"),
            "isDraft": pr.get("isDraft"),
            "headRefName": pr.get("headRefName"),
            "headRefOid": pr.get("headRefOid"),
            "baseRefName": pr.get("baseRefName"),
            "baseRefOid": pr.get("baseRefOid"),
            "reviewDecision": pr.get("reviewDecision"),
            "mergeable": mergeable,
            "mergeStateStatus": merge_state,
            "labels": [
                label.get("name")
                for label in pr.get("labels", [])
                if isinstance(label, dict)
            ],
        },
        "checks": {
            "source": "commit-pinned" if pinned_to_commit else "pr-checks",
            "headSha": head_sha,
            "total": len(canonical_checks),
            "failed": len(failed_checks),
            "pending": len(pending_checks),
            "passed": len(passed_checks),
            "failedChecks": [
                {
                    "name": check_name(c),
                    "url": check_url(c),
                    "class": c.get("class"),
                }
                for c in all_failed_checks
            ],
            "failedNames": [check_name(c) for c in failed_checks],
            "pendingNames": [check_name(c) for c in pending_checks],
            "passedNames": [check_name(c) for c in passed_checks],
            "requiredPassed": all(
                name in {check_name(c) for c in passed_checks}
                for name in REQUIRED_PASSED_CHECK_NAMES
            ),
            "outOfBand": [
                {
                    "name": check_name(c),
                    "state": out_of_band_state(c),
                    "url": check_url(c),
                }
                for c in out_of_band_checks
            ],
        },
        "comments": {
            "issueComments": len(issue_comment_data),
            "reviewComments": len(review_comment_data),
            "newIssueComments": len(new_issue_comments),
            "newIssueCommentIds": [c.get("id") for c in new_issue_comments],
            "newReviewComments": len(new_review_comments),
            "newReviewCommentIds": [c.get("id") for c in new_review_comments],
        },
        "threads": {
            "unresolved": len(unresolved_threads),
            "unresolvedNew": len(unresolved_new),
            "unresolvedUpdatedSinceHead": len(unresolved_updated_since_head),
            "unresolvedThreadIds": [t.get("id") for t in unresolved_threads],
            "unresolvedNewThreadIds": [t.get("id") for t in unresolved_new],
            "unresolvedUpdatedSinceHeadThreadIds": [
                t.get("id") for t in unresolved_updated_since_head
            ],
        },
        "reviews": {
            "total": len(review_data),
            "newReviewIds": [r.get("id") for r in new_reviews],
            "latestByReviewer": latest_by_reviewer,
            "effectiveDecision": effective_decision,
            "verifyOpenspec": verify_openspec,
        },
        "merge": {
            "blocked": merge_blocked,
            "hasConflicts": has_merge_conflicts,
            "conflictFiles": merge_conflict_data.get("files", []),
            "conflictAnalysisAvailable": merge_conflict_data.get(
                "analysisAvailable", False
            ),
            "mergeable": mergeable,
            "mergeStateStatus": merge_state,
        },
        "totals": {
            "issueComments": len(issue_comment_data),
            "reviewComments": len(review_comment_data),
            "reviews": len(review_data),
            "unresolvedThreads": len(unresolved_threads),
            "changesRequestedReviews": sum(
                1
                for r in review_data
                if (r.get("state") or "").upper() == "CHANGES_REQUESTED"
            ),
            "approvedReviews": sum(
                1 for r in review_data if (r.get("state") or "").upper() == "APPROVED"
            ),
        },
        "headPushedRecently": head_pushed_recently,
    }

    # Compute the single boolean agents read for verify-openspec label eligibility.
    # The caller must name the in-progress change, and the verify workflow file
    # must exist. Out-of-band acceptance does not count toward failed/pending;
    # an acceptance failure is still actionable, so it blocks the label.
    change_named = bool(openspec_change and str(openspec_change).strip())
    summary["reviews"]["verifyOpenspec"]["openspecChange"] = (
        str(openspec_change).strip() if change_named else None
    )
    summary["reviews"]["verifyOpenspec"]["verifyWorkflowPresent"] = verify_workflow_present
    # A passed in-band GitHub Actions check is required. checks.total counts
    # out-of-band statuses, so a lone pending acceptance run is not enough.
    # Pending acceptance still does not clear the flag once an in-band check
    # has passed.
    in_band_passed = any(
        c.get("class") == "auto-fixable" and c["derived"]["passed"]
        for c in counted_checks
    )
    summary["reviews"]["verifyOpenspec"]["requiresOpenspecVerification"] = (
        change_named
        and verify_workflow_present
        and verify_openspec["runState"] == "none"
        and in_band_passed
        and summary["checks"]["failed"] == 0
        and summary["checks"]["pending"] == 0
        and not actionable
        and effective_decision != "CHANGES_REQUESTED"
    )

    payload = {
        "repository": repo,
        "pr": pr,
        "checks": {
            "prChecks": raw_pr_checks,
            "prChecksPinned": [
                c
                for c in raw_pr_checks
                # gh pr checks doesn't expose per-row commit; we emit the same
                # list and rely on commit-pinned arrays below for canonical data
            ],
            "commitCheckRuns": pinned_check_runs,
            "commitStatuses": pinned_statuses,
            "statusCheckRollup": pr.get("statusCheckRollup", []),
            "headSha": head_sha,
        },
        "reviews": review_data,
        "issue_comments": issue_comment_data,
        "review_comments": review_comment_data,
        "review_threads": thread_data,
        "issue_events": event_data,
        "merge_conflicts": merge_conflict_data,
        "summary": summary,
    }

    # --- next state ------------------------------------------------------
    new_state = {
        "pr": pr.get("number"),
        "lastPolledAt": now_iso(),
        "lastHeadSha": head_sha,
        "seenIssueCommentIds": cap_seen_ids(
            list(seen_issue_ids)
            + [c.get("id") for c in issue_comment_data if c.get("id") is not None]
        ),
        "seenReviewCommentIds": cap_seen_ids(
            list(seen_review_comment_ids)
            + [c.get("id") for c in review_comment_data if c.get("id") is not None]
        ),
        "seenReviewIds": cap_seen_ids(
            list(seen_review_ids)
            + [r.get("id") for r in review_data if r.get("id") is not None]
        ),
        "seenReviewThreadIds": cap_seen_ids(
            list(seen_thread_ids)
            + [
                t.get("id")
                for t in thread_data
                if t.get("id") is not None
                and (t.get("isResolved") or t.get("isOutdated"))
            ]
        ),
        "lastVerifyOpenspecLabelAppliedAt": verify_openspec.get("lastLabelAppliedAt"),
        "lastVerifyOpenspecLabelRemovedAt": verify_openspec.get("lastLabelRemovedAt"),
        "lastMergeStateStatus": merge_state,
    }

    return payload, new_state


# ---------------------------------------------------------------------------
# Focused output formatter
# ---------------------------------------------------------------------------


def format_focused(payload: dict[str, Any]) -> dict[str, Any]:
    """Return a focused subset of the payload for agent consumption.

    The focused output omits raw data arrays (reviews, issue_comments,
    review_comments, review_threads, issue_events, merge_conflicts details,
    raw check arrays) and surfaces only actionable decision data.
    """
    summary = payload.get("summary", {})

    # PR metadata
    pr_summary = summary.get("pr", {})
    focused_pr = {
        "number": pr_summary.get("number"),
        "url": pr_summary.get("url"),
        "title": pr_summary.get("title"),
        "headRefName": pr_summary.get("headRefName"),
        "headRefOid": pr_summary.get("headRefOid"),
        "mergeable": pr_summary.get("mergeable"),
        "mergeStateStatus": pr_summary.get("mergeStateStatus"),
        "state": pr_summary.get("state"),
        "isDraft": pr_summary.get("isDraft"),
        "labels": pr_summary.get("labels"),
    }

    # Checks
    checks_summary = summary.get("checks", {})
    focused_checks = {
        "source": checks_summary.get("source"),
        "headSha": checks_summary.get("headSha"),
        "total": checks_summary.get("total"),
        "failed": checks_summary.get("failed"),
        "pending": checks_summary.get("pending"),
        "passed": checks_summary.get("passed"),
        "failedChecks": checks_summary.get("failedChecks", []),
        "failedNames": checks_summary.get("failedNames", []),
        "pendingNames": checks_summary.get("pendingNames", []),
        "passedNames": checks_summary.get("passedNames", []),
        "requiredPassed": checks_summary.get("requiredPassed", False),
        "outOfBand": checks_summary.get("outOfBand", []),
    }

    # Comments: build arrays with author + body for new items
    comments_summary = summary.get("comments", {})
    new_issue_comment_ids = set(comments_summary.get("newIssueCommentIds", []))
    new_review_comment_ids = set(comments_summary.get("newReviewCommentIds", []))

    new_issue_comments = []
    for c in payload.get("issue_comments", []):
        if c.get("id") in new_issue_comment_ids:
            new_issue_comments.append(
                {
                    "id": c.get("id"),
                    "author": (c.get("user") or {}).get("login"),
                    "body": c.get("body"),
                }
            )

    # Collect comment IDs that already appear in threadDetails so we don't
    # duplicate them in the standalone newReviewComments list.
    thread_detail_comment_ids: set[Any] = set()
    for thread in payload.get("review_threads", []):
        for comment in (thread.get("comments") or {}).get("nodes") or []:
            thread_detail_comment_ids.add(comment.get("databaseId"))

    new_review_comments = []
    for c in payload.get("review_comments", []):
        if (
            c.get("id") in new_review_comment_ids
            and c.get("id") not in thread_detail_comment_ids
        ):
            new_review_comments.append(
                {
                    "id": c.get("id"),
                    "author": (c.get("user") or {}).get("login"),
                    "body": c.get("body"),
                }
            )

    focused_comments = {
        "totalIssueComments": comments_summary.get("issueComments"),
        "totalReviewComments": comments_summary.get("reviewComments"),
        "newIssueCommentIds": comments_summary.get("newIssueCommentIds", []),
        "newReviewCommentIds": comments_summary.get("newReviewCommentIds", []),
        "newIssueComments": new_issue_comments,
        "newReviewComments": new_review_comments,
    }

    # Threads
    threads_summary = summary.get("threads", {})
    focused_threads = {
        "unresolved": threads_summary.get("unresolved"),
        "unresolvedNew": threads_summary.get("unresolvedNew"),
        "unresolvedUpdatedSinceHead": threads_summary.get("unresolvedUpdatedSinceHead"),
        "unresolvedThreadIds": threads_summary.get("unresolvedThreadIds", []),
        "unresolvedNewThreadIds": threads_summary.get("unresolvedNewThreadIds", []),
    }

    # Thread details: only new / updated unresolved threads
    thread_ids_to_include = set(
        threads_summary.get("unresolvedNewThreadIds", [])
        + threads_summary.get("unresolvedUpdatedSinceHeadThreadIds", [])
    )
    thread_details: dict[str, Any] = {}
    for thread in payload.get("review_threads", []):
        tid = thread.get("id")
        if tid not in thread_ids_to_include:
            continue
        comments = []
        for c in (thread.get("comments") or {}).get("nodes") or []:
            comments.append(
                {
                    "author": (c.get("author") or {}).get("login"),
                    "body": c.get("body"),
                    "databaseId": c.get("databaseId"),
                }
            )
        thread_details[tid] = {
            "path": thread.get("path"),
            "line": thread.get("line"),
            "resolved": thread.get("isResolved"),
            "outdated": thread.get("isOutdated"),
            "comments": comments,
        }

    # Reviews
    reviews_summary = summary.get("reviews", {})
    new_review_ids = set(reviews_summary.get("newReviewIds", []))
    new_reviews = []
    for r in payload.get("reviews", []):
        if r.get("id") in new_review_ids:
            entry = {
                "author": (r.get("user") or {}).get("login"),
                "state": r.get("state"),
                "submittedAt": r.get("submitted_at"),
                "id": r.get("id"),
            }
            if (review_body := (r.get("body") or "").strip()):
                entry["body"] = review_body
            new_reviews.append(entry)

    focused_reviews = {
        "total": reviews_summary.get("total"),
        "newReviewIds": reviews_summary.get("newReviewIds", []),
        "effectiveDecision": reviews_summary.get("effectiveDecision"),
        "latestByReviewer": reviews_summary.get("latestByReviewer", {}),
        "newReviews": new_reviews,
    }

    # Verify-openspec
    verify_summary = reviews_summary.get("verifyOpenspec", {})
    focused_verify = {
        "runState": verify_summary.get("runState"),
        "requiresOpenspecVerification": verify_summary.get(
            "requiresOpenspecVerification", False
        ),
        "openspecChange": verify_summary.get("openspecChange"),
        "verifyWorkflowPresent": verify_summary.get("verifyWorkflowPresent", False),
    }

    # Merge
    merge_summary = summary.get("merge", {})
    focused_merge = {
        "blocked": merge_summary.get("blocked"),
        "hasConflicts": merge_summary.get("hasConflicts"),
        "conflictFiles": merge_summary.get("conflictFiles", []),
        "conflictAnalysisAvailable": merge_summary.get("conflictAnalysisAvailable"),
        "mergeable": merge_summary.get("mergeable"),
        "mergeStateStatus": merge_summary.get("mergeStateStatus"),
    }

    return {
        "pr": focused_pr,
        "checks": focused_checks,
        "comments": focused_comments,
        "threads": focused_threads,
        "threadDetails": thread_details,
        "reviews": focused_reviews,
        "verifyOpenspec": focused_verify,
        "merge": focused_merge,
        "actionable": summary.get("actionable", []),
        "hasActionable": summary.get("hasActionable", False),
        "headPushedRecently": summary.get("headPushedRecently", False),
    }


# ---------------------------------------------------------------------------
# Top-level fetch + run
# ---------------------------------------------------------------------------


def fetch_all(
    pr_arg: str, head_sha_override: Optional[str] = None
) -> dict[str, Any]:
    """Fetch every piece of remote state once. Raises TransientGhError."""

    repo = repo_info()
    pr = pr_view(pr_arg)
    if not pr or not pr.get("number"):
        raise TransientGhError(
            json.dumps(
                {
                    "error": "pr view returned no data",
                    "pr": pr_arg,
                    "transient": True,
                }
            )
        )

    number = int(pr["number"])
    head_sha = head_sha_override or pr.get("headRefOid") or ""

    return {
        "repo": repo,
        "pr": pr,
        "pr_check_data": pr_checks(pr_arg),
        "commit_check_run_data": commit_check_runs(
            repo["owner"], repo["name"], head_sha
        ),
        "commit_status_data": commit_combined_status(
            repo["owner"], repo["name"], head_sha
        ),
        "issue_comment_data": issue_comments(repo["owner"], repo["name"], number),
        "review_comment_data": review_comments(repo["owner"], repo["name"], number),
        "review_data": reviews(repo["owner"], repo["name"], number),
        "thread_data": review_threads(repo["owner"], repo["name"], number),
        "event_data": issue_events(repo["owner"], repo["name"], number),
        "merge_conflict_data": merge_conflicts(pr),
    }


def run_once(
    args: argparse.Namespace, state_path: Optional[str]
) -> tuple[dict[str, Any], dict[str, Any]]:
    raw = fetch_all(args.pr, args.head_sha)
    state = load_state(state_path)
    payload, new_state = compute_payload(
        **raw,
        state=state,
        head_sha_override=args.head_sha,
        since_override=args.since,
        openspec_change=args.openspec_change,
        verify_workflow_present=detect_verify_workflow(),
    )
    save_state(state_path, new_state)
    return payload, new_state


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Return one JSON payload describing actionable PR state."
    )
    parser.add_argument("pr", help="Pull request number, URL, or branch accepted by gh")
    parser.add_argument(
        "--state-file",
        default=None,
        help=(
            "Path to the persistent state file. Defaults to "
            ".agents/skills/pr-monitoring-loop/scripts/state/.pr-monitor-<pr>.json."
        ),
    )
    parser.add_argument(
        "--no-state",
        action="store_true",
        help="Disable the state file entirely (every comment counts as new).",
    )
    parser.add_argument(
        "--since",
        default=None,
        help="ISO8601 timestamp; overrides the state file's lastPolledAt for new-since detection.",
    )
    parser.add_argument(
        "--head-sha",
        default=None,
        help="Override the SHA used to pin checks (defaults to pr.headRefOid).",
    )
    parser.add_argument(
        "--openspec-change",
        default=None,
        help=(
            "In-progress OpenSpec change id. requiresOpenspecVerification stays "
            "false unless this is set and the verify workflow file exists."
        ),
    )
    parser.add_argument(
        "--full-payload",
        action="store_true",
        help="Emit the full raw payload instead of the default focused output.",
    )
    watch_mode = parser.add_mutually_exclusive_group()
    watch_mode.add_argument(
        "--watch",
        action="store_true",
        help="Poll until actionable state appears or --max-duration elapses.",
    )
    watch_mode.add_argument(
        "--watch-acceptance",
        action="store_true",
        help=(
            "Poll until Buildkite acceptance (and release, if already posted) "
            "leaves pending, or --max-duration elapses. Ignores other actionable "
            f"signals. Defaults: interval {ACCEPTANCE_INTERVAL_SECONDS}s, "
            f"max {ACCEPTANCE_MAX_DURATION_SECONDS}s."
        ),
    )
    parser.add_argument(
        "--interval",
        type=int,
        default=None,
        help=(
            "Seconds between polls. Default "
            f"{DEFAULT_INTERVAL_SECONDS} for --watch, "
            f"{ACCEPTANCE_INTERVAL_SECONDS} for --watch-acceptance."
        ),
    )
    parser.add_argument(
        "--max-duration",
        type=int,
        default=None,
        help=(
            "Maximum total seconds. Default "
            f"{DEFAULT_MAX_DURATION_SECONDS} for --watch, "
            f"{ACCEPTANCE_MAX_DURATION_SECONDS} for --watch-acceptance."
        ),
    )
    return parser.parse_args(argv)


def resolve_state_path(args: argparse.Namespace) -> Optional[str]:
    if args.no_state:
        return None
    if args.state_file:
        return args.state_file
    # We need the PR number to construct the default. Defer until we have it.
    return None  # filled in after first fetch by main()


def emit_transient(exc: TransientGhError) -> None:
    try:
        body = json.loads(str(exc))
    except json.JSONDecodeError:
        body = {"error": str(exc), "transient": True}
    body.setdefault("transient", True)
    print(json.dumps(body, indent=2, sort_keys=True))


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)

    state_path = resolve_state_path(args)

    if args.watch or args.watch_acceptance:
        _resolve_watch_timing(args)
        return _run_watch(args, state_path)
    return _run_single(args, state_path)


def _resolve_watch_timing(args: argparse.Namespace) -> None:
    if args.watch_acceptance:
        if args.interval is None:
            args.interval = ACCEPTANCE_INTERVAL_SECONDS
        if args.max_duration is None:
            args.max_duration = ACCEPTANCE_MAX_DURATION_SECONDS
        return
    if args.interval is None:
        args.interval = DEFAULT_INTERVAL_SECONDS
    if args.max_duration is None:
        args.max_duration = DEFAULT_MAX_DURATION_SECONDS


def _ensure_state_path(
    args: argparse.Namespace, state_path: Optional[str], pr_number: int
) -> Optional[str]:
    if args.no_state:
        return None
    if state_path:
        return state_path
    return default_state_file_path(pr_number)


def _run_single(args: argparse.Namespace, state_path: Optional[str]) -> int:
    try:
        raw = fetch_all(args.pr, args.head_sha)
    except TransientGhError as exc:
        emit_transient(exc)
        return EXIT_TRANSIENT

    state_path = _ensure_state_path(args, state_path, int(raw["pr"]["number"]))
    state = load_state(state_path)
    payload, new_state = compute_payload(
        **raw,
        state=state,
        head_sha_override=args.head_sha,
        since_override=args.since,
        openspec_change=args.openspec_change,
        verify_workflow_present=detect_verify_workflow(),
    )
    save_state(state_path, new_state)
    out = payload if args.full_payload else format_focused(payload)
    print(json.dumps(out, indent=2, sort_keys=True))
    return EXIT_OK


def _run_watch(args: argparse.Namespace, state_path: Optional[str]) -> int:
    started = time.monotonic()
    tick = 0
    last_payload: Optional[dict[str, Any]] = None
    last_tick_transient = False
    while True:
        tick += 1
        last_tick_transient = False
        try:
            raw = fetch_all(args.pr, args.head_sha)
            state_path_resolved = _ensure_state_path(
                args, state_path, int(raw["pr"]["number"])
            )
            state = load_state(state_path_resolved)
            payload, new_state = compute_payload(
                **raw,
                state=state,
                head_sha_override=args.head_sha,
                since_override=args.since,
                openspec_change=args.openspec_change,
                verify_workflow_present=detect_verify_workflow(),
                retain_stale_since=args.watch_acceptance,
            )
            last_payload = payload
            focused = format_focused(payload) if not args.full_payload else payload
            out_of_band = payload["summary"]["checks"].get("outOfBand", [])
            settled = args.watch_acceptance and out_of_band_settled(out_of_band)
            # Intermediate acceptance ticks must not mark comments as seen.
            # The final payload is the one the watcher reports.
            if not args.watch_acceptance or settled:
                save_state(state_path_resolved, new_state)
            tick_line = json.dumps(
                {
                    "tick": tick,
                    "ts": now_iso(),
                    "payload": focused,
                },
                sort_keys=True,
            )
            print(tick_line, flush=True)
            if settled:
                print(
                    json.dumps(
                        {
                            "final": True,
                            "outcome": "acceptance_settled",
                            "payload": focused,
                        },
                        sort_keys=True,
                    ),
                    flush=True,
                )
                return EXIT_OK
            if not args.watch_acceptance and payload["summary"]["hasActionable"]:
                print(
                    json.dumps(
                        {"final": True, "outcome": "actionable", "payload": focused},
                        sort_keys=True,
                    ),
                    flush=True,
                )
                return EXIT_OK
        except TransientGhError as exc:
            try:
                err = json.loads(str(exc))
            except json.JSONDecodeError:
                err = {"error": str(exc), "transient": True}
            print(
                json.dumps(
                    {"tick": tick, "ts": now_iso(), "transient": True, "error": err},
                    sort_keys=True,
                ),
                flush=True,
            )
            # On transient errors, sleep and try again until max-duration.
            last_tick_transient = True
        elapsed = time.monotonic() - started
        if elapsed >= args.max_duration:
            last_summary = (
                format_focused(last_payload)
                if last_payload and not args.full_payload
                else last_payload
            )
            print(
                json.dumps(
                    {
                        "final": True,
                        "outcome": "timeout",
                        "elapsedSeconds": int(elapsed),
                        "lastSummary": last_summary,
                        "lastTickTransient": last_tick_transient,
                    },
                    sort_keys=True,
                ),
                flush=True,
            )
            return EXIT_TIMEOUT
        time.sleep(min(max(1, args.interval), args.max_duration - elapsed))


if __name__ == "__main__":
    sys.exit(main())
