#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RUNTIME_DIR="$(cd -- "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="${ENV_FILE:-$RUNTIME_DIR/.env}"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Missing env file: $ENV_FILE" >&2
  exit 1
fi

CP=http://localhost:8000
RT=http://localhost:8081
INTERNAL_KEY=$(grep -E '^INTERNAL_API_KEY=' "$ENV_FILE" | cut -d= -f2-)

echo "== Health =="
curl -sf "$CP/docs" >/dev/null && echo "control-plane docs OK"
curl -sf "$RT/health" | grep -q ok && echo "runtime health OK"
curl -sf "$RT/openapi.yaml" >/dev/null && echo "runtime openapi OK"

echo "== Developer auth (control-plane) =="
EMAIL="dev$(date +%s)@example.com"
curl -sf -X POST "$CP/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"secret123\",\"name\":\"Dev\"}" >/dev/null

LOGIN=$(curl -sf -X POST "$CP/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"secret123\"}")
DEV_TOKEN=$(echo "$LOGIN" | jq -r .access_token)
echo "developer token ok"

echo "== Create project =="
PROJ=$(curl -sf -X POST "$CP/api/v1/projects" \
  -H "Authorization: Bearer $DEV_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Shop","slug":"shop-'$(date +%s)'"}')
PROJECT_ID=$(echo "$PROJ" | jq -r .id)
echo "project=$PROJECT_ID"

echo "== Bootstrap schema =="
curl -sf -X POST "$RT/internal/bootstrap/$PROJECT_ID" \
  -H "X-Internal-Key: $INTERNAL_KEY" | jq .

echo "== End-user signup/signin =="
curl -sf -X POST "$RT/v1/auth/signup" \
  -H "X-Project-Id: $PROJECT_ID" \
  -H "Content-Type: application/json" \
  -d '{"email":"a@example.com","password":"secret123","fullName":"A"}' | jq -r .user.id

SIGNIN=$(curl -sf -X POST "$RT/v1/auth/signin" \
  -H "X-Project-Id: $PROJECT_ID" \
  -H "Content-Type: application/json" \
  -d '{"email":"a@example.com","password":"secret123"}')
TOKEN_A=$(echo "$SIGNIN" | jq -r .tokens.accessToken)

curl -sf "$RT/v1/auth/me" \
  -H "X-Project-Id: $PROJECT_ID" \
  -H "Authorization: Bearer $TOKEN_A" | jq -r .email

echo "== ALL SMOKE TESTS PASSED =="
