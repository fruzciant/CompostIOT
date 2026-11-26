# Despliegue de CompostIOT (Localhost todo-en-uno y Nube con HTTPS)

## 1. Modo Localhost sin Docker ni Vite (un solo comando)

El servidor Go detecta automáticamente `web/dist` (o la carpeta indicada con `-web-dir`) y sirve tanto la API REST como la interfaz web en `http://localhost:8080`:

```bash
cd backend
./bin/compostiot-server -port 8080 -db ./compostiot.db -web-dir ../web/dist -seed-demo
```

Y si conectas el **TTGO LoRa32 Receptor** por USB a un computador de aula sin WiFi, en otra terminal ejecuta el puente serial:

```bash
cd backend
./bin/compostiot-bridge -list
./bin/compostiot-bridge -port /dev/ttyUSB0 -url http://localhost:8080 -key <X-Device-Key>
```

## 2. Despliegue en Nube (Podman o Docker) para Laboratorios Remotos

1. Copia el archivo de variables y define un secreto JWT seguro:
   ```bash
   cd deploy
   cp .env.example .env
   ```
2. Levanta el servidor (HTTP en puerto `8080`):
   ```bash
   podman compose -f compose.yml up -d --build
   # o con Docker:
   docker compose -f compose.yml up -d --build
   ```
3. Con dominio propio y **HTTPS automático (Let's Encrypt via Caddy)**:
   ```bash
   docker compose -f compose.yml --profile https up -d --build
   ```
   Caddy incluye `flush_interval -1` para que las lecturas en tiempo real por SSE (`/api/v1/events`) lleguen al instante sin retardo.
