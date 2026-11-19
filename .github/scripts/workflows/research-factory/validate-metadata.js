/**
 * The single `schema_version` 1.1 constraint set for the research comment's pipeline metadata.
 *
 * Dual-use, deliberately: the agent runs it over its draft in the pre-flight
 * (`node .github/scripts/workflows/research-factory/validate-metadata.js <draft>`), and the
 * safe-output handler imports `validateMetadata` over the same block. One definition, so the
 * check the agent passes and the check the handler applies cannot drift.
 *
 * Violating the constraint set never discards a run: the handler publishes the comment and drops
 * the `schema_version` claim, because nothing consumes the metadata today and a 60-minute run is
 * worth more than a malformed optional field. The drop is best-effort rather than guaranteed — the
 * handler skips it in the two cases where applying it would cost the run instead (re-serialising
 * would introduce an HTML comment opener, or would breach the comment size limit), and warns.
 */
'use strict';

const { extractMetadata } = require('./locate-metadata.js');

const SCHEMA_VERSION = '1.1';
const KEBAB_CASE = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const CONFIDENCES = ['high', 'medium', 'low'];
const SCOPES = ['small', 'medium', 'large', 'unknown'];
const REFERENCE_TYPES = ['elastic-docs', 'repo-path', 'issue', 'pr', 'external'];
const CHECKLIST_ITEMS = ['grounded', 'mapped', 'compatible', 'versioned', 'testable', 'idiomatic'];
const CRITIC_STATUSES = ['ok', 'unavailable', 'error'];
const OUTCOMES = ['ready-for-change-factory', 'research-needs-human'];
const MAX_ROUNDS = 5;

function isObject(value) {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function isNonEmptyString(value) {
  return typeof value === 'string' && value.trim() !== '';
}

function isScore(value) {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 100;
}

function validateRecommendation(recommendation, errors) {
  if (!isObject(recommendation)) {
    errors.push('recommendation: must be an object');
    return;
  }
  if (!isNonEmptyString(recommendation.spine) || !KEBAB_CASE.test(recommendation.spine)) {
    errors.push(`recommendation.spine: must be a kebab-case identifier, got ${JSON.stringify(recommendation.spine)}`);
  }
  if (!Number.isInteger(recommendation.approach_index) || recommendation.approach_index < 0) {
    errors.push(
      `recommendation.approach_index: must be an integer >= 0, got ${JSON.stringify(recommendation.approach_index)}`,
    );
  }
  if (recommendation.confidence !== undefined && !CONFIDENCES.includes(recommendation.confidence)) {
    errors.push(
      `recommendation.confidence: must be one of ${CONFIDENCES.join('|')}, got ${JSON.stringify(recommendation.confidence)}`,
    );
  }
}

function validateOpenQuestions(openQuestions, errors) {
  if (openQuestions === undefined) {
    return;
  }
  if (!Array.isArray(openQuestions)) {
    errors.push('open_questions: must be an array');
    return;
  }
  openQuestions.forEach((question, i) => {
    if (!isObject(question)) {
      errors.push(`open_questions[${i}]: must be an object`);
      return;
    }
    if (!isNonEmptyString(question.id)) {
      errors.push(`open_questions[${i}].id: must be a non-empty string`);
    }
    if (!isNonEmptyString(question.text)) {
      errors.push(`open_questions[${i}].text: must be a non-empty string`);
    }
    if (typeof question.blocking !== 'boolean') {
      errors.push(`open_questions[${i}].blocking: must be a boolean, got ${JSON.stringify(question.blocking)}`);
    }
  });
}

function validateReferences(references, errors) {
  if (references === undefined) {
    return;
  }
  if (!Array.isArray(references)) {
    errors.push('references: must be an array');
    return;
  }
  references.forEach((reference, i) => {
    if (!isObject(reference)) {
      errors.push(`references[${i}]: must be an object`);
      return;
    }
    if (!REFERENCE_TYPES.includes(reference.type)) {
      errors.push(`references[${i}].type: must be one of ${REFERENCE_TYPES.join('|')}, got ${JSON.stringify(reference.type)}`);
    }
    if (!isNonEmptyString(reference.url) && !isNonEmptyString(reference.path)) {
      errors.push(`references[${i}]: must carry a non-empty url or path`);
    }
  });
}

function validateChecklist(checklist, errors) {
  if (!isObject(checklist)) {
    errors.push('gate.checklist: must be an object');
    return;
  }
  for (const item of CHECKLIST_ITEMS) {
    if (typeof checklist[item] !== 'boolean') {
      errors.push(`gate.checklist.${item}: must be a boolean, got ${JSON.stringify(checklist[item])}`);
    }
  }
  // Catches a capitalised or misspelled key, which would otherwise only surface as the
  // lowercase key being absent.
  const unknown = Object.keys(checklist).filter((key) => !CHECKLIST_ITEMS.includes(key));
  if (unknown.length > 0) {
    errors.push(`gate.checklist: unknown keys ${unknown.join(', ')} (the six keys are lowercase: ${CHECKLIST_ITEMS.join(', ')})`);
  }
}

function validateCritic(critic, errors) {
  if (!isObject(critic)) {
    errors.push('gate.critic: must be an object');
    return;
  }
  if (!isNonEmptyString(critic.model)) {
    errors.push('gate.critic.model: must be a non-empty string');
  }
  if (!CRITIC_STATUSES.includes(critic.status)) {
    errors.push(`gate.critic.status: must be one of ${CRITIC_STATUSES.join('|')}, got ${JSON.stringify(critic.status)}`);
  }
}

function validateGate(gate, errors) {
  if (!isObject(gate)) {
    errors.push('gate: must be an object');
    return;
  }
  if (!OUTCOMES.includes(gate.outcome)) {
    errors.push(`gate.outcome: must be one of ${OUTCOMES.join('|')}, got ${JSON.stringify(gate.outcome)}`);
  }
  validateChecklist(gate.checklist, errors);

  if (gate.score !== null && !isScore(gate.score)) {
    errors.push(`gate.score: must be a number 0-100 or null, got ${JSON.stringify(gate.score)}`);
  }

  let scoreCount = null;
  if (!Array.isArray(gate.scores)) {
    errors.push('gate.scores: must be an array of numbers');
  } else {
    scoreCount = gate.scores.length;
    gate.scores.forEach((score, i) => {
      if (!isScore(score)) {
        errors.push(`gate.scores[${i}]: must be a number 0-100, got ${JSON.stringify(score)}`);
      }
    });
  }

  if (typeof gate.converged !== 'boolean') {
    errors.push(`gate.converged: must be a boolean, got ${JSON.stringify(gate.converged)}`);
  }

  if (!Number.isInteger(gate.rounds) || gate.rounds < 0 || gate.rounds > MAX_ROUNDS) {
    errors.push(`gate.rounds: must be an integer 0-${MAX_ROUNDS}, got ${JSON.stringify(gate.rounds)}`);
  } else if (scoreCount !== null && gate.rounds !== scoreCount) {
    errors.push(`gate.rounds: ${gate.rounds} does not equal gate.scores.length (${scoreCount})`);
  }

  if (!Array.isArray(gate.outstanding_feedback)) {
    errors.push('gate.outstanding_feedback: must be an array of strings');
  } else {
    gate.outstanding_feedback.forEach((entry, i) => {
      if (typeof entry !== 'string') {
        errors.push(`gate.outstanding_feedback[${i}]: must be a string, got ${JSON.stringify(entry)}`);
      }
    });
  }

  if (!isNonEmptyString(gate.author_model)) {
    errors.push('gate.author_model: must be a non-empty string');
  }

  validateCritic(gate.critic, errors);
}

/**
 * Checks a parsed metadata object against the schema_version 1.1 constraint set.
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

  validateRecommendation(metadata.recommendation, errors);
  validateOpenQuestions(metadata.open_questions, errors);

  if (metadata.affected_capabilities !== undefined) {
    if (!Array.isArray(metadata.affected_capabilities)) {
      errors.push('affected_capabilities: must be an array of strings');
    } else {
      metadata.affected_capabilities.forEach((capability, i) => {
        if (typeof capability !== 'string') {
          errors.push(`affected_capabilities[${i}]: must be a string, got ${JSON.stringify(capability)}`);
        }
      });
    }
  }

  if (!SCOPES.includes(metadata.estimated_scope)) {
    errors.push(`estimated_scope: must be one of ${SCOPES.join('|')}, got ${JSON.stringify(metadata.estimated_scope)}`);
  }

  validateReferences(metadata.references, errors);
  validateGate(metadata.gate, errors);

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
  CHECKLIST_ITEMS,
  CRITIC_STATUSES,
  REFERENCE_TYPES,
  OUTCOMES,
  MAX_ROUNDS,
};
