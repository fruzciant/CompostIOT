# CompostIOT Web (React + TypeScript + Vite)

Interfaz web responsive (escritorio, tablet y teléfono) pensada para aprendices SENA y productores en campo. Este mismo código fuente alimenta también al cliente nativo de escritorio **Go/Wails** (`desktop/`) y a la app móvil **Capacitor** (`mobile/`).

## Desarrollo local

1. Arranca primero el backend en modo demo:
   ```bash
   cd ../backend && go run ./cmd/server -seed-demo
   ```
2. Inicia el servidor de desarrollo frontend:
   ```bash
   cd ../web
   pnpm dev
   ```
3. Abre `http://localhost:5173`, verifica que el servidor seleccionado sea `http://localhost:8080` e inicia sesión:
   - **Instructor (`admin`)**: `instructor` / `compost2026`
   - **Aprendiz (`aprendiz`)**: `aprendiz` / `aprendiz2026`

## Compilar para producción

```bash
pnpm build
```
Los archivos compilados quedan en `web/dist/`.
