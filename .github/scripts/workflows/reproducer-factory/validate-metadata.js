/**
 * The `schema_version` 1.0 constraint set for the reproducer comment's pipeline metadata.
 *
 * Dual-use, deliberately: the agent runs it over its draft in the pre-flight
 * (`node .github/scripts/workflows/reproducer-factory/validate-metadata.js <draft>`), and the
 * safe-output handler imports `validateMetadata` over the same block. One definition, so the check the
 * agent passes and the check the handler applies cannot drift.
 *
 * The `outcome` enum is the machine-checkable half of AC #1: an LLM spelling the enum is not an
 * acceptance criterion, so the handler fails closed on an `outcome` outside the three values. In
 * particular `reproduced` (the Elastic Stack provider's spelling) is rejected — without a live run the
 * stronger word over-claims, so #4384 names the three outcomes explicitly.
 *
 * There is no fourth `reproduced-live` outcome: the agent cannot know the acceptance result when it
 * emits, so the human reads the repro PR's CI for the live confirmation.
 *
 * `pr_url` is part of the schema but the agent does not write it today. The comment and the PR are
 * emitted in the same run, so the agent does not know the PR's URL; the prompt has it link the PR list
 * filtered by the deterministic repro branch instead. The field is validated shape-only when present,
 * so a later job that knows the URL can add it without a schema bump.
 */
'use strict';

const { extractMetadata } = require('./locate-metadata.js');

const SCHEMA_VERSION = '1.0';
const OUTCOMES = ['reproduced-by-inspection', 'cannot-reproduce', 'appears-fixed'];

function isObject(value) {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function isNonEmptyString(value) {
  return typeof value === 'string' && value.trim() !== '';
}

/**
 * Checks a parsed metadata object against the schema_version 1.0 constraint set.
 *
 * @param {unknown} metadata
 * @returns {{ errors: string[] }} Empty when the metadata conforms.
 */
function validateMetadata(metadata) {
  const errors = [];

  if (!isObject(metadata)) {
    return { errors: ['metadata: must be a JSON object'] };
  }

  if (metadata.schema_version !== SCHEMA_VERSION) {
    errors.push(`schema_version: must be "${SCHEMA_VERSION}", got ${JSON.stringify(metadata.schema_version)}`);
  }

  if (!OUTCOMES.includes(metadata.outcome)) {
    errors.push(`outcome: must be one of ${OUTCOMES.join('|')}, got ${JSON.stringify(metadata.outcome)}`);
  }

  // Optional, but if present it must be a usable name — an empty string is a drafting slip, not an
  // absent field, and would otherwise publish as a dangling "see test ``" in the comment prose.
  if (metadata.test_name !== undefined && !isNonEmptyString(metadata.test_name)) {
    errors.push(`test_name: when present, must be a non-empty string, got ${JSON.stringify(metadata.test_name)}`);
  }

  // Optional. Validated shape-only (a non-empty https URL) when present; never required, because the
  // agent cannot know the PR URL when it writes the comment (see the file header).
  if (metadata.pr_url !== undefined) {
    if (!isNonEmptyString(metadata.pr_url) || !/^https:\/\/\S+$/.test(metadata.pr_url)) {
      errors.push(`pr_url: when present, must be a non-empty https URL, got ${JSON.stringify(metadata.pr_url)}`);
    }
  }

  return { errors };
}

/**
 * Extracts the metadata block from a comment body and validates it.
 *
 * @param {string} body
 * @returns {{ errors: string[], metadata?: object }}
 */
function validateCommentBody(body) {
  const extracted = extractMetadata(body);
  if (extracted.error) {
    return { errors: [extracted.error] };
  }
  return { ...validateMetadata(extracted.metadata), metadata: extracted.metadata };
}

function main() {
  const fs = require('fs');
  const draftPath = process.argv[2];

  if (!draftPath) {
    console.error('usage: node validate-metadata.js <draft.md>');
    process.exit(2);
  }

  let body;
  try {
    body = fs.readFileSync(draftPath, 'utf8');
  } catch (err) {
    console.error(`cannot read ${draftPath}: ${err.message}`);
    process.exit(2);
  }

  const { errors } = validateCommentBody(body);
  if (errors.length > 0) {
    console.error(`${draftPath}: pipeline metadata does not satisfy schema_version ${SCHEMA_VERSION}:`);
    for (const error of errors) {
      console.error(`  - ${error}`);
    }
    process.exit(1);
  }

  console.log(`${draftPath}: pipeline metadata OK (schema_version ${SCHEMA_VERSION})`);
}

if (require.main === module) {
  main();
}

module.exports = {
  validateMetadata,
  validateCommentBody,
  SCHEMA_VERSION,
  OUTCOMES,
};
