#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}"
mkdir -p pdf

for src in 01_manual_ensamble_lego.typ 02_guia_aprendizaje_sena.typ 03_guia_plataforma_software.typ; do
  out="pdf/${src%.typ}.pdf"
  echo "Compilando ${src} -> ${out}..."
  typst compile "${src}" "${out}"
done

ls -lh pdf/
