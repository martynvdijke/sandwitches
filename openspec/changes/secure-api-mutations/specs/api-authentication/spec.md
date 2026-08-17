## ADDED Requirements

### Requirement: Public recipe reads

The service SHALL permit unauthenticated GET access to documented public `/api/v1` reads, including ping, settings, users, tags, recipes, recipe-of-the-day, and scaling endpoints.

#### Scenario: Public recipe read

- WHEN a client requests a public recipe GET endpoint without credentials
- THEN Sandwitches returns the read payload

### Requirement: Mutations require user and token

Recipe, tag, order, cart, rating, and every create/update/delete API operation MUST require an authenticated user and a valid bearer API token, while retaining existing cookie, admin, and CSRF checks.

#### Scenario: Anonymous order write

- WHEN an anonymous client posts an order or cart mutation
- THEN the service rejects it and changes no order state

### Requirement: Safe token lifecycle

Users SHALL create, list metadata for, revoke, and rotate owned tokens. Secrets MUST be shown once, stored only as hashes, and never logged or returned after creation.

#### Scenario: Revoked token write

- WHEN a revoked token is used on a recipe mutation
- THEN Sandwitches rejects the request without changing the recipe
