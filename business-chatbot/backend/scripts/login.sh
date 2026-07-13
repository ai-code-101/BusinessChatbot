#!/bin/bash
# Logs in as the admin and saves the JWT to .token, so other scripts/curl
# commands can reuse it without you copy-pasting it manually every time.
#
# Usage (run from the backend/ folder):
#   ./scripts/login.sh
#   curl -H "Authorization: Bearer $(cat .token)" http://localhost:8080/v1/admin/documents

set -e

# Pull ADMIN_EMAIL / ADMIN_PASSWORD straight from your .env file, so you
# never have to type credentials or set env vars manually.
if [ -f .env ]; then
  export $(grep -E '^ADMIN_EMAIL=|^ADMIN_PASSWORD=' .env | xargs)
fi

API_URL="${PORT:+http://localhost:$PORT}"
API_URL="${API_URL:-http://localhost:8080}"

if [ -z "$ADMIN_EMAIL" ] || [ -z "$ADMIN_PASSWORD" ]; then
  echo "Could not find ADMIN_EMAIL / ADMIN_PASSWORD in .env"
  exit 1
fi

RESPONSE=$(curl -s -X POST "$API_URL/v1/admin/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}")

TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Login failed. Response was:"
  echo "$RESPONSE"
  exit 1
fi

echo "$TOKEN" > .token
echo "Logged in as $ADMIN_EMAIL. Token saved to .token"