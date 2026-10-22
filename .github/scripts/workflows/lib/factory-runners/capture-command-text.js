/**
 * Captures the free text a maintainer typed after the `/<factory>` slash command.
 *
 * The result is the highest-precedence agent input, so it is sanitized here: it is the only
 * agent-visible value that does not pass through the sanitize-context runner. Issue events
 * carry no command, and the issue body is never substituted for it.
 */
const { getFactoryName } = require('./_factory-context.js');
const { sanitizeUserContent } = require('../sanitize-context.js');

module.exports = async function ({ github, context, core }) {

  const fs = require('fs');
  const crypto = require('crypto');
  const factoryName = getFactoryName();

  let humanDirection = '';
  if (context.eventName === 'issue_comment') {
    const body = context.payload.comment?.body ?? '';
    const stripped = body.replace(new RegExp(`^\\s*\\/${factoryName}\\s*`), '').trim();
    humanDirection = sanitizeUserContent(stripped);
  }

  // Multi-line directions are normal, so never core.setOutput here.
  const eofDelim = `EOF_${crypto.randomUUID().replace(/-/g, '')}`;
  fs.appendFileSync(process.env.GITHUB_OUTPUT, `human_direction<<${eofDelim}\n${humanDirection}\n${eofDelim}\n`);

  if (context.eventName === 'issue_comment') {
    core.info(`Captured ${humanDirection.length} characters of human direction from the slash command.`);
  } else {
    core.info(`Not an issue_comment event; human_direction is empty.`);
  }
};
