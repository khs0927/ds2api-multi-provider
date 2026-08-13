#!/usr/bin/env bash
set -euo pipefail

: "${DS2API_OPENAI_BASE_URL:?Set DS2API_OPENAI_BASE_URL, including /v1}"
: "${DS2API_API_KEY:?Set DS2API_API_KEY}"

MODEL="${OPENCODE_DS2API_MODEL:-ds2api/deepseek-v4-pro-0813}"
EXPECTED="OPENCODE_DS2API_PRO_OK"

if ! command -v opencode >/dev/null 2>&1; then
  echo "ERROR: opencode command not found" >&2
  exit 127
fi

if [[ ! -f opencode.json ]]; then
  echo "ERROR: opencode.json not found in current directory" >&2
  echo "Copy integrations/opencode/opencode.ds2api.example.json to opencode.json first." >&2
  exit 2
fi

echo "OpenCode: $(opencode --version 2>/dev/null || true)"
echo "Model: ${MODEL}"
echo "Checking configured DS2API models..."
opencode models ds2api

echo "Running Pro smoke test..."
OUTPUT="$(opencode run -m "${MODEL}" "Reply exactly with: ${EXPECTED}")"
printf '%s\n' "${OUTPUT}"

if [[ "${OUTPUT}" != *"${EXPECTED}"* ]]; then
  echo "ERROR: expected marker not found: ${EXPECTED}" >&2
  exit 1
fi

echo "SUCCESS: OpenCode -> DS2API -> Pro route responded with the expected marker."
