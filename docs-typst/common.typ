#import "@preview/obelisk:0.2.0": *

// Paleta institucional CompostIOT / SENA
#let sena-green = rgb("#15613B")
#let sena-soft = rgb("#EDF4F0")
#let sena-border = rgb("#C6D8CE")
#let ink-dark = rgb("#12261C")
#let ink-muted = rgb("#465C50")
#let phase-blue = rgb("#2563EB")
#let phase-orange = rgb("#D95F02")
#let phase-ochre = rgb("#B47B00")
#let phase-green = rgb("#15803D")

// Configuración global sobre Obelisk 0.2.0 (Bauhaus + español + sin advertencias de fuentes)
#let compost-doc(doc) = {
  set text(lang: "es", fill: ink-dark)
  show: init.with(
    paper: (
      width: 21cm,
      height: 29.7cm,
      two-sided: false,
    ),
    texts: (
      size: 10.5pt,
      ascender: 8.5pt,
      step: 14.5pt,
    ),
    fonts: (
      body: "Libertinus Serif",
      math: "New Computer Modern Math",
      sans: "Inter",
      mono: "DejaVu Sans Mono",
    ),
    deco: (
      line-number: false,
    ),
    headers: (
      h2: (sym: sym.circle.filled, dy: 0pt, size: 1.55),
      h3: (sym: sym.triangle.filled.small.b, dy: 0pt, size: 1.25),
    ),
  )
  doc
}

// Contenedor limpio para tablas, figuras y tarjetas (evita el bug de measure sin ancho de bblock)
#let card-block(breakable: false, body) = {
  block(
    width: 100%,
    breakable: breakable,
    above: 12pt,
    below: 14pt,
  )[
    #set text(top-edge: "ascender", bottom-edge: "descender")
    #set par(leading: 0.62em, spacing: 0.65em, justify: false)
    #body
  ]
}

// Ficha técnica de encabezado institucional (justo debajo del título Bauhaus de Obelisk)
#let cover-meta(code: "DOC-01", subtitle: "", audience: "", version: "v1.0") = {
  card-block[
    #box(
      width: 100%,
      stroke: (left: 4pt + sena-green, rest: 0.6pt + sena-border),
      fill: sena-soft,
      radius: (right: 4pt),
      inset: (x: 12pt, y: 10pt),
    )[
      #set text(font: "Inter")
      #grid(
        columns: (1fr, auto),
        column-gutter: 8pt,
        align: (left + horizon, right + horizon),
        [#text(size: 8pt, weight: "bold", fill: sena-green)[SENA · MAESTRÍA EN EDUCACIÓN SOSTENIBLE Y GESTIÓN AMBIENTAL TERRITORIAL]],
        [#text(size: 8pt, weight: "bold", fill: ink-muted)[#code · #version]],
      )
      #v(5pt)
      #text(size: 10pt, fill: ink-dark)[#subtitle]
      #v(6pt)
      #line(length: 100%, stroke: 0.5pt + sena-border)
      #v(5pt)
      #text(size: 8.5pt, fill: ink-muted)[*Dirigido a:* #text(fill: ink-dark)[#audience]]
    ]
  ]
}

// Bloque de paso visual estilo manual LEGO
#let lego-step(num: 1, title: "", parts: (), tools: "", check: "", body) = {
  card-block[
    #box(
      width: 100%,
      stroke: 0.8pt + sena-border,
      radius: 4pt,
      fill: white,
      clip: true,
    )[
      // Franja superior del paso
      #box(
        width: 100%,
        fill: sena-soft,
        inset: (x: 10pt, y: 6pt),
        stroke: (bottom: 0.6pt + sena-border),
      )[
        #set text(font: "Inter")
        #grid(
          columns: (auto, 1fr, auto),
          column-gutter: 8pt,
          align: (left + horizon, left + horizon, right + horizon),
          box(
            fill: sena-green,
            radius: 3pt,
            inset: (x: 7pt, y: 3.5pt),
          )[#text(fill: white, size: 10pt, weight: "extrabold")[PASO #num]],
          [#text(size: 10pt, weight: "bold", fill: ink-dark)[#title]],
          if tools != "" [
            #text(size: 7.5pt, fill: ink-muted)[*Herramienta:* #tools]
          ] else [],
        )
      ]

      // Cuerpo + inventario de piezas del paso
      #pad(x: 10pt, y: 8pt)[
        #if parts.len() > 0 {
          box(
            width: 100%,
            fill: rgb("#F8FAF9"),
            stroke: 0.5pt + sena-border,
            radius: 3pt,
            inset: (x: 8pt, y: 5pt),
          )[
            #set text(font: "Inter", size: 8pt)
            #text(weight: "bold", fill: sena-green)[Piezas: ]
            #parts.join("  ·  ")
          ]
          v(6pt)
        }

        #set text(size: 9.5pt)
        #body

        #if check != "" {
          v(6pt)
          box(
            width: 100%,
            fill: rgb("#F0FDF4"),
            stroke: (left: 3pt + phase-green, rest: 0.5pt + rgb("#BBF7D0")),
            inset: (x: 8pt, y: 4.5pt),
          )[
            #set text(font: "Inter", size: 8pt, fill: rgb("#14532D"))
            *Verificación [✓]:* #check
          ]
        }
      ]
    ]
  ]
}

// Tarjeta de alerta / consejo / actividad pedagógica
#let callout(kind: "info", title: "", body) = {
  let palette = if kind == "warning" {
    (bg: rgb("#FFFBEB"), border: rgb("#F59E0B"), fg: rgb("#92400E"))
  } else if kind == "danger" {
    (bg: rgb("#FEF2F2"), border: rgb("#DC2626"), fg: rgb("#991B1B"))
  } else if kind == "activity" {
    (bg: rgb("#EFF6FF"), border: rgb("#2563EB"), fg: rgb("#1E3A8A"))
  } else {
    (bg: sena-soft, border: sena-green, fg: ink-dark)
  }

  card-block[
    #box(
      width: 100%,
      fill: palette.bg,
      stroke: (left: 3.5pt + palette.border, rest: 0.5pt + palette.border.lighten(50%)),
      radius: (right: 3pt),
      inset: (x: 10pt, y: 8pt),
    )[
      #set text(font: "Inter", size: 9pt, fill: palette.fg)
      #text(weight: "bold")[#title]
      #v(4pt)
      #set text(font: "Libertinus Serif", size: 9.5pt, fill: ink-dark)
      #body
    ]
  ]
}
