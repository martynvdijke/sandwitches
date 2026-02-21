<p align="center">
  <img src="src/static/icons/banner.svg" alt="Sandwitches Banner" width="600px">
</p>

<h1 align="center">🥪 Sandwitches</h1>

<p align="center">
  <strong>Sandwiches so good, they haunt you!</strong>
</p>

<p align="center">
  <a href="https://github.com/martynvdijke/sandwitches/actions/workflows/ci.yaml">
    <img src="https://github.com/martynvdijke/sandwitches/actions/workflows/ci.yaml/badge.svg" alt="CI Status">
  </a>
  <a href="https://github.com/martynvdijke/sandwitches/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/martynvdijke/sandwitches" alt="License">
  </a>
  <img src="https://img.shields.io/badge/python-3.12+-blue.svg" alt="Python Version">
  <a href="https://github.com/astral-sh/ruff">
    <img src="https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/astral-sh/ruff/main/assets/badge/v2.json" alt="Ruff">
  </a>
</p>

---

## ✨ Overview

Sandwitches is a modern, recipe management platform built with **Django**.
It is made as a hobby project for my girlfriend, who likes to make what I call "fancy" sandwiches (sandwiches that go beyond the Dutch normals), lucky to be me :).
Sandwiches so good you will think they are haunted !.
See wanted to have a way to advertise and share those sandwiches with the family and so I started coding making it happen, in the hopes of getting more fancy sandwiches.

![](./docs/overview.png)

## 🎯 Features

Sandwitches comes packed with comprehensive features for recipe management, community engagement, and ordering:

- **🍞 Recipe Management** - Upload and create sandwich recipes with images, ingredients, and instructions
- **👥 Community Page** - Discover and browse sandwiches shared by community members
- **🛒 Ordering System** - Browse recipes and place orders with cart functionality and order tracking
- **⭐ Ratings & Reviews** - Rate recipes on a 0-10 scale with detailed comments
- **🔌 REST API** - Full API access for recipes, tags, ratings, orders, and user management
- **📊 Admin Dashboard** - Comprehensive admin interface for recipe approval and site management
- **🌍 Multi-language Support** - Internationalization for multiple languages
- **📱 Responsive Design** - Mobile-friendly interface with BeerCSS framework
- **🔔 Notifications** - Email and Gotify push notification integration
- **📈 Order Tracking** - Real-time order status tracking with unique tracking tokens
- **📊 Analytics** - Umami analytics integration for tracking user behavior

## 📥 Getting Started

```bash
services:
  sandwitches:
    image: martynvandijke/sandwitches:latest
    container_name: sandwitches
    environment:
     - ALLOWED_HOSTS=localhost,127.0.0.1,[::1]
     - CSRF_TRUSTED_ORIGINS=http://localhost:8000,http://127.0.0.1:8000
     - SECRET_KEY=superdupersecretkey
     - DATABASE_FILE=/config/db.sqlite3
     - MEDIA_ROOT=/config/media
    ports:
      - 6270:6270
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:6270/api/ping"]
      interval: 5s
      timeout: 10s
      retries: 3
    volumes:
      - /path/to/sandwitches:/config
    restart: always
```

### Environment variables

Below is a list of all supported environment variables.

| **Variable**         | **Required** | **Description**                                                                       |
| -------------------- | ------------ | ------------------------------------------------------------------------------------- |
| ALLOWED_HOSTS        | Yes          | A list of strings representing the host/domain names that this Django site can serve. |
| CSRF_TRUSTED_ORIGINS | Yes          | A list of trusted origins for safe cross-site requests (e.g., <https://example.com>). |
| SECRET_KEY           | Yes          | A unique, secret value used for cryptographic signing and session security.           |
| DATABASE_FILE        | Yes          | The file path to the SQLite database or the name of the database being used.          |
| MEDIA_ROOT           | Yes          | The absolute filesystem path to the directory that will hold user-uploaded files.     |
| SMTP_USE_TLS         | No           | Boolean (True/False) to enable/disable TLS encryption for outgoing emails.            |
| SMTP_HOST            | No           | The hostname or IP address of the mail server used to send emails.                    |
| SMTP_PORT            | No           | The port number to use for the SMTP server (usually 587 for TLS or 465 for SSL).      |
| SMTP_FROM_NAME       | No           | The display name that appears in the "From" field of outgoing emails.                 |
| SMTP_FROM_EMAIL      | No           | The actual email address used as the sender for system-generated messages.            |
| SMTP_USER            | No           | The username required to authenticate with the SMTP server.                           |
| SMTP_PASSWORD        | No           | The password required to authenticate with the SMTP server.                           |
| GOTIFY_URL           | No           | The base URL of your Gotify server instance for push notifications.                   |
| GOTIFY_TOKEN         | No           | The application-specific token used to authenticate with Gotify.                      |
| UMAMI_HOST           | No           | UMAMI analytics tracking host.                                                        |
| UMAMI_WEBSITE_ID     | No           | UMAMI analytics website id.                                                           |

## Development setup

### Prerequisites

- Python 3.12+
- [uv](https://github.com/astral-sh/uv) (recommended) or pip

### Installation

1. **Clone the repository**:

    ```bash
    git clone https://github.com/martynvdijke/sandwitches.git
    cd sandwitches
    ```

2. **Sync dependencies**:

    ```bash
    uv sync
    ```

3. **Run migrations and collect static files**:

    ```bash
    uv run src/manage.py migrate
    uv run src/manage.py collectstatic --noinput
    ```

4. **Start the development server**:

    ```bash
    uv run src/manage.py runserver
    ```

## 🧪 Testing & Quality

This project wraps all tests, linting and formatting in `invoke` tasks so you can run:

- **Run tests**: `uv run invoke tests`
- **Linting**: `uv run invoke linting`
- **Type checking**: `uv run invoke typecheck`

---

<p align="center">
  Made with ❤️ for sandwich enthusiasts.
</p>
