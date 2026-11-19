/**
 * Captures the free text a maintainer typed after the `/<factory>` slash command.
 *
 * The result is the highest-precedence agent input, so it is sanitized here: it is the only
 * agent-visible value that does not pass through the sanitize-context runner. Issue events
 * carry no command, and the issue body is never substituted for it.
 *
 * The direction is staged as a file in the context artifact, never as a job-output value: gh-aw
 * wires a job output into the agent step's env, and a large multibyte slash-command free-text
 * would exceed the kernel's MAX_ARG_STRLEN (~128 KB) and abort activation with E2BIG — after
 * intake has already removed the trigger label and set the phase label.
 */
const { getFactoryName, getFactoryContextDir } = require('./_factory-context.js');
const { sanitizeUserContent } = require('../sanitize-context.js');

module.exports = async function ({ github, context, core }) {

  const fs = require('fs');
  const factoryName = getFactoryName();

  let humanDirection = '';
  if (context.eventName === 'issue_comment') {
    const body = context.payload.comment?.body ?? '';
    const stripped = body.replace(new RegExp(`^\\s*\\/${factoryName}\\s*`), '').trim();
    humanDirection = sanitizeUserContent(stripped);
  }

  // Always write the file, even empty: the artifact upload uses if-no-files-found: error and
  // needs a stable file set, and the agent prompt references this path unconditionally.
  const dir = getFactoryContextDir();
  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(`${dir}/human_direction.md`, humanDirection);

  core.info(
    context.eventName === 'issue_comment'
      ? `Captured ${humanDirection.length} characters of human direction to ${dir}/human_direction.md`
      : `Not an issue_comment event; wrote an empty ${dir}/human_direction.md`,
  );
};
