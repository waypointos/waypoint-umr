# Security

This module holds the router owner's local portal credentials and talks to
the router over its self-signed-TLS LAN portal. Take anything touching that
credential path or release signing seriously.

## Reporting a vulnerability

Report vulnerabilities privately via GitHub:
https://github.com/waypointos/waypoint-umr/security/advisories/new

Do not open a public issue for a security problem. You should hear back
within a week.

## Verifying releases

Every release ships a `.raw` module image signed with cosign keyless
signing from this repository's release workflow. Verify with:

```
cosign verify-blob \
  --bundle umr-<version>.raw.cosign \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp '^https://github.com/waypointos/waypoint-umr/\.github/workflows/release\.yml@refs/tags/v' \
  umr-<version>.raw
```
