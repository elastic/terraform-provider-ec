/**
 * Detection pre-step: stages the exact research comment this run will post for gh-aw's threat
 * detector.
 *
 * Runs in the gh-aw `detection` job, as a `safe-outputs.threat-detection.steps` entry, after gh-aw
 * has prepared the detector's artifacts directory and before the detector executes. It never runs
 * in the `update_research_comment` (publish) job, and the publish job never reads what it writes.
 *
 * What it does:
 *
 *   take the draft from its canonical artifact path -> compute the published body with the
 *   publish handler's own `preparePublishedBody` -> bound it to `MAX_ENV_PAYLOAD_BYTES` (head and
 *   tail kept, a visible notice between them) -> write it to
 *   `<detectionDir>/comment-memory/research-comment.md`, where the detector lists `.md` files as
 *   content to read and analyse.
 *
 * Rules that are easy to "simplify" into bugs:
 *
 * - **Use the handler's exports, never a copy of them.** `resolveDraftPath` and
 *   `preparePublishedBody` are imported from `update-research-comment.js`. The staged file is only
 *   worth anything while it is the string the publish job posts; a re-implementation drifts.
 * - **Only `resolveDraftPath(...).candidates[0]` is accepted.** That path is where the agent
 *   artifact's entry always lands, and it holds the same artifact bytes in this job and in the
 *   publish job. The resolver's other routes (an absolute path, a search of the download root) can
 *   find a file in one job and not in the other, because the two jobs' download roots are not the
 *   same directory tree. A draft found by any other route is never read or staged here.
 * - **The staged file is never named `draft-final.md`.** The handler's fallback search matches on
 *   that name, so a copy with that name could be found by it.
 *
 * Outcome (the request is `AGENT_OUTPUT_TYPES` containing `update_research_comment`, the same test
 * the publish job's `if:` applies; it is used only to fail, never to skip staging):
 *
 *   | Situation                                        | Staged file              | Step     |
 *   | ------------------------------------------------ | ------------------------ | -------- |
 *   | AGENT_OUTPUT_TYPES undefined or empty            | none                     | fails    |
 *   | canonical draft, body within budget              | whole body               | succeeds |
 *   | canonical draft, body over budget                | bounded copy with notice | fails    |
 *   | canonical draft, handler rejects it              | none                     | succeeds |
 *   | no canonical draft, comment requested            | none                     | fails    |
 *   | no canonical draft, comment not requested        | none                     | succeeds |
 *   | any exception (cannot write, handler throws)     | none or partial          | fails    |
 *
 * A failed step fails the detection job, and the publish job requires a successful detection job.
 */
'use strict';

const fs = require('fs');
const path = require('path');

const { MAX_ENV_PAYLOAD_BYTES, byteLength, sliceToBytes } = require('../lib/env-payload.js');
const { preparePublishedBody, resolveDraftPath } = require('./update-research-comment.js');

const DETECTION_COPY_MAX_BYTES = MAX_ENV_PAYLOAD_BYTES;
const DETECTION_COPY_DIRNAME = 'comment-memory';
const DETECTION_COPY_FILENAME = 'research-comment.md';
const PUBLISH_OUTPUT_TYPE = 'update_research_comment';
/** The detection job downloads the agent artifact into `sourceRoot`; the detector reads `detectionDir`. */
const DEFAULT_PATHS = Object.freeze({ sourceRoot: '/tmp/gh-aw', detectionDir: '/tmp/gh-aw/threat-detection' });

/**
 * The notice placed where a bounded copy omits bytes.
 *
 * @param {number} omittedBytes
 * @param {number} totalBytes
 * @returns {string}
 */
function elisionNotice(omittedBytes, totalBytes) {
  return `\n\n[... detection copy truncated: ${omittedBytes} of ${totalBytes} bytes of the published comment omitted here ...]\n\n`;
}

/**
 * The longest suffix of whole code points whose UTF-8 length is at most `maxBytes`.
 *
 * @param {string} text
 * @param {number} maxBytes
 * @returns {string}
 */
function sliceTailToBytes(text, maxBytes) {
  const value = String(text ?? '');
  if (maxBytes <= 0) {
    return '';
  }
  if (byteLength(value) <= maxBytes) {
    return value;
  }
  // Spreading the string yields whole code points, so a surrogate pair is never cut in half.
  const points = [...value];
  let used = 0;
  let start = points.length;
  while (start > 0) {
    const size = byteLength(points[start - 1]);
    if (used + size > maxBytes) {
      break;
    }
    used += size;
    start -= 1;
  }
  return points.slice(start).join('');
}

/**
 * Bounds the detection copy to a UTF-8 byte budget, keeping the head and the tail.
 *
 * The tail is kept because the `### Quality gate` section and the pipeline metadata sit at the end
 * of the comment. The notice is reserved at its worst-case size before the text is cut, so the
 * result never exceeds `maxBytes`.
 *
 * @param {string} text
 * @param {number} [maxBytes]
 * @returns {{text: string, truncated: boolean, totalBytes: number, omittedBytes: number}}
 */
function boundDetectionCopy(text, maxBytes = DETECTION_COPY_MAX_BYTES) {
  const totalBytes = byteLength(text);
  if (totalBytes <= maxBytes) {
    return { text, truncated: false, totalBytes, omittedBytes: 0 };
  }

  const reserve = byteLength(elisionNotice(totalBytes, totalBytes));
  const room = maxBytes - reserve;
  if (room < 2) {
    throw new RangeError(`a ${maxBytes}-byte budget cannot hold the ${reserve}-byte truncation notice and any content`);
  }

  const head = sliceToBytes(text, Math.floor(room / 2));
  const tail = sliceTailToBytes(text, room - Math.floor(room / 2));
  const omittedBytes = totalBytes - byteLength(head) - byteLength(tail);

  return { text: `${head}${elisionNotice(omittedBytes, totalBytes)}${tail}`, truncated: true, totalBytes, omittedBytes };
}

/**
 * Computes the published body from the canonical draft and writes it into the detector's artifacts
 * directory. Knows nothing about whether a comment was requested. Exceptions are not caught.
 *
 * @param {{sourceRoot: string, detectionDir: string}} paths
 * @returns {{status: 'no-canonical-draft', canonicalPath: string, foundElsewhere: string | null}
 *   | {status: 'rejected', draftPath: string, fatals: string[]}
 *   | {status: 'staged', draftPath: string, stagedPath: string, publishedChars: number,
 *      publishedBytes: number, stagedBytes: number, truncated: boolean, omittedBytes: number}}
 */
function stageDetectionCopy({ sourceRoot, detectionDir }) {
  // Pointing the handler's resolver at `<sourceRoot>/agent_output.json` makes it treat
  // `sourceRoot` as the download root, as the publish job's GH_AW_AGENT_OUTPUT does for its own.
  const resolved = resolveDraftPath({ GH_AW_AGENT_OUTPUT: path.join(sourceRoot, 'agent_output.json') });
  const canonicalPath = resolved.candidates[0];
  if (resolved.path !== canonicalPath) {
    return { status: 'no-canonical-draft', canonicalPath, foundElsewhere: resolved.path || null };
  }

  const draft = fs.readFileSync(canonicalPath, 'utf8');
  const prepared = preparePublishedBody(draft);
  if (!prepared.body) {
    return { status: 'rejected', draftPath: canonicalPath, fatals: prepared.fatals };
  }

  const bounded = boundDetectionCopy(prepared.body);

  const dir = path.join(detectionDir, DETECTION_COPY_DIRNAME);
  fs.mkdirSync(dir, { recursive: true });
  // The detector refuses a symlink in this position, which would leave the copy unlisted.
  if (!fs.lstatSync(dir).isDirectory()) {
    throw new Error(`${dir} is not a directory`);
  }
  const target = path.join(dir, DETECTION_COPY_FILENAME);
  fs.rmSync(target, { force: true });
  fs.writeFileSync(target, bounded.text, { encoding: 'utf8', flag: 'wx' });

  return {
    status: 'staged',
    draftPath: canonicalPath,
    stagedPath: target,
    publishedChars: prepared.body.length,
    publishedBytes: bounded.totalBytes,
    stagedBytes: byteLength(bounded.text),
    truncated: bounded.truncated,
    omittedBytes: bounded.omittedBytes,
  };
}

/**
 * github-script entry. Never logs draft or body text.
 *
 * @param {{core: object}} ctx
 * @param {{paths?: {sourceRoot: string, detectionDir: string}, outputTypes?: string}} [options]
 */
module.exports = async function ({ core }, options = {}) {
  const paths = options.paths || DEFAULT_PATHS;
  const outputTypes = options.outputTypes !== undefined ? options.outputTypes : process.env.AGENT_OUTPUT_TYPES;

  if (typeof outputTypes !== 'string' || outputTypes.trim() === '') {
    core.setFailed('AGENT_OUTPUT_TYPES is empty: cannot tell whether a comment will be published');
    return;
  }
  const requested = outputTypes.includes(PUBLISH_OUTPUT_TYPE);

  let result;
  try {
    result = stageDetectionCopy(paths);
  } catch (err) {
    core.setFailed(`cannot stage the research comment body for detection: ${err.stack || err.message}`);
    return;
  }

  core.setOutput('detection_copy_status', result.status);

  switch (result.status) {
    case 'staged':
      if (result.truncated) {
        core.setFailed(
          `the comment is ${result.publishedBytes} bytes, over the ${DETECTION_COPY_MAX_BYTES}-byte detection budget; ` +
            `it will not be published (a bounded copy of ${result.stagedBytes} bytes is staged at ${result.stagedPath}). ` +
            'Re-running this job cannot fix it; re-apply the research-factory label for a fresh run.',
        );
      } else {
        core.info(
          `staged the research comment body for detection at ${result.stagedPath}: ${result.stagedBytes} bytes, ` +
            `${result.publishedChars} characters, from ${result.draftPath}`,
        );
      }
      return;
    case 'rejected':
      core.info(`the publish job will reject the draft at ${result.draftPath}; nothing staged: ${result.fatals.join('; ')}`);
      return;
    default:
      if (requested) {
        core.setFailed(
          `${PUBLISH_OUTPUT_TYPE} was requested but there is no draft at ${result.canonicalPath}; nothing will be published. ` +
            'A missing draft can also mean secret redaction failed, which drops the whole agent artifact. ' +
            'Re-apply the research-factory label for a fresh run.' +
            (result.foundElsewhere ? ` A draft-final.md was found at ${result.foundElsewhere}, which is not the canonical path.` : ''),
        );
      } else {
        core.info(`no draft at ${result.canonicalPath} and no ${PUBLISH_OUTPUT_TYPE} request; nothing staged`);
      }
  }
};

module.exports.DETECTION_COPY_MAX_BYTES = DETECTION_COPY_MAX_BYTES;
module.exports.DETECTION_COPY_DIRNAME = DETECTION_COPY_DIRNAME;
module.exports.DETECTION_COPY_FILENAME = DETECTION_COPY_FILENAME;
module.exports.PUBLISH_OUTPUT_TYPE = PUBLISH_OUTPUT_TYPE;
module.exports.DEFAULT_PATHS = DEFAULT_PATHS;
module.exports.elisionNotice = elisionNotice;
module.exports.sliceTailToBytes = sliceTailToBytes;
module.exports.boundDetectionCopy = boundDetectionCopy;
module.exports.stageDetectionCopy = stageDetectionCopy;
