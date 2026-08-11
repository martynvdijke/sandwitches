# Django → Go Port Guide

This document explains how Sandwitches was ported from Django (Python) to Go
(Gin + GORM + SQLite + HTMX). It is written so that:

1. You understand the mapping between the old Django app and the new Go app.
2. You can answer "how do I port X?" for any future feature.
3. You can re-run or extend the one-time data migration from a Django database.

The old Django codebase is **gone** from the repository (removed in commit
`56626eb`), but the schema and data survive: the Go server can auto-migrate an
existing Django `db.sqlite3` on first boot.

---

## Why Go?

- **Single static binary** — no Python runtime, no pip/uv, no gunicorn workers.
- **Faster startup and lower memory** — important for a hobby project running on
  a small VPS and for TRMNL e-ink displays polling the API.
- **HTMX + server-rendered templates** carry over 1:1 in spirit; Go's
  `html/template` replaces Django templates.
- The frontend (webpack + BeerCSS + htmx + chart.js + EasyMDE + cropperjs) was
  **kept as-is** — only the server layer was rewritten.

---

## Architecture Mapping

| Django (old)                          | Go (new)                                        |
| ------------------------------------- | ----------------------------------------------- |
| `manage.py` + `sandwitches/settings.py` | `main.go` + `internal/config/config.go` (env-driven) |
| `sandwitches/urls.py`                 | Route registration in `main.go` (gin engine)    |
| Django ORM models                     | GORM models in `internal/database/models.go`    |
| `sandwitches/migrations/`             | GORM `AutoMigrate` in `internal/database/database.go` |
| Django templates (`{% %}`)            | Go `html/template` (`{{ }}`) in `templates/`    |
| Django views                          | `internal/handlers/*.go`                        |
| Django middleware (auth, CSRF)        | `internal/middleware/auth.go`, `internal/middleware/csrf.go` |
| Django sessions                       | Signed cookie session (`sandwitches_session`)   |
| Django admin                          | Custom `/dashboard` in `internal/handlers/admin.go` |
| `static/` + `media/`                  | `static/` (webpack-built) + configurable `MEDIA_ROOT` |
| Django i18n (`locale/`)               | `internal/utils/i18n.go`                        |
| Celery/background tasks               | `internal/tasks/tasks.go` (goroutines)          |
| REST via DRF                          | `internal/api/api.go` (plain gin JSON)          |
| `asgi.py`/gunicorn                    | `supervisord.conf` running the compiled binary  |

### Project layout (new)

```
.
├── main.go                    # Entry point & route setup
├── internal/
│   ├── api/                   # REST API v1 (JSON)
│   ├── config/                # Environment configuration
│   ├── database/              # GORM init, models, Django migration
│   ├── handlers/              # Page handlers (public, auth, admin, profile, feed)
│   ├── middleware/            # Auth + CSRF middleware
│   ├── render/                # Template renderer that surfaces errors as 500s
│   ├── tasks/                 # Background tasks (notifications, resets)
│   └── utils/                 # Template helpers, flash, i18n, images, markdown
├── templates/                 # Go HTML templates (head/navbar/tail pattern)
├── static/                    # Frontend source (js/, css/, icons/) + dist/
├── trmnl/                     # TRMNL e-ink display Liquid templates
├── Dockerfile                 # Multi-stage: node → go → alpine runtime
├── supervisord.conf           # Process manager for the Docker runtime
└── Taskfile.yml               # Task runner (build/dev/test/clean)
```

---

## The One-Time Data Migration

The Go server can import an existing Django SQLite database **without losing
data**. This is how the live site was moved over.

### How it works

1. `main.go` checks `Django_DB_PATH` (env var). If unset, it probes
   `db.sqlite3` and `/config/db.sqlite3` for a Django database.
2. If found, `database.MigrateFromDjango(path)` runs before the server starts
   serving traffic.
3. The migration (in `internal/database/migrate_django.go`):
   - Opens the Django DB read-only with a GORM session.
   - Migrates users (password hashes are kept — Django `pbkdf2_sha256` hashes
     are detected and verified with Go's `crypto/pbkdf2`).
   - Migrates recipes, tags, ratings, orders, order items, cart items, and
     groups, preserving foreign-key relationships and image filenames.
   - Copies user media files into `MEDIA_ROOT` (avatars, recipe images).
   - Idempotent: users already present (matched by username) are skipped, so a
     re-run does not duplicate data.
4. On failure it logs `Django migration skipped: <err>` and the app boots on a
   fresh database — it never blocks startup.

### Running it yourself

```bash
# Point the app at a copy of the old Django db
Django_DB_PATH=/path/to/db.sqlite3 \
MEDIA_ROOT=/path/to/media \
DATABASE_FILE=/path/to/new/db.sqlite3 \
go run .
```

> **Always test against a copy first.** The migration is best-effort and was
> written for the production schema at the time of the port.

---

## Template Porting Reference

### Django → Go syntax

| Django                              | Go `html/template`                        |
| ----------------------------------- | ----------------------------------------- |
| `{% for r in recipes %}`            | `{{ range .recipes }}`                    |
| `{% if user.is_authenticated %}`    | `{{ if .user }}`                          |
| `{{ recipe.title }}`                | `{{ .Title }}` (dot is the loop/root var) |
| `{% url 'recipe' r.slug %}`         | hardcoded `/recipes/{{ .Slug }}`          |
| `{% include "x.html" %}`            | `{{ template "partials/x.html" . }}`      |
| `{% block content %}`               | `{{ template "content" . }}` or the head/navbar/tail pattern |
| filters (`|default`, `|date`)       | funcmap helpers: `default`, `add`, `sub`, `floatformat`, `iso8601_duration`, ... |

### Layout pattern

Every public page is a composition of three defines loaded from
`templates/`:

```
{{ template "head" . }}
{{ template "navbar" . }}
...page content...
{{ template "tail" . }}
```

Defines live in `head.html`, `components/navbar.html`, and `tail.html`.
Admin pages use `admin_base.html` + `admin_tail.html`.

### The #1 gotcha: nil users

Anonymous users render most pages. Any template that dereferences a `user`
field without guarding **silently truncates the whole page** — gin's default
renderer discards the `ExecuteTemplate` error and writes a partial 200
response. This was the actual homepage bug:

```gotemplate
{{! BAD — panics on nil *User for anonymous visitors }}
{{ if $.user.Username }}{{ if .Price }}

{{! GOOD }}
{{ if $.user }}{{ if .Price }}
```

Rules:

- Guard every `.user` dereference on public pages: `{{ if $.user }}`.
- Guard `{{ .User.Username }}` (ratings/orders): `{{ if .User }}`.
- Always use `internal/render` (registered as `router.HTMLRender`) so template
  errors become **HTTP 500 + a log line** instead of a silent partial page.
  Never switch back to `router.SetHTMLTemplate`.

---

## Environment Variables

The full table lives in `README.md`. Key ones:

| Variable           | Purpose                                            |
| ------------------ | -------------------------------------------------- |
| `SECRET_KEY`       | Cookie/session signing (required)                  |
| `ALLOWED_HOSTS`    | Host allow-list (required)                         |
| `CSRF_TRUSTED_ORIGINS` | Origins allowed to POST (required)             |
| `DATABASE_FILE`    | SQLite path (default `db.sqlite3`)                 |
| `MEDIA_ROOT`       | Uploaded files dir (default `media`)               |
| `PORT`             | Listen port (default `6270`)                       |
| `Django_DB_PATH`   | Legacy Django DB to auto-migrate from              |
| `GOTIFY_URL`/`TOKEN`, `SMTP_*`, `UMAMI_*` | Notifications & analytics |

---

## Deployment

The `Dockerfile` is a 3-stage build:

1. **node:24-alpine** — webpack bundles `static/js/index.js` → `static/dist/`.
2. **golang:1.26-alpine** — compiles the CGO-enabled binary (SQLite needs CGO).
3. **alpine:3.24** — runtime: `supervisord` runs the binary, `entrypoint.sh`
   validates env, healthcheck hits `/api/v1/ping`.

```bash
docker build -t martynvandijke/sandwitches:latest .
```

`docker-compose.yaml` mounts a config volume for the SQLite DB and media.
The `release.yaml` GitHub Action rebuilds and pushes the image on release.

---

## Development Workflow

```bash
npm install && npm run build   # frontend → static/dist
go mod download
go run .                        # or: task dev (air hot reload)
task build                      # both frontend + Go binary
task test                       # go test ./...
```

- UI tests (Playwright) live in `tests_go/`; they build the binary themselves
  and drive a server on port 6279. Run with `uv sync --dev && uv run pytest`.
- CI (`ci.yaml`) runs `go vet` + `go test` for the Go job and the Playwright
  suite for the UI job.
- `.pre-commit-config.yaml` runs `go vet` and `gofmt -l` locally.

---

## Porting a New Django Feature (checklist)

1. **Model** → add a GORM struct in `internal/database/models.go` + a slug
   helper if needed; `AutoMigrate` picks it up.
2. **View/URL** → add a handler in `internal/handlers/` and register the route
   in `main.go` (inside the CSRF group if it mutates state, behind
   `StaffRequired` if admin-only).
3. **Template** → create the `.html` in `templates/`, use the head/navbar/tail
   pattern, guard all `.user` derefs.
4. **Data** → if it must import from Django, extend `MigrateFromDjango`.
5. **Verify** → `go build ./... && go vet ./... && go test ./...`; render the
   page as an anonymous user AND as a logged-in user.
