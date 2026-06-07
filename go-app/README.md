# Sandwitches Go — Recipe Management Platform

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/Gin-1.12-00ADD8?style=flat&logo=go" alt="Gin">
  <img src="https://img.shields.io/badge/GORM-ORM-00ADD8?style=flat&logo=go" alt="GORM">
  <img src="https://img.shields.io/badge/SQLite3-003B57?style=flat&logo=sqlite" alt="SQLite">
  <img src="https://img.shields.io/badge/license-MIT-blue" alt="License">
  <img src="https://img.shields.io/badge/docker-ready-2496ED?style=flat&logo=docker" alt="Docker">
</p>

A full-featured recipe management platform with user accounts, shopping cart, order tracking, ratings, favorites, and a community recipe sharing system. Built with Go, Gin, GORM, and HTMX for a fast, dynamic user experience.

## Features

### Recipe Management
- **Full Recipe CRUD** — Create, edit, delete recipes with title, description, ingredients, instructions, servings, and pricing
- **Tag System** — Categorize recipes with tags, filterable and manageable through admin
- **Image Upload** — Upload recipe images with automatic thumbnail/small/medium/large size variants
- **Favorites** — Users can favorite recipes for quick access
- **Ratings & Reviews** — 1-5 star rating system with optional comments
- **Ingredient Scaling** — Auto-scale ingredient quantities to any serving size
- **Recipe of the Day** — Daily randomly selected featured recipe
- **Recipe History** — Audit trail of all recipe changes
- **Approval Workflow** — Admin moderation for community-submitted recipes
- **Markdown Support** — Write recipes in Markdown with custom template rendering helpers
- **Slug-Based URLs** — SEO-friendly recipe URLs

### User System
- **Registration & Login** — User signup, login, and logout with cookie sessions
- **User Profiles** — Avatar, bio, first/last name, language preference, theme selection
- **Dashboard** — Staff users get an admin dashboard with full management capabilities
- **Group-Based Permissions** — Admin and community groups with different access levels

### Shopping & Orders
- **Shopping Cart** — Add/remove/update quantities, persisted per user
- **Order System** — Place orders with tracking tokens for order status
- **Order Tracking** — Public order tracking via unique token URL
- **Admin Order Management** — View and update order statuses

### Notifications
- **Email (SMTP)** — Order confirmations with item details, new recipe alerts to all users with email
- **Gotify Push** — Real-time push notifications for new orders and new recipes
- **Background Tasks** — Async notification delivery, daily order count reset

### Administration
- **Admin Dashboard** — Overview of recipes, users, tags, orders, ratings
- **User Management** — Full CRUD with role/group assignment
- **Recipe Moderation** — Approve or reject community recipes
- **Tag Management** — Create, edit, delete tags
- **Rating Moderation** — View and remove ratings
- **Order Management** — View all orders, update statuses
- **Log Viewer** — Application log viewer with filtering
- **Task Management** — Run maintenance tasks from admin panel
- **Site Settings** — Global site configuration

### Integrations
- **Umami Analytics** — Optional privacy-friendly analytics
- **Gotify** — Self-hosted push notifications
- **SMTP Email** — Transactional emails
- **Django Migration** — Import users and recipes from a legacy Django database

### i18n
- **Internationalization** — Locale-based translations (locale files included)
- **User Language Preference** — Per-user language selection

### Tech Stack
- **Backend:** Go 1.26, Gin, GORM, SQLite (WAL mode)
- **Frontend:** Go templates, HTMX, custom CSS
- **Build Tools:** Webpack for static asset bundling, Supervisor for process management
- **Container:** Multi-stage Docker (Node.js build + Go build + Alpine runtime)

## Quick Start

### Docker (Recommended)

```bash
docker build -t sandwitches-go .
docker run -p 6270:6270 \
  -e SECRET_KEY=your-secret-key \
  sandwitches-go
```

Or use with docker-compose:

```yaml
services:
  sandwitches:
    build: .
    ports:
      - "6270:6270"
    environment:
      - SECRET_KEY=your-secret-key
      - DEBUG=true
    volumes:
      - sandwitches-data:/app/data
```

### Manual Setup

```bash
# Install Go dependencies
go mod download

# Install Node dependencies and build assets
npm install && npm run build

# Build
CGO_ENABLED=0 go build -o sandwitches .

# Run
SECRET_KEY=your-secret-key ./sandwitches
```

Open **[http://localhost:6270](http://localhost:6270)** in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `6270` | HTTP listen port |
| `SECRET_KEY` | — | Session encryption key **(required in production)** |
| `DEBUG` | `false` | Enable debug mode |
| `DATABASE_FILE` | `{BASE_DIR}/db.sqlite3` | SQLite database path |
| `MEDIA_ROOT` | `{BASE_DIR}/media` | Media upload directory |
| `ALLOWED_HOSTS` | — | Comma-separated allowed hostnames |
| `CSRF_TRUSTED_ORIGINS` | — | Trusted origins for CSRF |
| `BASE_DIR` | CWD | Base directory for defaults |
| `SEND_EMAIL` | `false` | Enable email notifications |
| `SMTP_HOST` | — | SMTP server hostname |
| `SMTP_PORT` | — | SMTP server port |
| `SMTP_USER` | — | SMTP username |
| `SMTP_PASSWORD` | — | SMTP password |
| `SMTP_FROM_EMAIL` | `noreply@sandwitches.local` | Sender email address |
| `GOTIFY_URL` | — | Gotify push notification server URL |
| `GOTIFY_TOKEN` | — | Gotify app token |
| `UMAMI_HOST` | — | Umami analytics host |
| `UMAMI_WEBSITE_ID` | — | Umami website ID |
| `LANGUAGE_CODE` | `en` | Default language |
| `BASE_URL` | `http://localhost` | Public base URL (for emails) |
| `Django_DB_PATH` | — | Path to legacy Django database for migration |

## Project Structure

```
go-app/
├── main.go                    # Application entry point & route setup
├── internal/
│   ├── api/
│   │   └── api.go             # REST API v1 (JSON)
│   ├── config/
│   │   └── config.go          # Environment configuration
│   ├── database/
│   │   ├── database.go        # GORM initialization & migrations
│   │   ├── models.go          # Data models (User, Recipe, Order, etc.)
│   │   └── migrate_django.go  # Django database migration
│   ├── handlers/
│   │   ├── public.go          # Public page handlers
│   │   ├── auth.go            # Authentication handlers
│   │   ├── profile.go         # User profile handlers
│   │   ├── feed.go            # Recipe feed handlers
│   │   └── admin.go           # Admin panel handlers
│   ├── middleware/
│   │   ├── auth.go            # Authentication middleware
│   │   └── csrf.go            # CSRF protection middleware
│   ├── tasks/
│   │   └── tasks.go           # Background tasks (notifications, resets)
│   └── utils/
│       ├── template.go        # Template helpers
│       ├── flash.go           # Flash message helpers
│       ├── pagination.go      # Pagination utilities
│       ├── ingredients.go     # Ingredient parsing & scaling
│       ├── markdown.go        # Markdown rendering
│       ├── image.go           # Image processing
│       └── i18n.go            # Internationalization
├── templates/                 # Go HTML templates
├── static/                    # Static assets
├── locale/                    # i18n translation files
├── migrations/                # Database migrations
├── Dockerfile                 # Multi-stage Docker build
├── supervisord.conf           # Supervisor process manager config
├── webpack.config.js          # Webpack asset bundling config
└── go.mod / go.sum            # Go module dependencies
```

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/api/v1/ping` | No | Health check |
| `GET` | `/api/v1/settings` | No | Public site settings |
| `POST` | `/api/v1/settings` | Staff | Update settings |
| `GET` | `/api/v1/me` | Yes | Current user profile |
| `GET` | `/api/v1/users` | No | List users |
| `GET` | `/api/v1/recipes` | No | List recipes |
| `POST` | `/api/v1/recipes` | Yes | Create recipe |
| `GET` | `/api/v1/recipes/:id` | No | Get recipe |
| `PATCH` | `/api/v1/recipes/:id` | Yes | Update recipe |
| `DELETE` | `/api/v1/recipes/:id` | Yes | Delete recipe |
| `GET` | `/api/v1/recipes/:id/scale-ingredients` | No | Scale ingredients |
| `GET` | `/api/v1/recipes/:id/rating` | No | Get average rating |
| `POST` | `/api/v1/recipes/:id/ratings` | Yes | Rate a recipe |
| `GET` | `/api/v1/recipe-of-the-day` | No | Daily recipe |
| `GET` | `/api/v1/tags` | No | List tags |
| `POST` | `/api/v1/tags` | Staff | Create tag |
| `GET` | `/api/v1/tags/:id` | No | Get tag |
| `PATCH` | `/api/v1/tags/:id` | Staff | Update tag |
| `DELETE` | `/api/v1/tags/:id` | Staff | Delete tag |
| `GET` | `/api/v1/orders` | Yes | List orders (own or all for staff) |
| `POST` | `/api/v1/orders` | Yes | Create order |
| `GET` | `/api/v1/orders/:id` | Yes | Get order |
| `PATCH` | `/api/v1/orders/:id` | Staff | Update order status |
| `GET` | `/api/v1/cart` | Yes | Get cart items |
| `POST` | `/api/v1/cart` | Yes | Add to cart |
| `PATCH` | `/api/v1/cart/:id` | Yes | Update cart item |
| `DELETE` | `/api/v1/cart/:id` | Yes | Remove from cart |

## License

MIT
