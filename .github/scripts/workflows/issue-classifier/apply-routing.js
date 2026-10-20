const fs = require('fs');
const { execFileSync } = require('child_process');

const allowed = new Set(['needs-research', 'needs-reproduction', 'needs-spec', 'needs-human']);

function applyRouting({ env, execFile }) {
  function fail(message) {
    const error = new Error(message);
    error.routingFailure = true;
    throw error;
  }
  if (env.DETECTION_RESULT !== 'success' || env.GH_AW_DETECTION_CONCLUSION === 'warning') {
    fail(`refusing to apply labels: detection result=${env.DETECTION_RESULT} conclusion=${env.GH_AW_DETECTION_CONCLUSION}`);
    return;
  }

  const raw = JSON.parse(fs.readFileSync(env.GH_AW_AGENT_OUTPUT, 'utf8'));
  const items = (raw.items || []).filter((item) => item.type === 'apply_routing');
  const byIssue = new Map();
  for (const item of items) {
    const text = typeof item.issue_number === 'number' ? String(item.issue_number) : String(item.issue_number).trim();
    const num = Number(text);
    if (!Number.isSafeInteger(num) || num <= 0 || String(num) !== text) {
      fail(`invalid issue number: ${text}`);
      return;
    }
    if (!allowed.has(item.category)) {
      fail(`invalid category for #${num}: ${item.category}`);
      return;
    }
    let comment = typeof item.comment === 'string' ? item.comment.trim() : '';
    if (!comment || comment.length > 65000) {
      fail(`invalid comment for #${num}`);
      return;
    }
    if (!comment.startsWith('<!-- gha-issue-classifier -->')) {
      comment = `<!-- gha-issue-classifier -->\n${comment}`;
    }
    const previous = byIssue.get(num);
    if (previous && (previous.category !== item.category || previous.comment !== comment)) {
      fail(`conflicting routing calls for #${num}`);
      return;
    }
    byIssue.set(num, { category: item.category, comment });
  }

  const [owner, repo] = env.GITHUB_REPOSITORY.split('/');
  function gh(args, input) {
    return execFile('gh', ['api', ...args], {
      input,
      encoding: 'utf8',
      stdio: ['pipe', 'pipe', 'inherit'],
    });
  }
  function listLabels(num) {
    const parsed = JSON.parse(gh(['--paginate', '--slurp', `repos/${owner}/${repo}/issues/${num}/labels`]) || '[]');
    const labels = Array.isArray(parsed[0]) ? parsed.flat() : parsed;
    return labels.map((label) => label.name);
  }
  function hasClassifierComment(num) {
    const parsed = JSON.parse(gh(['--paginate', '--slurp', `repos/${owner}/${repo}/issues/${num}/comments`]) || '[]');
    const comments = Array.isArray(parsed[0]) ? parsed.flat() : parsed;
    return comments.some((c) => typeof c.body === 'string' && c.body.startsWith('<!-- gha-issue-classifier -->'));
  }
  function assertIssue(num) {
    const data = JSON.parse(gh([`repos/${owner}/${repo}/issues/${num}`]));
    if (data.pull_request) {
      fail(`#${num} is a pull request`);
    }
  }
  function removeLabel(num, name) {
    try {
      gh(['--method', 'DELETE', `repos/${owner}/${repo}/issues/${num}/labels/${encodeURIComponent(name)}`]);
    } catch (err) {
      if (listLabels(num).includes(name)) {
        fail(`failed to remove ${name} from #${num}`);
      }
    }
  }

  for (const num of byIssue.keys()) {
    assertIssue(num);
  }
  for (const [num, routing] of byIssue) {
    const { category, comment } = routing;
    const current = listLabels(num);
    if (current.includes('triaged')) {
      console.error(`skipping #${num}: already triaged`);
      continue;
    }
    const stale = current.filter((name) => allowed.has(name) && name !== category);
    for (const name of stale) {
      removeLabel(num, name);
    }
    const afterRemoval = listLabels(num);
    const remaining = afterRemoval.filter((name) => allowed.has(name) && name !== category);
    if (remaining.length > 0) {
      fail(`routing labels still present on #${num}: ${remaining.join(', ')}`);
      return;
    }
    // Apply the category label, post the comment, then `triaged` LAST. `triaged`
    // is the completion marker that makes both pre-activation and this handler
    // skip the issue, so writing it last keeps a partial run recoverable: if the
    // comment POST fails, the issue stays untriaged and a rerun retries cleanly.
    if (!afterRemoval.includes(category)) {
      gh([
        '--method', 'POST',
        `repos/${owner}/${repo}/issues/${num}/labels`,
        '--input', '-',
      ], JSON.stringify({ labels: [category] }));
    }
    // Post the classification comment at most once; a rerun after a partial
    // failure must not duplicate it.
    if (!hasClassifierComment(num)) {
      gh([
        '--method', 'POST',
        `repos/${owner}/${repo}/issues/${num}/comments`,
        '--input', '-',
      ], JSON.stringify({ body: comment }));
    }
    gh([
      '--method', 'POST',
      `repos/${owner}/${repo}/issues/${num}/labels`,
      '--input', '-',
    ], JSON.stringify({ labels: ['triaged'] }));
  }
}

function main() {
  try {
    applyRouting({ env: process.env, execFile: execFileSync });
  } catch (err) {
    if (!err.routingFailure) {
      throw err;
    }
    console.error(err.message);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = { applyRouting };
