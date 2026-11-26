package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	serverURL := flag.String("url", "http://localhost:8080", "URL base del backend CompostIOT")
	deviceKey := flag.String("key", "ciot_demo_gateway_sena_2026", "Clave X-Device-Key del gateway")
	nodeID := flag.String("node", "NODO-01", "ID del nodo emisor a simular")
	withPH := flag.Bool("ph", false, "Incluir lecturas simuladas de sensor de pH")
	interval := flag.Duration("interval", 5*time.Second, "Intervalo entre envíos en modo continuo")
	count := flag.Int("count", 10, "Número de lecturas a enviar (0 = infinito)")
	flag.Parse()

	log.Printf("Simulador LoRa CompostIOT -> %s | nodo=%s | pH=%v | cada %s", *serverURL, *nodeID, *withPH, *interval)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	seq := 1
	for *count == 0 || seq <= *count {
		// Curva realista alrededor de fase termófila con leve variación natural
		temp := 54.0 + 6.5*math.Sin(float64(seq)*0.35) + (rng.Float64()-0.5)*1.2
		hum := 54.0 + 4.0*math.Cos(float64(seq)*0.25) + (rng.Float64()-0.5)*1.0
		temp = math.Round(temp*10) / 10
		hum = math.Round(hum*10) / 10

		payload := map[string]any{
			"node_id":       *nodeID,
			"temperature_c": temp,
			"humidity_pct":  hum,
			"power_source":  "mains",
			"rssi":          -72 - rng.Intn(14),
			"seq":           seq,
		}
		if *withPH {
			phVal := math.Round((6.8+(rng.Float64()-0.5)*0.4)*100) / 100
			payload["ph"] = phVal
		} else {
			payload["ph"] = nil
		}

		body, _ := json.Marshal(payload)
		req, err := http.NewRequest(http.MethodPost, *serverURL+"/api/v1/ingest/readings", bytes.NewReader(body))
		if err != nil {
			log.Fatalf("error construyendo petición: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Device-Key", *deviceKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("[seq=%d] error de red: %v", seq, err)
		} else {
			var out struct {
				Evaluation struct {
					Phase struct {
						Label string `json:"label"`
					} `json:"phase"`
				} `json:"evaluation"`
			}
			_ = json.NewDecoder(resp.Body).Decode(&out)
			resp.Body.Close()
			if resp.StatusCode == http.StatusCreated {
				fmt.Printf("[seq=%d] OK -> T=%.1f°C H=%.1f%% | Fase: %s\n", seq, temp, hum, out.Evaluation.Phase.Label)
			} else {
				log.Printf("[seq=%d] HTTP %d al enviar lectura", seq, resp.StatusCode)
			}
		}

		seq++
		if *count == 0 || seq <= *count {
			time.Sleep(*interval)
		}
	}
}
