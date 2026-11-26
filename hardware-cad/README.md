# CompostIOT — Diseño Mecánico / CAD 3D (FreeCAD)

Carcasas 100 % paramétricas diseñadas para las placas **LILYGO TTGO LoRa32 V2.1 (433 MHz, OLED 0.96")** y listas para impresión 3D. Todos los parámetros (diámetros de prensaestopas, tolerancias de tornillería y bahía modular) se pueden ajustar rápidamente en los scripts de `scripts/` una vez confirmados los modelos exactos de los sensores.

## 1. Nodo Emisor de Intemperie (`emisor/`)

Diseñado para instalación vertical sobre estaca o poste junto a la pila de compostaje:
- **Cara superior (`+Y`)**: collarín antigoteo para antena SMA 433 MHz (`Ø7.2 mm`).
- **Cara inferior (`-Y`)**: 4 pasos de prensaestopas orientados hacia el suelo para evitar entrada de lluvia por escurrimiento:
  1. **PG9 (`Ø15.5 mm`)**: alimentación 5 V (toma eléctrica cercana).
  2. **PG7 (`Ø12.5 mm`)**: sonda de Temperatura *(obligatoria)*.
  3. **PG7 (`Ø12.5 mm`)**: sonda de Humedad *(obligatoria)*.
  4. **PG7 (`Ø12.5 mm`)**: sonda de pH *(opcional — incluye tapón ciego imprimible `Emisor_Tapon_pH.stl`)*.
- **Tapa frontal (`Emisor_Tapa.stl`)**: visor OLED con rebaje interno (`34 × 22 mm`) para pegar ventana de acrílico/policarbonato transparente de 2 mm, lengüeta de compresión perimetral y visera superior contra sol y lluvia.
- **Sellado (`Emisor_Empaque_TPU.stl`)**: empaque perimetral imprimible en **TPU 95A** (o sustituible por cordón de silicona de 2 mm).
- **Bahía modular (`Emisor_Soporte_18650.stl`)**: bandeja intercambiable para bornera 5 V hoy o portapilas **1×18650** en el futuro.
- **Fijación a campo**: orejas laterales con 4 agujeros M5 y ranuras para amarres plásticos/abrazaderas.

### Archivos exportados (`emisor/`)
- `CompostIOT_NodoEmisor.FCStd` — Archivo nativo editable en FreeCAD.
- `CompostIOT_NodoEmisor_Imprimible.step` — Ensamble STEP de las 5 piezas imprimibles.
- `Emisor_Base.stl`, `Emisor_Tapa.stl`, `Emisor_Empaque_TPU.stl`, `Emisor_Tapon_pH.stl`, `Emisor_Soporte_18650.stl`.

## 2. Nodo Receptor / Gateway (`receptor/`)

Carcasa compacta para aula, laboratorio u oficina (escritorio o montaje en pared con ranuras tipo cerradura):
- Ventana alineada con la pantalla OLED 0.96" y orificio de acceso al botón Reset.
- Apertura lateral para puerto Micro-USB (energía y puente serial `localhost`) y ranura microSD.
- Collarín superior para antena SMA 433 MHz y ranuras inferiores de ventilación.

### Archivos exportados (`receptor/`)
- `CompostIOT_NodoReceptor.FCStd` — Archivo nativo editable en FreeCAD.
- `CompostIOT_NodoReceptor_Imprimible.step` — Ensamble STEP imprimible.
- `Receptor_Base.stl`, `Receptor_Tapa.stl`.

## 3. Recomendaciones de Impresión 3D y Ferretería

| Pieza | Material sugerido | Capa | Perímetros | Relleno | Soportes |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `Emisor_Base`, `Emisor_Tapa`, `Emisor_Soporte_18650`, `Emisor_Tapon_pH` | **PETG** o **ASA** (resistente a UV y calor) | 0.20 mm | 3–4 | 25 % giroide | Solo mínimos en visera/orejas si el puente supera 12 mm |
| `Emisor_Empaque_TPU` | **TPU 95A** | 0.20 mm | 2 | 100 % | No |
| `Receptor_Base`, `Receptor_Tapa` | **PLA+** o **PETG** | 0.20 mm | 3 | 20 % | No |

### Tornillería y accesorios estándar (BOM mecánico)
- **Tornillos tapa Emisor**: 4 × M3×16 mm (acero inoxidable recomendado para intemperie).
- **Tornillos tapa Receptor**: 4 × M3×12 mm.
- **Tornillos PCB TTGO y bandeja modular**: 4 × M2×6 mm + 2 × M3×8 mm.
- **Prensaestopas Emisor**: 1 × PG9 (5 V) + 2 o 3 × PG7 (sondas de temperatura, humedad y pH opcional).
- **Ventana transparente**: 1 × lámina de acrílico o policarbonato transparente de `33.5 × 21.5 × 2 mm` sellada con silicona neutra.

## 4. Renders para Manual de Ensamble estilo LEGO (`renders/`)
- `emisor_ensamblado.png` y `emisor_explosionado.png`
- `receptor_ensamblado.png` y `receptor_explosionado.png`
