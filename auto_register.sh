#!/usr/bin/env bash
set -euo pipefail

# ─── CONFIGURE THESE ──────────────────────────────────────────────────────────
# The Vault address to point at
export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"

# Path to your plugin’s main package, relative to the repo root
MODULE_PATH="cmd/db"

# Where Vault will look for your plugin binary
PLUGIN_DIR="vault/plugins"

# The name Vault will refer to this plugin as
PLUGIN_NAME="db2"
# ─────────────────────────────────────────────────────────────────────────────

# Build it!
echo "⛏  Building plugin ${PLUGIN_NAME} from ${MODULE_PATH}…"
go build -o "${PLUGIN_DIR}/${PLUGIN_NAME}" "./${MODULE_PATH}"

# Compute its SHA256
SHA256=$(shasum -a 256 "${PLUGIN_DIR}/${PLUGIN_NAME}" | cut -d ' ' -f 1)
echo "🔑 SHA256 = ${SHA256}"

# (Re-)register it with Vault
echo "🔌 Registering plugin with Vault…"
vault plugin register \
  -sha256="${SHA256}" \
  secret \
  "${PLUGIN_NAME}"

# Reload so Vault picks up the new code immediately
echo "♻️  Reloading plugin…"
vault plugin reload \
  -type=secret \
  -plugin="${PLUGIN_NAME}" \
  -scope=global

echo "✅ Plugin ${PLUGIN_NAME} built, registered, and reloaded!"