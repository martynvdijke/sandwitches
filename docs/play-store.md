# Google Play review and data-safety guide

This is the operator runbook for publishing the Sandwitches Android client and
for answering Google Play's account, privacy-policy and data-deletion checks.
The Android app is a native client for a self-hosted Sandwitches server; the
hosted instance is `https://sandwitches.vandijke.xyz`.

## Public URLs to enter in Play Console

Use the hosted instance unless you publish a fork that points somewhere else.
Every URL below loads without an account.

| Purpose | URL |
| --- | --- |
| Privacy policy | `https://sandwitches.vandijke.xyz/privacy` |
| Terms of use | `https://sandwitches.vandijke.xyz/terms` |
| Account deletion information | `https://sandwitches.vandijke.xyz/account/delete` |

Replace `sandwitches.vandijke.xyz` with your own host when reviewing a
self-hosted deployment.

The deletion page is deliberately public. `main.go` registers `/privacy`,
`/terms` and `/account/delete` before the authenticated CSRF group, and
`internal/handlers/legal.go` builds their template data without touching the
session store. A signed-out reviewer therefore sees the page, not a login
redirect or a 403, and no session cookie is created just by reading it.

The privacy policy URL must also be reachable from the app: Settings → Privacy
policy / Terms of use / Delete account information open the same `/privacy`,
`/terms` and `/account/delete` pages.

## Play Console setup

### Store listing

1. Open **Grow → Store presence → Main store listing** (older consoles:
   **Store listing**).
2. Set **Privacy policy** to `https://sandwitches.vandijke.xyz/privacy`.
3. Keep the listing text consistent with the app: it is a client for a
   self-hosted server and requires a server URL plus an account.

### App content → Data deletion

1. Open **Policy → App content → Data deletion**.
2. Enter the deletion URL: `https://sandwitches.vandijke.xyz/account/delete`.
3. Answer **Yes, users can request deletion in app** — the Android app exposes
   Settings → Delete account and calls `DELETE /api/v1/me`.
4. Optionally add a deletion contact address; the policy page publishes
   `LEGAL_CONTACT_EMAIL` when configured.
5. Do not invent a data-deletion file in `fastlane/`. Google Play's deletion URL
   is Console configuration, not fastlane metadata.

### App content → App access

The app requires a sign-in for most screens, so declare restricted access
rather than "all functionality is available without special access".

1. Open **Policy → App content → App access**.
2. Choose **All or some functionality is restricted**.
3. Add instructions: the app defaults to `https://sandwitches.vandijke.xyz`;
   open Settings → Privacy policy / Delete account for the legal pages.
4. Add the shared reviewer login:

   | Field | Value |
   | --- | --- |
   | Username | `demo` |
   | Password | `demo` |

   These are the defaults from `internal/config/config.go`
   (`DEMO_USER_USERNAME=demo`, `DEMO_USER_PASSWORD=demo`). If the deployment
   overrides either variable, enter the deployed values instead.

The login screen also advertises the demo login when it is enabled, and
`GET /api/v1/auth/config` returns the same public username/password for API
clients.

## Demo account

The demo login exists only so reviewers and visitors can try the instance
without registering. It is provisioned at boot by `internal/demo/demo.go`.

| Variable | Default | Meaning |
| --- | --- | --- |
| `DEMO_USER_ENABLED` | enabled only on the hosted host | Master switch. Unset means "enable when `LEGAL_BASE_URL`'s host is `sandwitches.vandijke.xyz`"; `true`/`false` forces it. |
| `DEMO_USER_USERNAME` | `demo` | Lower-cased shared username. |
| `DEMO_USER_PASSWORD` | `demo` | Shared password. Must be non-empty. |
| `DEMO_USER_EMAIL` | `demo@sandwitches.local` | Lower-cased email used only if the existing row has none. |

Properties enforced by the implementation:

- **Public and shared.** The username and password are shown on the login page
  and returned by `GET /api/v1/auth/config`; anyone can use and modify the
  account's ordinary content.
- **Non-admin and never a takeover.** The demo account is created as a regular
  active user in the `community` group. If the configured username already
  belongs to an account that is not already marked as the demo account
  (including any staff or superuser account), `demo.Ensure` refuses to touch it
  and logs `Demo account not available: ...` at startup. It never resets a real
  user's password.
- **Not deletable.** `account.Delete` returns `ErrDemoAccount` when
  `user.IsDemo` is true. Both the web/JSON self-service paths and the admin user
  deletion path surface "The demo account cannot be deleted." The intent is to
  keep the store-review login from disappearing.
- **Terms already accepted.** `accepted_terms_at` is set when the demo user is
  created or repaired.
- **Seed content.** On a database with no approved recipes, provisioning adds
  two approved text-only recipes ("Demo Cheese Sandwich" and "Demo Tomato
  Sandwich") tagged `demo` and `vegetarian`. Existing instances are not seeded.
- **Hosted default.** Because `DEMO_USER_ENABLED` defaults to
  `isDemoHost(LEGAL_BASE_URL)`, the hosted deployment gets the demo login when
  the variable is unset. The shipped `.env.example` leaves the variable
  commented out so that host-based default applies; set it to `true` or `false`
  to force the switch. A self-hoster should point `LEGAL_BASE_URL` at their own
  host (which disables the demo by default) or set `DEMO_USER_ENABLED=false`
  explicitly. A self-hosted instance that points `LEGAL_BASE_URL` at
  `sandwitches.vandijke.xyz` would also auto-enable the demo login.

Keep the demo login disabled on any instance holding private data: the
credentials are public by design.

## In-app account deletion path

1. Sign in with the reviewer account.
2. Open **Settings**.
3. Tap **Delete account** (subtitle: "Permanently close this account").
4. Type the account username exactly (`demo` for the reviewer login).
5. Confirm.

The app calls `DELETE /api/v1/me` with JSON body `{"confirm":"<username>"}`.
This is the same handler and the same `internal/account.Service.Delete` used by
the web form `POST /settings/account/delete`. On success the app clears its
stored auth state and cookie jar and returns to the login screen. The demo
account is the one account that cannot complete this flow; reviewers should try
deletion with a disposable account if they need to verify it end to end.

Web users take the same path from `/settings`: type the username, submit the
CSRF-protected form, and get redirected to `/account/delete?deleted=1`.

## Data safety questionnaire

Answer for the hosted instance and the app as shipped. The Android client has
no advertising SDK, no third-party analytics SDK and no crash-reporting SDK.
The server is the system of record.

### Data types

| Play category | What it maps to | Collected? | Purpose |
| --- | --- | --- | --- |
| Personal info → Name | First and last name | Yes, if the user enters them | App functionality / account management |
| Personal info → Email address | Account email | Yes | App functionality / account management |
| Personal info → User IDs | Username, numeric account ID, OIDC subject | Yes | App functionality / account management |
| Personal info → Other info | Bio, avatar reference, language/theme preference | Yes | App functionality |
| Photos and videos → Photos | Recipe images uploaded by users; avatar references when set | Yes, when provided | App functionality |
| User content → Other user-generated content | Recipes, ratings/comments, recipe edit history | Yes | App functionality |
| App activity → Other actions | Cart, favourites, ratings, page views if analytics is enabled | Yes | App functionality; analytics only when the operator enables Umami |
| App activity → In-app search history | Recipe search/filter queries | No dedicated history is stored by the server; requests may appear in transient logs | Security/operations |
| Identifiers → Device or other IDs | No advertising ID, IMEI or device ID is collected by the app | No | — |
| Location | No location permission or collection | No | — |
| Financial info → Purchase history | Order history, status, items and totals | Yes | App functionality and the instance's transaction record; no payment card data is processed by the app |
| Financial info → Other financial info | No payment-card, bank or credit data | No | — |
| Health and fitness | No | No | — |
| Messages | Ratings/comments are public user content, not private messages | See user content | — |

### Required answers

- **Is all data encrypted in transit?** Yes for the hosted service: it is served
  over HTTPS and the session cookie is marked `Secure` outside debug builds.
  Self-hosters must terminate TLS themselves.
- **Can users request that their data be deleted?** Yes. In-app Settings →
  Delete account and the public `/account/delete` page; the request is
  immediate. The shared demo account is the documented exception.
- **Is data sold or shared for advertising?** No. There are no ads, no data
  sale, no advertising cookies and no cross-site profiling.
- **Is data used for personalization?** No beyond remembering language/theme and
  session state.
- **Optional analytics.** Only if `UMAMI_HOST` and `UMAMI_WEBSITE_ID` are set on
  the instance: self-hosted, cookieless Umami page views, used for aggregate
  usage, with a per-account opt-out under Settings → Privacy choices. If you do
  not run Umami, answer that analytics is not collected.
- **Data sharing.** Data reaches the instance's hosting provider, and only when
  enabled: SMTP provider (email delivery), Gotify server (push), self-hosted
  Umami, OTLP collector, configured AI endpoint, and the OIDC provider used for
  sign-in. These are service providers, not advertising recipients.
- **Data retention.** Account data lives while the account exists; deletion
  removes the account immediately and anonymises retained recipes, edit history
  and orders. Backups rotate out on the operator's schedule.

Play's Data safety form is a declaration by the developer. Re-check it whenever
the operator enables a new processor (email, Gotify, OTLP, AI, OIDC), because
the privacy template's processor list is generated from the running config.

## What the deletion page promises

`templates/account_delete.html` and `templates/privacy.html` state the same
behaviour that `internal/account/account.go` implements.

**Deleted immediately**

- The user row: username, email, first/last name, bio, avatar reference,
  language, theme, password hash, OIDC subject, `accepted_terms_at`,
  `analytics_opt_out`, session version and active flag.
- Group memberships, favourites, ratings and their comments, cart rows,
  password-reset tokens and mobile one-time sign-in codes.
- Every session: the stored session version is bumped and the current session
  cookie is cleared, so cookies on other devices stop working.

**Retained, detached from the person**

- Recipes the user authored stay published with `uploaded_by_id = NULL`. They
  can still be viewed, ordered and rated. Operators may remove them.
- Recipe edit-history entries stay as anonymous history
  (`changed_by_id = NULL`).
- Orders stay as the instance's transaction records with `user_id = NULL`.
  Item names, quantities, prices and the tracking token are retained, so an
  already-shared tracking link keeps working.

**Other limits stated on the page**

- Backups taken before deletion may contain the old data until they rotate out;
  backups are not used to restore an individual deleted account.
- Uploaded recipe images and avatars may remain on disk after the account row is
  gone; the operator can remove them on request.
- OIDC/Authelia identities are not deleted at the identity provider.
- The last remaining superuser cannot be deleted through this route; promote
  another administrator first.
- The shared demo account cannot be deleted.

## fastlane metadata

`fastlane/metadata/android/en-US/privacy_url.txt` contains

```text
https://sandwitches.vandijke.xyz/privacy
```

If you publish listing metadata with fastlane (`supply`), that file becomes the
privacy-policy URL. The current CI does **not** do that:
`.github/workflows/play-store.yaml` builds the signed AAB and uploads only the
bundle with `r0adkll/upload-google-play` (`releaseFiles:
mobile/build/app/outputs/bundle/release/app-release.aab`). No metadata,
screenshots or store-listing fields are uploaded by CI, so Play Console remains
the source of truth for the privacy-policy URL and the data-deletion URL.

There is intentionally no `data_deletion_url.txt` or similar fastlane file.
Play Console configures data deletion under App content → Data deletion.
