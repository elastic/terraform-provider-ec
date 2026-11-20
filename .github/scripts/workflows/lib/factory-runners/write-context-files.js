const { getFactoryName, getFactoryContextDir, getFactoryDetectionInputsDir } = require('./_factory-context.js');
const { sanitizeUserContent } = require('../sanitize-context.js');
const { renderCensus } = require('../inputs-census.js');

module.exports = async function ({ github, context, core }) {

  const fs = require('fs');
  const factoryName = getFactoryName();
  const dir = getFactoryContextDir();
  const priorCommentFilename = `prior_${factoryName.replace(/-factory$/, '')}_comment.md`;

  // Sanitise each payload once and reuse it for both directories, so the detection copies are
  // byte-identical to the files the agent reads.
  const issueBody = sanitizeUserContent(process.env.ISSUE_BODY);
  const issueComments = sanitizeUserContent(process.env.ISSUE_COMMENTS);
  const priorComment = sanitizeUserContent(process.env.PRIOR_FACTORY_COMMENT || '');

  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(`${dir}/issue_body.md`, issueBody);
  fs.writeFileSync(`${dir}/issue_comments.md`, issueComments);
  fs.writeFileSync(`${dir}/${priorCommentFilename}`, priorComment);
  core.info(`Wrote sanitized issue context files to ${dir}/`);

  // Second copy for the threat-detection job: the same sanitised bytes plus a census.
  const detectionDir = getFactoryDetectionInputsDir();
  fs.mkdirSync(detectionDir, { recursive: true });
  for (const [filename, content] of [
    ['issue_body.md', issueBody],
    ['issue_comments.md', issueComments],
    [priorCommentFilename, priorComment],
  ]) {
    fs.writeFileSync(`${detectionDir}/${filename}`, content);
  }
  fs.writeFileSync(`${detectionDir}/census.md`, renderCensus({
    sources: [
      { name: 'issue_body.md', raw: process.env.ISSUE_BODY, sanitized: issueBody },
      { name: 'issue_comments.md', raw: process.env.ISSUE_COMMENTS, sanitized: issueComments },
      { name: priorCommentFilename, raw: process.env.PRIOR_FACTORY_COMMENT, sanitized: priorComment },
    ],
  }));
  core.info(`Wrote detection inputs (3 files + census.md) to ${detectionDir}/`);
};
