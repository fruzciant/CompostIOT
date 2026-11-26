# Protocolo LoRa 433 MHz y Arquitectura Hardware — CompostIOT

## 1. Hardware Confirmado

- **Placas (×2)**: **LILYGO TTGO LoRa32 V2.1** (ESP32, OLED SSD1306 0.96" I2C, LoRa **SX1278 433 MHz**, USB **CH9102F**, ranura microSD, conector SMA y circuito cargador LiPo).
- **Alimentación del Nodo Emisor**:
  - **Configuración principal**: Toma eléctrica cercana (adaptador 5 V DC regulado vía cable con prensaestopa estanco).
  - **Previsión futura**: Soporte mecánico en CAD y lectura opcional de voltaje (divisor resistivo en GPIO ADC) para migrar a batería **18650** + modo *deep sleep* sin rediseñar el protocolo ni el backend.
- **Sensores**:
  - Temperatura: **obligatorio** *(modelo exacto por confirmar antes de Fase 5)*.
  - Humedad: **obligatorio** *(modelo exacto por confirmar antes de Fase 5)*.
  - pH: **opcional** *(habilitable por constante/configuración de nodo; si no está conectado se envía `null` / flag apagado)*.

## 2. Parámetros Radio LoRa (SX1278 — 433 MHz)

| Parámetro | Valor inicial | Motivo |
| :--- | :--- | :--- |
| Frecuencia | `433.0 MHz` | Banda ISM soportada por módulo SX1278 |
| Spreading Factor (SF) | `SF9` | Buen equilibrio entre alcance rural (>1 km con línea de vista) y tiempo en aire |
| Bandwidth (BW) | `125 kHz` | Estándar estable con alta sensibilidad |
| Coding Rate (CR) | `4/5` | Corrección de errores ligera |
| Sync Word | `0x34` (o `0xC1` privado CompostIOT) | Evita interferencia con otros dispositivos 433 MHz |
| Intervalo de envío | `300 s` (5 min, configurable; `30 s` en modo prueba) | Suficiente resolución térmica para compostaje sin saturar almacenamiento |

## 3. Formato de Trama LoRa (Compacto, Legible y Verificable)

Para depuración rápida en entornos educativos SENA (incluso visible por monitor serial) y robustez frente a paquetes corruptos, usamos una trama ASCII compacta con checksum **CRC-16/CCITT**:

```text
CIOT|v1|<NODE_ID>|<SEQ>|<TEMP_C>|<HUM_PCT>|<PH_OR_NA>|<PWR>|<BATT_V_OR_NA>*<CRC16_HEX>
```

### Ejemplo con toma eléctrica y sin sensor de pH:
```text
CIOT|v1|NODO-01|142|58.4|54.2|NA|M|NA*9A3F
```

### Ejemplo futuro con batería 18650 y sensor de pH:
```text
CIOT|v1|NODO-01|143|59.1|53.8|7.12|B|3.94*4C12
```

### Campos
| Campo | Tipo | Descripción |
| :--- | :--- | :--- |
| `CIOT` | Constante | Prefijo identificador del protocolo CompostIOT |
| `v1` | Constante | Versión de trama |
| `NODE_ID` | String (`A-Z0-9-`) | Identificador único del nodo emisor (ej. `NODO-01`) |
| `SEQ` | Entero `0..65535` | Contador incremental para detectar paquetes perdidos |
| `TEMP_C` | Float (1 dec.) | Temperatura en °C (rango válido: `-10.0` a `110.0`) |
| `HUM_PCT` | Float (1 dec.) | Humedad relativa/sustrato en % (rango válido: `0.0` a `100.0`) |
| `PH_OR_NA` | Float (2 dec.) o `NA` | Valor de pH (`0.00` a `14.00`) o `NA` si no hay sonda conectada |
| `PWR` | `M` o `B` | `M` = Toma eléctrica (`mains`), `B` = Batería (`battery`) |
| `BATT_V_OR_NA` | Float (2 dec.) o `NA` | Voltaje de batería 18650 en voltios o `NA` |
| `CRC16_HEX` | 4 hex | Checksum CRC-16 sobre todo el contenido antes del `*` |

## 4. Respaldo y Reenvío en el Gateway (Nodo Receptor)

1. El **Nodo Receptor** valida el prefijo `CIOT|v1` y el `CRC16`.
2. Muestra en su pantalla OLED: `NODE_ID`, `T`, `H`, `pH` (si aplica), `RSSI` y estado del envío (`OK` / `RETRY`).
3. Traduce la trama al JSON de `POST /api/v1/ingest/readings` y la envía por WiFi con cabecera `X-Device-Key`.
4. **Modo Localhost sin WiFi**: también imprime cada trama validada como JSON por el puerto USB Serial (`115200 baud`), permitiendo usar un pequeño puente serial en el PC cuando el laboratorio no disponga de red WiFi.
5. **Respaldo microSD**: ambos nodos pueden registrar un archivo CSV local en la ranura microSD integrada de la TTGO LoRa32 V2.1 como respaldo científico ante cortes de energía o red.
