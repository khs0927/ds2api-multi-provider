# Profile account integration

DS2API supports an opaque profile-backed account record for deployments that use an external, user-driven login adapter.

## Core account shape

```json
{
  "name": "account-1",
  "auth_mode": "browser_session",
  "profile_id": "profile-1"
}
```

`profile_id` is an opaque identifier. DS2API uses it for account identity and pool selection. Existing email/mobile accounts continue to use the legacy login path.

## Resolver behavior

For `auth_mode=browser_session`, the managed-account resolver does not invoke the legacy email/mobile password login function. This lets an adapter own its own authentication lifecycle without placing provider-specific login logic in the DS2API core.

## Upstream adapter reference

The integration baseline is pinned in `integrations/sums001-deepseek-api.lock.json`.

The adapter runtime is intentionally separate from the core. It should provide user-driven sign-in, isolated state per profile, health/status reporting, and an OpenAI-compatible chat surface. DS2API should not expose adapter state through account-list responses.

## Operational boundary

Multiple profile records are for independently authorized accounts and ordinary availability/load distribution. The integration must not add ban evasion, IP rotation, CAPTCHA bypass, stealth fingerprinting, or mechanisms whose purpose is to defeat provider quotas or anti-abuse controls.

## Acceptance criteria

- Legacy email/mobile accounts keep working unchanged.
- Profile accounts have stable `profile:<profile_id>` identifiers.
- Profile accounts can be acquired and released by the existing account pool.
- Resolving a profile account does not invoke the legacy password login function.
- No provider-specific browser runtime is required by the DS2API core build.
- Go unit tests, lint/refactor gates, WebUI build, Windows tests, and macOS tests remain green.
