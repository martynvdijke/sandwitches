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
  <img src="https://img.shields.io/badge/django-6.0-green.svg" alt="Django Version">
  <a href="https://github.com/astral-sh/ruff">
    <img src="https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/astral-sh/ruff/main/assets/badge/v2.json" alt="Ruff">
  </a>
</p>

---

## ✨ Overview

Sandwitches is a modern, recipe management platform built with **Django 6.0**.
It is made as a hobby project for my girlfriend, who likes to make what I call "fancy" sandwiches (sandwiches that go beyond the Dutch normals), lucky to be me :).
See wanted to have a way to advertise and share those sandwtiches with the family and so I started coding making it happen, in the hopes of getting more fancy sandwiches.

## 📥 Getting Started

### Prerequisites

* Python 3.12+
* [uv](https://github.com/astral-sh/uv) (recommended) or pip

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
    uv run invoke setup-ci  # Sets up environment variables
    uv run src/manage.py migrate
    uv run src/manage.py collectstatic --noinput
    ```

4. **Start the development server**:

    ```bash
    uv run src/manage.py runserver
    ```

## 🧪 Testing & Quality

The project maintains high standards with over **80+ automated tests**.

* **Run tests**: `uv run invoke tests`
* **Linting**: `uv run invoke linting`
* **Type checking**: `uv run invoke typecheck`

---

<p align="center">
  Made with ❤️ for sandwich enthusiasts.
</p>
