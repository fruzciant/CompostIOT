# Modelo de Clasificación de Fases del Compostaje — CompostIOT

El motor de fases traduce las mediciones de **Temperatura (°C)**, **Humedad (%)**, **pH (opcional)** y el **historial térmico de la pila** en una fase comprensible para aprendices del SENA y personal de campo, acompañada de alertas y recomendaciones prácticas.

## 1. Las 4 Fases del Ciclo Aeróbico

| Código | Nombre | Rango térmico típico | Contexto en el ciclo | Color UI |
| :--- | :--- | :--- | :--- | :--- |
| `mesofila_inicial` | **Mesófila inicial** | `20 °C – 44.9 °C` (antes de alcanzar pico termófilo) | Arranque del proceso; microorganismos mesófilos descomponen azúcares y compuestos simples, elevando la temperatura. | Azul / Verde claro (`#3B82F6`) |
| `termofila` | **Termófila (Higienización)** | `≥ 45 °C` (óptimo `55 °C – 65 °C`) | Alta actividad microbiana, eliminación de patógenos y semillas de malezas. Suele observarse vapor al voltear. | Naranja / Ámbar intenso (`#EA580C`) |
| `enfriamiento` | **Enfriamiento (Mesófila II)** | `30 °C – 44.9 °C` **después** de haber superado fase termófila (`peak_temp ≥ 45 °C`) | Disminuye el material fácilmente degradable; hongos y actinomicetos degradan celulosa y lignina. | Amarillo / Ocre (`#CA8A04`) |
| `maduracion` | **Maduración** | `< 30 °C` **después** de fase termófila (cercana a temperatura ambiente, estable) | Estabilización del humus; el abono alcanza madurez apta para uso agrícola. | Verde tierra (`#16A34A`) |
| `sin_datos` | **Sin lecturas** | Sin mediciones aún | Pila recién creada esperando datos del nodo. | Gris (`#64748B`) |

## 2. Regla Determinista de Detección

Para evitar falsos saltos de fase por ruido puntual de una sola lectura:
1. **Temperatura de referencia (`T_ref`)**: promedio móvil de las últimas 3 lecturas válidas (o la lectura actual si hay menos de 3).
2. **Memoria de pico térmico (`has_reached_thermophilic`)**: `true` si la pila ha registrado al menos 2 lecturas `≥ 45.0 °C` desde su fecha de inicio (`started_at`).
3. **Clasificación**:
   - Si no hay lecturas → `sin_datos`
   - Si `T_ref ≥ 45.0` → `termofila`
   - Si `T_ref < 45.0` y `has_reached_thermophilic == false` → `mesofila_inicial`
   - Si `30.0 ≤ T_ref < 45.0` y `has_reached_thermophilic == true` → `enfriamiento`
   - Si `T_ref < 30.0` y `has_reached_thermophilic == true` → `maduracion`

## 3. Alertas y Recomendaciones Agronómicas

El motor evalúa condiciones críticas y devuelve una lista de recomendaciones en lenguaje claro de campo:

| Condición | Severidad | Mensaje sugerido al usuario |
| :--- | :--- | :--- |
| `temperature_c > 70.0` | `danger` | **Sobrecalentamiento (>70 °C)**: Riesgo de inhibir microorganismos benéficos. Realizar volteo inmediato para oxigenar y disipar calor. |
| `temperature_c < 20.0` (en `mesofila_inicial` > 3 días) | `warning` | **Arranque lento**: Revisar humedad, tamaño de pila o relación Carbono/Nitrógeno (agregar material rico en nitrógeno). |
| `humidity_pct < 40.0` | `warning` | **Humedad baja (<40 %)**: La actividad microbiana se detiene. Regar uniformemente la pila durante el próximo volteo (meta: 50–60 %). |
| `humidity_pct > 65.0` | `warning` | **Exceso de humedad (>65 %)**: Riesgo de pudrición anaeróbica y malos olores. Voltear y mezclar material seco estructurante (aserrín, hojarasca, paja). |
| `ph != null && ph < 5.5` | `warning` | **Sustrato muy ácido (pH < 5.5)**: Posible falta de aireación. Voltear la pila; si persiste, evaluar adición moderada de cal agrícola o ceniza. |
| `ph != null && ph > 8.5` | `warning` | **Sustrato muy alcalino (pH > 8.5)**: Riesgo de pérdida de nitrógeno (amoníaco). Incorporar material rico en carbono (hojarasca/aserrín). |
| Sin lecturas > 60 min (pila activa) | `info` | **Nodo sin reportar recientemente**: Verificar alimentación eléctrica del emisor o enlace LoRa del gateway. |
| Todo en rango óptimo | `ok` | **Proceso saludable**: Variables dentro de los rangos esperados para la fase actual. |

## 4. Progreso Visual del Ciclo (UI)

Para facilitar la lectura por estudiantes y agricultores, la interfaz muestra una barra de 4 etapas:
1. `Mesófila inicial` (25%) → 2. `Termófila` (50%) → 3. `Enfriamiento` (75%) → 4. `Maduración` (100%).
