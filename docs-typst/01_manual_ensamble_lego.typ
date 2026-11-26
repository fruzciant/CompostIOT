#import "common.typ": *

#show: compost-doc

= Manual de Ensamble CompostIOT

#cover-meta(
  code: "CIOT-MAN-01",
  subtitle: "Guía visual paso a paso estilo LEGO para imprimir, armar, sellar e instalar el Nodo Emisor de intemperie y el Nodo Receptor (LILYGO TTGO LoRa32 V2.1).",
  audience: "Aprendices tecnólogos del área ambiental, instructores SENA y equipo técnico de campo.",
  version: "v1.0 (Preliminar CAD)",
)

Este manual permite ensamblar desde cero los dos módulos físicos de *CompostIOT* sin experiencia previa en electrónica avanzada. #sidenote[Cada paso indica exactamente qué piezas tomar, qué herramienta usar y cómo verificar el montaje antes de avanzar.]

== Inventario visual de piezas (BOM)

Antes de comenzar, extiende todas las piezas sobre una mesa limpia y verifica que tu kit esté completo.

#card-block(breakable: true)[
  #set text(font: "Inter", size: 8pt)
  #table(
    columns: (auto, 1fr, auto, auto),
    inset: 5.5pt,
    stroke: 0.5pt + sena-border,
    fill: (_, row) => if row == 0 { sena-green } else if calc.odd(row) { rgb("#F8FAF9") } else { white },
    table.header(
      text(fill: white, weight: "bold")[Cant.],
      text(fill: white, weight: "bold")[Pieza / Componente],
      text(fill: white, weight: "bold")[Material / Tipo],
      text(fill: white, weight: "bold")[Módulo],
    ),
    [2×], [*LILYGO TTGO LoRa32 V2.1* (ESP32, OLED 0.96", SX1278 433 MHz + antena SMA)], [Electrónica], [Ambos],
    [1×], [`Emisor_Base.stl` — Caja principal con orejas y 4 pasos inferiores], [PETG / ASA], [Emisor],
    [1×], [`Emisor_Tapa.stl` — Tapa con visera, ventana OLED y grabado SENA], [PETG / ASA], [Emisor],
    [1×], [`Emisor_Empaque_TPU.stl` — Empaque perimetral sellante], [TPU 95A / Silicona], [Emisor],
    [1×], [`Emisor_Soporte_18650.stl` — Bandeja modular 5 V / futura 18650], [PETG / ASA], [Emisor],
    [1×], [`Emisor_Tapon_pH.stl` — Tapón ciego si no se usa sonda de pH], [PETG / ASA], [Emisor],
    [1×], [`Receptor_Base.stl` + `Receptor_Tapa.stl` — Carcasa de aula], [PLA+ / PETG], [Receptor],
    [1×], [Lámina transparente acrílico/policarbonato (`33.5 × 21.5 × 2 mm`)], [Acrílico + silicona], [Emisor],
    [1× / 3×], [Prensaestopas: 1× *PG9* (5 V) + 2 o 3× *PG7* (sondas)], [Nylon IP68], [Emisor],
    [14×], [Tornillos inox: 4× M3×16, 4× M3×12, 2× M3×8, 4× M2×6 mm], [Acero inox], [Ambos],
    [1×], [Fuente regulada 5 V DC (toma cercana) + cable encauchetado], [Alimentación], [Emisor],
    [2 o 3], [Sensores: *Temperatura* (oblig.), *Humedad* (oblig.), *pH* (opcional)], [Sondas campo], [Emisor],
  )
]

#callout(
  kind: "info",
  title: "Nota sobre sensores (actualizable en Fase 5)",
)[
  La carcasa reserva *3 puertos PG7 independientes* para Temperatura, Humedad y pH opcional. Cuando confirmes los modelos exactos de tus sensores, añadiremos aquí la tabla de pines por color de cable y el diagrama eléctrico exacto.
]

== Vista explosionada — Nodo Emisor (Intemperie)

Observa cómo encajan las 5 piezas impresas, la placa TTGO LoRa32 V2.1, la antena 433 MHz y la bandeja modular inferior (lista hoy para bornera 5 V y preparada a futuro para celda 18650).

#card-block[
  #align(center)[
    #image("assets/emisor_explosionado.png", width: 88%)
    #v(2pt)
    #text(font: "Inter", size: 8pt, fill: ink-muted)[*Figura 1.* Vista explosionada del Nodo Emisor CompostIOT con empaque TPU, visor OLED, tapón opcional de pH y bahía modular.]
  ]
]

== Ensamble paso a paso — Nodo Emisor

#lego-step(
  num: 1,
  title: "Instalar prensaestopas en la cara inferior",
  parts: ("1× Emisor_Base", "1× PG9 (5 V)", "2× PG7 (Temp/Hum)", "1× PG7 o Emisor_Tapon_pH"),
  tools: "Llave fija o mano firme",
  check: "Los 4 puertos quedan apuntando hacia el suelo y sus contratuercas internas firmes.",
)[
  Coloca `Emisor_Base` con los 4 agujeros mirando hacia ti e inserta de izquierda a derecha: (1) *PG9* para alimentación 5 V; (2) *PG7* para Temperatura; (3) *PG7* para Humedad; (4) *PG7* para pH o, si no usas pH, presiona el tapón `Emisor_Tapon_pH` con una gota de silicona neutra.
]

#lego-step(
  num: 2,
  title: "Fijar la bandeja modular inferior (5 V hoy / 18650 futuro)",
  parts: ("1× Emisor_Soporte_18650", "2× Tornillos M3×8 mm"),
  tools: "Destornillador / Allen",
  check: "La bandeja queda nivelada sobre los dos postes inferiores sin tocar los prensaestopas.",
)[
  Alinea los dos agujeros de `Emisor_Soporte_18650` sobre los postes internos inferiores de `Emisor_Base` y atornilla con M3×8 mm. #sidenote[Esta misma bandeja aloja hoy la bornera de 5 V y permite migrar luego a portapilas 1×18650 sin reimprimir.]
]

#lego-step(
  num: 3,
  title: "Montar la placa LILYGO TTGO LoRa32 V2.1 y su antena",
  parts: ("1× TTGO LoRa32 V2.1", "1× Antena SMA 433 MHz", "4× Tornillos M2×6 mm"),
  tools: "Destornillador pequeño",
  check: "La pantalla OLED mira hacia arriba y la antena sale vertical por el collarín superior.",
)[
  Apoya la TTGO LoRa32 V2.1 sobre los 4 postes altos superiores y entre los rieles guía. Fíjala con los 4 tornillos M2 sin apretar de más. Enrosca la antena *433 MHz* a través del collarín superior antigoteo.
]

#callout(
  kind: "warning",
  title: "¡Nunca energices la placa LoRa sin la antena conectada!",
)[
  Transmitir con el chip SX1278 (433 MHz) sin la antena enroscada puede dañar el amplificador de radiofrecuencia. Confirma siempre la antena antes de aplicar 5 V.
]

#lego-step(
  num: 4,
  title: "Pasar cables, conectar sensores y sellar prensaestopas",
  parts: ("Cable 5 V", "Sondas de Temperatura, Humedad y (opcional) pH"),
  tools: "Pelacables y perillero",
  check: "Al halar suavemente cada cable hacia afuera, el prensaestopas lo retiene firmemente.",
)[
  Introduce cada cable desde abajo dejando 5 cm de holgura interna, realiza las conexiones al PCB/bornera y aprieta la tuerca exterior de cada prensaestopas hasta comprimir el sello interno de goma.
]

#lego-step(
  num: 5,
  title: "Sellar la ventana acrílica en la tapa frontal",
  parts: ("1× Emisor_Tapa", "1× Acrílico 33.5×21.5×2 mm", "Silicona neutra"),
  tools: "Aplicador limpio",
  check: "El acrílico queda al ras dentro del rebaje trasero de la tapa, limpio frente al OLED.",
)[
  Voltea `Emisor_Tapa` boca abajo, aplica un hilo fino de silicona neutra en el escalón interno alrededor de la ventana OLED, presiona la lámina acrílica y deja curar 30 minutos.
]

#lego-step(
  num: 6,
  title: "Colocar el empaque TPU y cerrar con 4 tornillos M3",
  parts: ("1× Emisor_Empaque_TPU", "1× Emisor_Tapa", "4× M3×16 mm inox"),
  tools: "Llave Allen",
  check: "La visera queda en el borde superior (lado antena) y el empaque bien asentado en su canal.",
)[
  Encaja `Emisor_Empaque_TPU` en la ranura perimetral de `Emisor_Base`, coloca `Emisor_Tapa` con la visera hacia arriba y aprieta los 4 tornillos M3 en cruz (X).
  #v(5pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 8pt,
    align: center,
    [#image("assets/emisor_tapa_top.png", width: 98%)],
    [#image("assets/emisor_ensamblado.png", width: 98%)],
  )
]

#lego-step(
  num: 7,
  title: "Fijar en estaca junto a la pila de compostaje",
  parts: ("Nodo Emisor cerrado", "2× Amarres plásticos o 4× tornillos M5", "Estaca o tubo PVC"),
  tools: "Alicate de corte",
  check: "Caja vertical, antena al cielo y cables formando curva de goteo (U) hacia abajo.",
)[
  Sujeta el Nodo Emisor a una estaca firme a 40–60 cm del suelo usando las orejas laterales. Forma una *curva de goteo* en todos los cables antes de entrar a los prensaestopas para impedir que el agua escurra hacia la caja.
]

== Ensamble del Nodo Receptor (Gateway de Aula)

El segundo TTGO LoRa32 V2.1 recibe los datos de campo a 433 MHz y los envía al backend por WiFi o por cable USB hacia `localhost`.

#card-block[
  #align(center)[
    #image("assets/receptor_explosionado.png", width: 82%)
    #v(2pt)
    #text(font: "Inter", size: 8pt, fill: ink-muted)[*Figura 2.* Vista explosionada del Nodo Receptor / Gateway con acceso Micro-USB/microSD y grabado SENA.]
  ]
]

#lego-step(
  num: 8,
  title: "Armar y cerrar el Nodo Receptor / Gateway",
  parts: ("1× Receptor_Base", "1× Receptor_Tapa", "1× TTGO LoRa32 V2.1", "1× Antena 433 MHz", "4× M3×12 mm"),
  tools: "Destornillador",
  check: "Puerto Micro-USB accesible por el costado derecho y pantalla OLED visible en la ventana superior.",
)[
  Apoya la placa TTGO en los 4 postes internos de `Receptor_Base` con el puerto Micro-USB hacia el lado derecho, enrosca la antena 433 MHz, encaja `Receptor_Tapa` y asegura las 4 esquinas con M3×12 mm. Conecta Micro-USB a 5 V (WiFi) o al PC del laboratorio (`localhost`).
  #v(5pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 8pt,
    align: center,
    [#image("assets/receptor_tapa_top.png", width: 98%)],
    [#image("assets/receptor_ensamblado.png", width: 98%)],
  )
]

== Checklist rápido pre-despliegue

#card-block[
  #set text(font: "Inter", size: 9pt)
  - $square$ Ambas antenas de *433 MHz* están enroscadas antes de energizar.
  - $square$ Las pantallas OLED muestran temperatura, humedad y señal LoRa (`RSSI`).
  - $square$ El Emisor está vertical con prensaestopas abajo y tapón puesto si no hay sonda de pH.
  - $square$ El Receptor reporta lecturas válidas visibles en el tablero de CompostIOT.
]
