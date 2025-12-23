#!/usr/bin/env bash
set -euo pipefail
shopt -s nocasematch

# Simple truthy check: 1, true, yes, on => true
is_truthy() {
  case "${1:-}" in
    1|true|yes|on) return 0 ;;
    *) return 1 ;;
  esac
}

DEBUG=${DEBUG:-0}

if is_truthy "$DEBUG"; then
  debug=true
else
  debug=false
fi

# SECRET_KEY must be set in production
if [ -z "${SECRET_KEY:-}" ]; then
  if ! $debug; then
    echo "ERROR: SECRET_KEY is not set and DEBUG is false. Set SECRET_KEY in environment for production." >&2
    exit 1
  else
    echo "WARNING: SECRET_KEY not set — using development fallback (insecure)." >&2
  fi
fi

# Warn if ALLOWED_HOSTS is empty
if [ -z "${ALLOWED_HOSTS:-}" ]; then
  echo "WARNING: ALLOWED_HOSTS is empty. Consider setting ALLOWED_HOSTS to a comma-separated list (e.g., 'example.com,127.0.0.1')." >&2
fi

# Warn if CSRF_TRUSTED_ORIGINS is empty
if [ -z "${CSRF_TRUSTED_ORIGINS:-}" ]; then
  echo "WARNING: CSRF_TRUSTED_ORIGINS is empty. Add your site origins (scheme + host) if necessary." >&2
fi

# All checks passed; run the requested command
exec "$@"
