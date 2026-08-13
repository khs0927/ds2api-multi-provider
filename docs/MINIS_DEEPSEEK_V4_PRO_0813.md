# Minis + DeepSeek V4 Pro 0813 compatibility

This repository already supports the DeepSeek Web **Pro/Expert** route.

- `deepseek-v4-pro` resolves to DeepSeek Web `model_type: "expert"`.
- `deepseek-v4-pro-search` resolves to `model_type: "expert"` with search enabled.
- The sample configuration adds client-facing aliases:
  - `deepseek-v4-pro-0813` -> `deepseek-v4-pro`
  - `deepseek-v4-pro-0813-search` -> `deepseek-v4-pro-search`

## Important: alias vs. backend pinning

`ds2api` talks to DeepSeek Web using the Web routing type (`expert`); it does not send a public API snapshot ID such as `0813` to DeepSeek Web. Therefore the `*-0813` names in this repository are **compatibility aliases for clients such as Minis**, not a cryptographic or protocol-level guarantee that DeepSeek Web is pinned forever to a particular backend snapshot.

If DeepSeek changes the Web Expert backend, the alias will continue to follow whatever backend DeepSeek serves for the Expert route.

## Minis configuration

Use an OpenAI-compatible connection:

- Base URL: `https://YOUR_DS2API_HOST/v1`
- API key: a DS2API client API key (do not use an admin secret or DeepSeek account password)
- Model: `deepseek-v4-pro-0813`
- Search model: `deepseek-v4-pro-0813-search`

If Minis only accepts models returned by `/v1/models`, use the canonical model name `deepseek-v4-pro`. Configured aliases are accepted by request resolution but are not necessarily advertised in the canonical model list.

## Required model aliases

Copy these entries into the active `config.json` if it predates this change:

```json
{
  "model_aliases": {
    "deepseek-v4-pro-0813": "deepseek-v4-pro",
    "deepseek-v4-pro-0813-search": "deepseek-v4-pro-search"
  }
}
```

Merge them with any existing `model_aliases` instead of replacing unrelated aliases.

## Smoke test

```bash
export DS2API_BASE_URL="https://YOUR_DS2API_HOST"
export DS2API_API_KEY="YOUR_DS2API_CLIENT_KEY"

curl -sS "$DS2API_BASE_URL/v1/chat/completions" \
  -H "Authorization: Bearer $DS2API_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-v4-pro-0813",
    "messages": [{"role":"user","content":"Reply with exactly: PRO_OK"}],
    "stream": false
  }'
```

A successful response proves that the client alias resolves and that the request reaches the Pro/Expert route. It does **not** independently prove the remote DeepSeek Web backend build number; that requires DeepSeek to expose verifiable backend-version metadata.

## Operational notes

- Keep DeepSeek account credentials and DS2API admin secrets out of mobile clients.
- Expose only the OpenAI-compatible API through HTTPS when Minis connects over the Internet.
- Expect DeepSeek Web protocol, rate limits, or authentication behavior to change independently of this repository.
