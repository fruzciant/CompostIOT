#import "common.typ": *

#show: compost-doc

= Manual de ensamble

#doc-meta(
  code: "CIOT-MAN-01 · v1.1",
  subtitle: "Guía paso a paso para imprimir, ensamblar, sellar e instalar el Nodo Emisor de intemperie y el Nodo Receptor basados en LILYGO TTGO LoRa32 V2.1.",
  audience: "Aprendices tecnólogos del área ambiental, instructores SENA y personal de campo.",
)

Este manual describe el armado completo de los dos módulos físicos de *CompostIOT*. En cada paso, la columna lateral izquierda enumera las piezas exactas y la herramienta requerida, al estilo de los manuales visuales de montaje.

== Inventario de piezas

Extiende todos los componentes sobre una mesa limpia antes de empezar y confirma las cantidades del kit.

#editorial-table(
  columns: (auto, 1fr, auto, auto),
  headers: ("Cant.", "Pieza o componente", "Material", "Módulo"),
  rows: (
    ([2×], [Placa *LILYGO TTGO LoRa32 V2.1* (ESP32, OLED 0.96", SX1278 433 MHz y antena SMA)], [Electrónica], [Ambos]),
    ([1×], [`Emisor_Base.stl`: carcasa principal con orejas y 4 pasos inferiores], [PETG / ASA], [Emisor]),
    ([1×], [`Emisor_Tapa.stl`: tapa frontal con visera, visor OLED y grabado SENA], [PETG / ASA], [Emisor]),
    ([1×], [`Emisor_Empaque_TPU.stl`: empaque perimetral de sellado], [TPU 95A], [Emisor]),
    ([1×], [`Emisor_Soporte_18650.stl`: bandeja modular para bornera 5 V o futura 18650], [PETG / ASA], [Emisor]),
    ([1×], [`Emisor_Tapon_pH.stl`: tapón ciego si la pila no lleva sonda de pH], [PETG / ASA], [Emisor]),
    ([2×], [`Receptor_Base.stl` y `Receptor_Tapa.stl`: carcasa de escritorio o pared], [PLA+ / PETG], [Receptor]),
    ([1×], [Lámina transparente de acrílico o policarbonato de `33.5 × 21.5 × 2 mm`], [Acrílico], [Emisor]),
    ([1× / 3×], [Prensaestopas IP68: 1× *PG9* (cable 5 V) y 2 o 3× *PG7* (sondas)], [Nylon], [Emisor]),
    ([14×], [Tornillos inoxidables: 4× M3×16, 4× M3×12, 2× M3×8 y 4× M2×6 mm], [Acero inox], [Ambos]),
    ([1×], [Fuente regulada 5 V DC para toma eléctrica cercana y cable encauchetado], [Energía], [Emisor]),
    ([2 o 3], [Sensores: temperatura (obligatorio), humedad (obligatorio) y pH (opcional)], [Sondas], [Emisor]),
  ),
)

#nota[Sensores definitivos][
  La carcasa deja *tres puertos PG7 independientes* para las sondas de temperatura, humedad y pH opcional. Una vez confirmados los modelos exactos de sensores en la Fase 5, aquí se incorporará la tabla definitiva de pines por color de cable.
]

== Vista general del Nodo Emisor

La @fig:emisor-exp muestra el orden exacto de apilado de las cinco piezas impresas en 3D, la placa TTGO LoRa32 V2.1, la antena de 433 MHz y la bandeja modular inferior.

#clean-figure(
  caption: [Despiece completo del Nodo Emisor de intemperie con empaque TPU, visor OLED, tapón opcional de pH y bahía modular.],
)[
  #image("assets/emisor_explosionado.png", width: 88%)
]<fig:emisor-exp>

== Montaje del Nodo Emisor

#assembly-step(
  num: 1,
  title: "Instalar los prensaestopas en el borde inferior",
  parts: ("1× Emisor_Base", "1× PG9 (5 V)", "2× PG7 (Temp/Hum)", "1× PG7 o Emisor_Tapon_pH"),
  tool: "Llave fija o ajuste manual",
  check: "Los cuatro puertos miran hacia abajo y sus contratuercas internas quedan firmes.",
)[
  Ubica `Emisor_Base` con los cuatro orificios mirando hacia ti e instala de izquierda a derecha: el *PG9* de mayor diámetro para el cable de 5 V, el primer *PG7* para la sonda de temperatura y el segundo *PG7* para la sonda de humedad. En el cuarto orificio monta el tercer *PG7* si vas a medir pH; de lo contrario, inserta a presión el tapón `Emisor_Tapon_pH` con un sello fino de silicona neutra.
]

#assembly-step(
  num: 2,
  title: "Fijar la bandeja modular inferior",
  parts: ("1× Emisor_Soporte_18650", "2× Tornillos M3×8 mm"),
  tool: "Destornillador o llave Allen",
  check: "La bandeja reposa nivelada sobre los dos postes inferiores sin tocar los prensaestopas.",
)[
  Alinea las dos perforaciones de `Emisor_Soporte_18650` sobre los postes internos de la zona inferior de `Emisor_Base` y sujétala con los tornillos M3×8 mm. Esta bandeja aloja hoy la bornera de distribución de 5 V y permite migrar en el futuro a un portapilas 1×18650 sin modificar la carcasa.
]

#assembly-step(
  num: 3,
  title: "Fijar la placa TTGO LoRa32 V2.1 y conectar la antena",
  parts: ("1× TTGO LoRa32 V2.1", "1× Antena SMA 433 MHz", "4× Tornillos M2×6 mm"),
  tool: "Destornillador de precisión",
  check: "La pantalla OLED queda orientada al frente y la antena sobresale recta por el collarín superior.",
)[
  Asienta la placa TTGO LoRa32 V2.1 sobre los cuatro postes elevados superiores y entre los rieles guía laterales. Asegúrala suavemente con los tornillos M2×6 mm y enrosca la antena de *433 MHz* atravesando el collarín superior antigoteo.
]

#v(14pt)
#precaucion[Nunca energices sin antena][
  Encender o transmitir con el módulo LoRa SX1278 (433 MHz) sin la antena conectada puede dañar de forma irreversible su etapa de salida de radiofrecuencia. Comprueba la antena antes de aplicar alimentación de 5 V.
]

#assembly-step(
  num: 4,
  title: "Pasar cables, conectar sondas y ajustar sellos",
  parts: ("Cable de 5 V", "Sondas de temperatura, humedad y pH opcional"),
  tool: "Pelacables y perillero",
  check: "Ningún cable se desliza al halarlo suavemente desde el exterior.",
)[
  Introduce cada cable desde la parte inferior dejando unos 5 cm de holgura interna para mantenimiento. Realiza las conexiones en la bornera y en los pines definidos en el firmware, y aprieta la tuerca externa de cada prensaestopas hasta comprimir el anillo de caucho alrededor de la chaqueta del cable.
]

#assembly-step(
  num: 5,
  title: "Pegar la ventana transparente en la tapa frontal",
  parts: ("1× Emisor_Tapa", "1× Acrílico 33.5×21.5×2 mm", "Silicona neutra"),
  tool: "Aplicador limpio",
  check: "La lámina queda al ras dentro del rebaje trasero y la zona del OLED permanece limpia.",
)[
  Coloca `Emisor_Tapa` boca abajo, aplica un cordón delgado y continuo de silicona neutra sobre el rebaje interno alrededor de la ventana OLED, asienta la lámina acrílica y deja curar durante 30 minutos antes de cerrar la caja.
]

#assembly-step(
  num: 6,
  title: "Asentar el empaque TPU y cerrar la carcasa",
  parts: ("1× Emisor_Empaque_TPU", "1× Emisor_Tapa", "4× Tornillos M3×16 mm"),
  tool: "Llave Allen M3",
  check: "La visera queda arriba junto a la antena y el empaque se comprime parejo en todo el perímetro.",
)[
  Presiona `Emisor_Empaque_TPU` dentro de la ranura perimetral de `Emisor_Base`. Presenta `Emisor_Tapa` con la visera orientada hacia la antena y aprieta los cuatro tornillos M3×16 mm en orden cruzado para obtener un sellado uniforme contra lluvia y vapor.

  #clean-figure(
    caption: [Vista frontal con el grabado institucional en bajo relieve y vista isométrica del Nodo Emisor cerrado.],
  )[
    #grid(
      columns: (1fr, 1fr),
      column-gutter: 10pt,
      [#image("assets/emisor_tapa_top.png", width: 100%)],
      [#image("assets/emisor_ensamblado.png", width: 100%)],
    )
  ]
]

#assembly-step(
  num: 7,
  title: "Instalar en estaca junto a la pila de compostaje",
  parts: ("Nodo Emisor sellado", "2× Amarres plásticos o 4× M5", "Estaca o tubo PVC"),
  tool: "Alicate de corte",
  check: "Antena apuntando al cielo, tapa vertical y cables con curva de goteo hacia el suelo.",
)[
  Sujeta el Nodo Emisor a una estaca firme a 40 o 60 cm sobre el suelo usando las orejas laterales. Antes de entrar a los prensaestopas, deja una *curva de goteo en U* en todos los cables para que el agua de lluvia escurra hacia el suelo y nunca hacia la caja.

  #clean-figure(
    caption: [Orientación vertical obligatoria en campo y curva de goteo en los cables inferiores.],
  )[
    #box(width: 240pt, height: 112pt)[
      // Estaca vertical
      #place(left + top, dx: 42pt, dy: 8pt, rect(width: 10pt, height: 98pt, fill: rgb("#E2E8F0"), stroke: 0.7pt + ink-soft))
      // Antena apuntando al cielo
      #place(left + top, dx: 96pt, dy: 2pt, rect(width: 5pt, height: 24pt, fill: ink, radius: 1pt))
      #place(left + top, dx: 110pt, dy: 6pt, text(font: "Inter", size: 7.5pt, fill: ink)[Antena 433 MHz arriba])
      // Carcasa emisor
      #place(left + top, dx: 52pt, dy: 26pt, rect(width: 68pt, height: 46pt, fill: rgb("#EDF4F0"), stroke: 1pt + accent, radius: 3pt))
      #place(left + top, dx: 63pt, dy: 34pt, rect(width: 26pt, height: 13pt, fill: rgb("#1E3A5F"), radius: 1pt))
      #place(left + top, dx: 62pt, dy: 53pt, text(font: "Inter", size: 6.5pt, weight: "bold", fill: accent)[Nodo Emisor])
      // Curva de goteo (cables saliendo por debajo y bajando antes de subir a la pila)
      #place(left + top, dx: 70pt, dy: 72pt, line(start: (0pt, 0pt), end: (0pt, 22pt), stroke: 1pt + ink))
      #place(left + top, dx: 70pt, dy: 94pt, line(start: (0pt, 0pt), end: (26pt, 0pt), stroke: 1pt + ink))
      #place(left + top, dx: 96pt, dy: 94pt, line(start: (0pt, 0pt), end: (44pt, -16pt), stroke: 1pt + ink))
      #place(left + top, dx: 70pt, dy: 99pt, text(font: "Inter", size: 7pt, fill: accent)[Curva de goteo (agua cae aquí)])
      #place(left + top, dx: 145pt, dy: 72pt, text(font: "Inter", size: 7.5pt, fill: ink)[Hacia sondas en la pila])
    ]
  ]
]

== Montaje del Nodo Receptor (Gateway)

El segundo módulo TTGO LoRa32 V2.1 opera bajo techo (aula, oficina o laboratorio), recibe las tramas de 433 MHz y las envía al servidor por WiFi o por cable USB hacia `localhost`.

#clean-figure(
  caption: [Despiece del Nodo Receptor con ranura lateral Micro-USB/microSD y ventilación inferior.],
)[
  #image("assets/receptor_explosionado.png", width: 80%)
]

#assembly-step(
  num: 8,
  title: "Ensamblar y cerrar el Nodo Receptor",
  parts: ("1× Receptor_Base", "1× Receptor_Tapa", "1× TTGO LoRa32 V2.1", "1× Antena 433 MHz", "4× M3×12 mm"),
  tool: "Destornillador M3",
  check: "El puerto Micro-USB queda alineado en el costado derecho y la pantalla OLED visible arriba.",
)[
  Apoya la segunda placa TTGO LoRa32 V2.1 sobre los cuatro postes internos de `Receptor_Base` con el puerto Micro-USB orientado hacia la apertura derecha. Enrosca la antena de 433 MHz por el collarín superior, encaja `Receptor_Tapa` y asegura las cuatro esquinas con los tornillos M3×12 mm.

  #clean-figure(
    caption: [Vista superior e isométrica del Nodo Receptor listo para escritorio o montaje en pared.],
  )[
    #grid(
      columns: (1fr, 1fr),
      column-gutter: 10pt,
      [#image("assets/receptor_tapa_top.png", width: 100%)],
      [#image("assets/receptor_ensamblado.png", width: 100%)],
    )
  ]
]

== Verificación previa a la puesta en marcha

Antes de dejar el prototipo operando junto a la pila, confirma los cuatro puntos de control:

- $square$ Ambas antenas de 433 MHz están firmemente enroscadas antes de energizar las placas.
- $square$ Las pantallas OLED muestran temperatura, humedad y nivel de señal LoRa (`RSSI`).
- $square$ El Nodo Emisor está vertical, con los prensaestopas hacia abajo, curva de goteo formada y tapón puesto si no se usa sonda de pH.
- $square$ El Nodo Receptor entrega lecturas visibles en el tablero de CompostIOT.
