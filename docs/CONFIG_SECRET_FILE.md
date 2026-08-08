# Docker Secret configuration

This fork supports `DS2API_CONFIG_JSON_FILE` in addition to the upstream
`DS2API_CONFIG_JSON` variable. The file may contain the normal DS2API JSON
configuration, including managed DeepSeek email/mobile accounts and passwords.

The fork also supports `DS2API_ADMIN_KEY_FILE` and `DS2API_JWT_SECRET_FILE`.
Direct environment variables take precedence for backwards compatibility; when
they are blank, the service reads the corresponding root-only secret file.

The service reads the file at startup, clears persisted runtime tokens from the
in-memory bootstrap config, and treats the configuration as env-backed. Set
`DS2API_ENV_WRITEBACK=0` for a read-only Docker Secret. The service does not
print the file contents or copy the secret into Git.

Example container wiring:

```yaml
services:
  ds2api:
    image: ghcr.io/OWNER/ds2api-multi-provider:PINNED_DIGEST
    environment:
      DS2API_CONFIG_JSON_FILE: /run/secrets/ds2api_config
      DS2API_ENV_WRITEBACK: "0"
      DS2API_ADMIN_KEY_FILE: /run/secrets/ds2api_admin_key
      DS2API_JWT_SECRET_FILE: /run/secrets/ds2api_jwt_secret
    secrets:
      - ds2api_config
      - ds2api_admin_key
      - ds2api_jwt_secret

secrets:
  ds2api_config:
    file: /root/minis-secrets/ds2api-config.json
  ds2api_admin_key:
    file: /root/minis-secrets/ds2api-admin-key
  ds2api_jwt_secret:
    file: /root/minis-secrets/ds2api-jwt-secret
```

Keep the host file outside Git with mode `0600`. Do not use this file as an
admin export endpoint, and do not enable writeback against the mounted secret.
The DS2API account/session manager still owns login and token refresh; this
feature only changes how its initial configuration is injected.

When the gateway uses `DS2API_API_KEY_FILE`, its value must match one of the
managed keys in the configuration JSON's `keys` array. Otherwise DS2API treats
the bearer value as a direct upstream token instead of selecting a managed
account.
