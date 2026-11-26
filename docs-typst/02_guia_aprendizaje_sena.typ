#import "common.typ": *

#show: compost-doc

= Guía de aprendizaje SENA

#doc-meta(
  code: "CIOT-GUIA-02 · v1.1",
  subtitle: "Análisis científico y gestión productiva del compostaje asistido por IoT: del registro manual en cuaderno a la toma de decisiones con datos reales.",
  audience: "Aprendices tecnólogos del área ambiental (SENA), instructores y productores de abono orgánico.",
)

== Propósito formativo

En la práctica habitual de compostaje, los aprendices invierten buena parte de la jornada tomando datos a mano y estimando las fases por observación visual, como la salida de vapor durante la fase termófila. Este flujo limita la formación a labores operativas y favorece errores de transcripción o vacíos durante fines de semana. #sidenote[Con *CompostIOT*, los nodos capturan los datos automáticamente para que el aprendiz dedique su tiempo a *analizar curvas, diagnosticar desviaciones y mejorar la calidad del abono*.]

Al completar esta guía podrás:
1. Diferenciar las cuatro fases del compostaje aeróbico a partir de la curva térmica histórica.
2. Evaluar en conjunto temperatura, humedad y pH (cuando exista sonda) para prevenir condiciones anaeróbicas o sobrecalentamiento.
3. Exportar series históricas en CSV para informes técnicos, artículos científicos y seguimiento productivo.

== Las cuatro fases del compostaje aeróbico

Para evitar falsos cambios de fase ocasionados por ruido puntual de un sensor, el motor de *CompostIOT* calcula el promedio móvil de las últimas tres lecturas válidas:

$ T_"ref" = (T_n + T_(n-1) + T_(n-2)) / 3 $

#concepto[Memoria de pico térmico][
  Una pila solo puede ingresar a *Enfriamiento* o *Maduración* después de haber registrado al menos dos lecturas en régimen termófilo ($T >= 45 degree"C"$). Antes de ese umbral, toda temperatura inferior a $45 degree"C"$ se clasifica como *Mesófila inicial*.
]

#editorial-table(
  columns: (auto, auto, 1fr, 1fr),
  headers: ("Fase", "Temperatura", "Actividad biológica", "Acción en campo"),
  rows: (
    ([*1. Mesófila inicial*], [20 a 44.9 °C\ (previa al pico)], [Bacterias mesófilas degradan azúcares y compuestos simples, elevando la temperatura.], [Verificar humedad inicial entre 50 y 60 % y relación C/N cercana a 30:1.]),
    ([*2. Termófila*], [$>= 45 degree"C"$\ (óptimo 55 a 65 °C)], [Microorganismos termófilos eliminan patógenos y semillas de malezas; aparece vapor al voltear.], [Voltear de inmediato si supera 70 °C y reponer agua si la humedad cae de 45 %.]),
    ([*3. Enfriamiento*], [30 a 44.9 °C\ (tras el pico)], [Hongos y actinomicetos degradan celulosa y lignina mientras el calor desciende.], [Espaciar volteos y evitar que el sustrato se reseque por debajo de 40 %.]),
    ([*4. Maduración*], [$< 30 degree"C"$\ estable], [Estabilización del humus con aroma a tierra húmeda y abono apto para uso agrícola.], [Dejar reposar de 2 a 4 semanas, tamizar y exportar el CSV del lote.]),
  ),
)

=== Curva térmica de referencia

#clean-figure(
  caption: [Curva térmica típica de una pila aeróbica bien manejada y umbrales de clasificación empleados por CompostIOT.],
)[
  #box(width: 320pt, height: 128pt)[
    // Ejes principales
    #place(left + top, dx: 36pt, dy: 12pt, line(start: (0pt, 0pt), end: (0pt, 92pt), stroke: 0.8pt + ink))
    #place(left + top, dx: 36pt, dy: 104pt, line(start: (0pt, 0pt), end: (272pt, 0pt), stroke: 0.8pt + ink))
    // Banda termófila sutil (45 a 65 °C)
    #place(left + top, dx: 36.5pt, dy: 24pt, rect(width: 270pt, height: 26pt, fill: rgb("#FFF7ED")))
    // Líneas de referencia 70 °C y 45 °C
    #place(left + top, dx: 36pt, dy: 18pt, line(start: (0pt, 0pt), end: (270pt, 0pt), stroke: (paint: rgb("#DC2626"), thickness: 0.5pt, dash: "dashed")))
    #place(left + top, dx: 36pt, dy: 50pt, line(start: (0pt, 0pt), end: (270pt, 0pt), stroke: (paint: warn-ink, thickness: 0.5pt, dash: "dotted")))
    // Etiquetas eje Y
    #place(left + top, dx: 4pt, dy: 14pt, text(font: "Inter", size: 7pt, fill: rgb("#DC2626"))[70 °C])
    #place(left + top, dx: 4pt, dy: 46pt, text(font: "Inter", size: 7pt, fill: warn-ink)[45 °C])
    #place(left + top, dx: 4pt, dy: 72pt, text(font: "Inter", size: 7pt, fill: ink-soft)[30 °C])
    #place(left + top, dx: 4pt, dy: 96pt, text(font: "Inter", size: 7pt, fill: ink-soft)[15 °C])
    // Curva térmica suave (4 fases) usando curve nativo de Typst 0.15
    #place(left + top, dx: 36pt, dy: 12pt, curve(
      stroke: 1.6pt + accent,
      curve.move((4pt, 76pt)),
      curve.cubic((22pt, 66pt), (32pt, 52pt), (46pt, 38pt)),
      curve.cubic((64pt, 20pt), (88pt, 18pt), (112pt, 18pt)),
      curve.cubic((142pt, 18pt), (160pt, 34pt), (182pt, 48pt)),
      curve.cubic((208pt, 64pt), (236pt, 72pt), (264pt, 74pt)),
    ))
    // Etiquetas de fases sobre la curva
    #place(left + top, dx: 40pt, dy: 109pt, text(font: "Inter", size: 7.2pt, fill: blue-ink)[Mesófila I])
    #place(left + top, dx: 112pt, dy: 109pt, text(font: "Inter", size: 7.2pt, weight: "bold", fill: warn-ink)[Termófila])
    #place(left + top, dx: 186pt, dy: 109pt, text(font: "Inter", size: 7.2pt, fill: ink-soft)[Enfriamiento])
    #place(left + top, dx: 255pt, dy: 109pt, text(font: "Inter", size: 7.2pt, weight: "bold", fill: accent)[Maduración])
  ]
]

== Diagnóstico rápido de alertas

Cuando el tablero reporte una desviación en cualquiera de las variables monitoreadas, aplica el ajuste agronómico indicado:

#editorial-table(
  columns: (auto, 1fr, 1fr),
  headers: ("Condición detectada", "Causa probable", "Intervención recomendada"),
  rows: (
    ([Temperatura > 70 °C], [Actividad excesiva o pila demasiado alta y compacta.], [Voltear de inmediato para oxigenar y disipar el exceso de calor.]),
    ([Humedad < 40 %], [Evaporación intensa en fase termófila sin reposición.], [Regar uniformemente durante el volteo buscando 50 a 60 %.]),
    ([Humedad > 65 %], [Exceso de residuos húmedos o entrada de agua lluvia.], [Voltear y mezclar material seco estructurante como aserrín u hojarasca.]),
    ([pH < 5.5 (si aplica)], [Acumulación de ácidos orgánicos por falta de aireación.], [Aumentar aireación con volteo y evaluar adición ligera de cal agrícola.]),
    ([pH > 8.5 (si aplica)], [Exceso de nitrógeno y pérdida de amoníaco.], [Incorporar material rico en carbono seco para equilibrar la mezcla.]),
  ),
)

== Actividades prácticas para aprendices

#practica[Validación de higienización termófila][
  Inicia sesión con tu cuenta `aprendiz` en la web, el cliente de escritorio o tu teléfono y selecciona una pila activa. Observa la curva de *Temperatura* y documenta: cuántos días tardó en superar los $45 degree"C"$, si sostuvo el rango de $55 " a " 65 degree"C"$ durante al menos tres días continuos y qué labor de campo explica los descensos breves seguidos de un nuevo repunte térmico.
]

#practica[Análisis comparativo con datos CSV][
  Usa el botón *Exportar CSV* en dos pilas con sustratos diferentes (por ejemplo, residuos de cafetería frente a estiércol bovino con viruta). Abre ambos archivos en LibreOffice Calc, Excel o Python para calcular temperatura máxima, humedad promedio y duración de la fase termófila. Redacta una recomendación técnica para optimizar el siguiente lote productivo.
]

#practica[Laboratorio remoto entre sedes SENA][
  Cuando el servidor esté desplegado en la nube del centro de formación, grupos de aprendices de otras sedes o en modalidad virtual pueden conectarse simultáneamente con rol `aprendiz`, analizar en tiempo real las pilas de la granja y proponer decisiones de manejo sustentadas únicamente en la telemetría.
]
