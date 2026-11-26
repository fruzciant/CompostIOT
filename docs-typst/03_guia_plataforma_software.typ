#import "common.typ": *

#show: compost-doc

= Guía de plataforma

#doc-meta(
  code: "CIOT-SW-03 · v1.1",
  subtitle: "Manual operativo para desplegar el servidor local o en la nube, configurar múltiples servidores, trabajar sin internet en Wails y usar la app móvil.",
  audience: "Instructores administradores, soporte técnico y aprendices SENA.",
)

== Operación dual: Localhost y nube

*CompostIOT* ejecuta el mismo servidor compilado en Go + SQLite tanto en aulas o fincas *sin acceso a internet* como en un servidor institucional en la nube. #sidenote[Todos los clientes (Web, escritorio Wails y móvil) guardan múltiples perfiles de servidor para alternar entre `localhost` y la nube sin reinstalar.]

#editorial-table(
  columns: (auto, 1fr, 1fr),
  headers: ("Escenario", "Modo de conexión", "Ventaja principal"),
  rows: (
    ([Localhost o red LAN], [Ejecutar `compostiot-server` en un PC o Raspberry Pi (`http://localhost:8080` o `http://192.168.x.x:8080`).], [Cero dependencia de internet; ideal para zonas rurales y pruebas de aula.]),
    ([Nube institucional], [Desplegar el mismo binario detrás de HTTPS (`https://compost.centro.edu.co`).], [Habilita laboratorios remotos y consulta desde cualquier sede o ciudad.]),
  ),
)

== Arranque rápido del servidor

Ejecuta el binario desde la carpeta `backend/` del repositorio:

```bash
./bin/compostiot-server -port 8080 -db ./compostiot.db -seed-demo
```

Al inicializarse sobre una base de datos vacía, el servidor crea dos cuentas base que el instructor puede editar o ampliar desde la vista *Ajustes*:

- *Instructor (`admin`)*: usuario `instructor`, contraseña `compost2026`.
- *Aprendiz (`aprendiz`)*: usuario `aprendiz`, contraseña `aprendiz2026`.

== Cliente nativo de escritorio (Go + Wails)

El cliente nativo (`desktop/build/bin/CompostIOT`, compilable para Windows, Linux y macOS) está pensado para trabajar de forma segura incluso en entornos con conectividad intermitente:

1. *Múltiples servidores*: en la pantalla de ingreso selecciona *Servidor local (Localhost)* o pulsa *Agregar servidor* para registrar una instancia remota. El botón *Probar* verifica `/api/v1/health` antes de iniciar sesión.
2. *Llavero del sistema operativo*: tras autenticarse, el backend devuelve un token Bearer JWT que el cliente almacena en *Windows Credential Manager*, *macOS Keychain* o *Linux Secret Service* (`go-keyring`) y envía automáticamente en cada petición REST.
3. *Caché offline en SQLite*: cada sincronización guarda pilas y lecturas en una base local SQLite (`cache.db`). Ante una caída de red o al apagar el servidor local, el cliente muestra el aviso *Modo sin conexión* y permite seguir consultando gráficas y exportando CSV desde el caché local.

== Aplicación móvil para teléfono y tablet

La interfaz se adapta automáticamente al tamaño de pantalla y orientación del dispositivo:

- *Teléfono*: navegación inferior fija (*Tablero*, *Guía*, *Ajustes*) y selector horizontal deslizable de pilas para operar con una mano junto a la compostera.
- *Tablet*: tarjetas amplias en 2 y 3 columnas con soporte de rotación vertical y horizontal.
- *Instalación*: proyecto nativo Android en `mobile/android` (Capacitor 7 con tráfico HTTP LAN habilitado para servidores locales de finca) o instalación directa *PWA* desde Chrome o Safari mediante *Agregar a la pantalla de inicio*.

== Roles y permisos

#editorial-table(
  columns: (1fr, auto, auto),
  headers: ("Acción en CompostIOT", "Instructor (admin)", "Aprendiz"),
  rows: (
    ([Consultar pilas, variables en tiempo real, fase y alertas], [Sí], [Sí]),
    ([Analizar gráficas históricas y consultar la guía de fases], [Sí], [Sí]),
    ([Exportar históricos completos a CSV para Excel, Python o R], [Sí], [Sí]),
    ([Crear, editar, archivar o eliminar pilas de compostaje], [Sí], [No]),
    ([Registrar mediciones manuales de calibración o prueba], [Sí], [No]),
    ([Generar o revocar claves `X-Device-Key` de gateways LoRa], [Sí], [No]),
    ([Crear o eliminar cuentas de usuarios], [Sí], [No]),
  ),
)

#nota[Vincular un nuevo Gateway TTGO LoRa32][
  Inicia sesión como *Instructor*, abre la sección *Ajustes*, asigna un nombre al receptor (por ejemplo, *TTGO Aula Ambiental*) y pulsa *Generar clave*. Copia la clave `ciot_...` que se muestra una sola vez y configúrala en el firmware del Gateway como cabecera `X-Device-Key`.
]
