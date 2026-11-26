"""
CompostIOT — Generador CAD Paramétrico: Nodo Emisor de Intemperie (LILYGO TTGO LoRa32 V2.1)
Diseñado para impresión 3D en PETG o ASA (exterior) + TPU 95A (empaque).

Orientación de montaje en campo (vertical sobre estaca/poste):
  +Y : Arriba (salida de antena SMA 433 MHz + visera superior)
  -Y : Abajo  (4 prensaestopas orientados hacia el suelo: 5V, Temp, Hum, pH opcional)
  +Z : Frente (tapa sellada con visor OLED y visera solar/lluvia)
  -Z : Fondo  (orejas de sujeción con agujeros M5 y ranuras para amarres/abrazaderas)
"""

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

# ==============================================================================
# PARÁMETROS EDITABLES (ajustar cuando se confirmen dimensiones finales de sensores)
# ==============================================================================
OUT_DIR = "/home/fruzciant/Documents/Proyectos/CompostIOT/hardware-cad/emisor"
RENDER_DIR = "/home/fruzciant/Documents/Proyectos/CompostIOT/hardware-cad/renders"

# Caja principal (mm)
L = 136.0            # Largo exterior (X)
W = 92.0             # Alto exterior en montaje vertical (Y)
H_BASE = 46.0        # Profundidad de la base (Z)
WALL = 3.2           # Espesor de pared principal (PETG/ASA)
FLOOR = 3.2          # Espesor de fondo
R_OUT = 8.0          # Radio de esquinas exteriores

# Labio superior para canal de empaque
RIM_W = 6.0          # Ancho total del borde superior (permite canal de empaque robusto)
RIM_H = 5.5          # Altura del refuerzo interno superior
GROOVE_W = 2.0       # Ancho del canal del empaque
GROOVE_D = 2.2       # Profundidad del canal del empaque
GROOVE_OFFSET = 3.0  # Distancia desde cara exterior al centro del canal

# Tornillería tapa (4x M3)
BOSS_X = L / 2.0 - RIM_W - 2.5   # 59.5 mm
BOSS_Y = W / 2.0 - RIM_W - 2.5   # 37.5 mm
BOSS_R = 4.4
M3_INSERT_R = 1.65   # Rosca directa M3 o inserto M3 (ajustar a 2.05 si usas inserto brass M3x5)
M3_CLEAR_R = 1.75
M3_HEAD_R = 3.3

# Prensacables inferiores (-Y)
GLAND_Z = 21.0
GLAND_5V_X = -42.0   # PG9 toma eléctrica 5V
GLAND_5V_R = 7.75    # Ø15.5 mm
GLAND_T_X = -14.0    # PG7 sonda temperatura (obligatorio)
GLAND_H_X = 14.0     # PG7 sonda humedad (obligatorio)
GLAND_PH_X = 42.0    # PG7 sonda pH (opcional / tapón ciego)
GLAND_PG7_R = 6.25   # Ø12.5 mm

# Salida antena SMA (+Y)
SMA_X = 18.0
SMA_Z = 24.0
SMA_BORE_R = 3.6     # Ø7.2 mm (SMA bulkhead / pasamuros)

# TTGO LoRa32 V2.1 (65 x 27 mm)
TTGO_CX = -4.0
TTGO_CY = 18.0
TTGO_L = 65.0
TTGO_W = 27.0
TTGO_Z_PCB = 30.0    # PCB elevado para que el OLED quede cerca de la ventana frontal

# Bahía modular inferior (5V bornera / futura batería 18650)
BAY_CX = 0.0
BAY_CY = -19.0
BAY_L = 80.0
BAY_W = 24.0

# Tapa frontal
LID_T = 4.2
OLED_WIN_CX = TTGO_CX - 6.0
OLED_WIN_CY = TTGO_CY
OLED_WIN_L = 28.0
OLED_WIN_W = 16.0


# ==============================================================================
# UTILIDADES GEOMÉTRICAS ROBUSTAS
# ==============================================================================
def make_rounded_rect_face(length, width, radius, z=0.0):
    """Crea una cara rectangular con esquinas redondeadas centrada en (0, 0, z)."""
    hx = length / 2.0
    hy = width / 2.0
    r = max(0.2, min(radius, hx - 0.1, hy - 0.1))

    p1 = App.Vector(-hx + r, -hy, z)
    p2 = App.Vector(hx - r, -hy, z)
    p3 = App.Vector(hx, -hy + r, z)
    p4 = App.Vector(hx, hy - r, z)
    p5 = App.Vector(hx - r, hy, z)
    p6 = App.Vector(-hx + r, hy, z)
    p7 = App.Vector(-hx, hy - r, z)
    p8 = App.Vector(-hx, -hy + r, z)

    import math
    k = r * (1.0 - math.sqrt(2.0) / 2.0)
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
    wire = Part.Wire(edges)
    return Part.Face(wire)


def rounded_box(length, width, height, radius, z0=0.0):
    face = make_rounded_rect_face(length, width, radius, z0)
    return face.extrude(App.Vector(0, 0, height))


def centered_box(cx, cy, z0, lx, ly, lz):
    return Part.makeBox(lx, ly, lz, App.Vector(cx - lx / 2.0, cy - ly / 2.0, z0))


def cyl_z(cx, cy, z0, radius, height):
    return Part.makeCylinder(radius, height, App.Vector(cx, cy, z0), App.Vector(0, 0, 1))


def cyl_y(cx, y0, cz, radius, length):
    return Part.makeCylinder(radius, length, App.Vector(cx, y0, cz), App.Vector(0, 1, 0))


# ==============================================================================
# CONSTRUCCIÓN DE PIEZAS
# ==============================================================================
def build_emisor_base():
    # 1. Cuerpo principal redondeado
    outer = rounded_box(L, W, H_BASE, R_OUT, 0.0)

    # 2. Orejas laterales de montaje a estaca/poste (X = +/- L/2)
    ear_ext = 18.0
    ear_y = 62.0
    ear_t = 6.0
    flange = rounded_box(L + 2.0 * ear_ext, ear_y, ear_t, 5.0, 0.0)
    body = outer.fuse(flange)

    # 3. Collarín superior antigoteo para antena SMA (+Y)
    sma_boss = cyl_y(SMA_X, W / 2.0 - 2.0, SMA_Z, 7.2, 7.0)
    body = body.fuse(sma_boss)

    # 4. Vaciar cavidad interior dejando borde superior más grueso (RIM_W) para empaque
    inner_lower = rounded_box(
        L - 2.0 * WALL, W - 2.0 * WALL, H_BASE - FLOOR - RIM_H, max(2.0, R_OUT - WALL), FLOOR
    )
    inner_upper = rounded_box(
        L - 2.0 * RIM_W, W - 2.0 * RIM_W, RIM_H + 2.0, max(2.0, R_OUT - RIM_W), H_BASE - RIM_H - 0.5
    )
    body = body.cut(inner_lower).cut(inner_upper)

    # 5. Canal perimetral para empaque TPU/silicona en la cara superior (Z = H_BASE)
    g_out_l = L - 2.0 * (GROOVE_OFFSET - GROOVE_W / 2.0)
    g_out_w = W - 2.0 * (GROOVE_OFFSET - GROOVE_W / 2.0)
    g_in_l = L - 2.0 * (GROOVE_OFFSET + GROOVE_W / 2.0)
    g_in_w = W - 2.0 * (GROOVE_OFFSET + GROOVE_W / 2.0)
    groove_ring = rounded_box(g_out_l, g_out_w, GROOVE_D + 0.5, max(1.5, R_OUT - 2.0), H_BASE - GROOVE_D).cut(
        rounded_box(g_in_l, g_in_w, GROOVE_D + 1.5, max(1.0, R_OUT - 4.0), H_BASE - GROOVE_D - 0.5)
    )
    body = body.cut(groove_ring)

    # 6. Cuatro torres M3 en esquinas internas
    for sx in (-1.0, 1.0):
        for sy in (-1.0, 1.0):
            bx = sx * BOSS_X
            by = sy * BOSS_Y
            boss = cyl_z(bx, by, FLOOR - 0.2, BOSS_R, H_BASE - FLOOR + 0.2)
            body = body.fuse(boss)

    # 7. Soportes internos para placa TTGO LoRa32 V2.1 (4 torres + tope lateral)
    ttgo_post_h = TTGO_Z_PCB - FLOOR
    for dx in (-TTGO_L / 2.0 + 3.0, TTGO_L / 2.0 - 3.0):
        for dy in (-TTGO_W / 2.0 + 2.8, TTGO_W / 2.0 - 2.8):
            post = cyl_z(TTGO_CX + dx, TTGO_CY + dy, FLOOR - 0.2, 3.0, ttgo_post_h + 0.2)
            body = body.fuse(post)
    # Rieles guía en los bordes cortos del PCB
    for sx in (-1.0, 1.0):
        rail = centered_box(
            TTGO_CX + sx * (TTGO_L / 2.0 + 1.4), TTGO_CY, TTGO_Z_PCB - 3.0, 2.4, 18.0, 6.5
        )
        body = body.fuse(rail)

    # 8. Dos postes M3 para bandeja modular inferior (bornera 5V / futura 18650)
    for sx in (-1.0, 1.0):
        bay_post = cyl_z(BAY_CX + sx * 34.0, BAY_CY, FLOOR - 0.2, 4.2, 8.0)
        body = body.fuse(bay_post)

    # 9. Perforaciones M3 en esquinas, TTGO y bandeja modular
    for sx in (-1.0, 1.0):
        for sy in (-1.0, 1.0):
            hole = cyl_z(sx * BOSS_X, sy * BOSS_Y, H_BASE - 13.0, M3_INSERT_R, 14.0)
            body = body.cut(hole)

    for dx in (-TTGO_L / 2.0 + 3.0, TTGO_L / 2.0 - 3.0):
        for dy in (-TTGO_W / 2.0 + 2.8, TTGO_W / 2.0 - 2.8):
            hole = cyl_z(TTGO_CX + dx, TTGO_CY + dy, TTGO_Z_PCB - 8.0, 1.15, 9.0)
            body = body.cut(hole)

    for sx in (-1.0, 1.0):
        hole = cyl_z(BAY_CX + sx * 34.0, BAY_CY, FLOOR + 1.0, M3_INSERT_R, 11.0)
        body = body.cut(hole)

    # 10. Agujeros de prensaestopas en cara inferior (-Y)
    glands = [
        (GLAND_5V_X, GLAND_5V_R),
        (GLAND_T_X, GLAND_PG7_R),
        (GLAND_H_X, GLAND_PG7_R),
        (GLAND_PH_X, GLAND_PG7_R),
    ]
    for gx, gr in glands:
        bore = cyl_y(gx, -W / 2.0 - 2.0, GLAND_Z, gr, WALL + 6.0)
        body = body.cut(bore)

    # 11. Agujero SMA en cara superior (+Y)
    sma_bore = cyl_y(SMA_X, W / 2.0 - WALL - 3.0, SMA_Z, SMA_BORE_R, WALL + 12.0)
    body = body.cut(sma_bore)

    # 12. Perforaciones y ranuras para amarres en orejas laterales
    for sx in (-1.0, 1.0):
        ex = sx * (L / 2.0 + ear_ext / 2.0)
        for sy in (-1.0, 1.0):
            body = body.cut(cyl_z(ex, sy * 20.0, -1.0, 2.6, ear_t + 2.0))
        slot = centered_box(ex, 0.0, -1.0, 4.5, 16.0, ear_t + 2.0)
        body = body.cut(slot)

    return body.removeSplitter()


def build_emisor_tapa():
    # 1. Placa principal de tapa
    lid = rounded_box(L, W, LID_T, R_OUT, H_BASE)

    # 2. Lengüeta perimetral de compresión hacia el canal de empaque
    t_w = 1.6
    t_out_l = L - 2.0 * (GROOVE_OFFSET - t_w / 2.0)
    t_out_w = W - 2.0 * (GROOVE_OFFSET - t_w / 2.0)
    t_in_l = L - 2.0 * (GROOVE_OFFSET + t_w / 2.0)
    t_in_w = W - 2.0 * (GROOVE_OFFSET + t_w / 2.0)
    tongue = rounded_box(t_out_l, t_out_w, 1.8, max(1.5, R_OUT - 2.2), H_BASE - 1.6).cut(
        rounded_box(t_in_l, t_in_w, 2.4, max(1.0, R_OUT - 3.8), H_BASE - 1.9)
    )
    lid = lid.fuse(tongue)

    # 3. Visera protectora superior contra sol y lluvia (sobre el borde +Y, rasante y robusta)
    visor_l = 104.0
    visor_depth = 15.0
    visor_top = centered_box(0.0, W / 2.0 - 4.0, H_BASE + LID_T - 0.5, visor_l, 4.0, visor_depth + 0.5)
    visor_left = centered_box(-visor_l / 2.0 + 2.0, W / 2.0 - 12.0, H_BASE + LID_T - 0.5, 4.0, 18.0, visor_depth + 0.5)
    visor_right = centered_box(visor_l / 2.0 - 2.0, W / 2.0 - 12.0, H_BASE + LID_T - 0.5, 4.0, 18.0, visor_depth + 0.5)
    lid = lid.fuse(visor_top).fuse(visor_left).fuse(visor_right)

    # 4. Ventana OLED + rebaje interno para acrílico/policarbonato transparente
    win_cut = centered_box(OLED_WIN_CX, OLED_WIN_CY, H_BASE - 2.0, OLED_WIN_L, OLED_WIN_W, LID_T + 4.0)
    lens_pocket = centered_box(OLED_WIN_CX, OLED_WIN_CY, H_BASE - 0.5, OLED_WIN_L + 6.0, OLED_WIN_W + 6.0, 2.3)
    lid = lid.cut(win_cut).cut(lens_pocket)

    # 5. Cuatro agujeros avellanados M3 en esquinas
    for sx in (-1.0, 1.0):
        for sy in (-1.0, 1.0):
            bx = sx * BOSS_X
            by = sy * BOSS_Y
            lid = lid.cut(cyl_z(bx, by, H_BASE - 2.0, M3_CLEAR_R, LID_T + 4.0))
            lid = lid.cut(cyl_z(bx, by, H_BASE + LID_T - 2.2, M3_HEAD_R, 3.0))

    # 6. Vaciado en bajo relieve (0.65 mm): Logo SENA + Maestría en Educación Sostenible y Gestión Ambiental Territorial
    mark = branding.make_institutional_mark(
        cx_logo=-42.0,
        cy_logo=-19.0,
        cx_text=-30.0,
        cy_text=-19.0,
        z_top=H_BASE + LID_T,
        logo_size=13.0,
        sena_font=3.0,
        text_font=3.4,
        line_gap=5.2,
        text_align="left",
        max_text_width=78.0,
        depth=0.65,
    )
    lid = lid.cut(mark)

    return lid.removeSplitter()


def build_emisor_empaque():
    g_w = 1.8
    g_out_l = L - 2.0 * (GROOVE_OFFSET - g_w / 2.0)
    g_out_w = W - 2.0 * (GROOVE_OFFSET - g_w / 2.0)
    g_in_l = L - 2.0 * (GROOVE_OFFSET + g_w / 2.0)
    g_in_w = W - 2.0 * (GROOVE_OFFSET + g_w / 2.0)
    gasket = rounded_box(g_out_l, g_out_w, 2.2, max(1.5, R_OUT - 2.1), H_BASE - 2.0).cut(
        rounded_box(g_in_l, g_in_w, 3.0, max(1.0, R_OUT - 3.9), H_BASE - 2.4)
    )
    return gasket


def build_tapon_ph():
    """Tapón ciego imprimible para sellar el puerto PG7 de pH cuando la pila no usa sonda de pH."""
    cap = cyl_y(GLAND_PH_X, -W / 2.0 - 3.0, GLAND_Z, GLAND_PG7_R + 3.0, 3.2)
    shank = cyl_y(GLAND_PH_X, -W / 2.0 - 0.2, GLAND_Z, GLAND_PG7_R - 0.2, 9.0)
    return cap.fuse(shank)


def build_soporte_modular_18650():
    """Bandeja modular inferior: soporta bornera 5V hoy o portapilas 1x18650 en el futuro."""
    z0 = FLOOR + 7.8
    plate = centered_box(BAY_CX, BAY_CY, z0, BAY_L, BAY_W, 2.8)
    wall_f = centered_box(BAY_CX, BAY_CY + BAY_W / 2.0 - 1.3, z0 + 2.6, BAY_L - 12.0, 2.6, 13.0)
    wall_b = centered_box(BAY_CX, BAY_CY - BAY_W / 2.0 + 1.3, z0 + 2.6, BAY_L - 12.0, 2.6, 13.0)
    tray = plate.fuse(wall_f).fuse(wall_b)
    for sx in (-1.0, 1.0):
        tray = tray.cut(cyl_z(BAY_CX + sx * 34.0, BAY_CY, z0 - 1.0, 1.75, 5.0))
    return tray.removeSplitter()


def build_ref_hardware():
    """Sólidos de referencia visual (TTGO LoRa32 V2.1 + antena 433 MHz + celda 18650 opcional)."""
    pcb = centered_box(TTGO_CX, TTGO_CY, TTGO_Z_PCB, TTGO_L, TTGO_W, 1.6)
    oled = centered_box(OLED_WIN_CX, OLED_WIN_CY, TTGO_Z_PCB + 1.6, 27.3, 19.0, 4.0)
    sma_hex = cyl_y(SMA_X, W / 2.0 + 5.0, SMA_Z, 4.2, 8.0)
    antenna = cyl_y(SMA_X, W / 2.0 + 13.0, SMA_Z, 5.0, 52.0)
    cell_18650 = Part.makeCylinder(
        9.2, 65.0, App.Vector(BAY_CX - 32.5, BAY_CY, FLOOR + 10.6 + 9.2), App.Vector(1, 0, 0)
    )
    return pcb, oled, sma_hex.fuse(antenna), cell_18650


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

    doc_name = "CompostIOT_NodoEmisor"
    if doc_name in App.listDocuments():
        App.closeDocument(doc_name)
    doc = App.newDocument(doc_name)

    base = add_part(doc, "Emisor_Base", build_emisor_base(), (0.12, 0.40, 0.25))          # Verde SENA exterior
    tapa = add_part(doc, "Emisor_Tapa", build_emisor_tapa(), (0.92, 0.94, 0.93), 15)      # Gris claro solar
    empaque = add_part(doc, "Emisor_Empaque_TPU", build_emisor_empaque(), (0.85, 0.37, 0.05))  # Naranja técnico
    tapon = add_part(doc, "Emisor_Tapon_pH", build_tapon_ph(), (0.85, 0.37, 0.05))
    soporte = add_part(doc, "Emisor_Soporte_18650", build_soporte_modular_18650(), (0.25, 0.32, 0.30))

    pcb_s, oled_s, ant_s, cell_s = build_ref_hardware()
    ref_group = doc.addObject("App::DocumentObjectGroup", "Ref_Hardware_NoImprimir")
    ref_pcb = add_part(doc, "Ref_TTGO_LoRa32_PCB", pcb_s, (0.10, 0.12, 0.14))
    ref_oled = add_part(doc, "Ref_OLED_096", oled_s, (0.05, 0.25, 0.45))
    ref_ant = add_part(doc, "Ref_Antena_433MHz", ant_s, (0.15, 0.15, 0.15))
    ref_cell = add_part(doc, "Ref_Celda_18650_Opcional", cell_s, (0.20, 0.65, 0.35), 25)
    for r in (ref_pcb, ref_oled, ref_ant, ref_cell):
        ref_group.addObject(r)

    doc.recompute()

    # Guardar documento paramétrico .FCStd
    fcstd_path = os.path.join(OUT_DIR, "CompostIOT_NodoEmisor.FCStd")
    doc.saveAs(fcstd_path)

    # Exportar STEP conjunto imprimible y STLs individuales
    printable = [base, tapa, empaque, tapon, soporte]
    step_path = os.path.join(OUT_DIR, "CompostIOT_NodoEmisor_Imprimible.step")
    Part.export(printable, step_path)

    for obj in printable:
        stl_path = os.path.join(OUT_DIR, f"{obj.Name}.stl")
        mesh = MeshPart.meshFromShape(Shape=obj.Shape, LinearDeflection=0.1, AngularDeflection=0.2, Relative=False)
        mesh.write(stl_path)

    # Renderizar vista ensamblada y vista explosionada tipo LEGO para Typst
    if App.GuiUp:
        view = Gui.ActiveDocument.ActiveView
        view.viewIsometric()
        view.fitAll()
        view.saveImage(os.path.join(RENDER_DIR, "emisor_ensamblado.png"), 1400, 1000, "White")

        # Vista explosionada tipo LEGO (separación amplia para ver todas las piezas sin solape)
        tapa.ViewObject.Transparency = 0
        tapa.Placement.Base.z = 92.0
        empaque.Placement.Base.z = 62.0
        ref_pcb.Placement.Base.z = 32.0
        ref_oled.Placement.Base.z = 32.0
        soporte.Placement.Base.z = 22.0
        ref_cell.Placement.Base.z = 44.0
        ref_ant.Placement.Base.y = 26.0
        tapon.Placement.Base.y = -24.0
        doc.recompute()

        view.viewIsometric()
        view.fitAll()
        view.saveImage(os.path.join(RENDER_DIR, "emisor_explosionado.png"), 1400, 1000, "White")

        # Restaurar posición ensamblada y guardar
        for obj in (tapa, empaque, ref_pcb, ref_oled, soporte, ref_cell, ref_ant, tapon):
            obj.Placement.Base = App.Vector(0, 0, 0)
        tapa.ViewObject.Transparency = 15
        doc.recompute()
        doc.save()

    print("OK Nodo Emisor generado:")
    print(" - FCStd:", fcstd_path)
    print(" - STEP :", step_path)
    print(" - STLs :", len(printable), "piezas en", OUT_DIR)
    print(" - Renders en", RENDER_DIR)


if __name__ == "__main__":
    main()
