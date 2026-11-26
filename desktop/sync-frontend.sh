#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_DIR="${SCRIPT_DIR}/../web"

echo "[1/2] Compilando frontend web..."
(cd "${WEB_DIR}" && pnpm build)

echo "[2/2] Sincronizando web/dist -> desktop/frontend/dist..."
rm -rf "${SCRIPT_DIR}/frontend/dist"
mkdir -p "${SCRIPT_DIR}/frontend"
cp -r "${WEB_DIR}/dist" "${SCRIPT_DIR}/frontend/dist"
echo "Frontend listo para empaquetar en Wails."
