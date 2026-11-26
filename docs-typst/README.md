# CompostIOT — Documentación y Guías de Aprendizaje (Typst)

Material educativo, técnico y manual visual de ensamble estilo LEGO construidos con **Typst** sobre la plantilla **`@preview/obelisk:0.2.0`**.

## Documentos incluidos

1. **`01_manual_ensamble_lego.typ`** → `pdf/01_manual_ensamble_lego.pdf`
   - Inventario visual de piezas (BOM), vistas explosionadas CAD del Nodo Emisor y Nodo Receptor, y secuencia de armado paso a paso estilo LEGO con checklist pre-despliegue.
2. **`02_guia_aprendizaje_sena.typ`** → `pdf/02_guia_aprendizaje_sena.pdf`
   - Guía formativa para aprendices tecnólogos del área ambiental (SENA): las 4 fases del compostaje, lectura científica de curvas térmicas, matriz de diagnóstico en campo y 3 actividades prácticas (incluyendo laboratorio remoto).
3. **`03_guia_plataforma_software.typ`** → `pdf/03_guia_plataforma_software.pdf`
   - Guía operativa y de despliegue: modo Localhost vs. Nube, cliente nativo Go/Wails (llavero del SO + caché SQLite offline), app móvil teléfono/tablet y permisos por rol.

## Compilar todos los PDFs

```bash
cd docs-typst
bash build-all.sh
```
