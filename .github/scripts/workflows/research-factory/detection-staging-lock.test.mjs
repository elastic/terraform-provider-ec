// Asserts the compiled research-factory lock still wires the detection staging step the way
// stage-detection-copy.js expects, so a gh-aw bump or a source edit that breaks one of these
// couplings fails `make workflow-test`.
import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { PUBLISH_OUTPUT_TYPE, DEFAULT_PATHS } = require('./stage-detection-copy.js');

const REPO_ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..', '..', '..');
const LOCK_PATH = path.join(REPO_ROOT, '.github', 'workflows', 'research-factory-issue.lock.yml');
const LOCK = fs.readFileSync(LOCK_PATH, 'utf8');
const LINES = LOCK.split('\n');

const PREPARE = 'Prepare threat detection files';
const CHECKOUT = 'Checkout the publish handler for detection';
const STAGE = 'Stage the research comment body for detection';
const DOWNLOAD_INPUTS = 'Download research-factory grounding inputs for detection';
const VERIFY_INPUTS = 'Verify research-factory grounding inputs for detection';
const SETUP = 'Setup threat detection';
const EXECUTE = 'Execute threat detection with AWF';

const JOB_LINE = /^ {2}[a-z_]+:$/;
const STEP_PREFIX = '      - name:';
const PINNED_USES = /^\s+uses: ([\w.-]+\/[\w./-]+)@([0-9a-f]{40})$/;

/** Lines of one job: from `  <jobId>:` (under `jobs:`) to the next job line. */
function jobLines(jobId) {
  const jobsStart = LINES.indexOf('jobs:');
  assert.notEqual(jobsStart, -1, 'the lock has no jobs: key');
  const start = LINES.indexOf(`  ${jobId}:`, jobsStart);
  assert.notEqual(start, -1, `the lock has no ${jobId} job`);
  let end = start + 1;
  while (end < LINES.length && !JOB_LINE.test(LINES[end])) {
    end += 1;
  }
  return LINES.slice(start, end);
}

/** The steps of a job, each as `{ name, lines }`. */
function steps(jobId) {
  const lines = jobLines(jobId);
  const found = [];
  for (const line of lines) {
    if (line.startsWith(STEP_PREFIX)) {
      found.push({ name: line.slice(STEP_PREFIX.length).trim(), lines: [line] });
    } else if (found.length > 0) {
      found[found.length - 1].lines.push(line);
    }
  }
  return found;
}

function step(jobId, name) {
  const matches = steps(jobId).filter((s) => s.name === name);
  assert.equal(matches.length, 1, `expected exactly one "${name}" step in ${jobId}, found ${matches.length}`);
  return matches[0];
}

function stepText(s) {
  return s.lines.join('\n');
}

/** The `owner/repo@sha` of a step's `uses:` line, whatever trails it. */
function usesRef(s) {
  const line = s.lines.find((l) => /^\s+uses: /.test(l));
  assert.ok(line, `step "${s.name}" has no uses:`);
  const match = /^\s+uses: ([\w.-]+\/[\w./-]+)@([0-9a-f]{40})\b/.exec(line);
  assert.ok(match, `step "${s.name}" does not use a 40-hex SHA: ${line.trim()}`);
  return { repo: match[1], sha: match[2], line };
}

/** Steps of the detection job strictly between "Prepare threat detection files" and "Setup threat detection". */
function customDetectionSteps() {
  const all = steps('detection');
  const from = all.findIndex((s) => s.name === PREPARE);
  const to = all.findIndex((s) => s.name === SETUP);
  assert.ok(from !== -1 && to !== -1 && from < to, 'detection job lost its prepare/setup steps');
  return all.slice(from + 1, to);
}

describe('research-factory lock: detection staging', () => {
  test('the staging steps run after the detector files are prepared and before the detector', () => {
    const names = steps('detection').map((s) => s.name);
    const order = [PREPARE, CHECKOUT, STAGE, SETUP, EXECUTE].map((name) => {
      const index = names.indexOf(name);
      assert.notEqual(index, -1, `detection job has no "${name}" step`);
      return index;
    });
    for (let i = 1; i < order.length; i += 1) {
      assert.ok(order[i - 1] < order[i], `step order broken: ${names.join(' | ')}`);
    }
  });

  test('the grounding-input steps run after the detector files are prepared and before the detector', () => {
    const names = steps('detection').map((s) => s.name);
    const order = [PREPARE, DOWNLOAD_INPUTS, VERIFY_INPUTS, SETUP, EXECUTE].map((name) => {
      const index = names.indexOf(name);
      assert.notEqual(index, -1, `detection job has no "${name}" step`);
      return index;
    });
    for (let i = 1; i < order.length; i += 1) {
      assert.ok(order[i - 1] < order[i], `step order broken: ${names.join(' | ')}`);
    }
  });

  test('every custom detection step pins its action to a SHA gh-aw pins elsewhere in the lock', () => {
    const custom = customDetectionSteps();
    assert.ok(custom.some((s) => s.name === CHECKOUT) && custom.some((s) => s.name === STAGE));
    assert.ok(custom.some((s) => s.name === DOWNLOAD_INPUTS));
    for (const s of custom) {
      const line = s.lines.find((l) => /^\s+uses: /.test(l));
      if (!line) continue;
      const match = PINNED_USES.exec(line);
      assert.ok(match, `step "${s.name}" must use owner/repo@<40-hex SHA> with no trailing comment: ${line.trim()}`);
      assert.ok(LOCK.includes(`${match[1]}@${match[2]} # v`), `${match[1]}@${match[2]} is not pinned to a version anywhere else in the lock`);
    }
  });

  test('the staging steps use the same actions as the publish job', () => {
    const stage = usesRef(step('detection', STAGE));
    const publish = usesRef(step('update_research_comment', 'Publish the research comment'));
    assert.equal(stage.repo, 'actions/github-script');
    assert.equal(`${stage.repo}@${stage.sha}`, `${publish.repo}@${publish.sha}`);

    const checkout = usesRef(step('detection', CHECKOUT));
    const publishCheckout = usesRef(step('update_research_comment', 'Checkout repository'));
    assert.equal(checkout.repo, 'actions/checkout');
    assert.equal(`${checkout.repo}@${checkout.sha}`, `${publishCheckout.repo}@${publishCheckout.sha}`);
  });

  test('no custom detection step continues on error', () => {
    for (const s of customDetectionSteps()) {
      assert.ok(!stepText(s).includes('continue-on-error'), `step "${s.name}" has continue-on-error`);
    }
  });

  test('the checkout brings only the workflow scripts', () => {
    const s = step('detection', CHECKOUT);
    const text = stepText(s);
    assert.match(text, /^\s+sparse-checkout-cone-mode: false$/m);
    assert.match(text, /^\s+persist-credentials: false$/m);

    const keyIndex = s.lines.findIndex((l) => /^\s+sparse-checkout: \|$/.test(l));
    assert.notEqual(keyIndex, -1, 'checkout step has no sparse-checkout: | block');
    const keyIndent = s.lines[keyIndex].search(/\S/);
    const patterns = [];
    for (const line of s.lines.slice(keyIndex + 1)) {
      if (line.trim() !== '' && line.search(/\S/) <= keyIndent) break;
      if (line.trim() !== '') patterns.push(line.trim());
    }
    assert.deepEqual(patterns, ['/.github/scripts/workflows/']);
  });

  test('the staging step reads the same output types the publish job tests', () => {
    assert.ok(stepText(step('detection', STAGE)).includes('AGENT_OUTPUT_TYPES: ${{ needs.agent.outputs.output_types }}'));
    const publishJob = jobLines('update_research_comment').join('\n');
    const match = /contains\(needs\.agent\.outputs\.output_types, '([^']+)'\)/.exec(publishJob);
    assert.ok(match, 'the publish job no longer tests needs.agent.outputs.output_types');
    assert.equal(match[1], PUBLISH_OUTPUT_TYPE);
  });

  test('the staging step requires stage-detection-copy.js from the checkout', () => {
    const match = /require\('\$\{\{ github\.workspace \}\}\/([^']+)'\)/.exec(stepText(step('detection', STAGE)));
    assert.ok(match, 'staging step has no require of a workspace path');
    assert.equal(path.basename(match[1]), 'stage-detection-copy.js');
    assert.ok(fs.statSync(path.join(REPO_ROOT, match[1])).isFile(), `${match[1]} does not exist`);
    assert.ok(match[1].startsWith('.github/scripts/workflows/'));
  });

  test('the staged copy lands in the directory the detector analyses', () => {
    const text = stepText(step('detection', EXECUTE));
    const match = /threat-detect [^']*?(\S+)'/.exec(text);
    assert.ok(match, 'no threat-detect invocation in the detector step');
    assert.equal(match[1], DEFAULT_PATHS.detectionDir);
    assert.ok(text.includes(`--mount ${DEFAULT_PATHS.detectionDir}:${DEFAULT_PATHS.detectionDir}:rw`));
  });

  test('the draft is read from where the detection job downloads the agent artifact', () => {
    const text = stepText(step('detection', 'Download agent output artifact'));
    assert.match(text, /^\s+path: (.+)$/m);
    assert.equal(/^\s+path: (.+)$/m.exec(text)[1].trim(), `${DEFAULT_PATHS.sourceRoot}/`);
    assert.ok(text.includes('pattern: "{agent,agent-output-fallback}"'));
    assert.ok(jobLines('agent').some((l) => l.trim() === '/tmp/gh-aw/agent/research/**/*.md'));
  });

  test('the detector is told about the staged file on both detection steps', () => {
    const prompts = LINES.filter((l) => /^\s+CUSTOM_PROMPT:/.test(l));
    assert.equal(prompts.length, 2);
    for (const line of prompts) {
      assert.ok(line.includes('comment-memory/research-comment.md'));
      assert.ok(line.includes('/tmp/gh-aw/threat-detection/grounding-inputs/'));
    }
  });

  test('the publish job still requires a successful detection job', () => {
    assert.ok(jobLines('update_research_comment').join('\n').includes("needs.detection.result == 'success'"));
  });
});
