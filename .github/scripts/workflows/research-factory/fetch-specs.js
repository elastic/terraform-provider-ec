/**
 * Agent-job pre-step: fetches the four API specifications the research agent and its critic ground
 * on, before the agent sandbox is built.
 *
 * Runs in the gh-aw `agent` job through `actions/github-script`, as the `Fetch Elastic Cloud API
 * specifications` step of `research-factory-issue.md`. The URL list is the step's `SPECS` env
 * value, one `<url> <file name>` pair per line; the files land in `SPECS_DIR`. It makes no GitHub
 * API call, reads no secret and takes no input from the issue.
 *
 * **Every spec or no run.** The research needs all of the specs, so the step fails when any of them
 * could not be fetched, whatever the reason: a dead link, a 5xx, a rate limit, a timeout, DNS, or a
 * failed write. A failure here costs a re-run; research grounded without a spec costs a comment
 * nobody can trust (cp-hosted-team#5441).
 *
 * The URLs are fetched in parallel. Each gets up to `ATTEMPTS` attempts, `RETRY_DELAYS_MS` apart;
 * a `Retry-After` header lengthens the wait, up to `MAX_RETRY_AFTER_MS`. An attempt succeeds only
 * on 2xx with a body that is an API specification; anything else is tried again.
 *
 * Only the HTTP fetch is retried. Writing a fetched spec to SPECS_DIR is not: a local write fails
 * for reasons waiting does not fix (disk full, permissions, read-only file system), so that spec
 * fails at once with `could not save <file>`; the step fails after the other URLs finish.
 *
 *   | URL result        | File in SPECS_DIR | Log                          | Step                      |
 *   | ----------------- | ----------------- | ---------------------------- | ------------------------- |
 *   | spec fetched      | written           | `fetched <file> (<n> bytes)` | succeeds                  |
 *   | any failure       | removed           | error                        | fails, after all URLs ran |
 *
 * Rules that are easy to "simplify" into bugs:
 *
 * - **Keep the `fetched <file> (<n> bytes)` line as it is.** cp-hosted-team#5441's acceptance
 *   check greps for it in the run log.
 * - **A file exists in SPECS_DIR only after a completed write of a real spec.** A partial file or
 *   an HTML page would be read as if it were the spec.
 * - **Keep the worst case well inside the agent job's 60-minute timeout.** The URLs run in
 *   parallel, so one URL bounds the step: `ATTEMPTS` × `TIMEOUT_MS` plus the retry waits.
 */
'use strict';

const fs = require('node:fs');
const path = require('node:path');

const RETRY_DELAYS_MS = [5000, 30000, 120000];
const ATTEMPTS = RETRY_DELAYS_MS.length + 1;
const MAX_RETRY_AFTER_MS = 120000;
const TIMEOUT_MS = 60000;

const FILE_NAME = /^[A-Za-z0-9][A-Za-z0-9._-]*\.json$/;

const WORKFLOW_FILE = '.github/workflows/research-factory-issue.md';

const defaultSleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

/**
 * Parses the `SPECS` list: one `<https url> <file name>` pair per non-blank line.
 *
 * @param {string} text
 * @returns {{url: string, file: string}[]}
 */
function parseSpecList(text) {
  const pairs = [];
  const seen = new Set();
  for (const raw of String(text).split('\n')) {
    const line = raw.trim();
    if (line === '') {
      continue;
    }
    const tokens = line.split(/\s+/);
    if (tokens.length !== 2) {
      throw new Error(`expected "<url> <file name>", got: ${line}`);
    }
    const [url, file] = tokens;
    let parsed;
    try {
      parsed = new URL(url);
    } catch {
      throw new Error(`not a URL in line: ${line}`);
    }
    if (parsed.protocol !== 'https:') {
      throw new Error(`not an https URL in line: ${line}`);
    }
    if (!FILE_NAME.test(file)) {
      throw new Error(`file name must match ${FILE_NAME} in line: ${line}`);
    }
    if (seen.has(file)) {
      throw new Error(`file name ${file} repeats in line: ${line}`);
    }
    seen.add(file);
    pairs.push({ url, file });
  }
  if (pairs.length === 0) {
    throw new Error('the list is empty');
  }
  return pairs;
}

/**
 * Whether `text` is an API specification: a JSON object with an own `openapi` or `swagger` key and
 * a non-empty `paths` object. Never throws.
 *
 * @param {string} text
 * @returns {boolean}
 */
function isApiSpec(text) {
  let value;
  try {
    value = JSON.parse(text);
  } catch {
    return false;
  }
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    return false;
  }
  if (!Object.hasOwn(value, 'openapi') && !Object.hasOwn(value, 'swagger')) {
    return false;
  }
  const paths = value.paths;
  return paths !== null && typeof paths === 'object' && !Array.isArray(paths) && Object.keys(paths).length > 0;
}

/**
 * A `Retry-After` header value in milliseconds (delay-seconds or an HTTP date), or 0 when it is
 * absent or unreadable. Never throws.
 *
 * @param {string|null|undefined} value
 * @param {number} now
 * @returns {number}
 */
function retryAfterMs(value, now) {
  if (typeof value !== 'string' || value.trim() === '') {
    return 0;
  }
  const text = value.trim();
  if (/^\d+$/.test(text)) {
    return Number(text) * 1000;
  }
  const date = Date.parse(text);
  return Number.isNaN(date) ? 0 : Math.max(0, date - now);
}

function errorReason(err) {
  const code = err?.cause?.code;
  return `${err?.name}: ${err?.message}${code ? ` [${code}]` : ''}`;
}

/** One attempt: `{ ok: true, body }` or `{ ok: false, reason, retryAfter }`. */
async function attempt(url, fetchImpl, now) {
  let status;
  let body;
  let retryAfter = 0;
  try {
    const response = await fetchImpl(url, { redirect: 'follow', signal: AbortSignal.timeout(TIMEOUT_MS) });
    status = response.status;
    try {
      retryAfter = retryAfterMs(response.headers?.get?.('retry-after'), now());
    } catch {
      retryAfter = 0;
    }
    body = await response.text();
  } catch (err) {
    return { ok: false, reason: errorReason(err), retryAfter };
  }
  const success = status >= 200 && status <= 299;
  if (success && isApiSpec(body)) {
    return { ok: true, body };
  }
  return { ok: false, reason: success ? `HTTP ${status} but not an API specification` : `HTTP ${status}`, retryAfter };
}

/**
 * Fetches one URL with up to `ATTEMPTS` attempts. On failure the reason is that of the last
 * attempt.
 *
 * @param {{url: string, fetchImpl: Function, sleep: (ms: number) => Promise<void>, now?: () => number}} args
 * @returns {Promise<{ok: boolean, body: string|undefined, reason: string|undefined}>}
 */
async function fetchOne({ url, fetchImpl, sleep, now = Date.now }) {
  let reason;
  for (let i = 0; i < ATTEMPTS; i += 1) {
    const result = await attempt(url, fetchImpl, now);
    if (result.ok) {
      return { ok: true, body: result.body, reason: undefined };
    }
    reason = result.reason;
    if (i < RETRY_DELAYS_MS.length) {
      await sleep(Math.max(RETRY_DELAYS_MS[i], Math.min(result.retryAfter, MAX_RETRY_AFTER_MS)));
    }
  }
  return { ok: false, body: undefined, reason };
}

/**
 * Removes `target` if it exists. Never throws: a failure here does not change the outcome, because
 * the pair is reported as missing and the step fails either way.
 */
function removeQuietly(fsImpl, target) {
  try {
    fsImpl.rmSync(target, { force: true });
  } catch {
    // Ignored, see above.
  }
}

/** Fetches and writes one pair. Never throws: `{ file, url, bytes }` or `{ file, url, error }`. */
async function fetchPair({ url, file, specsDir, fetchImpl, sleep, now, fsImpl }) {
  const target = path.join(specsDir, file);
  try {
    const { ok, body, reason } = await fetchOne({ url, fetchImpl, sleep, now });
    if (ok) {
      fsImpl.writeFileSync(target, body);
      return { file, url, bytes: Buffer.byteLength(body) };
    }
    removeQuietly(fsImpl, target);
    return {
      file,
      url,
      error: `could not fetch ${url} after ${ATTEMPTS} attempts (${reason}); if the link is dead, fix the URL in ${WORKFLOW_FILE}, otherwise run the research again later`,
    };
  } catch (err) {
    removeQuietly(fsImpl, target);
    return { file, url, error: `could not save ${file}: ${err?.message}` };
  }
}

/**
 * Fetches every pair of `env.SPECS` into `env.SPECS_DIR`, then fails the step unless every spec
 * was fetched.
 *
 * @param {{core: object, env?: object, fetchImpl?: Function, sleep?: Function, now?: () => number, fs?: typeof import('node:fs')}} args
 */
module.exports = async function fetchSpecs({
  core,
  env = process.env,
  fetchImpl = fetch,
  sleep = defaultSleep,
  now = Date.now,
  fs: fsImpl = fs,
}) {
  const specsDir = env.SPECS_DIR;
  if (specsDir === undefined || specsDir === null || String(specsDir).trim() === '') {
    core.setFailed('SPECS_DIR is not set');
    return;
  }
  let pairs;
  try {
    pairs = parseSpecList(env.SPECS ?? '');
  } catch (err) {
    core.setFailed(`invalid SPECS: ${err.message}`);
    return;
  }
  try {
    fsImpl.mkdirSync(specsDir, { recursive: true });
  } catch (err) {
    core.setFailed(`could not create ${specsDir}: ${err.message}`);
    return;
  }

  const results = await Promise.all(
    pairs.map(({ url, file }) => fetchPair({ url, file, specsDir, fetchImpl, sleep, now, fsImpl })),
  );

  const fetched = [];
  const missing = [];
  for (const result of results) {
    if (result.error === undefined) {
      fetched.push(`${result.file} (${result.bytes} bytes)`);
      core.info(`fetched ${result.file} (${result.bytes} bytes)`);
    } else {
      missing.push(result.file);
      core.error(result.error);
    }
  }

  core.info(`specs in ${specsDir}: ${fetched.length > 0 ? fetched.join(', ') : 'none'}`);

  if (missing.length > 0) {
    core.setFailed(`${missing.length} of ${pairs.length} spec(s) not fetched: ${missing.join(', ')}; the research needs every spec`);
  }
};

module.exports.parseSpecList = parseSpecList;
module.exports.isApiSpec = isApiSpec;
module.exports.retryAfterMs = retryAfterMs;
module.exports.fetchOne = fetchOne;
module.exports.ATTEMPTS = ATTEMPTS;
module.exports.RETRY_DELAYS_MS = RETRY_DELAYS_MS;
module.exports.MAX_RETRY_AFTER_MS = MAX_RETRY_AFTER_MS;
module.exports.TIMEOUT_MS = TIMEOUT_MS;
