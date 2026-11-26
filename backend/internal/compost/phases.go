package compost

import (
	"math"
	"time"
)

type PhaseCode string

const (
	PhaseSinDatos        PhaseCode = "sin_datos"
	PhaseMesofilaInicial PhaseCode = "mesofila_inicial"
	PhaseTermofila       PhaseCode = "termofila"
	PhaseEnfriamiento    PhaseCode = "enfriamiento"
	PhaseMaduracion      PhaseCode = "maduracion"
)

type PhaseInfo struct {
	Code        PhaseCode `json:"code"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	ProgressPct int       `json:"progress_pct"`
}

type Alert struct {
	Code     string `json:"code"`
	Severity string `json:"severity"` // ok | info | warning | danger
	Title    string `json:"title"`
	Message  string `json:"message"`
}

type ReadingSample struct {
	TemperatureC float64
	HumidityPct  float64
	PH           *float64
	RecordedAt   time.Time
}

type Evaluation struct {
	Phase                  PhaseInfo `json:"phase"`
	RefTemperatureC        *float64  `json:"ref_temperature_c"`
	HasReachedThermophilic bool      `json:"has_reached_thermophilic"`
	PeakTemperatureC       *float64  `json:"peak_temperature_c"`
	DaysElapsed            int       `json:"days_elapsed"`
	Alerts                 []Alert   `json:"alerts"`
}

func InfoForPhase(code PhaseCode) PhaseInfo {
	switch code {
	case PhaseMesofilaInicial:
		return PhaseInfo{
			Code:        PhaseMesofilaInicial,
			Label:       "Mesófila inicial",
			Description: "Arranque del proceso: microorganismos mesófilos descomponen compuestos simples y elevan la temperatura.",
			Color:       "#3B82F6",
			ProgressPct: 25,
		}
	case PhaseTermofila:
		return PhaseInfo{
			Code:        PhaseTermofila,
			Label:       "Termófila (Higienización)",
			Description: "Alta actividad microbiana (≥45 °C): eliminación de patógenos y semillas de malezas.",
			Color:       "#EA580C",
			ProgressPct: 50,
		}
	case PhaseEnfriamiento:
		return PhaseInfo{
			Code:        PhaseEnfriamiento,
			Label:       "Enfriamiento (Mesófila II)",
			Description: "Descenso gradual tras el pico térmico: hongos y actinomicetos degradan celulosa y lignina.",
			Color:       "#CA8A04",
			ProgressPct: 75,
		}
	case PhaseMaduracion:
		return PhaseInfo{
			Code:        PhaseMaduracion,
			Label:       "Maduración",
			Description: "Temperatura estable y cercana al ambiente: estabilización del humus listo para uso agrícola.",
			Color:       "#16A34A",
			ProgressPct: 100,
		}
	default:
		return PhaseInfo{
			Code:        PhaseSinDatos,
			Label:       "Sin lecturas",
			Description: "Pila registrada esperando las primeras mediciones del nodo.",
			Color:       "#64748B",
			ProgressPct: 0,
		}
	}
}

// Evaluate recibe las lecturas ordenadas cronológicamente (de más antigua a más reciente).
func Evaluate(startedAt time.Time, readings []ReadingSample, pileStatus string, now time.Time) Evaluation {
	daysElapsed := int(now.Sub(startedAt).Hours() / 24)
	if daysElapsed < 0 {
		daysElapsed = 0
	}

	if len(readings) == 0 {
		return Evaluation{
			Phase:       InfoForPhase(PhaseSinDatos),
			DaysElapsed: daysElapsed,
			Alerts: []Alert{
				{
					Code:     "NO_DATA",
					Severity: "info",
					Title:    "Esperando lecturas",
					Message:  "Aún no se han recibido mediciones para esta pila. Verifica el nodo LoRa o registra una lectura de prueba.",
				},
			},
		}
	}

	// Calcular pico térmico y memoria de fase termófila (>= 45 °C al menos 2 lecturas).
	thermoCount := 0
	peak := readings[0].TemperatureC
	for _, r := range readings {
		if r.TemperatureC >= 45.0 {
			thermoCount++
		}
		if r.TemperatureC > peak {
			peak = r.TemperatureC
		}
	}
	hasReachedThermo := thermoCount >= 2

	// Promedio móvil de las últimas 3 lecturas (o menos si hay < 3).
	windowStart := len(readings) - 3
	if windowStart < 0 {
		windowStart = 0
	}
	var sumTemp float64
	for _, r := range readings[windowStart:] {
		sumTemp += r.TemperatureC
	}
	refTemp := round1(sumTemp / float64(len(readings)-windowStart))
	peakRounded := round1(peak)

	var phaseCode PhaseCode
	switch {
	case refTemp >= 45.0:
		phaseCode = PhaseTermofila
	case !hasReachedThermo:
		phaseCode = PhaseMesofilaInicial
	case refTemp >= 30.0:
		phaseCode = PhaseEnfriamiento
	default:
		phaseCode = PhaseMaduracion
	}

	latest := readings[len(readings)-1]
	alerts := buildAlerts(phaseCode, latest, daysElapsed, pileStatus, now)

	return Evaluation{
		Phase:                  InfoForPhase(phaseCode),
		RefTemperatureC:        &refTemp,
		HasReachedThermophilic: hasReachedThermo,
		PeakTemperatureC:       &peakRounded,
		DaysElapsed:            daysElapsed,
		Alerts:                 alerts,
	}
}

func buildAlerts(phase PhaseCode, latest ReadingSample, daysElapsed int, pileStatus string, now time.Time) []Alert {
	var alerts []Alert

	if latest.TemperatureC > 70.0 {
		alerts = append(alerts, Alert{
			Code:     "OVERHEAT",
			Severity: "danger",
			Title:    "Sobrecalentamiento (>70 °C)",
			Message:  "Riesgo de inhibir microorganismos benéficos. Realizar volteo inmediato para oxigenar y disipar calor.",
		})
	} else if phase == PhaseMesofilaInicial && latest.TemperatureC < 20.0 && daysElapsed >= 3 {
		alerts = append(alerts, Alert{
			Code:     "SLOW_START",
			Severity: "warning",
			Title:    "Arranque lento",
			Message:  "Temperatura por debajo de 20 °C tras 3 días. Revisar humedad, tamaño de pila o relación Carbono/Nitrógeno.",
		})
	}

	if latest.HumidityPct < 40.0 {
		alerts = append(alerts, Alert{
			Code:     "LOW_HUMIDITY",
			Severity: "warning",
			Title:    "Humedad baja (<40 %)",
			Message:  "La actividad microbiana disminuye. Regar uniformemente la pila durante el próximo volteo (meta: 50–60 %).",
		})
	} else if latest.HumidityPct > 65.0 {
		alerts = append(alerts, Alert{
			Code:     "HIGH_HUMIDITY",
			Severity: "warning",
			Title:    "Exceso de humedad (>65 %)",
			Message:  "Riesgo de condiciones anaeróbicas y malos olores. Voltear y mezclar material seco estructurante (aserrín u hojarasca).",
		})
	}

	if latest.PH != nil {
		if *latest.PH < 5.5 {
			alerts = append(alerts, Alert{
				Code:     "ACID_PH",
				Severity: "warning",
				Title:    "Sustrato muy ácido (pH < 5.5)",
				Message:  "Posible falta de aireación. Voltear la pila y evaluar adición moderada de cal agrícola o ceniza.",
			})
		} else if *latest.PH > 8.5 {
			alerts = append(alerts, Alert{
				Code:     "ALKALINE_PH",
				Severity: "warning",
				Title:    "Sustrato muy alcalino (pH > 8.5)",
				Message:  "Riesgo de pérdida de nitrógeno en forma de amoníaco. Incorporar material rico en carbono (hojarasca o aserrín).",
			})
		}
	}

	if pileStatus == "active" && now.Sub(latest.RecordedAt) > 60*time.Minute {
		alerts = append(alerts, Alert{
			Code:     "STALE_NODE",
			Severity: "info",
			Title:    "Nodo sin reportar recientemente",
			Message:  "Han pasado más de 60 minutos desde la última lectura. Verificar alimentación del emisor o enlace LoRa.",
		})
	}

	if len(alerts) == 0 {
		alerts = append(alerts, Alert{
			Code:     "HEALTHY",
			Severity: "ok",
			Title:    "Proceso saludable",
			Message:  "Variables dentro de los rangos óptimos para la fase actual.",
		})
	}

	return alerts
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
