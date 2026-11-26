package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/compostiot/backend/internal/lora"
	"go.bug.st/serial"
)

func main() {
	portName := flag.String("port", "", "Puerto USB-Serial del Nodo Receptor TTGO (ej. /dev/ttyUSB0 o COM3)")
	baud := flag.Int("baud", 115200, "Velocidad en baudios del puerto serial")
	serverURL := flag.String("url", "http://localhost:8080", "URL del backend CompostIOT")
	deviceKey := flag.String("key", "ciot_demo_gateway_sena_2026", "Clave X-Device-Key generada en Ajustes")
	listOnly := flag.Bool("list", false, "Listar puertos seriales disponibles y salir")
	flag.Parse()

	if *listOnly || *portName == "" {
		ports, err := serial.GetPortsList()
		if err != nil {
			log.Fatalf("error listando puertos seriales: %v", err)
		}
		if len(ports) == 0 {
			fmt.Println("No se detectaron puertos USB-Serial conectados.")
		} else {
			fmt.Println("Puertos USB-Serial detectados:")
			for _, p := range ports {
				fmt.Printf("  - %s\n", p)
			}
		}
		if *listOnly {
			return
		}
		if len(ports) == 1 {
			*portName = ports[0]
			log.Printf("Usando único puerto detectado automáticamente: %s", *portName)
		} else {
			log.Fatal("Especifica el puerto con -port (ej. -port /dev/ttyUSB0 o -port COM3)")
		}
	}

	target := strings.TrimRight(*serverURL, "/") + "/api/v1/ingest/readings"
	httpClient := &http.Client{Timeout: 8 * time.Second}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	log.Printf("CompostIOT Puente USB-Serial -> %s (puerto=%s @ %d baud)", target, *portName, *baud)

	go func() {
		for {
			mode := &serial.Mode{BaudRate: *baud}
			port, err := serial.Open(*portName, mode)
			if err != nil {
				log.Printf("Esperando puerto %s (%v)... reintentando en 3s", *portName, err)
				time.Sleep(3 * time.Second)
				continue
			}
			log.Printf("Conectado a %s. Escuchando tramas CIOT|v1 y JSON...", *portName)

			scanner := bufio.NewScanner(port)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}
				pkt, err := lora.ParseLine(line)
				if err != nil {
					// Ignorar logs de arranque del ESP32, pero mostrar si parecía una trama CIOT
					if strings.HasPrefix(line, "CIOT|") || strings.HasPrefix(line, "{") {
						log.Printf("[DESCARTADA] %v | trama=%q", err, line)
					}
					continue
				}
				if err := forwardPacket(httpClient, target, *deviceKey, pkt); err != nil {
					log.Printf("[ERROR ENVÍO] nodo=%s: %v", pkt.NodeID, err)
				} else {
					log.Printf("[OK] nodo=%s T=%.1f°C H=%.1f%% pwr=%s", pkt.NodeID, pkt.TemperatureC, pkt.HumidityPct, pkt.PowerSource)
				}
			}

			_ = port.Close()
			log.Printf("Conexión serial cerrada (%v). Reintentando en 3s...", scanner.Err())
			time.Sleep(3 * time.Second)
		}
	}()

	<-stop
	log.Println("Puente USB-Serial detenido.")
}

func forwardPacket(client *http.Client, endpoint, deviceKey string, pkt lora.Packet) error {
	body, err := json.Marshal(pkt)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-Key", deviceKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("servidor respondió HTTP %d", resp.StatusCode)
	}
	return nil
}
