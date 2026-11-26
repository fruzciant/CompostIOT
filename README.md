# CompostIOT — Sistema IoT de Monitoreo de Compostaje (SENA)

Prototipo funcional y de bajo costo para el monitoreo automatizado de pilas de compostaje orientado a aprendices del área ambiental del **SENA** y proyectos productivos en campo.

## Arquitectura del Monorepo

```text
CompostIOT/
├── specs/            # Contratos técnicos (API REST, fases del compostaje, trama LoRa)
├── backend/          # Servidor REST en Go + SQLite (modo nube y localhost)
├── web/              # Frontend Responsive en React + TypeScript + Vite (compartido)
├── desktop/          # Cliente nativo Windows/Linux/macOS en Go (Wails) + Keyring + SQLite offline
├── mobile/           # App móvil Android/iOS (Capacitor) adaptada a teléfono y tablet
├── firmware/         # Firmware ESP32 (LILYGO TTGO LoRa32 V2.1 SX1278 433 MHz)
├── hardware-cad/     # Diseño 3D paramétrico (FreeCAD .FCStd/.STEP/.STL) para intemperie
└── docs-typst/       # Guías de aprendizaje SENA y manual tipo LEGO (@preview/obelisk:0.2.0)
```

## Variables Monitoreadas

- **Temperatura (°C)** — Obligatoria
- **Humedad (%)** — Obligatoria
- **pH** — Opcional (detección/configuración por pila)

## Especificaciones de Fase 0

- [Contrato de API REST](specs/api-contract.md)
- [Modelo de Fases del Compostaje y Alertas](specs/composting-phases.md)
- [Protocolo LoRa 433 MHz y Hardware](specs/lora-protocol.md)

## Documentación y Guías (Typst)

- [Manual Visual de Ensamble estilo LEGO (PDF)](docs-typst/pdf/01_manual_ensamble_lego.pdf)
- [Guía de Aprendizaje SENA (PDF)](docs-typst/pdf/02_guia_aprendizaje_sena.pdf)
- [Guía de Plataforma: Backend, Web, Wails y Móvil (PDF)](docs-typst/pdf/03_guia_plataforma_software.pdf)

## Licencia

Distribuido bajo licencia **MIT**. Consulta el archivo [LICENSE](LICENSE) para más detalles.
