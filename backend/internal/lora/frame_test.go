package lora

import "testing"

func TestFormatAndParseFrame(t *testing.T) {
	seq := 42
	ph := 7.15
	batt := 3.92

	t.Run("trama toma 5V sin pH", func(t *testing.T) {
		raw := FormatFrame(Packet{
			NodeID:       "NODO-01",
			Seq:          &seq,
			TemperatureC: 58.4,
			HumidityPct:  54.2,
			PowerSource:  "mains",
		})
		got, err := ParseLine(raw + " RSSI=-76")
		if err != nil {
			t.Fatalf("ParseLine falló: %v (raw=%s)", err, raw)
		}
		if got.NodeID != "NODO-01" || got.TemperatureC != 58.4 || got.HumidityPct != 54.2 || got.PH != nil || got.PowerSource != "mains" {
			t.Fatalf("paquete inesperado: %+v", got)
		}
		if got.RSSI == nil || *got.RSSI != -76 {
			t.Fatalf("RSSI esperado -76, obtuvo %v", got.RSSI)
		}
	})

	t.Run("trama futura batería 18650 con pH", func(t *testing.T) {
		raw := FormatFrame(Packet{
			NodeID:       "NODO-02",
			Seq:          &seq,
			TemperatureC: 61.2,
			HumidityPct:  51.0,
			PH:           &ph,
			PowerSource:  "battery",
			BatteryV:     &batt,
		})
		got, err := ParseLine(raw)
		if err != nil {
			t.Fatalf("ParseLine falló: %v", err)
		}
		if got.PH == nil || *got.PH != 7.15 || got.BatteryV == nil || *got.BatteryV != 3.92 || got.PowerSource != "battery" {
			t.Fatalf("paquete inesperado: %+v", got)
		}
	})

	t.Run("rechaza CRC corrupto", func(t *testing.T) {
		if _, err := ParseLine("CIOT|v1|NODO-01|1|50.0|50.0|NA|M|NA*0000"); err == nil {
			t.Fatal("debía rechazar trama con CRC16 inválido")
		}
	})

	t.Run("acepta línea JSON del gateway", func(t *testing.T) {
		got, err := ParseLine(`{"node_id":"nodo-03","temperature_c":49.5,"humidity_pct":56.0,"ph":null,"rssi":-81}`)
		if err != nil {
			t.Fatalf("ParseLine JSON falló: %v", err)
		}
		if got.NodeID != "NODO-03" || got.TemperatureC != 49.5 || got.PowerSource != "mains" {
			t.Fatalf("paquete JSON inesperado: %+v", got)
		}
	})
}
