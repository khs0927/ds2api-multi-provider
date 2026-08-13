# OpenCode + DS2API

This repository exposes DeepSeek Web through an OpenAI-compatible API. OpenCode can use that API as a custom provider.

## Supported route

The DS2API model `deepseek-v4-pro` resolves to the DeepSeek Web Expert route (`model_type: "expert"`). The compatibility alias `deepseek-v4-pro-0813` resolves to `deepseek-v4-pro`.

Important: `deepseek-v4-pro-0813` is a compatibility alias. It does not protocol-pin DeepSeek Web to a specific backend snapshot. The effective backend follows whatever DeepSeek Web currently serves for the Expert route.

## 1. Start DS2API

Start DS2API first and confirm its OpenAI-compatible endpoint is reachable.

Example local base URL:

```text
http://127.0.0.1:5001/v1
```

For a remote deployment, use the HTTPS URL ending in `/v1`.

Never commit DeepSeek credentials, cookies, session tokens, or DS2API client API keys to Git.

## 2. Configure OpenCode

Copy the template into the root of the project where OpenCode will run:

```bash
cp integrations/opencode/opencode.ds2api.example.json opencode.json
```

Set these environment variables in the shell that launches OpenCode:

```bash
export DS2API_OPENAI_BASE_URL="http://127.0.0.1:5001/v1"
export DS2API_API_KEY="YOUR_DS2API_CLIENT_KEY"
```

`DS2API_OPENAI_BASE_URL` must include `/v1`.

The template uses OpenCode's current custom OpenAI-compatible provider contract:

- provider id: `ds2api`
- package: `@ai-sdk/openai-compatible`
- default model: `ds2api/deepseek-v4-pro-0813`
- small model: `ds2api/deepseek-v4-flash`

## 3. Confirm models

```bash
opencode models ds2api
```

Expected configured model IDs include:

```text
ds2api/deepseek-v4-pro-0813
ds2api/deepseek-v4-pro
ds2api/deepseek-v4-pro-0813-search
ds2api/deepseek-v4-flash
```

## 4. Run a direct smoke test

```bash
opencode run -m ds2api/deepseek-v4-pro-0813 \
  "Reply exactly with: OPENCODE_DS2API_PRO_OK"
```

Or run the included script:

```bash
bash scripts/opencode-ds2api-smoke.sh
```

## 5. TUI usage

Start OpenCode:

```bash
opencode
```

Then use `/models` and select:

```text
ds2api/deepseek-v4-pro-0813
```

## Troubleshooting

### Provider/model does not appear

Check:

```bash
opencode models ds2api
```

and verify `opencode.json` is in the current project root or a parent Git project root.

### Authentication failure

Check that `DS2API_API_KEY` is the DS2API client API key. Do not use or expose the DeepSeek account password as the OpenCode API key.

### 401 / 403

Check DS2API client authentication and the upstream DeepSeek login/session state. Do not print secrets into logs.

### 429

This can be caused by DeepSeek Web account/rate limits. It is not automatically an OpenCode configuration bug.

### Tool calling problems

Keep the provider package as `@ai-sdk/openai-compatible`. DS2API exposes `/v1/chat/completions` and supports the OpenAI-compatible tool-call path. Do not switch to an unverified provider package unless the DS2API endpoint contract changes.

### 0813 verification

The alias only guarantees this routing inside DS2API:

```text
deepseek-v4-pro-0813
  -> deepseek-v4-pro
  -> model_type: expert
  -> DeepSeek Web Expert backend
```

It does not independently prove which server-side Expert snapshot DeepSeek Web is serving.
