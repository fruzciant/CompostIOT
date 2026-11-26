package lora

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Packet representa una lectura recibida desde el Nodo Receptor TTGO LoRa32 (por trama CIOT o JSON).
type Packet struct {
	NodeID       string   `json:"node_id"`
	Seq          *int     `json:"seq,omitempty"`
	TemperatureC float64  `json:"temperature_c"`
	HumidityPct  float64  `json:"humidity_pct"`
	PH           *float64 `json:"ph"`
	PowerSource  string   `json:"power_source"`
	BatteryV     *float64 `json:"battery_v"`
	RSSI         *int     `json:"rssi,omitempty"`
}

// CRC16CCITT calcula el checksum CRC-16/CCITT-FALSE (polinomio 0x1021, semilla 0xFFFF).
func CRC16CCITT(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// FormatFrame construye una trama CIOT|v1 con CRC-16 lista para pruebas o transmisión.
func FormatFrame(p Packet) string {
	seqVal := 0
	if p.Seq != nil {
		seqVal = *p.Seq
	}
	phStr := "NA"
	if p.PH != nil {
		phStr = fmt.Sprintf("%.2f", *p.PH)
	}
	pwr := "M"
	if p.PowerSource == "battery" {
		pwr = "B"
	}
	battStr := "NA"
	if p.BatteryV != nil {
		battStr = fmt.Sprintf("%.2f", *p.BatteryV)
	}
	body := fmt.Sprintf("CIOT|v1|%s|%d|%.1f|%.1f|%s|%s|%s",
		strings.ToUpper(strings.TrimSpace(p.NodeID)),
		seqVal,
		p.TemperatureC,
		p.HumidityPct,
		phStr,
		pwr,
		battStr,
	)
	crc := CRC16CCITT([]byte(body))
	return fmt.Sprintf("%s*%04X", body, crc)
}

// ParseLine acepta tanto una línea JSON enviada por el Gateway como una trama cruda CIOT|v1|...*CRC16.
func ParseLine(line string) (Packet, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return Packet{}, errors.New("línea vacía")
	}

	// Modo 1: línea JSON directa del Gateway ESP32
	if strings.HasPrefix(trimmed, "{") {
		var p Packet
		if err := json.Unmarshal([]byte(trimmed), &p); err != nil {
			return Packet{}, fmt.Errorf("JSON inválido: %w", err)
		}
		p.NodeID = strings.ToUpper(strings.TrimSpace(p.NodeID))
		if p.NodeID == "" {
			return Packet{}, errors.New("JSON sin node_id")
		}
		if p.PowerSource != "battery" {
			p.PowerSource = "mains"
		}
		return p, nil
	}

	// Modo 2: trama compacta CIOT|v1|<NODE>|<SEQ>|<TEMP>|<HUM>|<PH|NA>|<M|B>|<BATT|NA>*<CRC16>
	// También soporta sufijo opcional "|RSSI=-78" añadido por el receptor tras el CRC.
	var rssiPtr *int
	if idx := strings.Index(trimmed, " RSSI="); idx != -1 {
		if rssiVal, err := strconv.Atoi(strings.TrimSpace(trimmed[idx+6:])); err == nil {
			rssiPtr = &rssiVal
		}
		trimmed = strings.TrimSpace(trimmed[:idx])
	}

	starIdx := strings.LastIndex(trimmed, "*")
	if starIdx == -1 || starIdx+5 > len(trimmed) {
		return Packet{}, errors.New("trama sin checksum *CRC16")
	}
	body := trimmed[:starIdx]
	crcHex := strings.ToUpper(strings.TrimSpace(trimmed[starIdx+1:]))
	expectedCRC := fmt.Sprintf("%04X", CRC16CCITT([]byte(body)))
	if crcHex != expectedCRC {
		return Packet{}, fmt.Errorf("CRC16 inválido: recibió %s, esperaba %s", crcHex, expectedCRC)
	}

	parts := strings.Split(body, "|")
	if len(parts) != 9 || parts[0] != "CIOT" || parts[1] != "v1" {
		return Packet{}, errors.New("formato de trama CIOT|v1 inválido")
	}

	nodeID := strings.ToUpper(strings.TrimSpace(parts[2]))
	if nodeID == "" {
		return Packet{}, errors.New("node_id vacío")
	}
	seqVal, err := strconv.Atoi(parts[3])
	if err != nil {
		return Packet{}, fmt.Errorf("seq inválido: %w", err)
	}
	temp, err := strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return Packet{}, fmt.Errorf("temperatura inválida: %w", err)
	}
	hum, err := strconv.ParseFloat(parts[5], 64)
	if err != nil {
		return Packet{}, fmt.Errorf("humedad inválida: %w", err)
	}

	var phPtr *float64
	if parts[6] != "NA" && parts[6] != "" {
		phVal, err := strconv.ParseFloat(parts[6], 64)
		if err != nil {
			return Packet{}, fmt.Errorf("pH inválido: %w", err)
		}
		phPtr = &phVal
	}

	pwr := "mains"
	if strings.ToUpper(parts[7]) == "B" {
		pwr = "battery"
	}

	var battPtr *float64
	if parts[8] != "NA" && parts[8] != "" {
		battVal, err := strconv.ParseFloat(parts[8], 64)
		if err != nil {
			return Packet{}, fmt.Errorf("voltaje de batería inválido: %w", err)
		}
		battPtr = &battVal
	}

	return Packet{
		NodeID:       nodeID,
		Seq:          &seqVal,
		TemperatureC: temp,
		HumidityPct:  hum,
		PH:           phPtr,
		PowerSource:  pwr,
		BatteryV:     battPtr,
		RSSI:         rssiPtr,
	}, nil
}
