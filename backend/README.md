# CompostIOT Backend (Go + SQLite)

Servidor REST liviano de binario único, diseñado para correr tanto en `localhost` (PC del laboratorio o Raspberry Pi sin internet) como en un servidor en la nube.

## Arranque rápido (Localhost con datos de demostración)

```bash
cd backend
go run ./cmd/server -seed-demo
```

El servidor quedará escuchando en `http://localhost:8080`.

### Credenciales iniciales creadas automáticamente
- **Instructor / Administrador (`admin`)**: `instructor` / `compost2026`
- **Aprendiz (`aprendiz`)**: `aprendiz` / `aprendiz2026`
- **Clave Gateway Demo (`-seed-demo`)**: `ciot_demo_gateway_sena_2026`

## Simulador de lecturas LoRa (sin hardware físico todavía)

En otra terminal:

```bash
cd backend
go run ./cmd/simulator -url http://localhost:8080 -node NODO-01 -interval 3s -count 10
```

## Compilar binario único portable

```bash
cd backend
go build -o bin/compostiot-server ./cmd/server
./bin/compostiot-server -port 8080 -db ./compostiot.db -seed-demo
```

## Variables de entorno (opcionales para nube)

| Variable | Defecto | Descripción |
| :--- | :--- | :--- |
| `COMPOST_PORT` | `8080` | Puerto HTTP |
| `COMPOST_DB_PATH` | `./compostiot.db` | Ruta de la base de datos SQLite |
| `COMPOST_JWT_SECRET` | *(local dev)* | Secreto HS256 para firmar tokens Bearer |
| `COMPOST_SERVER_NAME` | `CompostIOT Localhost (SENA)` | Nombre visible en `/api/v1/health` |
| `COMPOST_SEED_DEMO` | `false` | Cargar pilas y curvas de ejemplo si la BD está vacía |

## Ejecutar pruebas

```bash
cd backend
go test -v ./...
```
