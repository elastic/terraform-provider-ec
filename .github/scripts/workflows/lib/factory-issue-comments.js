/**
 * Shared comment helpers for research-factory issue intake workflows.
 */

const { byteLength, sliceToBytes, dropOrphanedCommentOpener } = require('./env-payload.js');

/**
 * Fetches human-authored comments for an issue, paginated, with bot filtering and a hard cap.
 *
 * @param {{ github: object, owner: string, repo: string, issueNumber: number }} params
 * @returns {Promise<{ comments: Array<{author: string, createdAt: string, body: string}>, truncated: boolean }>}
 */
async function factoryFetchIssueComments({ github, owner, repo, issueNumber }) {
  const MAX_COMMENTS = 200;
  const allComments = await github.paginate(github.rest.issues.listComments, {
    owner,
    repo,
    issue_number: issueNumber,
    per_page: 100,
  });

  const humanComments = [];
  let truncated = false;
  for (const comment of allComments) {
    if (comment.user?.login?.endsWith('[bot]')) {
      continue;
    }
    if (humanComments.length >= MAX_COMMENTS) {
      truncated = true;
      break;
    }
    humanComments.push({
      author: comment.user?.login ?? '',
      createdAt: comment.created_at ?? '',
      body: comment.body ?? '',
    });
  }

  return {
    comments: humanComments,
    truncated,
  };
}

/**
 * Budget in **UTF-8 bytes**, not characters. The serialized history is handed between
 * pre-activation steps as an environment variable, and Linux caps a single entry at 128 KiB — a
 * 50 000-character budget is 150 000 bytes of CJK, which kills the exporting step with `E2BIG`
 * after the trigger label has already been consumed. See `lib/env-payload.js`.
 */
const COMMENT_CONTEXT_BUDGET = 50_000;
/** Overhead reserved for truncation markers appended after the loop. */
const COMMENT_CONTEXT_MARKER_OVERHEAD = 200;
/**
 * Appended inside a comment whose body did not fit. Without it a sliced comment is indistinguishable
 * from a complete one — the agent and the critic would read a half-stated requirement as the whole
 * requirement — and the trailing "N more comments truncated" marker does not fire when the sliced
 * comment is the last one.
 */
const COMMENT_BODY_TRUNCATED_NOTICE = '\n\n[... this comment was truncated for context budget]';
/**
 * Appended when a comment body is cut at an unterminated `<!--`. The whole serialized history is
 * passed through `sanitizeUserContent` downstream, whose `stripHtmlComments` deletes from an
 * orphaned opener to the end of the *string* — so one comment ending mid-comment silently erases
 * every comment after it, maintainer feedback included. Trimming the one body that carries the
 * opener costs that comment's tail; not trimming costs the rest of the history.
 */
const COMMENT_BODY_OPENER_NOTICE = '\n\n[... this comment was cut at an unterminated HTML comment]';

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
  let usedBytes = 0;
  let includedCount = 0;
  let bodyTruncated = false;

  for (const comment of comments) {
    const header = `**@${comment.author || ''}** (${comment.createdAt || ''}):\n\n`;
    const raw = comment.body || '';
    // Applied to every body, not only to the ones the budget slices: an orphaned opener erases the
    // rest of the history whether or not this comment was truncated.
    const trimmed = dropOrphanedCommentOpener(raw);
    const body = trimmed === raw ? raw : `${trimmed}${COMMENT_BODY_OPENER_NOTICE}`;
    const footer = '\n\n---\n';
    const available = bodyBudget - usedBytes;

    if (available <= 0) {
      break;
    }

    const frameBytes = byteLength(header) + byteLength(footer);
    if (frameBytes > available) {
      break;
    }
    const fullBlock = header + body + footer;
    if (byteLength(fullBlock) <= available) {
      result += fullBlock;
      usedBytes += byteLength(fullBlock);
    } else {
      // Truncate this comment's body so the output stays within budget, and say so inside the
      // block: a silently sliced comment reads as a complete one. If the frame plus the notice
      // alone would not fit, stop instead of emitting a block that overshoots `available` — the
      // trailing marker below still reports the comment as dropped.
      const room = available - frameBytes - byteLength(COMMENT_BODY_TRUNCATED_NOTICE);
      if (room < 0) {
        break;
      }
      // Same hazard as the env payloads: a cut landing inside an HTML comment orphans the opener,
      // and `sanitizeUserContent` downstream then deletes from there to the end of the string —
      // taking this notice and the trailing drop marker with it. Issue templates leave such
      // comments in bodies routinely.
      const truncatedBody = dropOrphanedCommentOpener(sliceToBytes(body, room));
      const block = header + truncatedBody + COMMENT_BODY_TRUNCATED_NOTICE + footer;
      result += block;
      usedBytes += byteLength(block);
      bodyTruncated = true;
    }
    includedCount++;
  }

  const remaining = comments.length - includedCount;
  if (remaining > 0) {
    result += `[... ${remaining} more comments truncated for context budget]\n`;
  } else if (bodyTruncated) {
    // "a comment", not "the last comment": the sliced one is not necessarily the final entry — a
    // short comment after it can still fit whole.
    result += '[... a comment above was truncated for context budget]\n';
  }

  if (truncated) {
    result += '[... comment history truncated at 200 comments]\n';
  }

  return result;
}

if (typeof module !== 'undefined') {
  module.exports = {
    factoryFetchIssueComments,
    serializeIssueComments,
    COMMENT_CONTEXT_BUDGET,
    COMMENT_CONTEXT_MARKER_OVERHEAD,
    COMMENT_BODY_TRUNCATED_NOTICE,
    COMMENT_BODY_OPENER_NOTICE,
  };
}
