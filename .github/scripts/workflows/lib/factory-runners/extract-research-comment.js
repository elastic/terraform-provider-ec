/**
 * Reads the staged comment history, extracts the research-factory sticky comment, and writes
 * the sanitized result into the factory context directory the agent artifact is built from.
 *
 * The serialized human comments travel to sanitize-context by PATH, not through a GITHUB_OUTPUT
 * value, so a large multibyte comment history cannot overflow a step env var and abort intake
 * with E2BIG. The research comment body is likewise never an output: the agent reads
 * research_comment.md from the artifact.
 */
const { getFactoryName, getFactoryContextDir } = require('./_factory-context.js');
const { sanitizeForConcatenation, findResearchComment } = require('../sanitize-context.js');
const { serializeIssueComments } = require('../factory-issue-comments.js');
const { RESEARCH_COMMENT_MARKER } = require('../intake/research-factory-constants.js');

module.exports = async function ({ github, context, core }) {

  const fs = require('fs');

  let comments = [];
  let truncated = false;
  const commentsPath = process.env.INPUT_COMMENTS_PATH || '';
  if (commentsPath) {
    try {
      const parsed = JSON.parse(fs.readFileSync(commentsPath, 'utf8'));
      comments = Array.isArray(parsed?.comments) ? parsed.comments : [];
      truncated = parsed?.truncated === true;
    } catch (err) {
      core.info(`Could not read staged comments from ${commentsPath}: ${err.message}`);
    }
  } else {
    core.info('INPUT_COMMENTS_PATH is empty; treating the comment history as empty.');
  }

  const researchComment = findResearchComment(comments, RESEARCH_COMMENT_MARKER);
  // Same strength as the comment blob below, so one file does not carry two sanitisation levels.
  const sanitizedBody = researchComment ? sanitizeForConcatenation(researchComment.body) : '';

  // The staged JSON is parsed from disk, so drop nullish entries and read `author` defensively:
  // serializeIssueComments dereferences every element it is handed.
  const humanComments = comments.filter(c => c != null && !String(c.author ?? '').endsWith('[bot]'));
  // Sanitize per field, never once over the concatenation, and to a fixed point: a single pass
  // can leave — or reassemble — a literal `<!--`, and the downstream pass over the joined blob
  // would then delete from it to the end, erasing every later comment and its `**@author**`
  // attribution from the agent's context.
  const serialized = serializeIssueComments({
    comments: humanComments.map(c => ({
      author: sanitizeForConcatenation(String(c.author ?? '')),
      createdAt: sanitizeForConcatenation(String(c.createdAt ?? '')),
      body: sanitizeForConcatenation(String(c.body ?? '')),
    })),
    truncated,
  });

  // Stage the serialized human comments next to the raw comment JSON (RUNNER_TEMP, not the
  // uploaded context dir) and hand sanitize-context the path, not the value — byte-safe transport.
  const runnerTemp = process.env.RUNNER_TEMP || require('os').tmpdir();
  const rawDir = `${runnerTemp}/${getFactoryName()}-raw`;
  const humanCommentsPath = `${rawDir}/human_comments.txt`;
  fs.mkdirSync(rawDir, { recursive: true });
  fs.writeFileSync(humanCommentsPath, serialized);
  core.setOutput('human_comments_path', humanCommentsPath);
  core.setOutput('research_comment_found', researchComment ? 'true' : 'false');

  // Always write the file, even empty: the artifact upload uses if-no-files-found: error and
  // needs a stable file set.
  const dir = getFactoryContextDir();
  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(`${dir}/research_comment.md`, sanitizedBody);
  core.info(
    researchComment
      ? `Wrote the research comment to ${dir}/research_comment.md`
      : `No research comment found; wrote an empty ${dir}/research_comment.md`,
  );
};
