#import "common.typ": *

#show: compost-doc

= Guía de Plataforma CompostIOT

#cover-meta(
  code: "CIOT-SW-03",
  subtitle: "Manual operativo para desplegar el servidor local o en nube, configurar múltiples servidores, trabajar sin internet en Wails y usar la app móvil.",
  audience: "Instructores administradores, soporte técnico del centro de formación y aprendices SENA.",
  version: "v1.0",
)

== Dos modos de operación: Localhost y Nube

*CompostIOT* funciona tanto en aulas y fincas *sin acceso a internet* como en un servidor institucional en la nube, ejecutando exactamente el mismo binario compilado en Go + SQLite. #sidenote[Todos los clientes (Web, Wails y Móvil) guardan *múltiples perfiles de servidor* para alternar entre localhost y la nube sin reinstalar.]

#card-block[
  #set text(font: "Inter", size: 8.2pt)
  #table(
    columns: (auto, 1fr, 1fr),
    inset: 5.5pt,
    stroke: 0.5pt + sena-border,
    fill: (_, row) => if row == 0 { sena-green } else if calc.odd(row) { rgb("#F8FAF9") } else { white },
    table.header(
      text(fill: white, weight: "bold")[Escenario],
      text(fill: white, weight: "bold")[Cómo se conecta],
      text(fill: white, weight: "bold")[Ventaja principal],
    ),
    [*Localhost / Red LAN*], [Ejecutar `compostiot-server` en un PC o Raspberry Pi (`http://localhost:8080` o `http://192.168.x.x:8080`).], [Cero dependencia de internet; ideal para zonas rurales y prácticas en aula.],
    [*Nube institucional*], [Desplegar el mismo binario tras HTTPS (`https://compost.centro.edu.co`).], [Habilita laboratorios remotos y monitoreo desde cualquier lugar.],
  )
]

== Arranque rápido del servidor Backend

Desde una terminal en la carpeta `backend/` del repositorio:

#card-block[
  #set text(font: "DejaVu Sans Mono", size: 8.2pt)
  #box(width: 100%, fill: rgb("#12261C"), radius: 4pt, inset: 8pt)[
    #text(fill: rgb("#A7F3D0"))[\$ ./bin/compostiot-server -port 8080 -db ./compostiot.db -seed-demo]
  ]
]

En una base de datos nueva se crean dos cuentas iniciales (modificables desde *Ajustes*):
- *Instructor / Administrador (`admin`)*: `instructor` / `compost2026`.
- *Aprendiz (`aprendiz`)*: `aprendiz` / `aprendiz2026`.

== Cliente nativo de escritorio en Go (Wails)

El ejecutable nativo (`desktop/build/bin/CompostIOT`, compilable para Windows, Linux y macOS) integra seguridad y operación offline:

1. *Múltiples servidores*: selecciona *Servidor local (Localhost)* o pulsa *Agregar servidor* para guardar la URL de la nube. Usa *Probar* (`/api/v1/health`) antes de iniciar sesión.
2. *Llavero del sistema operativo*: el backend devuelve un *token Bearer JWT* que el cliente guarda de forma segura en *Windows Credential Manager*, *macOS Keychain* o *Linux Secret Service* (`go-keyring`) y envía en cada petición REST.
3. *Caché offline en SQLite*: cada sincronización respalda pilas y lecturas en `cache.db`. Si falla la red o se apaga el servidor, el cliente activa el aviso *Modo sin conexión* y permite consultar gráficas e incluso *exportar CSV* desde SQLite.

== Aplicación móvil para teléfono y tablet

La interfaz se adapta automáticamente a cualquier pantalla:
- *Teléfono*: navegación inferior fija (*Tablero*, *Guía*, *Ajustes*) y selector horizontal de pilas para operar con una mano en campo.
- *Tablet*: tarjetas en 2 y 3 columnas con rotación vertical/horizontal.
- *Opciones de instalación*: APK nativo Android desde `mobile/android` (con tráfico HTTP LAN habilitado para redes de finca) o instalación directa *PWA* desde Chrome/Safari (*Agregar a la pantalla de inicio*).

== Roles y permisos en la plataforma

#card-block[
  #set text(font: "Inter", size: 8.2pt)
  #table(
    columns: (1fr, auto, auto),
    inset: 5.5pt,
    stroke: 0.5pt + sena-border,
    fill: (_, row) => if row == 0 { sena-green } else if calc.odd(row) { rgb("#F8FAF9") } else { white },
    table.header(
      text(fill: white, weight: "bold")[Acción en CompostIOT],
      text(fill: white, weight: "bold")[Instructor (`admin`)],
      text(fill: white, weight: "bold")[Aprendiz (`aprendiz`)],
    ),
    [Ver pilas, variables en tiempo real, fase y alertas], [Sí], [Sí],
    [Analizar gráficas históricas y consultar Guía de fases], [Sí], [Sí],
    [Exportar históricos a CSV (Excel / Python / R)], [Sí], [Sí],
    [Crear, editar, archivar o eliminar pilas], [Sí], [No],
    [Registrar mediciones manuales de calibración], [Sí], [No],
    [Generar claves `X-Device-Key` para gateways ESP32], [Sí], [No],
    [Crear o eliminar usuarios], [Sí], [No],
  )
]

#callout(
  kind: "info",
  title: "Enlazar un nuevo Gateway ESP32 LoRa32",
)[
  Inicia sesión como *Instructor*, abre *Ajustes*, escribe el nombre del receptor (ej. *TTGO Aula Ambiental*) y pulsa *Generar clave*. Copia la clave `ciot_...` (se muestra una única vez) para configurarla en el firmware del Gateway como cabecera `X-Device-Key`.
]
