/**
 * Shared comment helpers for factory issue intake workflows.
 *
 * Fetching lives in `factory-runners/fetch-issue-comments.js`, which keeps bot comments:
 * the research-factory sticky comment a change-factory run consumes is bot-authored.
 */
'use strict';

const COMMENT_CONTEXT_BUDGET = 50_000;
/** Overhead reserved for truncation markers appended after the loop. */
const COMMENT_CONTEXT_MARKER_OVERHEAD = 200;

/**
 * Serializes captured issue comments into a deterministic markdown string for agent prompts.
 *
 * @param {{ comments: Array<{author: string, createdAt: string, body: string}>, truncated: boolean }} params
 * @returns {string}
 */
function serializeIssueComments({ comments, truncated }) {
  if (!Array.isArray(comments) || comments.length === 0) {
    return '';
  }

  const bodyBudget = COMMENT_CONTEXT_BUDGET - COMMENT_CONTEXT_MARKER_OVERHEAD;
  let result = '';
  let includedCount = 0;
  let partialTruncation = false;

  for (const comment of comments) {
    const header = `**@${comment.author || ''}** (${comment.createdAt || ''}):\n\n`;
    const body = comment.body || '';
    const footer = '\n\n---\n';
    const available = bodyBudget - result.length;

    if (available <= 0) {
      break;
    }

    const frameLength = header.length + footer.length;
    if (frameLength > available) {
      break;
    }
    const fullBlock = header + body + footer;
    if (fullBlock.length <= available) {
      result += fullBlock;
    } else {
      // Truncate this comment's body so the output stays within budget. Record it: the comment
      // still counts as included, so without this flag an oversized last/only comment would lose
      // its tail (e.g. trailing requirements) and the agent would see it as complete.
      const truncatedBody = body.slice(0, available - frameLength);
      result += header + truncatedBody + footer;
      partialTruncation = true;
    }
    includedCount++;
  }

  const remaining = comments.length - includedCount;
  if (remaining > 0) {
    result += `[... ${remaining} more comments truncated for context budget]\n`;
  }

  if (partialTruncation) {
    result += '[... last included comment truncated for context budget]\n';
  }

  if (truncated) {
    result += '[... comment history truncated at 200 comments]\n';
  }

  return result;
}

if (typeof module !== 'undefined') {
  module.exports = {
    serializeIssueComments,
    COMMENT_CONTEXT_BUDGET,
    COMMENT_CONTEXT_MARKER_OVERHEAD,
  };
}
