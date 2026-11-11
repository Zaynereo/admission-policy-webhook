#!/usr/bin/env bash
# Generate local-only TLS material for webhook development. Private keys are gitignored — never commit them.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${ROOT}/certs"
mkdir -p "${OUT}"
CRT="${OUT}/tls.crt"
KEY="${OUT}/tls.key"

openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout "${KEY}" \
  -out "${CRT}" \
  -days 365 \
  -subj "/CN=admission-policy-webhook.admission-policy-webhook.svc"

chmod 600 "${KEY}"
echo "Wrote ${CRT} and ${KEY}"
echo "CA bundle (for ValidatingWebhookConfiguration.caBundle):"
base64 -w0 "${CRT}" 2>/dev/null || base64 < "${CRT}" | tr -d '\n'
echo
