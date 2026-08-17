## 1. Route audit and storage

- [ ] 1.1 Inventory all `/api/v1` GET reads and recipe/tag/order/cart/rating mutations.
- [ ] 1.2 Add reversible hashed-token migration and indexes.

## 2. Enforcement

- [ ] 2.1 Implement bearer validation, owner matching, expiry, and revocation.
- [ ] 2.2 Allowlist public reads and protect all create/update/delete routes with existing auth/CSRF.
- [ ] 2.3 Implement owner-only create/list/revoke/rotate token operations.

## 3. Verification

- [ ] 3.1 Test public reads, anonymous/token-only writes, malformed/expired/revoked tokens, and ownership.
- [ ] 3.2 Test admin/cookie/CSRF behavior and secret non-disclosure.
- [ ] 3.3 Document client migration and run the full Sandwitches suite.
