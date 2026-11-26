package compost

import (
	"testing"
	"time"
)

func floatPtr(v float64) *float64 { return &v }

func TestEvaluatePhasesAndAlerts(t *testing.T) {
	now := time.Date(2026, 11, 26, 14, 0, 0, 0, time.UTC)
	started := now.Add(-10 * 24 * time.Hour)

	t.Run("sin datos", func(t *testing.T) {
		ev := Evaluate(started, nil, "active", now)
		if ev.Phase.Code != PhaseSinDatos {
			t.Fatalf("esperaba %s, obtuvo %s", PhaseSinDatos, ev.Phase.Code)
		}
		if len(ev.Alerts) != 1 || ev.Alerts[0].Code != "NO_DATA" {
			t.Fatalf("esperaba alerta NO_DATA, obtuvo %+v", ev.Alerts)
		}
	})

	t.Run("mesofila inicial saludable", func(t *testing.T) {
		readings := []ReadingSample{
			{TemperatureC: 28.0, HumidityPct: 55.0, RecordedAt: now.Add(-30 * time.Minute)},
			{TemperatureC: 32.5, HumidityPct: 54.0, RecordedAt: now.Add(-15 * time.Minute)},
			{TemperatureC: 36.0, HumidityPct: 53.0, RecordedAt: now.Add(-5 * time.Minute)},
		}
		ev := Evaluate(started, readings, "active", now)
		if ev.Phase.Code != PhaseMesofilaInicial {
			t.Fatalf("esperaba mesofila_inicial, obtuvo %s", ev.Phase.Code)
		}
		if ev.HasReachedThermophilic {
			t.Fatalf("no debía marcar pico termófilo aún")
		}
		if ev.Alerts[0].Code != "HEALTHY" {
			t.Fatalf("esperaba alerta HEALTHY, obtuvo %+v", ev.Alerts)
		}
	})

	t.Run("termofila con sobrecalentamiento y humedad baja", func(t *testing.T) {
		readings := []ReadingSample{
			{TemperatureC: 52.0, HumidityPct: 50.0, RecordedAt: now.Add(-20 * time.Minute)},
			{TemperatureC: 64.0, HumidityPct: 45.0, RecordedAt: now.Add(-10 * time.Minute)},
			{TemperatureC: 72.5, HumidityPct: 35.0, RecordedAt: now.Add(-2 * time.Minute)},
		}
		ev := Evaluate(started, readings, "active", now)
		if ev.Phase.Code != PhaseTermofila {
			t.Fatalf("esperaba termofila, obtuvo %s", ev.Phase.Code)
		}
		if !ev.HasReachedThermophilic {
			t.Fatalf("debía marcar pico termófilo")
		}
		if len(ev.Alerts) < 2 {
			t.Fatalf("esperaba al menos 2 alertas (OVERHEAT y LOW_HUMIDITY), obtuvo %+v", ev.Alerts)
		}
	})

	t.Run("enfriamiento y maduracion tras pico termofilo", func(t *testing.T) {
		cooling := []ReadingSample{
			{TemperatureC: 55.0, HumidityPct: 55.0, RecordedAt: now.Add(-48 * time.Hour)},
			{TemperatureC: 60.0, HumidityPct: 54.0, RecordedAt: now.Add(-36 * time.Hour)},
			{TemperatureC: 41.0, HumidityPct: 52.0, RecordedAt: now.Add(-20 * time.Minute)},
			{TemperatureC: 39.0, HumidityPct: 51.0, RecordedAt: now.Add(-10 * time.Minute)},
			{TemperatureC: 37.5, HumidityPct: 50.0, RecordedAt: now.Add(-2 * time.Minute)},
		}
		evCool := Evaluate(started, cooling, "active", now)
		if evCool.Phase.Code != PhaseEnfriamiento {
			t.Fatalf("esperaba enfriamiento, obtuvo %s", evCool.Phase.Code)
		}

		maturing := append(cooling,
			ReadingSample{TemperatureC: 27.0, HumidityPct: 48.0, PH: floatPtr(7.1), RecordedAt: now.Add(-3 * time.Minute)},
			ReadingSample{TemperatureC: 26.2, HumidityPct: 47.0, PH: floatPtr(7.0), RecordedAt: now.Add(-2 * time.Minute)},
			ReadingSample{TemperatureC: 25.8, HumidityPct: 47.0, PH: floatPtr(7.0), RecordedAt: now.Add(-1 * time.Minute)},
		)
		evMat := Evaluate(started, maturing, "active", now)
		if evMat.Phase.Code != PhaseMaduracion {
			t.Fatalf("esperaba maduracion, obtuvo %s", evMat.Phase.Code)
		}
		if evMat.Phase.ProgressPct != 100 {
			t.Fatalf("esperaba progreso 100%%, obtuvo %d%%", evMat.Phase.ProgressPct)
		}
	})

	t.Run("alertas de pH y nodo inactivo", func(t *testing.T) {
		readings := []ReadingSample{
			{TemperatureC: 34.0, HumidityPct: 70.0, PH: floatPtr(5.1), RecordedAt: now.Add(-90 * time.Minute)},
		}
		ev := Evaluate(started, readings, "active", now)
		codes := map[string]bool{}
		for _, a := range ev.Alerts {
			codes[a.Code] = true
		}
		for _, expected := range []string{"HIGH_HUMIDITY", "ACID_PH", "STALE_NODE"} {
			if !codes[expected] {
				t.Fatalf("faltó alerta esperada %s; alertas=%+v", expected, ev.Alerts)
			}
		}
	})
}
