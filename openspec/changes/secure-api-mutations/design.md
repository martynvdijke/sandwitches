## Context

Sandwitches already routes `/api/v1` through authentication middleware that recognizes API traffic and uses cookies for browser auth. The change adds a bearer token as a second credential for writes while preserving public recipe reads.

## Goals / Non-Goals

### Goals

- Preserve public GET recipe/integration reads.
- Protect recipe, tag, order, cart, rating, and future mutations with user plus token.
- Preserve admin, cookie, and CSRF controls.

### Non-Goals

- Making public reads private.
- Removing browser authentication or admin authorization.
- Committing a real API token.

## Decisions

- Add a user-owned token store with hash-only secret storage and lifecycle metadata.
- Compose bearer validation with existing cookie/session middleware; require identity agreement before writes.
- Explicitly allowlist public GET routes and protect non-GET routes by default.
- Provide owner-only token lifecycle operations with one-time secret display.

## Risks / Trade-offs

Automation must gain user sessions and tokens. A public-read allowlist requires maintenance as new routes are added, but avoids accidental exposure of writes.

## Migration Plan

1. Audit `/api/v1` route methods and add token persistence.
2. Implement lifecycle endpoints and middleware.
3. Protect mutations and add authorization/CSRF tests.
4. Provision tokens to API clients and document the new contract.

## Open Questions

- Should token scopes distinguish recipe management from order management?
