## Why

Sandwitches routes its API through existing middleware and supports recipes, tags, orders, carts, and ratings. Read access should remain easy for public integrations while writes require a user and API token consistently.

## What Changes

- Keep public GET reads such as ping, settings, recipes, recipe-of-the-day, tags, users, and ingredient scaling public where currently public.
- Require an authenticated user plus bearer API token for recipe/tag/order/cart/rating and every create/update/delete operation.
- Add owner-scoped one-time-visible token management while retaining cookie, admin, and CSRF protections.

## Capabilities

### New Capabilities

- `api-authentication`

### Modified Capabilities

- `public-api-reads`

## Impact

Sandwitches route middleware, token storage/migration, token UI/API, API tests, and documentation are affected. Existing `/api/v1` routing and browser auth remain compatible.
