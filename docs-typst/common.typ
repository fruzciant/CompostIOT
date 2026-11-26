#import "@preview/obelisk:0.2.0": *

// Paleta editorial mínima: tinta oscura, grafito y un único acento verde SENA
#let ink = rgb("#141916")
#let ink-soft = rgb("#4A5550")
#let rule-color = rgb("#C8D0CC")
#let accent = rgb("#15613B")
#let warn-ink = rgb("#9A3412")
#let blue-ink = rgb("#1D4ED8")

// Entornos nativos de Obelisk (sin cajas redondeadas ni rellenos pastel)
#let precaucion = make-environment("Precaución", "PRE", rgb("#991B1B"))
#let practica = make-environment("Práctica", "LAB", rgb("#1D4ED8"))
#let concepto = make-environment("Concepto", "DEF", rgb("#15613B"))
#let nota = make-environment("Nota", "NOTA", rgb("#475569"))

// Configuración base estrictamente fiel a @preview/obelisk:0.2.0
#let compost-doc(doc) = {
  set text(lang: "es", fill: ink)
  show: init.with(
    paper: (
      width: 21cm,
      height: 29.7cm,
      two-sided: false,
    ),
    texts: (
      size: 10.5pt,
      ascender: 8.5pt,
      step: 15pt,
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
      h2: (sym: sym.circle.filled, dy: 0pt, size: 1.5),
      h3: (sym: sym.triangle.filled.small.b, dy: 0pt, size: 1.18),
    ),
  )
  doc
}

// Encabezado tipográfico editorial (cero cajas, solo jerarquía y filete 0.6pt)
#let doc-meta(code: "", subtitle: "", audience: "") = {
  block(width: 100%, above: 4pt, below: 18pt)[
    #set text(top-edge: "ascender", bottom-edge: "descender")
    #set par(leading: 0.6em, spacing: 0.5em, justify: false)
    #text(font: "Inter", size: 8.8pt, weight: "semibold", fill: accent)[
      SENA · Maestría en Educación Sostenible y Gestión Ambiental Territorial
    ]
    #v(6pt)
    #text(size: 10.5pt, fill: ink-soft)[#subtitle]
    #v(7pt)
    #line(length: 100%, stroke: 0.6pt + rule-color)
    #v(6pt)
    #grid(
      columns: (1fr, auto),
      column-gutter: 12pt,
      align: (left + top, right + top),
      [#text(font: "Inter", size: 8.4pt, fill: ink-soft)[Dirigido a: #text(fill: ink)[#audience]]],
      [#text(font: "DejaVu Sans Mono", size: 8pt, fill: ink-soft)[#code]],
    )
  ]
}

// Tabla limpia estilo Booktabs (sin rellenos sólidos ni rejillas verticales)
#let editorial-table(columns: auto, headers: (), rows: ()) = {
  block(width: 100%, breakable: true, above: 10pt, below: 14pt)[
    #set text(font: "Inter", size: 8.5pt, top-edge: "ascender", bottom-edge: "descender")
    #set par(leading: 0.55em, spacing: 0.45em, justify: false)
    #table(
      columns: columns,
      inset: (x: 4pt, y: 6pt),
      stroke: (x, y) => (
        top: if y == 0 { 1.1pt + ink } else if y == 1 { 0.6pt + ink } else { 0.3pt + rule-color },
        bottom: if y == rows.len() { 1.1pt + ink } else { none },
        left: none,
        right: none,
      ),
      table.header(
        ..headers.map(h => text(weight: "bold", fill: ink)[#h]),
      ),
      ..rows.flatten(),
    )
  ]
}

// Paso de ensamble estilo Bauhaus / LEGO aprovechando el margen lateral nativo de Obelisk
#let assembly-step(num: 1, title: "", parts: (), tool: "", check: "", body) = {
  block(width: 100%, breakable: false, above: 18pt, below: 14pt)[
    #set text(top-edge: "ascender", bottom-edge: "descender")
    #set par(leading: 0.55em, justify: false)
    #line(length: 100%, stroke: 0.45pt + rule-color)
    #v(8pt)
    #sidenote[
      #set text(font: "Inter", size: 7.6pt, top-edge: "ascender", bottom-edge: "descender")
      #set par(leading: 0.46em, spacing: 0.3em, justify: false)
      #text(weight: "bold", fill: accent)[Paso #num]\
      #for p in parts {
        [#p\ ]
      }
      #if tool != "" {
        v(2pt)
        text(fill: ink-soft)[#tool]
      }
    ]
    #text(font: "Inter", size: 11.5pt, weight: "bold", fill: ink)[
      #text(fill: accent)[#num.] #title
    ]
  ]

  body

  if check != "" {
    block(width: 100%, above: 8pt, below: 14pt)[
      #set text(font: "Inter", size: 8.8pt, fill: accent, top-edge: "ascender", bottom-edge: "descender")
      #set par(leading: 0.55em, justify: false)
      *Verificación:* #text(fill: ink)[#check]
    ]
  }
}

// Figura limpia sin marcos pesados, centrada en el espacio blanco
#let clean-figure(caption: "", body) = {
  figure(
    supplement: [Figura],
    caption: text(font: "Inter", size: 8.2pt, fill: ink-soft)[#caption],
    block(width: 100%, breakable: false, above: 8pt, below: 4pt)[
      #set text(top-edge: "ascender", bottom-edge: "descender")
      #set par(leading: 0.55em, justify: false)
      #align(center)[#body]
    ],
  )
}
