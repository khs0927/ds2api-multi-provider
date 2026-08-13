# OCI Always Free runtime plan for profile adapters

This document records the preferred zero-cost runtime for the optional user-driven browser-login adapter referenced by `docs/profile-account-integration.md`.

## Decision

Primary candidate: Oracle Cloud Infrastructure (OCI) Ampere A1 Always Free in the tenancy home region.

Use the current conservative Always Free tenancy limit from Oracle's OCI documentation:

- Shape: `VM.Standard.A1.Flex`
- Total allocation: up to 2 OCPUs and 12 GB RAM across Always Free A1 instances
- Block storage: up to 200 GB total Always Free block volume storage
- Region: South Korea Central (Seoul) is eligible; South Korea North (Chuncheon) is excluded for Always Free A1
- Persistent state must live on OCI block storage, not temporary container layers

Oracle may temporarily have no A1 host capacity. Provisioning therefore needs a capacity fallback procedure rather than assuming an instance can always be recreated immediately.

## Why this runtime

The profile adapter needs more than a stateless web function:

1. Chromium/Playwright must run on Linux.
2. Browser profile directories need persistent storage.
3. A user-driven sign-in flow can take minutes and cannot be treated as a short serverless invocation.
4. The adapter should remain separate from the DS2API core so the Go API does not own provider-specific browser state.

Playwright publishes Linux ARM64 browser support and ARM64 Docker images, so Ampere A1 is compatible with the browser runtime architecture.

## Recommended topology

```text
Internet / mobile client
        |
        v
   reverse proxy
        |
        +---------------------+
        |                     |
        v                     v
     DS2API             profile adapter
     Go API              Playwright/Python
        |                     |
        | opaque profile_id   | user-driven sign-in
        +-------------------->|
                              |
                              v
                        persistent volume
                        /data/profiles/<id>
```

The core account record remains provider-agnostic:

```json
{
  "name": "account-1",
  "auth_mode": "browser_session",
  "profile_id": "profile-1"
}
```

Do not place Google passwords, DeepSeek browser state, or browser-profile files in Git, DS2API config exports, or application logs.

## Upstream baseline

The optional browser adapter reference is pinned by:

`integrations/sums001-deepseek-api.lock.json`

The pinned upstream already provides:

- user-driven Playwright sign-in
- persistent browser-profile reuse
- a refresh path from a saved profile
- a DeepSeek client
- OpenAI-compatible chat handling
- streaming support

Integration work should adapt the minimum necessary profile selection layer instead of reimplementing authentication or DeepSeek's web protocol.

## OCI provisioning checklist

- [ ] Create or use one OCI account only; OCI Free Tier permits one free account per person.
- [ ] Select **South Korea Central (Seoul)** as the home region if Korean hosting is required.
- [ ] Create `VM.Standard.A1.Flex` within the current Always Free allocation.
- [ ] Start conservatively at 1 OCPU / 6 GB; increase only if browser memory pressure requires it and remain within the free limit.
- [ ] Use a sufficiently large boot/block volume while keeping total Always Free block storage within the OCI limit.
- [ ] Install Docker and Compose (or Podman) on the VM.
- [ ] Use an ARM64-compatible Playwright runtime.
- [ ] Mount profile state under a persistent `/data` path.
- [ ] Expose only the required public entrypoint; keep internal adapter/control ports private.
- [ ] Create an OCI budget/alert at the minimum practical threshold even when only Always Free resources are intended.
- [ ] Back up non-secret configuration separately from browser-profile state.

## Idle-instance caveat

OCI documents an idle-resource reclamation policy for Always Free compute. An A1 VM can be considered idle when CPU, network, and memory utilization all remain below Oracle's thresholds over a seven-day period.

Do **not** add artificial traffic or workloads solely to defeat reclamation. Treat reclamation as a platform limitation: keep reproducible deployment instructions and persistent-volume recovery procedures, and use real application activity only.

## Capacity caveat

`Out of host capacity` is an expected OCI Free Tier failure mode. If it occurs:

1. Do not delete a working instance just to resize or rebuild it.
2. Preserve the boot/block volume.
3. Try another availability domain when the home region exposes one.
4. Otherwise wait for capacity and retry later.
5. Do not move Always Free compute to a non-home region assuming it will remain free.

## Adapter acceptance criteria

The optional adapter is ready only when all of the following pass:

- [ ] `profile-1` and `profile-2` use isolated browser-profile directories.
- [ ] A user can explicitly start sign-in for one profile without entering a Google password into DS2API.
- [ ] The adapter reports only coarse states such as `pending`, `connected`, `reconnect_required`, and `error`.
- [ ] No browser credential/session material appears in adapter API responses or logs.
- [ ] Restarting the adapter reuses the persistent profile.
- [ ] One profile failure does not invalidate another profile.
- [ ] DS2API's existing account pool can acquire/release the opaque profile account normally.
- [ ] Existing email/mobile DS2API accounts remain unchanged.
- [ ] Normal load distribution is used only for independently authorized accounts; no quota, ban, CAPTCHA, fingerprint, or anti-abuse circumvention is added.

## References

- Oracle OCI Free Tier: https://docs.oracle.com/iaas/Content/FreeTier/freetier.htm
- Oracle Always Free resources: https://docs.oracle.com/en-us/iaas/Content/FreeTier/freetier_topic-Always_Free_Resources.htm
- Oracle public cloud regions: https://www.oracle.com/cloud/public-cloud-regions/
- Playwright release notes: https://playwright.dev/docs/release-notes
- Microsoft Playwright container registry: https://mcr.microsoft.com/artifact/mar/playwright
