# Legal pages, GDPR operations and EU launch checklist

This document is for the operator of a Sandwitches instance. It maps the public
legal pages to code, lists the deployment facts they publish, and records the
deletion, export and retention behaviour the operator must be able to explain
to users and supervisors. It is an engineering/operations reference, not legal
advice; have the operator's counsel review the published text and the records
below before an EU launch.

## Public legal pages

| Route | Handler | Template | Notes |
| --- | --- | --- | --- |
| `GET /privacy` | `handlers.PrivacyPage` | `templates/privacy.html` | Public, no session required |
| `GET /terms` | `handlers.TermsPage` | `templates/terms.html` | Public, no session required |
| `GET /account/delete` | `handlers.AccountDeletePage` | `templates/account_delete.html` | Public; `?deleted=1` shows the success notice |
| `POST /settings/account/delete` | `handlers.DeleteAccount` | redirect only | Signed-in web form, requires CSRF and the typed username |
| `DELETE /api/v1/me` | `api.deleteMe` | JSON | Signed-in API, body `{"confirm":"<username>"}` |
| `GET /api/v1/me/export` | `api.exportMe` | JSON attachment | Signed-in API; also linked from `/settings` |

`main.go` registers the three GET routes before the authenticated/CSRF group
with the explicit comment that store reviewers and users who cannot sign in
must reach them anonymously. `legalTemplateData` in
`internal/handlers/legal.go` avoids the session store so reading a policy does
not create a cookie.

The footer and settings page link all three pages. The Android app opens the
same routes from Settings.

## Dynamic `.legal` facts

`legalData()` runs once per request and exposes a `legal` object to the
templates. Values come from the loaded config (`SetLegalConfig`), then the
environment, then neutral defaults, so a misconfigured instance publishes
honest fallback wording rather than claiming a processor it does not run.

| Template field | Env variable | Fallback / behaviour |
| --- | --- | --- |
| `.legal.Controller` | `LEGAL_CONTROLLER_NAME` | `the operator of this <SiteName> instance` |
| `.legal.ContactEmail` | `LEGAL_CONTACT_EMAIL` | The admin-settings contact email; if neither is set, the page directs users to self-service deletion or the administrator |
| `.legal.ContactAddress` | `LEGAL_CONTACT_ADDRESS` | Empty; postal address is omitted |
| `.legal.DPOEmail` | `LEGAL_DPO_EMAIL` | Empty; the DPO line is omitted |
| `.legal.SupervisoryAuthority` | `LEGAL_SUPERVISORY_AUTHORITY` | `your local data protection authority` |
| `.legal.HostingRegion` | `LEGAL_HOSTING_REGION` | `the European Union` |
| `.legal.BaseURL` | `LEGAL_BASE_URL` | `https://sandwitches.vandijke.xyz`; trailing slash trimmed |
| `.legal.SessionCookie` | — | `middleware.SessionCookieName` = `sandwitches_session` |
| `.legal.Analytics` / `.legal.AnalyticsURL` | `UMAMI_HOST` + `UMAMI_WEBSITE_ID` | Shown only when both are set |
| `.legal.Traces` | `OTEL_EXPORTER_OTLP_*` / admin OTel settings | Shown only when telemetry is configured |
| `.legal.Mail` | `SEND_EMAIL=true` or `SMTP_HOST` / admin SMTP | Shown only when email delivery is enabled |
| `.legal.Gotify` | `GOTIFY_URL` / admin Gotify | Shown only when push is configured |
| `.legal.AI` | Admin AI connection point | Shown only when an AI endpoint is configured |
| `.legal.OIDC` | OIDC configuration | Shown only when OIDC is valid/configured |

`LEGAL_BASE_URL` is also rendered as the canonical public address of the
deletion page. Set it to the https origin users and reviewers actually reach;
otherwise the page advertises the wrong canonical URL.

The policy is written so that the processor list is conditional. Enabling
SMTP, Gotify, Umami, OTLP, an AI endpoint or OIDC changes the published page on
the next request. If the operator adds a processor the template does not know
about, update `templates/privacy.html` and this document together.

## GDPR operating checklist

The code supplies self-service primitives (notice, export, erasure, opt-out).
The operator remains responsible for the organisational records below.

### 1. Controller identity and contact

- Set `LEGAL_CONTROLLER_NAME` to the legal person/entity, not a nickname.
- Set `LEGAL_CONTACT_EMAIL` to a monitored privacy mailbox (or ensure the admin
  settings email is one). Without it, the policy tells users to use the account
  pages or the administrator, which is weaker for Article 13/14 notice.
- Set `LEGAL_CONTACT_ADDRESS` if the operator has a postal address to publish.
- Set `LEGAL_DPO_EMAIL` if a DPO is designated; the policy then prints the
  contact. If no DPO is required, leaving it empty is intentional.
- Set `LEGAL_SUPERVISORY_AUTHORITY` to the competent authority (for the
  Netherlands: Autoriteit Persoonsgegevens).
- Set `LEGAL_HOSTING_REGION` to the real region of the server **and backups**.

### 2. Processor agreements

- Sign a data processing agreement (Article 28) with the hosting provider and
  with any backup/storage provider.
- Keep the SMTP provider, Gotify host, Umami host, OTLP collector and AI
  provider in the same inventory. The policy discloses each only when enabled.
- If a provider processes outside the EEA, record the adequacy decision or the
  standard contractual clauses relied on. The policy states transfers rely on
  one of those two mechanisms.

### 3. Record of processing (Article 30)

Keep a record covering at least:

| Processing | Data | Purpose | Legal basis | Retention |
| --- | --- | --- | --- | --- |
| Accounts and profiles | username, email, name, bio, avatar, language/theme, password hash, OIDC subject, timestamps | Provide the account and sign-in | Contract | Life of account, then deleted |
| Recipes and contributions | recipe text, images, tags, ratings, comments, edit history | Community recipe service | Contract / legitimate interest | Life of content; author link removed on deletion |
| Orders and cart | order items, quantities, prices, status, tracking token | Fulfil and track orders; transaction record | Contract / legal obligation | Retained after account deletion, detached from the user |
| Security | session cookie, IP for rate limiting, request logs with path/status/timing/request ID, reset-token and mobile-code hashes | Secure operation and abuse prevention | Legitimate interest | Log rotation window; tokens 24 h / 2 min |
| Optional analytics | page views in self-hosted Umami | Aggregate usage | Legitimate interest with opt-out | Umami retention configured by operator |
| Optional notifications | email address and message content to SMTP; order/recipe text to Gotify | Deliver notifications | Contract / legitimate interest | Provider retention |
| Optional AI feature | recipe text submitted to the configured endpoint | Provide the requested feature | Contract | Provider retention |
| OIDC sign-in | subject identifier, name/email/groups returned by provider | Single sign-on | Contract / legitimate interest | Life of account |

### 4. Retention schedule

The published policy commits to:

- Account/profile: while the account exists, then deleted.
- Password-reset tokens: 24 hours or until used. Mobile sign-in codes: two
  minutes or until used.
- Session cookie: up to 30 days, or until sign-out/password change/deletion.
- Recipes: while the recipe exists; on account deletion they stay published
  without an author unless an administrator removes them.
- Ratings, comments, favourites, cart, group memberships, reset/mobile tokens:
  deleted immediately on account deletion.
- Orders: retained as the transaction record with the account link cleared;
  accounting retention may require it.
- Logs and backups: kept for the operator's normal rotation window, then
  overwritten; backups are not used to restore an individual deleted account.

Document the actual number of days for logs, Umami, backups and accounting
records. The templates say "the operator's normal window"; a concrete internal
schedule is what an authority will ask for.

### 5. Backups

The application does **not** run a scheduled backup daemon. The only automatic
backup in the tree is the one-off verified SQLite copy made before legacy-table
cleanup (`internal/database/cleanup_legacy.go`). The operator must provision
backups independently and:

- Encrypt backups at rest and restrict access to the operator.
- Record the rotation window and make it match the policy's "normal log and
  backup rotation window" wording.
- Never restore an individual deleted account from backup as a way to undo an
  erasure request; the policy promises backups are not used that way.
- Include uploaded `MEDIA_ROOT` files in the retention decision. Deleting a
  user row does not delete avatar/recipe image files; the deletion page warns
  that files may remain and can be removed on request.

### 6. Breach process (72 hours)

Maintain a runbook:

1. Detect and contain; preserve logs (request IDs, timestamps) before rotating
   anything.
2. Assess risk to rights and freedoms; record the facts, affected data
   categories and approximate number of users.
3. Notify the supervisory authority without undue delay and within 72 hours of
   becoming aware, unless the breach is unlikely to result in a risk
   (Article 33). Record any decision not to notify.
4. Notify affected users without undue delay when there is a high risk
   (Article 34), including what happened and what they should do.
5. Rotate `SECRET_KEY` if session integrity is in doubt (this invalidates all
   sessions), force password resets where credentials may be exposed, and
   review OIDC/client secrets.
6. Document remediation and update this runbook.

### 7. Data subject requests (access, portability, rectification)

- **Self-service export:** `GET /api/v1/me/export`, linked from Settings →
  "Your data" and from the Android Settings screen. It is a JSON attachment
  named `sandwitches-export-<username>.json` with `Cache-Control: no-store`.
- **Rectification:** profile fields are editable under `/settings` and
  `/profile`; administrators can also correct accounts.
- **Deadline:** answer requests not covered by self-service within one month
  (Article 12), extendable by two further months for complex requests with
  notice.
- **Identity:** for email requests, verify control of the account email before
  disclosing or deleting data. Record the request, verification, action and
  response date.

### 8. Erasure

- Self-service: `/settings` → Delete account (web) or Settings → Delete account
  (Android), typing the username to confirm, or the API
  `DELETE /api/v1/me` with `{"confirm":"<username>"}`.
- Fallback when the user cannot sign in: the public `/account/delete` page
  directs them to `LEGAL_CONTACT_EMAIL`, or to the administrator if no contact
  is configured.
- Administrators can delete users from `/dashboard/users`; that path calls the
  same `account.Service.Delete` and therefore applies the same demo and
  last-superuser guards.
- The shared demo account cannot be erased. Document this exception in the
  privacy policy (it is already stated) and in any internal erasure log.

### 9. Analytics objection

When Umami is enabled, `/settings` → **Privacy choices** stores
`AnalyticsOptOut` on the user. The analytics script is then not rendered for
that user (`utils.TemplateData` / `legalTemplateData`). Signed-out visitors can
use the browser's Do Not Track signal, which the Umami script is configured to
honour (`data-do-not-track` in `main.go`). This is the operational objection
path referenced in the policy. The analytics is self-hosted and cookieless; no
advertising cookie consent banner is implemented because there are no
advertising or cross-site cookies.

### 10. Consent record

`database.User.AcceptedTermsAt` (`accepted_terms_at`) records when the user
accepted the terms and privacy policy. It is set on signup, on first setup, when
an OIDC login provisions a new account (the login page carries the notice),
repaired for the demo account, and included in the JSON export. There is **no
terms version column**:
if the terms change materially, the timestamp alone does not prove which
version was accepted. Re-prompt and record acceptance separately if the
operator needs versioned consent.

### 11. Children

The policy states the instance is not intended for anyone under 16 and that
accounts for younger users are not knowingly created. Do not market the app to
children. If an under-16 account is discovered, delete it via the admin path
(the same anonymising deletion) and record the action.

### 12. International transfers

Record the location of the host, backup storage, SMTP relay, Gotify server,
Umami instance, OTLP collector, AI endpoint and OIDC provider. For any
non-EEA processing, keep the adequacy decision or signed SCCs. `LEGAL_HOSTING_REGION`
must not claim an EU region if backups or subprocessors sit elsewhere.

## Exact deletion behaviour

`internal/account/account.go` is the single implementation used by the web
form, the JSON API and the admin user-delete action.

Preconditions and guards:

1. A missing/deleted account returns `ErrNotFound` (API 404; web treats it as
   already deleted and clears the session).
2. `user.IsDemo` returns `ErrDemoAccount` before any transaction. The shared
   reviewer login is never deleted.
3. If the user is a superuser, the transaction counts other superusers. Zero
   others returns `ErrLastSuperuser` (API 409). This guard runs inside the
   transaction so two concurrent deletions cannot remove the last two admins.

Inside one transaction:

1. Set `is_active = false` and bump `session_version` (`1 → 2`, otherwise
   `+1`). Existing cookies carry the old version and stop authenticating on
   every request. The handler also clears and saves the current session, so the
   initiating browser/app is signed out immediately.
2. Detach content other people may depend on:
   - `UPDATE recipes SET uploaded_by_id = NULL`
   - `UPDATE orders SET user_id = NULL`
   - `UPDATE recipe_histories SET changed_by_id = NULL`
3. Delete account-owned references:
   - `user_groups` (group memberships)
   - `user_favorites`
   - `ratings` (comments are part of the rating row)
   - `cart_items`
   - `password_reset_tokens`
   - `mobile_auth_codes`
4. Delete the `users` row itself, removing username, email, name, bio, avatar
   reference, language, theme, password hash, OIDC subject, consent timestamp
   and the rest of the profile.
5. Commit. A failure rolls the whole transaction back.

What survives, intentionally:

- Authored recipes (public, authorless).
- Recipe edit history (authorless).
- Orders (anonymous; item names, quantities, prices and tracking token remain).
- Uploaded image files on disk, until the operator removes them.
- Logs and backups, until their rotation window expires.
- The OIDC/Authelia identity at the provider; deleting Sandwitches does not
  delete the upstream identity.

## Data export contents

`GET /api/v1/me/export` returns the `account.Export` JSON document:

- `exported_at`
- `profile`: id, username, email, first_name, last_name, bio, avatar,
  language, theme, is_staff, is_superuser, `oidc_subject` (when present),
  date_joined, last_login, `accepted_terms_at`, `analytics_opt_out`
- `groups`: names of groups the user belongs to
- `recipes`: authored recipes with title, slug, description, ingredients,
  instructions, servings, price, image reference, approval flag, tags and
  timestamps
- `ratings`: recipe id/title, score, comment, image reference, timestamps
- `favorites`: recipe id/title/slug
- `cart`: recipe id/title, quantity, line total
- `orders`: id, status, completed, total price, tracking token, timestamps and
  line items (recipe id/title, quantity, price)
- `recipe_edits`: recipe id, field name, old/new value, changed_at

Excluded by design:

- Password hash and session version.
- Password-reset token hashes and mobile sign-in code hashes.
- `is_active` / `is_demo` internal flags.
- The settings singleton and any other user's data.
- Binary image files; only image references/paths are exported.
- Server logs, request traces and backups.

The export is not rate-limited beyond the normal API limiter and is available
to the demo account as well. Treat downloads as personal data: they are served
with `Cache-Control: no-store` but the operator should not retain copies.

## EU launch checklist

- [ ] Set all `LEGAL_*` values, especially controller, contact email, hosting
      region and supervisory authority. Verify `LEGAL_BASE_URL` is the public
      HTTPS origin.
- [ ] Open `/privacy`, `/terms` and `/account/delete` in a private window with
      no cookies and confirm they render without login.
- [ ] Confirm the deletion page's canonical URL and the `/privacy` link are the
      public host.
- [ ] Terminate TLS; keep `DEBUG=false` so the session cookie is `Secure`.
- [ ] Publish the privacy-policy URL in Play Console; configure Data deletion
      and App access with the demo credentials.
- [ ] Decide whether the demo account should exist on this instance. Keep it
      off for private data; verify `DEMO_USER_ENABLED` is not accidentally
      inherited from `LEGAL_BASE_URL` pointing at the hosted host.
- [ ] Confirm a disposable account can export and delete itself in the app and
      on the web; confirm the demo account is refused.
- [ ] Sign the hosting/backup DPAs and record subprocessors and transfer
      safeguards.
- [ ] Write the Article 30 record and a concrete retention schedule for logs,
      backups, Umami and accounting records.
- [ ] Configure encrypted backups and test a restore into a scratch environment.
- [ ] Publish and rehearse the 72-hour breach runbook, including who notifies
      the authority and users.
- [ ] Monitor the privacy mailbox and record DSAR/erasure handling.
- [ ] Decide whether terms-versioned consent is required; the code stores only
      `accepted_terms_at`.
- [ ] Verify the analytics opt-out under Settings → Privacy choices actually
      stops the Umami script for that account.
- [ ] If OIDC, SMTP, Gotify, OTLP or AI is enabled, re-check the generated
      policy text and the Play Data safety answers.
- [ ] Keep `docs/play-store.md` and this document in sync with the deployed
      configuration.
