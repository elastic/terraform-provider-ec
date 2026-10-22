/**
 * Shared comment helpers for factory issue intake workflows.
 *
 * Fetches human-authored issue comments (paginated, bot-filtered, hard-capped) and serializes
 * them into a deterministic markdown string bounded by a context budget. Used by the
 * fetch-issue-comments runner so an agent sees the discussion without unbounded prompt growth.
 */

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

// Budget is measured in UTF-8 BYTES, not JS string length (UTF-16 code units): the serialized
// value is transported to the sanitize step through the HUMAN_COMMENTS environment variable, and a
// single env-var string above Linux's MAX_ARG_STRLEN (~128 KB) makes the next process launch fail
// with E2BIG. A CJK-heavy thread is ~3-4 bytes/char, so a code-unit budget could exceed that; a
// byte budget keeps it safe regardless of script.
const COMMENT_CONTEXT_BUDGET = 50_000;
/** Overhead reserved for truncation markers appended after the loop. */
const COMMENT_CONTEXT_MARKER_OVERHEAD = 200;

function utf8Len(str) {
  return Buffer.byteLength(str, 'utf8');
}

// Returns the longest prefix of `str` whose UTF-8 encoding is <= maxBytes, without splitting a
// multi-byte character.
function sliceToByteBudget(str, maxBytes) {
  if (maxBytes <= 0) return '';
  if (utf8Len(str) <= maxBytes) return str;
  const buf = Buffer.from(str, 'utf8').subarray(0, maxBytes);
  // Drop any trailing bytes that form an incomplete UTF-8 sequence.
  let end = buf.length;
  while (end > 0 && (buf[end - 1] & 0xc0) === 0x80) end--;
  if (end > 0 && (buf[end - 1] & 0x80) !== 0) end--;
  return buf.subarray(0, end).toString('utf8');
}

/**
 * Serializes captured issue comments into a deterministic markdown string for agent prompts,
 * bounded by a UTF-8 byte budget.
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

  for (const comment of comments) {
    const header = `**@${comment.author || ''}** (${comment.createdAt || ''}):\n\n`;
    const body = comment.body || '';
    const footer = '\n\n---\n';
    const available = bodyBudget - usedBytes;

    if (available <= 0) {
      break;
    }

    const frameBytes = utf8Len(header) + utf8Len(footer);
    if (frameBytes > available) {
      break;
    }
    const fullBlock = header + body + footer;
    const fullBlockBytes = utf8Len(fullBlock);
    if (fullBlockBytes <= available) {
      result += fullBlock;
      usedBytes += fullBlockBytes;
    } else {
      // Truncate this comment's body (by bytes, without splitting a character) to fit the budget.
      const truncatedBody = sliceToByteBudget(body, available - frameBytes);
      const block = header + truncatedBody + footer;
      result += block;
      usedBytes += utf8Len(block);
    }
    includedCount++;
  }

  const remaining = comments.length - includedCount;
  if (remaining > 0) {
    result += `[... ${remaining} more comments truncated for context budget]\n`;
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
  };
}
