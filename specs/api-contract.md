# Contrato de API REST — CompostIOT v1

Base URL configurable por el cliente (ej. `http://localhost:8080` o `https://compost.ejemplo.edu.co`).
Todos los endpoints de la API se versionan bajo el prefijo `/api/v1`.

## 1. Autenticación y Roles

### Roles de usuario
| Rol | Descripción | Permisos |
| :--- | :--- | :--- |
| `admin` | Instructor / Administrador | CRUD de pilas, nodos/gateways, usuarios, configuración y consulta/exportación total |
| `aprendiz` | Aprendiz SENA / Operador | Consulta de pilas, lectura de métricas, análisis de fases, anotaciones opcionales y exportación CSV/JSON |

### Esquemas de autenticación
1. **Clientes de usuario (Web, Wails, Móvil)**:
   - `POST /api/v1/auth/login` con `username` y `password`.
   - Devuelve `access_token` (JWT firmado HS256) que el cliente envía en cada petición:
     `Authorization: Bearer <token>`
   - En **Wails** el token se guarda en el **llavero del sistema operativo** (`go-keyring`) por perfil de servidor.
2. **Nodos Gateway (ESP32 Receptor)**:
   - Envía una clave de dispositivo en la cabecera `X-Device-Key: <api_key>` al endpoint de ingesta `POST /api/v1/ingest/readings`.

---

## 2. Endpoints

### Salud e información del servidor (público)
- `GET /api/v1/health`
  - Permite a los clientes (Web/Wails/Móvil) validar la URL de un servidor antes de iniciar sesión.
  - Respuesta `200 OK`:
    ```json
    {
      "status": "ok",
      "service": "compostiot-backend",
      "version": "1.0.0",
      "server_name": "CompostIOT Localhost",
      "server_time": "2026-11-26T14:00:00-05:00"
    }
    ```

### Autenticación
- `POST /api/v1/auth/login`
  - Request:
    ```json
    { "username": "instructor", "password": "secret" }
    ```
  - Response `200 OK`:
    ```json
    {
      "access_token": "eyJhbGciOi...",
      "token_type": "Bearer",
      "expires_in": 86400,
      "user": {
        "id": 1,
        "username": "instructor",
        "full_name": "Instructor Ambiental",
        "role": "admin"
      }
    }
    ```
- `GET /api/v1/auth/me` *(requiere Bearer)* — devuelve el perfil del usuario autenticado.
- `POST /api/v1/auth/change-password` *(requiere Bearer)* — cambia la contraseña propia.

### Usuarios (solo `admin`)
- `GET /api/v1/users` — lista usuarios.
- `POST /api/v1/users` — crea un usuario (`username`, `full_name`, `password`, `role`).
- `PATCH /api/v1/users/:id` — actualiza nombre, rol, estado activo o resetea contraseña.
- `DELETE /api/v1/users/:id` — desactiva o elimina usuario (no permite auto-eliminarse).

### Pilas de Compost (`piles`)
- `GET /api/v1/piles?status=active|archived|all` *(Bearer)*
  - Devuelve listado de pilas junto con su **última lectura**, **fase calculada actual** y **alertas activas**.
- `GET /api/v1/piles/:id` *(Bearer)*
  - Detalle completo de una pila, configuración (si incluye sensor de pH o no), resumen de fase actual, días transcurridos y recomendaciones.
- `POST /api/v1/piles` *(solo `admin`)*
  - Request:
    ```json
    {
      "name": "Pila #1 — Residuos Orgánicos Cafetería",
      "location": "Invernadero Norte - Lote A",
      "substrate": "Restos vegetales + aserrín + estiércol bovino",
      "node_id": "NODO-01",
      "has_ph_sensor": false,
      "started_at": "2026-11-20T08:00:00-05:00",
      "notes": "Relación C/N inicial estimada 30:1"
    }
    ```
- `PUT /api/v1/piles/:id` *(solo `admin`)* — edita metadatos, `node_id`, `has_ph_sensor` o `status` (`active` / `archived`).
- `DELETE /api/v1/piles/:id` *(solo `admin`)* — elimina una pila y sus lecturas asociadas.

### Lecturas e Histórico (`readings`)
- `GET /api/v1/piles/:id/readings?from=ISO8601&to=ISO8601&limit=500` *(Bearer)*
  - Devuelve lecturas ordenadas cronológicamente para graficar curvas térmicas, humedad y pH.
  - Cada registro incluye: `id`, `pile_id`, `temperature_c`, `humidity_pct`, `ph` (`null` si no aplica), `phase`, `battery_v` (`null` si alimentado por red), `power_source` (`mains` | `battery`), `rssi`, `recorded_at`.
- `POST /api/v1/piles/:id/readings` *(solo `admin`)*
  - Permite registrar una medición manual o de calibración/prueba desde la interfaz.
- `GET /api/v1/piles/:id/export?format=csv&from=...&to=...` *(Bearer)*
  - Descarga archivo CSV listo para Excel/LibreOffice/R/Python con encabezados claros en español y metadatos de fase.

### Ingesta desde Gateway ESP32 (`ingest`)
- `POST /api/v1/ingest/readings` *(cabecera `X-Device-Key`)*
  - Request enviado por el Nodo Receptor TTGO LoRa32:
    ```json
    {
      "node_id": "NODO-01",
      "temperature_c": 58.4,
      "humidity_pct": 54.2,
      "ph": null,
      "power_source": "mains",
      "battery_v": null,
      "rssi": -78,
      "seq": 142
    }
    ```
  - El backend busca la pila activa asociada a `node_id`, valida rangos físicos, calcula la fase actual con el motor de fases, guarda la lectura y responde `201 Created` con la fase resultante.

### Gestión de Claves de Gateway (`devices`, solo `admin`)
- `GET /api/v1/devices` — lista gateways registrados y última actividad (`last_seen_at`).
- `POST /api/v1/devices` — genera un nuevo gateway y muestra su `api_key` una única vez.
- `DELETE /api/v1/devices/:id` — revoca un gateway.

### Sincronización Offline (para Wails y App Móvil)
- `GET /api/v1/sync/snapshot?since=ISO8601` *(Bearer)*
  - Endpoint optimizado que devuelve en una sola respuesta todas las pilas actualizadas y las lecturas nuevas desde `since`, facilitando la actualización rápida del caché local SQLite.

---

## 3. Estándar de Errores

Todas las respuestas de error siguen este formato uniforme:
```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Usuario o contraseña incorrectos"
  }
}
```
Códigos HTTP estándar: `400` (validación), `401` (sin autenticación / token vencido), `403` (sin permiso de rol), `404` (no encontrado), `409` (conflicto, ej. `node_id` duplicado en pila activa).
