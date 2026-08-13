#!/usr/bin/env sh
set -eu

: "${DS2API_BASE_URL:?Set DS2API_BASE_URL, for example https://example.com}"
: "${DS2API_API_KEY:?Set DS2API_API_KEY to a DS2API client key}"

MODEL="${DS2API_MODEL:-deepseek-v4-pro-0813}"
BASE_URL="${DS2API_BASE_URL%/}"

curl -fsS "$BASE_URL/v1/chat/completions" \
  -H "Authorization: Bearer $DS2API_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"model\":\"$MODEL\",\"messages\":[{\"role\":\"user\",\"content\":\"Reply with exactly: PRO_OK\"}],\"stream\":false}"

printf '\n'
