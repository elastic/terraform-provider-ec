const { getFactoryContextDir } = require('./_factory-context.js');
const { sanitizeUserContent } = require('../sanitize-context.js');

/**
 * Reads a staged file, returning '' when the env var is unset/empty or the file is missing or
 * unreadable. Used so the issue body and comment history travel by PATH, not by value: a single
 * GitHub Actions step env var is bounded by the kernel's MAX_ARG_STRLEN (~128 KB), and the comment
 * budget / issue body are measured in characters, so multibyte (e.g. CJK) content can exceed that
 * byte limit and abort the step with E2BIG before the context upload and trigger-label cleanup.
 */
function readStagedFile(envName, core) {
  const path = process.env[envName] || '';
  if (!path) return '';
  try {
    return require('fs').readFileSync(path, 'utf8');
  } catch (err) {
    core.info(`Could not read ${envName} (${path}): ${err.message}; treating as empty.`);
    return '';
  }
}

module.exports = async function ({ github, context, core }) {

  const fs = require('fs');

  const dir = getFactoryContextDir();

  // Body and comments arrive as file paths (byte-safe). The title is bounded by GitHub to a short
  // single line, so it is fine to pass by value.
  const body = readStagedFile('INPUT_ISSUE_BODY_PATH', core);
  const comments = readStagedFile('INPUT_HUMAN_COMMENTS_PATH', core);

  const sanitizedBody = sanitizeUserContent(body);
  const sanitizedComments = sanitizeUserContent(comments);

  // A title is single-line, so plain setOutput is enough. Unset ISSUE_TITLE yields an empty
  // output, which keeps callers that do not pass a title working unchanged.
  core.setOutput('sanitized_issue_title', sanitizeUserContent(process.env.ISSUE_TITLE || ''));

  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(`${dir}/issue_body.md`, sanitizedBody);
  fs.writeFileSync(`${dir}/issue_comments.md`, sanitizedComments);
  core.info(`Wrote sanitized issue context files to ${dir}/`);
};
