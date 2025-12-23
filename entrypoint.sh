#!/usr/bin/env bash
set -euo pipefail
shopt -s nocasematch


# SECRET_KEY must be set in production
if [ -z "${SECRET_KEY:-}" ]; then
  echo "WARNING: SECRET_KEY not set — using development fallback (insecure)."
  exit 1
fi

# Warn if ALLOWED_HOSTS is empty
if [ -z "${ALLOWED_HOSTS:-}" ]; then
  echo "WARNING: ALLOWED_HOSTS is empty. Consider setting ALLOWED_HOSTS to a comma-separated list (e.g., 'example.com,127.0.0.1')."
  exit 1
fi

# Validate CSRF_TRUSTED_ORIGINS entries (must start with http:// or https://)
if [ -n "${CSRF_TRUSTED_ORIGINS:-}" ]; then
  IFS=',' read -ra _origins <<< "$CSRF_TRUSTED_ORIGINS"
  bad=()
  for o in "${_origins[@]}"; do
    # trim whitespace
    origin="$(echo "$o" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
    if [ -z "$origin" ]; then
      continue
    fi
    case "$origin" in
      http://*|https://*) ;;
      *)
        bad+=("$origin")
        ;;
    esac
  done
  if [ "${#bad[@]}" -ne 0 ]; then
    echo "ERROR: Invalid CSRF_TRUSTED_ORIGINS entries (must start with http:// or https://): ${bad[*]}" >&2
    exit 1
  fi
else
  echo "WARNING: CSRF_TRUSTED_ORIGINS is empty. Add your site origins (scheme + host) if necessary." >&2
fi

exec "$@"
