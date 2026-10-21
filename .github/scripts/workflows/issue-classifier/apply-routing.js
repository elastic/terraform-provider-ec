const fs = require('fs');
const { execFileSync } = require('child_process');

const allowed = new Set(['needs-research', 'needs-reproduction', 'needs-spec', 'needs-human']);

function applyRouting({ env, execFile }) {
  function fail(message) {
    const error = new Error(message);
    error.routingFailure = true;
    throw error;
  }
  // Detection steps run with continue-on-error, so the job can report
  // result=success while the detector warned (threat), failed (engine/parse
  // error), or emitted nothing. threat-detect concludes `success` on a clean run
  // and `skipped` when detection was not required; allowlist exactly those two so
  // warning, failure, an empty conclusion, or any unknown value fails closed.
  // Never write labels or a comment without an explicit clean verdict.
  if (env.DETECTION_RESULT !== 'success' || !['success', 'skipped'].includes(env.GH_AW_DETECTION_CONCLUSION)) {
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
  // The safe-output job posts as this identity (GH_TOKEN = github.token). Only a
  // comment authored by it counts as our own, so a marker-bearing comment from a
  // reporter cannot suppress the real classification explanation.
  const botLogin = env.GH_AW_BOT_LOGIN || 'github-actions[bot]';
  function gh(args, input) {
    return execFile('gh', ['api', ...args], {
      input,
      encoding: 'utf8',
      // A large paginated label/comment history can exceed execFileSync's 1 MiB
      // default stdout limit and abort routing after the label is applied; allow
      // a generous ceiling so a busy issue still completes.
      maxBuffer: 64 * 1024 * 1024,
      stdio: ['pipe', 'pipe', 'inherit'],
    });
  }
  function listLabels(num) {
    const parsed = JSON.parse(gh(['--paginate', '--slurp', `repos/${owner}/${repo}/issues/${num}/labels`]) || '[]');
    const labels = Array.isArray(parsed[0]) ? parsed.flat() : parsed;
    return labels.map((label) => label.name);
  }
  function botClassifierComment(num) {
    const parsed = JSON.parse(gh(['--paginate', '--slurp', `repos/${owner}/${repo}/issues/${num}/comments`]) || '[]');
    const comments = Array.isArray(parsed[0]) ? parsed.flat() : parsed;
    return comments.find((c) =>
      c?.user?.login === botLogin &&
      typeof c.body === 'string' &&
      c.body.startsWith('<!-- gha-issue-classifier -->'),
    ) || null;
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
    // Post the classification comment, or reconcile the bot's existing one. A
    // clean rerun (identical comment) is a no-op, so the comment is never
    // duplicated; a rerun that reclassified to a different category edits the
    // existing comment so the explanation always matches the final routing label.
    const existingComment = botClassifierComment(num);
    if (!existingComment) {
      gh([
        '--method', 'POST',
        `repos/${owner}/${repo}/issues/${num}/comments`,
        '--input', '-',
      ], JSON.stringify({ body: comment }));
    } else if (existingComment.body !== comment) {
      gh([
        '--method', 'PATCH',
        `repos/${owner}/${repo}/issues/comments/${existingComment.id}`,
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
