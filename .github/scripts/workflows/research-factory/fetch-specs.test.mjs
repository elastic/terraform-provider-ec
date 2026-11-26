// Tests for fetch-specs.js, the runner-side fetch of the four API specifications, and for the way
// the compiled research-factory lock wires it into the agent job. No test touches the network or
// waits on a real timer: `fetch`, `sleep` and (where a failure is needed) `fs` are injected.
import assert from 'node:assert/strict';
import test, { afterEach, beforeEach, describe } from 'node:test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const fetchSpecs = require('./fetch-specs.js');
const { parseSpecList, isApiSpec, retryAfterMs, fetchOne, ATTEMPTS, RETRY_DELAYS_MS, MAX_RETRY_AFTER_MS, TIMEOUT_MS } =
  fetchSpecs;

const REPO_ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..', '..', '..');
const WORKFLOW_PATH = path.join(REPO_ROOT, '.github', 'workflows', 'research-factory-issue.md');
const LOCK_PATH = path.join(REPO_ROOT, '.github', 'workflows', 'research-factory-issue.lock.yml');

const SWAGGER_URL = 'https://api.elastic-cloud.com/api/v1/api-docs-user/swagger.json';
const CLOUD_URL = 'https://www.elastic.co/docs/api/doc/cloud.json';
const ECE_URL = 'https://www.elastic.co/docs/api/doc/cloud-enterprise.json';
const SERVERLESS_URL = 'https://www.elastic.co/docs/api/cloud-serverless.json';

const PAIRS = [
  { url: SWAGGER_URL, file: 'api-elastic-cloud-swagger.json' },
  { url: CLOUD_URL, file: 'cloud.json' },
  { url: ECE_URL, file: 'cloud-enterprise.json' },
  { url: SERVERLESS_URL, file: 'elastic-cloud-serverless.json' },
];
const SPECS = PAIRS.map((p) => `${p.url} ${p.file}`).join('\n');

const OPENAPI_BODY = (title) => JSON.stringify({ openapi: '3.1.2', info: { title }, paths: { '/a': {} } });
const SWAGGER_BODY = JSON.stringify({ swagger: '2.0', info: { title: 'ECH' }, paths: { '/a': {} } });
const HTML_BODY = '<!doctype html><html><body>Not found</body></html>';

/** Healthy answers for the four URLs: one Swagger 2.0, three OpenAPI. */
function healthyResponses() {
  return {
    [SWAGGER_URL]: [{ status: 200, body: SWAGGER_BODY }],
    [CLOUD_URL]: [{ status: 200, body: OPENAPI_BODY('cloud') }],
    [ECE_URL]: [{ status: 200, body: OPENAPI_BODY('ece') }],
    [SERVERLESS_URL]: [{ status: 200, body: OPENAPI_BODY('serverless') }],
  };
}

/** A `core` double that records every call. */
function makeCore() {
  const calls = { info: [], warning: [], error: [], setFailed: [] };
  return {
    calls,
    info: (m) => calls.info.push(m),
    warning: (m) => calls.warning.push(m),
    error: (m) => calls.error.push(m),
    setFailed: (m) => calls.setFailed.push(m),
  };
}

/**
 * A `fetch` double. `responses` maps a URL to one response per attempt; the last one repeats.
 * A response is `{ status, body, retryAfter? }`, `{ throws: Error }`, or `{ status, textThrows: Error }`.
 */
function makeFetch(responses) {
  const calls = {};
  const options = [];
  const impl = async (url, opts) => {
    calls[url] = (calls[url] ?? 0) + 1;
    options.push(opts);
    const list = responses[url];
    assert.ok(list, `unexpected fetch of ${url}`);
    const r = list[Math.min(calls[url] - 1, list.length - 1)];
    if (r.throws) {
      throw r.throws;
    }
    return {
      status: r.status,
      headers: { get: (name) => (name.toLowerCase() === 'retry-after' ? (r.retryAfter ?? null) : null) },
      text: async () => {
        if (r.textThrows) {
          throw r.textThrows;
        }
        return r.body;
      },
    };
  };
  return { impl, calls, options };
}

/** A `sleep` double that records the delay and resolves at once. */
function makeSleep() {
  const delays = [];
  const impl = async (ms) => {
    delays.push(ms);
  };
  return { impl, delays };
}

function timeoutError() {
  const err = new Error('The operation was aborted due to timeout');
  err.name = 'TimeoutError';
  return err;
}

function dnsError() {
  const err = new TypeError('fetch failed');
  err.cause = { code: 'ENOTFOUND' };
  return err;
}

let specsDir;

beforeEach(() => {
  specsDir = fs.mkdtempSync(path.join(os.tmpdir(), 'fetch-specs-'));
});

afterEach(() => {
  fs.rmSync(specsDir, { recursive: true, force: true });
});

/** Runs the main function with doubles; returns the doubles for assertions. */
async function run({ responses, specs = SPECS, env, fsImpl, now } = {}) {
  const core = makeCore();
  const fetchDouble = makeFetch(responses ?? healthyResponses());
  const sleepDouble = makeSleep();
  const args = {
    core,
    env: env ?? { SPECS_DIR: specsDir, SPECS: specs },
    fetchImpl: fetchDouble.impl,
    sleep: sleepDouble.impl,
  };
  if (fsImpl) {
    args.fs = fsImpl;
  }
  if (now) {
    args.now = now;
  }
  await fetchSpecs(args);
  return { core, fetch: fetchDouble, sleep: sleepDouble };
}

function filesIn(dir) {
  return fs.readdirSync(dir).sort();
}

describe('fetch-specs: constants', () => {
  test('four attempts, 5 s / 30 s / 2 min apart, 60 s per attempt, Retry-After capped at 2 min', () => {
    assert.equal(ATTEMPTS, 4);
    assert.deepEqual(RETRY_DELAYS_MS, [5000, 30000, 120000]);
    assert.equal(TIMEOUT_MS, 60000);
    assert.equal(MAX_RETRY_AFTER_MS, 120000);
  });

  test('the worst case for one URL stays under a quarter of the 60-minute agent job', () => {
    const worst = ATTEMPTS * TIMEOUT_MS + RETRY_DELAYS_MS.map((d) => Math.max(d, MAX_RETRY_AFTER_MS)).reduce((a, b) => a + b, 0);
    assert.ok(worst <= 15 * 60 * 1000, `worst case ${worst} ms`);
  });
});

describe('fetch-specs: parseSpecList', () => {
  test('four valid lines give four pairs, in order', () => {
    assert.deepEqual(parseSpecList(SPECS), PAIRS);
  });

  test('blank lines and surrounding spaces are ignored', () => {
    const text = `\n  ${PAIRS[0].url}   ${PAIRS[0].file}  \n\n\t${PAIRS[1].url} ${PAIRS[1].file}\n   \n`;
    assert.deepEqual(parseSpecList(text), PAIRS.slice(0, 2));
  });

  const bad = [
    ['one token', `${SWAGGER_URL}`],
    ['three tokens', `${SWAGGER_URL} a.json extra`],
    ['an http:// URL', 'http://www.elastic.co/docs/api/cloud.json cloud.json'],
    ['a token that is not a URL', 'not-a-url x.json'],
    ['a file name with /', `${SWAGGER_URL} specs/x.json`],
    ['a file name without .json', `${SWAGGER_URL} x.yaml`],
  ];
  for (const [label, line] of bad) {
    test(`throws on ${label}, quoting the line`, () => {
      assert.throws(() => parseSpecList(line), (err) => {
        assert.ok(err instanceof Error);
        assert.ok(err.message.includes(line.trim()), `message should quote the line: ${err.message}`);
        return true;
      });
    });
  }

  test('throws on a repeated file name, quoting the second line', () => {
    assert.throws(
      () => parseSpecList(`${SWAGGER_URL} a.json\n${CLOUD_URL} a.json`),
      (err) => err.message.includes('repeats') && err.message.includes(`${CLOUD_URL} a.json`),
    );
  });

  test('throws on an empty list', () => {
    assert.throws(() => parseSpecList(''), /empty/);
    assert.throws(() => parseSpecList('  \n \n'), /empty/);
  });
});

describe('fetch-specs: isApiSpec', () => {
  test('true for an OpenAPI and a Swagger document with paths', () => {
    assert.equal(isApiSpec('{"openapi":"3.1.2","paths":{"/a":{}}}'), true);
    assert.equal(isApiSpec('{"swagger":"2.0","paths":{"/a":{}}}'), true);
  });

  const notSpecs = [
    ['an HTML string', HTML_BODY],
    ['an empty string', ''],
    ['[]', '[]'],
    ['null', 'null'],
    ['"x"', '"x"'],
    ['an object without paths', '{"openapi":"3.1.2"}'],
    ['paths: {}', '{"openapi":"3.1.2","paths":{}}'],
    ['paths: []', '{"openapi":"3.1.2","paths":[]}'],
    ['no openapi/swagger key', '{"info":{},"paths":{"/a":{}}}'],
  ];
  for (const [label, text] of notSpecs) {
    test(`false for ${label}`, () => {
      assert.equal(isApiSpec(text), false);
    });
  }
});

describe('fetch-specs: retryAfterMs', () => {
  const NOW = Date.parse('2026-11-27T12:00:00Z');

  test('delay-seconds', () => {
    assert.equal(retryAfterMs('0', NOW), 0);
    assert.equal(retryAfterMs('45', NOW), 45000);
    assert.equal(retryAfterMs(' 7 ', NOW), 7000);
  });

  test('an HTTP date in the future or the past', () => {
    assert.equal(retryAfterMs('Fri, 27 Nov 2026 12:01:30 GMT', NOW), 90000);
    assert.equal(retryAfterMs('Fri, 27 Nov 2026 11:00:00 GMT', NOW), 0);
  });

  for (const value of [undefined, null, '', '  ', 'soon', '-5', '1.5']) {
    test(`0 for ${JSON.stringify(value)}`, () => {
      assert.equal(retryAfterMs(value, NOW), 0);
    });
  }
});

describe('fetch-specs: fetchOne', () => {
  test('passes redirect: follow and an abort signal to fetch', async () => {
    const f = makeFetch({ [SERVERLESS_URL]: [{ status: 200, body: OPENAPI_BODY('s') }] });
    const s = makeSleep();
    const result = await fetchOne({ url: SERVERLESS_URL, fetchImpl: f.impl, sleep: s.impl });
    assert.equal(result.ok, true);
    assert.equal(result.reason, undefined);
    assert.equal(f.options[0].redirect, 'follow');
    assert.ok(f.options[0].signal instanceof AbortSignal);
    assert.deepEqual(s.delays, []);
  });

  test('a failure reports the reason of the last attempt', async () => {
    const f = makeFetch({ [SERVERLESS_URL]: [{ status: 404 }, { status: 503 }, { status: 404 }, { throws: timeoutError() }] });
    const result = await fetchOne({ url: SERVERLESS_URL, fetchImpl: f.impl, sleep: makeSleep().impl });
    assert.equal(result.ok, false);
    assert.ok(result.reason.startsWith('TimeoutError: '), result.reason);
    assert.equal(result.body, undefined);
  });

  test('Retry-After longer than the scheduled wait is honoured, capped at MAX_RETRY_AFTER_MS', async () => {
    const f = makeFetch({
      [SERVERLESS_URL]: [
        { status: 429, retryAfter: '60' },
        { status: 503, retryAfter: '600' },
        { status: 429, retryAfter: '1' },
        { status: 200, body: OPENAPI_BODY('s') },
      ],
    });
    const s = makeSleep();
    const result = await fetchOne({ url: SERVERLESS_URL, fetchImpl: f.impl, sleep: s.impl });
    assert.equal(result.ok, true);
    assert.deepEqual(s.delays, [60000, MAX_RETRY_AFTER_MS, RETRY_DELAYS_MS[2]]);
  });

  test('a Retry-After date is read against the injected clock', async () => {
    const now = () => Date.parse('2026-11-27T12:00:00Z');
    const f = makeFetch({
      [SERVERLESS_URL]: [{ status: 429, retryAfter: 'Fri, 27 Nov 2026 12:00:45 GMT' }, { status: 200, body: OPENAPI_BODY('s') }],
    });
    const s = makeSleep();
    await fetchOne({ url: SERVERLESS_URL, fetchImpl: f.impl, sleep: s.impl, now });
    assert.deepEqual(s.delays, [45000]);
  });
});

describe('fetch-specs: main function', () => {
  test('1. four valid specs: four files, four fetched lines in order, a summary, nothing else', async () => {
    const responses = healthyResponses();
    const { core, fetch: f, sleep } = await run({ responses });
    assert.deepEqual(filesIn(specsDir), PAIRS.map((p) => p.file).sort());
    for (const p of PAIRS) {
      assert.equal(fs.readFileSync(path.join(specsDir, p.file), 'utf8'), responses[p.url][0].body);
    }
    const fetched = core.calls.info.filter((m) => m.startsWith('fetched '));
    assert.deepEqual(
      fetched,
      PAIRS.map((p) => `fetched ${p.file} (${Buffer.byteLength(responses[p.url][0].body)} bytes)`),
    );
    const last = core.calls.info.at(-1);
    assert.ok(last.startsWith('specs in '), last);
    for (const p of PAIRS) {
      assert.ok(last.includes(p.file), last);
    }
    assert.deepEqual(core.calls.warning, []);
    assert.deepEqual(core.calls.error, []);
    assert.deepEqual(core.calls.setFailed, []);
    for (const p of PAIRS) {
      assert.equal(f.calls[p.url], 1);
    }
    assert.deepEqual(sleep.delays, []);
  });

  // Every way an attempt can fail: each one, on every attempt, fails the step.
  const failures = [
    ['HTTP 404', { status: 404, body: HTML_BODY }, 'HTTP 404'],
    ['HTTP 410', { status: 410 }, 'HTTP 410'],
    ['HTTP 403', { status: 403, body: HTML_BODY }, 'HTTP 403'],
    ['HTTP 408', { status: 408 }, 'HTTP 408'],
    ['HTTP 429', { status: 429 }, 'HTTP 429'],
    ['HTTP 500', { status: 500 }, 'HTTP 500'],
    ['HTTP 503', { status: 503, body: HTML_BODY }, 'HTTP 503'],
    ['HTTP 200 with HTML', { status: 200, body: HTML_BODY }, 'HTTP 200 but not an API specification'],
    ['HTTP 200 with empty paths', { status: 200, body: '{"openapi":"3.1.2","paths":{}}' }, 'HTTP 200 but not an API specification'],
    ['a timeout', { throws: timeoutError() }, 'TimeoutError: '],
    ['a DNS failure', { throws: dnsError() }, 'TypeError: fetch failed [ENOTFOUND]'],
    ['response.text() rejecting', { status: 200, textThrows: timeoutError() }, 'TimeoutError: '],
  ];
  for (const [label, response, reason] of failures) {
    test(`2. ${label} on every attempt fails the step and names the URL and the reason`, async () => {
      const responses = { ...healthyResponses(), [SERVERLESS_URL]: [response] };
      const { core, fetch: f, sleep } = await run({ responses });
      assert.equal(f.calls[SERVERLESS_URL], ATTEMPTS);
      assert.deepEqual(sleep.delays, RETRY_DELAYS_MS);
      assert.equal(core.calls.error.length, 1);
      const e = core.calls.error[0];
      assert.ok(e.startsWith(`could not fetch ${SERVERLESS_URL} after ${ATTEMPTS} attempts (`), e);
      assert.ok(e.includes(reason), e);
      assert.ok(
        e.endsWith('if the link is dead, fix the URL in .github/workflows/research-factory-issue.md, otherwise run the research again later'),
        e,
      );
      assert.deepEqual(core.calls.warning, []);
      assert.deepEqual(filesIn(specsDir), ['api-elastic-cloud-swagger.json', 'cloud-enterprise.json', 'cloud.json']);
      assert.deepEqual(core.calls.setFailed, ['1 of 4 spec(s) not fetched: elastic-cloud-serverless.json; the research needs every spec']);
    });
  }

  test('3. failures on early attempts, then a valid spec: ok, no error', async () => {
    const body = OPENAPI_BODY('serverless');
    const responses = {
      ...healthyResponses(),
      [SERVERLESS_URL]: [{ status: 404 }, { throws: timeoutError() }, { status: 503 }, { status: 200, body }],
    };
    const { core, fetch: f, sleep } = await run({ responses });
    assert.equal(fs.readFileSync(path.join(specsDir, 'elastic-cloud-serverless.json'), 'utf8'), body);
    assert.deepEqual(core.calls.error, []);
    assert.deepEqual(core.calls.setFailed, []);
    assert.equal(f.calls[SERVERLESS_URL], ATTEMPTS);
    assert.deepEqual(sleep.delays, RETRY_DELAYS_MS);
  });

  test('4. two failed URLs: two errors in list order, one failure naming both, healthy files written', async () => {
    const responses = {
      ...healthyResponses(),
      [CLOUD_URL]: [{ status: 404 }],
      [SERVERLESS_URL]: [{ status: 503 }],
    };
    const { core } = await run({ responses });
    assert.equal(core.calls.error.length, 2);
    assert.ok(core.calls.error[0].includes(CLOUD_URL), core.calls.error[0]);
    assert.ok(core.calls.error[1].includes(SERVERLESS_URL), core.calls.error[1]);
    assert.deepEqual(core.calls.setFailed, [
      '2 of 4 spec(s) not fetched: cloud.json, elastic-cloud-serverless.json; the research needs every spec',
    ]);
    assert.deepEqual(filesIn(specsDir), ['api-elastic-cloud-swagger.json', 'cloud-enterprise.json']);
  });

  test('5. all four fail: the summary says none', async () => {
    const responses = Object.fromEntries(PAIRS.map((p) => [p.url, [{ status: 503 }]]));
    const { core } = await run({ responses });
    assert.equal(core.calls.info.at(-1), `specs in ${specsDir}: none`);
    assert.equal(core.calls.error.length, 4);
    assert.equal(core.calls.setFailed.length, 1);
    assert.ok(core.calls.setFailed[0].startsWith('4 of 4 spec(s) not fetched: '), core.calls.setFailed[0]);
  });

  test('6. the URLs are fetched in parallel: a slow URL does not delay the first attempt of the others', { timeout: 5000 }, async () => {
    const order = [];
    let releaseSlow;
    const slow = new Promise((resolve) => {
      releaseSlow = resolve;
    });
    const fetchImpl = async (url) => {
      order.push(url);
      if (url === SWAGGER_URL) {
        await slow;
      }
      if (order.length === PAIRS.length) {
        releaseSlow();
      }
      return { status: 200, headers: { get: () => null }, text: async () => OPENAPI_BODY(url) };
    };
    const core = makeCore();
    await fetchSpecs({ core, env: { SPECS_DIR: specsDir, SPECS }, fetchImpl, sleep: makeSleep().impl });
    assert.deepEqual(order, PAIRS.map((p) => p.url));
    assert.deepEqual(core.calls.setFailed, []);
    assert.deepEqual(
      core.calls.info.filter((m) => m.startsWith('fetched ')).map((m) => m.split(' ')[1]),
      PAIRS.map((p) => p.file),
    );
  });

  test('7. a multi-byte body logs its byte length, not its string length', async () => {
    const body = JSON.stringify({ openapi: '3.1.2', info: { title: 'Élastic — ☁' }, paths: { '/a': {} } });
    assert.notEqual(Buffer.byteLength(body), body.length);
    const responses = { ...healthyResponses(), [SERVERLESS_URL]: [{ status: 200, body }] };
    const { core } = await run({ responses });
    assert.ok(core.calls.info.includes(`fetched elastic-cloud-serverless.json (${Buffer.byteLength(body)} bytes)`));
    assert.equal(fs.readFileSync(path.join(specsDir, 'elastic-cloud-serverless.json'), 'utf8'), body);
  });

  test('8. SPECS missing fails the step and fetches nothing', async () => {
    const { core, fetch: f } = await run({ env: { SPECS_DIR: specsDir } });
    assert.equal(core.calls.setFailed.length, 1);
    assert.ok(core.calls.setFailed[0].startsWith('invalid SPECS: '), core.calls.setFailed[0]);
    assert.deepEqual(f.calls, {});
  });

  test('8. SPECS_DIR missing or blank fails the step and fetches nothing', async () => {
    for (const env of [{ SPECS }, { SPECS_DIR: '  ', SPECS }]) {
      const { core, fetch: f } = await run({ env });
      assert.deepEqual(core.calls.setFailed, ['SPECS_DIR is not set']);
      assert.deepEqual(f.calls, {});
    }
  });

  test('8. a malformed SPECS line fails the step and fetches nothing', async () => {
    const { core, fetch: f } = await run({ specs: `${SPECS}\nnot-a-url x.json` });
    assert.equal(core.calls.setFailed.length, 1);
    assert.ok(core.calls.setFailed[0].startsWith('invalid SPECS: '), core.calls.setFailed[0]);
    assert.ok(core.calls.setFailed[0].includes('not-a-url x.json'), core.calls.setFailed[0]);
    assert.deepEqual(f.calls, {});
  });

  test('9. a failed write fails the step and leaves no partial file; the other pairs still run', async () => {
    const thirdTarget = path.join(specsDir, 'cloud-enterprise.json');
    const fsImpl = {
      ...fs,
      writeFileSync: (target, ...rest) => {
        if (target === thirdTarget) {
          fs.writeFileSync(target, 'partial');
          throw new Error('ENOSPC: no space left on device');
        }
        return fs.writeFileSync(target, ...rest);
      },
    };
    const { core, fetch: f } = await run({ fsImpl });
    assert.deepEqual(filesIn(specsDir), ['api-elastic-cloud-swagger.json', 'cloud.json', 'elastic-cloud-serverless.json']);
    assert.deepEqual(core.calls.error, ['could not save cloud-enterprise.json: ENOSPC: no space left on device']);
    assert.equal(f.calls[SERVERLESS_URL], 1);
    assert.deepEqual(core.calls.setFailed, ['1 of 4 spec(s) not fetched: cloud-enterprise.json; the research needs every spec']);
  });

  test('10. mkdirSync throws: the step fails and nothing is fetched', async () => {
    const fsImpl = {
      ...fs,
      mkdirSync: () => {
        throw new Error('EACCES: permission denied');
      },
    };
    const { core, fetch: f } = await run({ fsImpl });
    assert.deepEqual(core.calls.setFailed, [`could not create ${specsDir}: EACCES: permission denied`]);
    assert.deepEqual(f.calls, {});
  });

  test('11. a file already at the target is removed when its URL fails', async () => {
    const target = path.join(specsDir, 'elastic-cloud-serverless.json');
    fs.writeFileSync(target, 'stale');
    const responses = { ...healthyResponses(), [SERVERLESS_URL]: [{ status: 503 }] };
    await run({ responses });
    assert.equal(fs.existsSync(target), false);
  });

  test('12. a response without headers is still read', async () => {
    const fetchImpl = async () => ({ status: 503, text: async () => '' });
    const s = makeSleep();
    const result = await fetchOne({ url: SERVERLESS_URL, fetchImpl, sleep: s.impl });
    assert.equal(result.ok, false);
    assert.equal(result.reason, 'HTTP 503');
    assert.deepEqual(s.delays, RETRY_DELAYS_MS);
  });

  test('13. a cleanup that throws after a failed fetch still reports the fetch failure', async () => {
    const fsImpl = {
      ...fs,
      rmSync: () => {
        throw new Error('ERR_FS_EISDIR');
      },
    };
    const responses = { ...healthyResponses(), [SERVERLESS_URL]: [{ status: 404 }] };
    const { core } = await run({ responses, fsImpl });
    assert.equal(core.calls.error.length, 1);
    assert.ok(core.calls.error[0].startsWith(`could not fetch ${SERVERLESS_URL} after ${ATTEMPTS} attempts (HTTP 404)`), core.calls.error[0]);
    assert.deepEqual(core.calls.setFailed, ['1 of 4 spec(s) not fetched: elastic-cloud-serverless.json; the research needs every spec']);
  });
});

// --- Lock wiring -------------------------------------------------------------------------------

const FETCH_STEP = 'Fetch Elastic Cloud API specifications';
const JOB_LINE = /^ {2}[a-z_]+:$/;
const STEP_START = '      - ';
const PINNED_GITHUB_SCRIPT = /uses: actions\/github-script@([0-9a-f]{40})\b/;

describe('research-factory lock: spec fetch step', () => {
  const LINES = fs.readFileSync(LOCK_PATH, 'utf8').split('\n');

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

  /**
   * One step of a job, found by name. gh-aw writes a step's keys in alphabetical order, so a step
   * can start with `      - env:` and carry its name on a later `        name:` line: find the name
   * line, walk back to the nearest step start, and forward to the line before the next one.
   */
  function stepByName(lines, name) {
    const nameLines = [`      - name: ${name}`, `        name: ${name}`];
    const hits = lines.flatMap((l, i) => (nameLines.includes(l) ? [i] : []));
    assert.equal(hits.length, 1, `expected exactly one "${name}" step, found ${hits.length}`);
    let start = hits[0];
    while (start > 0 && !lines[start].startsWith(STEP_START)) {
      start -= 1;
    }
    let end = hits[0] + 1;
    while (end < lines.length && !lines[end].startsWith(STEP_START)) {
      end += 1;
    }
    return { start, text: lines.slice(start, end).join('\n') };
  }

  const agent = jobLines('agent');

  test('the agent job has exactly one fetch step name line', () => {
    assert.equal(agent.filter((l) => l === `        name: ${FETCH_STEP}`).length, 1);
  });

  test('the step runs research-factory/fetch-specs.js', () => {
    assert.ok(stepByName(agent, FETCH_STEP).text.includes('research-factory/fetch-specs.js'));
  });

  test('the step carries the four <url> <file> pairs', () => {
    const { text } = stepByName(agent, FETCH_STEP);
    for (const p of PAIRS) {
      assert.ok(text.includes(`${p.url} ${p.file}`), `missing pair: ${p.url} ${p.file}`);
    }
    assert.ok(text.includes('https://www.elastic.co/docs/api/cloud-serverless.json elastic-cloud-serverless.json'));
  });

  test('the step pins actions/github-script to the SHA the pre-activation job uses', () => {
    const ours = stepByName(agent, FETCH_STEP).text.match(PINNED_GITHUB_SCRIPT);
    assert.ok(ours, 'the fetch step does not use a SHA-pinned actions/github-script');
    const qualify = stepByName(jobLines('pre_activation'), 'Qualify trigger event').text.match(PINNED_GITHUB_SCRIPT);
    assert.ok(qualify, 'Qualify trigger event does not use a SHA-pinned actions/github-script');
    assert.equal(ours[1], qualify[1]);
  });

  test('a spec that could not be fetched fails the run, so the step has no continue-on-error', () => {
    assert.ok(!stepByName(agent, FETCH_STEP).text.includes('continue-on-error'));
  });

  test('the step runs after the checkout and before the agent', () => {
    const { start } = stepByName(agent, FETCH_STEP);
    const checkout = agent.indexOf('      - name: Checkout repository (gh-aw default)');
    const execute = agent.indexOf('      - name: Execute Claude Code CLI');
    assert.notEqual(checkout, -1);
    assert.notEqual(execute, -1);
    assert.ok(checkout < start, 'the fetch step runs before the checkout that provides the helper');
    assert.ok(start < execute, 'the fetch step runs after the agent starts');
  });

  test('the lock does not contain the dead serverless URL', () => {
    assert.ok(!LINES.some((l) => l.includes('doc/elastic-cloud-serverless.json')));
  });

  test('parseSpecList accepts the SPECS block of research-factory-issue.md', () => {
    const md = fs.readFileSync(WORKFLOW_PATH, 'utf8').split('\n');
    const start = md.indexOf('      SPECS: |');
    assert.notEqual(start, -1, 'research-factory-issue.md has no SPECS block');
    const end = md.indexOf('    with:', start);
    assert.notEqual(end, -1, 'the SPECS block is not followed by with:');
    const block = md.slice(start + 1, end).map((l) => l.replace(/^ {8}/, '')).join('\n');
    assert.deepEqual(parseSpecList(block), PAIRS);
  });
});
