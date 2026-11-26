import { Flame, Leaf, Sprout, ThermometerSnowflake } from 'lucide-react'
import type { Evaluation, PhaseCode } from '../types'
import { PHASE_STYLES } from '../utils/format'

interface PhaseProgressProps {
  evaluation: Evaluation
}

const STAGES: { code: PhaseCode; label: string; range: string; icon: typeof Sprout }[] = [
  { code: 'mesofila_inicial', label: 'Mesófila inicial', range: '20 a 44 °C', icon: Sprout },
  { code: 'termofila', label: 'Termófila', range: '≥ 45 °C', icon: Flame },
  { code: 'enfriamiento', label: 'Enfriamiento', range: '30 a 44 °C', icon: ThermometerSnowflake },
  { code: 'maduracion', label: 'Maduración', range: '< 30 °C estable', icon: Leaf },
]

export function PhaseProgress({ evaluation }: PhaseProgressProps) {
  const currentCode = evaluation.phase.code
  const style = PHASE_STYLES[currentCode]
  const activeIdx = STAGES.findIndex((s) => s.code === currentCode)

  return (
    <section
      className="rounded-2xl border p-5 sm:p-6 space-y-5"
      style={{
        backgroundColor: style.bgSoft,
        borderColor: style.border,
      }}
    >
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div className="space-y-1.5 max-w-2xl">
          <div className="flex items-center gap-2.5">
            <span
              className="inline-block h-3 w-3 rounded-full shrink-0"
              style={{ backgroundColor: style.accent }}
            />
            <h2 className="text-xl font-bold tracking-tight" style={{ color: style.text }}>
              Fase actual: {evaluation.phase.label}
            </h2>
          </div>
          <p className="text-base text-[#12261C]/85 leading-relaxed">
            {evaluation.phase.description}
          </p>
        </div>

        <div className="flex flex-wrap sm:flex-col sm:items-end gap-x-5 gap-y-1 text-sm text-[#465C50] tabular shrink-0">
          <span>
            Día del ciclo: <strong className="text-[#12261C]">{evaluation.days_elapsed}</strong>
          </span>
          {evaluation.peak_temperature_c !== null && (
            <span>
              Pico térmico: <strong className="text-[#12261C]">{evaluation.peak_temperature_c.toFixed(1)} °C</strong>
            </span>
          )}
          {evaluation.ref_temperature_c !== null && (
            <span>
              Promedio móvil: <strong className="text-[#12261C]">{evaluation.ref_temperature_c.toFixed(1)} °C</strong>
            </span>
          )}
        </div>
      </div>

      {/* Línea de 4 etapas del compostaje */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
        {STAGES.map((stage, idx) => {
          const Icon = stage.icon
          const isCurrent = idx === activeIdx
          const isCompleted = activeIdx > idx
          const stageStyle = PHASE_STYLES[stage.code]

          return (
            <div
              key={stage.code}
              className={`rounded-xl px-3.5 py-3 border transition-[background-color,border-color] duration-150 ${
                isCurrent
                  ? 'bg-white shadow-xs'
                  : isCompleted
                    ? 'bg-white/75 border-[#DCE5DF]'
                    : 'bg-white/45 border-[#DCE5DF]/80 opacity-75'
              }`}
              style={isCurrent ? { borderColor: stageStyle.accent } : undefined}
            >
              <div className="flex items-center gap-2 mb-1">
                <Icon
                  className="h-4 w-4 shrink-0"
                  style={{ color: isCurrent || isCompleted ? stageStyle.accent : '#64748B' }}
                />
                <span
                  className={`text-sm font-bold leading-tight ${
                    isCurrent ? 'text-[#12261C]' : 'text-[#465C50]'
                  }`}
                >
                  {stage.label}
                </span>
              </div>
              <p className="text-sm text-[#465C50] tabular">{stage.range}</p>
            </div>
          )
        })}
      </div>
    </section>
  )
}
