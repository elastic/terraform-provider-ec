---
model: "anthropic/claude-sonnet-5"
engine:
  id: claude
  args:
    - "--effort"
    - "high"
  env:
    ANTHROPIC_BASE_URL: "https://openrouter.ai/api"
    ANTHROPIC_API_KEY: ${{ secrets.OPENROUTER_API_KEY }}
---
