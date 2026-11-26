#import "common.typ": *

#show: compost-doc

= Guía de Aprendizaje SENA

#cover-meta(
  code: "CIOT-GUIA-02",
  subtitle: "Análisis científico y productivo del proceso de compostaje asistido por IoT: del registro manual a la toma de decisiones basada en datos.",
  audience: "Aprendices tecnólogos del área ambiental (SENA), instructores y productores de abono orgánico.",
  version: "v1.0",
)

== Propósito formativo

El seguimiento tradicional de pilas de compostaje obliga a los aprendices a tomar mediciones manuales diarias y deducir la fase por señales visuales (como vapor en la fase termófila). Ese flujo consume horas operativas y genera errores de transcripción o vacíos durante fines de semana. #sidenote[Con *CompostIOT*, los sensores automatizan la captura para que el aprendiz enfoque su criterio técnico en *interpretar curvas, diagnosticar anomalías y mejorar la calidad del abono*.]

Al finalizar esta guía podrás:
1. Identificar las *cuatro fases del compostaje aeróbico* a partir de la curva térmica histórica.
2. Evaluar en conjunto *Temperatura*, *Humedad* y *pH (opcional)* para prevenir fallas del proceso.
3. Exportar datos en CSV para informes técnicos, artículos científicos y control productivo.

== Las cuatro fases del ciclo aeróbico

El motor de *CompostIOT* calcula el promedio móvil de las últimas 3 lecturas válidas:
$ T_"ref" = (T_n + T_(n-1) + T_(n-2)) / 3 $
combinado con la *memoria de pico térmico* (al menos 2 mediciones con $T >= 45 degree"C"$), evitando falsos cambios de fase por ruido puntual.

#card-block(breakable: true)[
  #set text(font: "Inter", size: 8.2pt)
  #table(
    columns: (auto, auto, 1fr, 1fr),
    inset: 5.5pt,
    stroke: 0.5pt + sena-border,
    fill: (_, row) => if row == 0 { sena-green } else if calc.odd(row) { rgb("#F8FAF9") } else { white },
    table.header(
      text(fill: white, weight: "bold")[Fase],
      text(fill: white, weight: "bold")[Temperatura],
      text(fill: white, weight: "bold")[Proceso biológico],
      text(fill: white, weight: "bold")[Acción práctica del aprendiz],
    ),
    [#text(fill: phase-blue, weight: "bold")[1. Mesófila inicial]], [20–44.9 °C\ (previo al pico)], [Bacterias mesófilas degradan azúcares y compuestos simples; la pila eleva su temperatura.], [Verificar humedad inicial (50–60 %) y relación C/N aproximada de 25:1 a 30:1.],
    [#text(fill: phase-orange, weight: "bold")[2. Termófila\ (Higienización)]], [$>= 45 degree"C"$\ (óptimo 55–65 °C)], [Microorganismos termófilos eliminan patógenos y semillas de malezas; suele liberarse vapor.], [Controlar picos $> 70 degree"C"$ para voltear de inmediato y reponer agua si humedad $< 45\%$.],
    [#text(fill: phase-ochre, weight: "bold")[3. Enfriamiento\ (Mesófila II)]], [30–44.9 °C\ (tras el pico)], [Hongos y actinomicetos degradan celulosa y lignina; el calor desciende gradualmente.], [Espaciar volteos y evitar que el sustrato se reseque por debajo del 40 %.],
    [#text(fill: phase-green, weight: "bold")[4. Maduración]], [$< 30 degree"C"$\ estable], [Estabilización del humus; aroma a tierra húmeda de bosque y abono apto para uso agrícola.], [Dejar reposar 2–4 semanas, tamizar y exportar CSV como respaldo del lote.],
  )
]

=== Progresión visual del ciclo en el tablero

#card-block[
  #box(width: 100%, inset: 8pt, stroke: 0.6pt + sena-border, radius: 4pt, fill: white)[
    #set text(font: "Inter", size: 7.5pt)
    #grid(
      columns: (1fr, 1.25fr, 1.05fr, 1fr),
      column-gutter: 5pt,
      box(width: 100%, fill: rgb("#EFF6FF"), stroke: 0.6pt + rgb("#BFDBFE"), radius: 3pt, inset: 6pt)[
        #text(fill: phase-blue, weight: "bold")[1. Mesófila inicial]\
        #v(2pt)
        20 → 44 °C · Días 1–5\
        #text(fill: ink-muted)[Arranque térmico]
      ],
      box(width: 100%, fill: rgb("#FFF7ED"), stroke: 0.8pt + rgb("#FDBA74"), radius: 3pt, inset: 6pt)[
        #text(fill: phase-orange, weight: "bold")[2. Termófila (Higienización)]\
        #v(2pt)
        *45 → 65 °C* (alerta >70 °C)\
        #text(fill: ink-muted)[Eliminación de patógenos]
      ],
      box(width: 100%, fill: rgb("#FEFCE8"), stroke: 0.6pt + rgb("#FDE047"), radius: 3pt, inset: 6pt)[
        #text(fill: phase-ochre, weight: "bold")[3. Enfriamiento]\
        #v(2pt)
        44 → 30 °C · Días 26–45\
        #text(fill: ink-muted)[Hongos y lignina]
      ],
      box(width: 100%, fill: rgb("#F0FDF4"), stroke: 0.6pt + rgb("#86EFAC"), radius: 3pt, inset: 6pt)[
        #text(fill: phase-green, weight: "bold")[4. Maduración]\
        #v(2pt)
        < 30 °C estable\
        #text(fill: ink-muted)[Humus listo]
      ],
    )
    #v(5pt)
    #align(center)[#text(fill: ink-muted, size: 7.8pt)[*Figura 1.* Las cuatro etapas detectadas automáticamente por CompostIOT según temperatura y memoria térmica.]]
  ]
]

== Matriz de diagnóstico y decisiones en campo

Cuando el tablero muestre una alerta amarilla o roja, aplica la intervención correspondiente:

#card-block(breakable: true)[
  #set text(font: "Inter", size: 8.2pt)
  #table(
    columns: (auto, 1fr, 1fr),
    inset: 5.5pt,
    stroke: 0.5pt + sena-border,
    fill: (_, row) => if row == 0 { ink-dark } else if calc.odd(row) { rgb("#F8FAF9") } else { white },
    table.header(
      text(fill: white, weight: "bold")[Alerta],
      text(fill: white, weight: "bold")[Causa probable],
      text(fill: white, weight: "bold")[Intervención recomendada],
    ),
    [*Temperatura > 70 °C*], [Actividad excesiva o pila demasiado alta y compacta.], [Voltear de inmediato para oxigenar y disipar calor excesivo.],
    [*Humedad < 40 %*], [Evaporación alta durante fase termófila sin reposición.], [Regar uniformemente durante el volteo (meta: 50–60 %).],
    [*Humedad > 65 %*], [Exceso de residuos húmedos o filtración de lluvia.], [Voltear y mezclar material seco estructurante (aserrín u hojarasca).],
    [*pH < 5.5* (opcional)], [Acumulación de ácidos orgánicos por baja aireación.], [Oxigenar con volteo; si persiste, evaluar cal agrícola moderada.],
    [*pH > 8.5* (opcional)], [Exceso de nitrógeno y pérdida de amoníaco.], [Incorporar material carbonado seco para equilibrar C/N.],
  )
]

== Actividades prácticas para aprendices SENA

#callout(
  kind: "activity",
  title: "Práctica 1 — Validación de la fase termófila (Higienización)",
)[
  1. Inicia sesión con tu usuario `aprendiz` desde Web, cliente de escritorio o celular.
  2. Selecciona una pila activa y analiza la curva de *Temperatura*.
  3. Responde en tu informe técnico:
     - ¿Cuántos días tardó la pila en superar $45 degree"C"$?
     - ¿Se mantuvo en la banda de $55 degree"C" " a " 65 degree"C"$ durante al menos 3 días continuos?
     - Si observas caídas breves seguidas de un nuevo repunte térmico, ¿qué labor de campo explicaría ese patrón?
]

#callout(
  kind: "activity",
  title: "Práctica 2 — Estudio comparativo exportando CSV",
)[
  1. Pulsa *Exportar CSV* en dos pilas con sustratos diferentes (ej. residuos vegetales de cafetería vs. estiércol bovino con viruta).
  2. Abre ambos archivos en LibreOffice Calc, Excel o Python y calcula: temperatura máxima, humedad media y días en fase termófila.
  3. Formula una recomendación productiva para mejorar el rendimiento del siguiente lote.
]

#callout(
  kind: "activity",
  title: "Práctica 3 — Laboratorio remoto entre centros SENA",
)[
  Con el backend desplegado en la nube institucional, aprendices en modalidad virtual o de otra sede pueden conectarse simultáneamente con cuenta `aprendiz`, analizar en tiempo real las pilas de la granja y emitir un diagnóstico agronómico fundamentado únicamente en la telemetría.
]
