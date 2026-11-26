"""
Utilidades de grabado en bajo relieve (deboss) para carcasas CompostIOT:
- Isotipo SENA + texto "SENA"
- Leyenda: "Maestría en Educación Sostenible y Gestión Ambiental Territorial"
"""

import FreeCAD as App
import Part

FONT_DIR = "/home/fruzciant/Documents/Proyectos/CompostIOT/hardware-cad/scripts/fonts/"
FONT_FILE = "NotoSans-Bold.ttf"


def _thick_segment(p1, p2, thickness, z0, height):
    """Crea una barra rectangular con extremos redondeados entre p1 y p2 en el plano XY."""
    v = p2.sub(p1)
    length = v.Length
    if length < 1e-4:
        return Part.makeCylinder(thickness / 2.0, height, App.Vector(p1.x, p1.y, z0))

    u = App.Vector(v.x / length, v.y / length, 0)
    n = App.Vector(-u.y * (thickness / 2.0), u.x * (thickness / 2.0), 0)

    c1 = App.Vector(p1.x + n.x, p1.y + n.y, z0)
    c2 = App.Vector(p2.x + n.x, p2.y + n.y, z0)
    c3 = App.Vector(p2.x - n.x, p2.y - n.y, z0)
    c4 = App.Vector(p1.x - n.x, p1.y - n.y, z0)

    poly = Part.makePolygon([c1, c2, c3, c4, c1])
    bar = Part.Face(poly).extrude(App.Vector(0, 0, height))
    cap1 = Part.makeCylinder(thickness / 2.0, height, App.Vector(p1.x, p1.y, z0))
    cap2 = Part.makeCylinder(thickness / 2.0, height, App.Vector(p2.x, p2.y, z0))
    return bar.fuse(cap1).fuse(cap2)


def make_sena_logo_cutout(cx, cy, z_top, size=14.0, depth=0.65):
    """
    Genera el sólido de corte (bajo relieve) del isotipo SENA:
    círculo superior + dos brazos angulados simétricos + chevrón inferior central.
    """
    z0 = z_top - depth
    h = depth + 0.35
    s = float(size)
    t = s * 0.115

    # 1. Círculo superior (cabeza)
    head = Part.makeCylinder(s * 0.115, h, App.Vector(cx, cy + s * 0.34, z0))

    # 2. Brazo izquierdo: horizontal superior -> diagonal inferior izquierda
    l_outer = App.Vector(cx - s * 0.44, cy + s * 0.16, 0)
    l_elbow = App.Vector(cx - s * 0.11, cy + s * 0.16, 0)
    l_foot = App.Vector(cx - s * 0.35, cy - s * 0.36, 0)
    left_arm = _thick_segment(l_outer, l_elbow, t, z0, h).fuse(_thick_segment(l_elbow, l_foot, t, z0, h))

    # 3. Brazo derecho: horizontal superior -> diagonal inferior derecha
    r_outer = App.Vector(cx + s * 0.44, cy + s * 0.16, 0)
    r_elbow = App.Vector(cx + s * 0.11, cy + s * 0.16, 0)
    r_foot = App.Vector(cx + s * 0.35, cy - s * 0.36, 0)
    right_arm = _thick_segment(r_outer, r_elbow, t, z0, h).fuse(_thick_segment(r_elbow, r_foot, t, z0, h))

    # 4. Chevrón inferior central (tercer camino)
    c_apex = App.Vector(cx, cy - s * 0.02, 0)
    c_left = App.Vector(cx - s * 0.19, cy - s * 0.36, 0)
    c_right = App.Vector(cx + s * 0.19, cy - s * 0.36, 0)
    chevron = _thick_segment(c_left, c_apex, t * 0.92, z0, h).fuse(_thick_segment(c_apex, c_right, t * 0.92, z0, h))

    return head.fuse(left_arm).fuse(right_arm).fuse(chevron)


def make_text_cutout(text, cx, cy, z_top, font_size=3.6, depth=0.65, align="center", max_width=None):
    """
    Convierte una cadena UTF-8 (con tildes) en sólidos 3D listos para vaciado en bajo relieve.
    Usa FaceMakerBullseye para preservar huecos internos de letras (o, a, e, ó, í).
    Si max_width está definido y el texto excede ese ancho, escala proporcionalmente en XY para que siempre encaje.
    """
    char_wire_lists = Part.makeWireString(text, FONT_DIR, FONT_FILE, float(font_size), 0.0)
    faces = []

    for wires in char_wire_lists:
        if not wires:
            continue
        try:
            faces.append(Part.makeFace(wires, "Part::FaceMakerBullseye"))
        except Exception:
            for w in wires:
                try:
                    faces.append(Part.Face(w))
                except Exception:
                    pass

    if not faces:
        return None

    comp_2d = Part.makeCompound(faces)
    bb = comp_2d.BoundBox
    if max_width and bb.XLength > max_width:
        scale = float(max_width) / bb.XLength
        m = App.Matrix()
        m.scale(scale, scale, 1.0)
        comp_2d = comp_2d.transformGeometry(m)
        bb = comp_2d.BoundBox

    z0 = z_top - depth
    h = depth + 0.35

    if align == "left":
        dx = cx - bb.XMin
    elif align == "right":
        dx = cx - bb.XMax
    else:
        dx = cx - (bb.XMin + bb.XMax) / 2.0

    dy = cy - (bb.YMin + bb.YMax) / 2.0
    dz = z0 - bb.ZMin
    comp_2d.translate(App.Vector(dx, dy, dz))
    return comp_2d.extrude(App.Vector(0, 0, h))


def make_institutional_mark(
    cx_logo,
    cy_logo,
    cx_text,
    cy_text,
    z_top,
    logo_size=13.5,
    sena_font=3.0,
    text_font=3.5,
    line_gap=4.9,
    text_align="left",
    max_text_width=None,
    depth=0.65,
):
    """
    Construye el compuesto completo de corte para el logo SENA +
    'Maestría en Educación Sostenible' / 'y Gestión Ambiental Territorial'.
    """
    parts = [make_sena_logo_cutout(cx_logo, cy_logo + sena_font * 0.48, z_top, size=logo_size, depth=depth)]

    sena_txt = make_text_cutout(
        "SENA",
        cx_logo,
        cy_logo - logo_size * 0.48,
        z_top,
        font_size=sena_font,
        depth=depth,
        align="center",
        max_width=logo_size * 0.95,
    )
    if sena_txt:
        parts.append(sena_txt)

    # Ambas líneas usan la misma escala (determinada por la línea más larga) para uniformidad tipográfica
    ref_len_1 = len("Maestría en Educación Sostenible")
    ref_len_2 = len("y Gestión Ambiental Territorial")
    w1 = max_text_width
    w2 = max_text_width * (float(ref_len_2) / float(ref_len_1)) if max_text_width else None

    l1 = make_text_cutout(
        "Maestría en Educación Sostenible",
        cx_text,
        cy_text + line_gap / 2.0,
        z_top,
        font_size=text_font,
        depth=depth,
        align=text_align,
        max_width=w1,
    )
    l2 = make_text_cutout(
        "y Gestión Ambiental Territorial",
        cx_text,
        cy_text - line_gap / 2.0,
        z_top,
        font_size=text_font,
        depth=depth,
        align=text_align,
        max_width=w2,
    )
    if l1:
        parts.append(l1)
    if l2:
        parts.append(l2)

    return Part.makeCompound(parts)
