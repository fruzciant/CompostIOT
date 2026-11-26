"""
CompostIOT — Generador CAD Paramétrico: Nodo Receptor / Gateway (LILYGO TTGO LoRa32 V2.1)
Carcasa compacta de escritorio o montaje en pared para aula, oficina o estación local.
Diseñada para impresión 3D en PLA+ o PETG sin soportes complejos.
"""

import math
import os
import sys
import FreeCAD as App
import FreeCADGui as Gui
import Part
import MeshPart

SCRIPTS_DIR = "/home/fruzciant/Documents/Proyectos/CompostIOT/hardware-cad/scripts"
if SCRIPTS_DIR not in sys.path:
    sys.path.insert(0, SCRIPTS_DIR)
import importlib
import branding
importlib.reload(branding)

OUT_DIR = "/home/fruzciant/Documents/Proyectos/CompostIOT/hardware-cad/receptor"
RENDER_DIR = "/home/fruzciant/Documents/Proyectos/CompostIOT/hardware-cad/renders"

# Caja compacta (mm)
L = 94.0
W = 50.0
H_BASE = 27.0
WALL = 2.6
FLOOR = 2.6
R_OUT = 6.0
LID_T = 3.2

# Tornillos tapa (4x M3)
BOSS_X = L / 2.0 - WALL - 4.2
BOSS_Y = W / 2.0 - WALL - 4.2
BOSS_R = 3.8

# TTGO LoRa32 V2.1 (65 x 27 mm)
TTGO_CX = -2.0
TTGO_CY = 0.0
TTGO_L = 65.0
TTGO_W = 27.0
TTGO_Z_PCB = 13.5

# Ventana OLED
OLED_WIN_CX = TTGO_CX - 6.0
OLED_WIN_CY = TTGO_CY
OLED_WIN_L = 28.0
OLED_WIN_W = 16.0

# Salida antena SMA (+Y)
SMA_X = 18.0
SMA_Z = 16.0


def make_rounded_rect_face(length, width, radius, z=0.0):
    hx = length / 2.0
    hy = width / 2.0
    r = max(0.2, min(radius, hx - 0.1, hy - 0.1))
    k = r * (1.0 - math.sqrt(2.0) / 2.0)

    p1 = App.Vector(-hx + r, -hy, z)
    p2 = App.Vector(hx - r, -hy, z)
    p3 = App.Vector(hx, -hy + r, z)
    p4 = App.Vector(hx, hy - r, z)
    p5 = App.Vector(hx - r, hy, z)
    p6 = App.Vector(-hx + r, hy, z)
    p7 = App.Vector(-hx, hy - r, z)
    p8 = App.Vector(-hx, -hy + r, z)

    a1 = App.Vector(hx - k, -hy + k, z)
    a2 = App.Vector(hx - k, hy - k, z)
    a3 = App.Vector(-hx + k, hy - k, z)
    a4 = App.Vector(-hx + k, -hy + k, z)

    edges = [
        Part.LineSegment(p1, p2).toShape(),
        Part.Arc(p2, a1, p3).toShape(),
        Part.LineSegment(p3, p4).toShape(),
        Part.Arc(p4, a2, p5).toShape(),
        Part.LineSegment(p5, p6).toShape(),
        Part.Arc(p6, a3, p7).toShape(),
        Part.LineSegment(p7, p8).toShape(),
        Part.Arc(p8, a4, p1).toShape(),
    ]
    return Part.Face(Part.Wire(edges))


def rounded_box(length, width, height, radius, z0=0.0):
    return make_rounded_rect_face(length, width, radius, z0).extrude(App.Vector(0, 0, height))


def centered_box(cx, cy, z0, lx, ly, lz):
    return Part.makeBox(lx, ly, lz, App.Vector(cx - lx / 2.0, cy - ly / 2.0, z0))


def cyl_z(cx, cy, z0, radius, height):
    return Part.makeCylinder(radius, height, App.Vector(cx, cy, z0), App.Vector(0, 0, 1))


def cyl_y(cx, y0, cz, radius, length):
    return Part.makeCylinder(radius, length, App.Vector(cx, y0, cz), App.Vector(0, 1, 0))


def build_receptor_base():
    body = rounded_box(L, W, H_BASE, R_OUT, 0.0)
    # Collarín SMA en borde +Y
    body = body.fuse(cyl_y(SMA_X, W / 2.0 - 1.5, SMA_Z, 6.2, 5.5))

    # Cavidad interna
    cavity = rounded_box(L - 2.0 * WALL, W - 2.0 * WALL, H_BASE - FLOOR + 1.0, max(2.0, R_OUT - WALL), FLOOR)
    body = body.cut(cavity)

    # 4 torres de tornillos M3 en esquinas
    for sx in (-1.0, 1.0):
        for sy in (-1.0, 1.0):
            body = body.fuse(cyl_z(sx * BOSS_X, sy * BOSS_Y, FLOOR - 0.2, BOSS_R, H_BASE - FLOOR + 0.2))

    # 4 postes soporte para TTGO LoRa32 V2.1
    post_h = TTGO_Z_PCB - FLOOR
    for dx in (-TTGO_L / 2.0 + 3.0, TTGO_L / 2.0 - 3.0):
        for dy in (-TTGO_W / 2.0 + 2.8, TTGO_W / 2.0 - 2.8):
            body = body.fuse(cyl_z(TTGO_CX + dx, TTGO_CY + dy, FLOOR - 0.2, 2.8, post_h + 0.2))

    # Perforaciones M3 esquinas y M2.2 PCB
    for sx in (-1.0, 1.0):
        for sy in (-1.0, 1.0):
            body = body.cut(cyl_z(sx * BOSS_X, sy * BOSS_Y, H_BASE - 11.0, 1.6, 12.0))
    for dx in (-TTGO_L / 2.0 + 3.0, TTGO_L / 2.0 - 3.0):
        for dy in (-TTGO_W / 2.0 + 2.8, TTGO_W / 2.0 - 2.8):
            body = body.cut(cyl_z(TTGO_CX + dx, TTGO_CY + dy, TTGO_Z_PCB - 7.0, 1.15, 8.0))

    # Puerto lateral (+X) para Micro-USB y acceso microSD
    usb_slot = centered_box(L / 2.0, 0.0, 10.5, WALL + 6.0, 22.0, 10.5)
    body = body.cut(usb_slot)

    # Agujero SMA en +Y
    body = body.cut(cyl_y(SMA_X, W / 2.0 - WALL - 3.0, SMA_Z, 3.6, WALL + 10.0))

    # Ranuras de ventilación en fondo + 2 agujeros tipo cerradura (pared)
    for vx in (-18.0, -6.0, 6.0, 18.0):
        vent = centered_box(vx, 0.0, -1.0, 3.0, 22.0, FLOOR + 2.0)
        body = body.cut(vent)
    for kx in (-32.0, 32.0):
        body = body.cut(cyl_z(kx, -4.0, -1.0, 3.5, FLOOR + 2.0))
        body = body.cut(centered_box(kx, 1.5, -1.0, 3.8, 10.0, FLOOR + 2.0))
        body = body.cut(cyl_z(kx, 6.5, -1.0, 1.9, FLOOR + 2.0))

    return body.removeSplitter()


def build_receptor_tapa():
    lid = rounded_box(L, W, LID_T, R_OUT, H_BASE)

    # Labio interno de centrado rápido
    lip_out = rounded_box(L - 2.0 * WALL - 0.6, W - 2.0 * WALL - 0.6, 2.0, max(1.5, R_OUT - WALL - 0.3), H_BASE - 1.8)
    lip_in = rounded_box(L - 2.0 * WALL - 3.6, W - 2.0 * WALL - 3.6, 2.6, max(1.0, R_OUT - WALL - 1.8), H_BASE - 2.1)
    lip = lip_out.cut(lip_in)
    # Despejar esquinas de las torres M3 para que el labio no choque con los bosses
    for sx in (-1.0, 1.0):
        for sy in (-1.0, 1.0):
            lip = lip.cut(cyl_z(sx * BOSS_X, sy * BOSS_Y, H_BASE - 3.0, BOSS_R + 0.8, 4.0))
    lid = lid.fuse(lip)

    # Ventana OLED 0.96"
    lid = lid.cut(centered_box(OLED_WIN_CX, OLED_WIN_CY, H_BASE - 2.0, OLED_WIN_L, OLED_WIN_W, LID_T + 4.0))

    # Agujero de acceso a botón Reset (pinhole Ø3 mm, desplazado a la derecha para despejar el logo SENA)
    lid = lid.cut(cyl_z(TTGO_CX + 34.0, TTGO_CY + 1.0, H_BASE - 2.0, 1.5, LID_T + 4.0))

    # 4 agujeros avellanados M3
    for sx in (-1.0, 1.0):
        for sy in (-1.0, 1.0):
            lid = lid.cut(cyl_z(sx * BOSS_X, sy * BOSS_Y, H_BASE - 2.0, 1.75, LID_T + 4.0))
            lid = lid.cut(cyl_z(sx * BOSS_X, sy * BOSS_Y, H_BASE + LID_T - 1.8, 3.2, 2.5))

    # Vaciado en bajo relieve (0.6 mm): Logo SENA a la derecha del OLED + leyenda centrada despejando tornillos
    mark = branding.make_institutional_mark(
        cx_logo=19.5,
        cy_logo=1.5,
        cx_text=0.0,
        cy_text=-16.0,
        z_top=H_BASE + LID_T,
        logo_size=10.0,
        sena_font=2.4,
        text_font=2.4,
        line_gap=3.6,
        text_align="center",
        max_text_width=62.0,
        depth=0.6,
    )
    lid = lid.cut(mark)

    return lid.removeSplitter()


def build_ref_hardware():
    pcb = centered_box(TTGO_CX, TTGO_CY, TTGO_Z_PCB, TTGO_L, TTGO_W, 1.6)
    oled = centered_box(OLED_WIN_CX, OLED_WIN_CY, TTGO_Z_PCB + 1.6, 27.3, 19.0, 4.0)
    sma_hex = cyl_y(SMA_X, W / 2.0 + 4.0, SMA_Z, 4.2, 8.0)
    antenna = cyl_y(SMA_X, W / 2.0 + 12.0, SMA_Z, 5.0, 52.0)
    return pcb, oled, sma_hex.fuse(antenna)


def add_part(doc, name, shape, color, transparency=0):
    obj = doc.addObject("Part::Feature", name)
    obj.Shape = shape
    if App.GuiUp:
        obj.ViewObject.ShapeColor = color
        obj.ViewObject.Transparency = transparency
    return obj


def main():
    os.makedirs(OUT_DIR, exist_ok=True)
    os.makedirs(RENDER_DIR, exist_ok=True)

    doc_name = "CompostIOT_NodoReceptor"
    if doc_name in App.listDocuments():
        App.closeDocument(doc_name)
    doc = App.newDocument(doc_name)

    base = add_part(doc, "Receptor_Base", build_receptor_base(), (0.16, 0.24, 0.20))
    tapa = add_part(doc, "Receptor_Tapa", build_receptor_tapa(), (0.93, 0.95, 0.94), 15)

    pcb_s, oled_s, ant_s = build_ref_hardware()
    ref_group = doc.addObject("App::DocumentObjectGroup", "Ref_Hardware_NoImprimir")
    ref_pcb = add_part(doc, "Ref_TTGO_LoRa32_PCB", pcb_s, (0.10, 0.12, 0.14))
    ref_oled = add_part(doc, "Ref_OLED_096", oled_s, (0.05, 0.25, 0.45))
    ref_ant = add_part(doc, "Ref_Antena_433MHz", ant_s, (0.15, 0.15, 0.15))
    for r in (ref_pcb, ref_oled, ref_ant):
        ref_group.addObject(r)

    doc.recompute()

    fcstd_path = os.path.join(OUT_DIR, "CompostIOT_NodoReceptor.FCStd")
    doc.saveAs(fcstd_path)

    printable = [base, tapa]
    step_path = os.path.join(OUT_DIR, "CompostIOT_NodoReceptor_Imprimible.step")
    Part.export(printable, step_path)

    for obj in printable:
        stl_path = os.path.join(OUT_DIR, f"{obj.Name}.stl")
        mesh = MeshPart.meshFromShape(Shape=obj.Shape, LinearDeflection=0.1, AngularDeflection=0.2, Relative=False)
        mesh.write(stl_path)

    if App.GuiUp:
        view = Gui.ActiveDocument.ActiveView
        view.viewIsometric()
        view.fitAll()
        view.saveImage(os.path.join(RENDER_DIR, "receptor_ensamblado.png"), 1400, 1000, "White")

        tapa.ViewObject.Transparency = 0
        tapa.Placement.Base.z = 42.0
        ref_pcb.Placement.Base.z = 16.0
        ref_oled.Placement.Base.z = 16.0
        ref_ant.Placement.Base.y = 22.0
        doc.recompute()

        view.viewIsometric()
        view.fitAll()
        view.saveImage(os.path.join(RENDER_DIR, "receptor_explosionado.png"), 1400, 1000, "White")

        for obj in (tapa, ref_pcb, ref_oled, ref_ant):
            obj.Placement.Base = App.Vector(0, 0, 0)
        tapa.ViewObject.Transparency = 15
        doc.recompute()
        doc.save()

    print("OK Nodo Receptor generado:")
    print(" - FCStd:", fcstd_path)
    print(" - STEP :", step_path)
    print(" - STLs :", len(printable), "piezas en", OUT_DIR)


if __name__ == "__main__":
    main()
