import { useState } from 'react'
import {
  Area,
  AreaChart,
  CartesianGrid,
  ReferenceArea,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import type { Reading } from '../types'
import { formatChartAxisTime, formatShortDateTime, PHASE_STYLES } from '../utils/format'

interface ReadingsChartProps {
  readings: Reading[]
  showPH: boolean
}

type MetricTab = 'temperature' | 'humidity' | 'ph'

export function ReadingsChart({ readings, showPH }: ReadingsChartProps) {
  const hasAnyPH = showPH || readings.some((r) => r.ph !== null)
  const [metric, setMetric] = useState<MetricTab>('temperature')

  const activeMetric: MetricTab = metric === 'ph' && !hasAnyPH ? 'temperature' : metric

  const chartData = readings.map((r) => ({
    ...r,
    axisLabel: formatChartAxisTime(r.recorded_at),
    fullTime: formatShortDateTime(r.recorded_at),
  }))

  const config = {
    temperature: {
      key: 'temperature_c',
      label: 'Temperatura',
      unit: '°C',
      stroke: '#D95F02',
      fill: '#D95F02',
      domain: [10, 80] as [number, number],
      bandMin: 45,
      bandMax: 65,
      bandLabel: 'Zona termófila óptima (45 a 65 °C)',
    },
    humidity: {
      key: 'humidity_pct',
      label: 'Humedad',
      unit: '%',
      stroke: '#0284C7',
      fill: '#0284C7',
      domain: [0, 100] as [number, number],
      bandMin: 40,
      bandMax: 65,
      bandLabel: 'Rango óptimo de humedad (40 a 65 %)',
    },
    ph: {
      key: 'ph',
      label: 'pH del sustrato',
      unit: '',
      stroke: '#15613B',
      fill: '#15613B',
      domain: [4, 10] as [number, number],
      bandMin: 5.5,
      bandMax: 8.5,
      bandLabel: 'Rango saludable de pH (5.5 a 8.5)',
    },
  }[activeMetric]

  return (
    <section className="bg-white border border-[#DCE5DF] rounded-2xl p-5 sm:p-6 space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h3 className="text-xl font-bold text-[#12261C]">Comportamiento histórico</h3>
          <p className="text-sm text-[#465C50]">{config.bandLabel}</p>
        </div>

        {/* Selector segmentado no-botón */}
        <div
          role="tablist"
          aria-label="Variable de gráfica"
          className="inline-flex rounded-xl bg-[#EDF4F0] p-1 border border-[#DCE5DF] self-start"
        >
          {(
            [
              { id: 'temperature', label: 'Temperatura' },
              { id: 'humidity', label: 'Humedad' },
              ...(hasAnyPH ? [{ id: 'ph', label: 'pH' }] : []),
            ] as { id: MetricTab; label: string }[]
          ).map((tab) => {
            const active = activeMetric === tab.id
            return (
              <button
                key={tab.id}
                role="tab"
                type="button"
                aria-selected={active}
                onClick={() => setMetric(tab.id)}
                className={`px-3.5 py-1.5 rounded-lg text-sm font-semibold transition-[background-color,color] duration-150 ${
                  active ? 'bg-white text-[#12261C] shadow-xs' : 'text-[#465C50] hover:text-[#12261C]'
                }`}
              >
                {tab.label}
              </button>
            )
          })}
        </div>
      </div>

      {readings.length === 0 ? (
        <div className="h-64 flex flex-col items-center justify-center rounded-xl bg-[#F8FAF9] border border-dashed border-[#DCE5DF] text-center px-4">
          <p className="text-base font-semibold text-[#12261C]">Aún no hay lecturas para graficar</p>
          <p className="text-sm text-[#465C50] mt-1">
            En cuanto el nodo LoRa transmita o se registre una lectura manual, la curva aparecerá aquí.
          </p>
        </div>
      ) : (
        <div className="h-72 sm:h-80 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData} margin={{ top: 10, right: 12, left: -12, bottom: 0 }}>
              <defs>
                <linearGradient id={`grad-${activeMetric}`} x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor={config.fill} stopOpacity={0.22} />
                  <stop offset="95%" stopColor={config.fill} stopOpacity={0.02} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#E5ECE8" vertical={false} />
              <XAxis
                dataKey="axisLabel"
                tick={{ fill: '#465C50', fontSize: 13 }}
                axisLine={{ stroke: '#DCE5DF' }}
                tickLine={false}
                minTickGap={28}
              />
              <YAxis
                domain={config.domain}
                tick={{ fill: '#465C50', fontSize: 13 }}
                axisLine={false}
                tickLine={false}
                unit={config.unit ? ` ${config.unit}` : ''}
              />
              <ReferenceArea
                y1={config.bandMin}
                y2={config.bandMax}
                fill={config.fill}
                fillOpacity={0.07}
              />
              {activeMetric === 'temperature' && (
                <ReferenceLine
                  y={70}
                  stroke="#DC2626"
                  strokeDasharray="4 4"
                  label={{ value: 'Límite 70 °C', position: 'insideTopRight', fill: '#DC2626', fontSize: 12 }}
                />
              )}
              <Tooltip
                content={({ active, payload }) => {
                  if (!active || !payload?.length) return null
                  const row = payload[0].payload as Reading & { fullTime: string }
                  const phaseStyle = PHASE_STYLES[row.phase]
                  return (
                    <div className="rounded-xl bg-white border border-[#DCE5DF] shadow-md px-3.5 py-2.5 text-sm space-y-1">
                      <p className="font-semibold text-[#12261C] tabular">{row.fullTime}</p>
                      <p className="tabular text-[#12261C]">
                        Temperatura: <strong>{row.temperature_c.toFixed(1)} °C</strong>
                      </p>
                      <p className="tabular text-[#12261C]">
                        Humedad: <strong>{row.humidity_pct.toFixed(1)} %</strong>
                      </p>
                      {row.ph !== null && (
                        <p className="tabular text-[#12261C]">
                          pH: <strong>{row.ph.toFixed(2)}</strong>
                        </p>
                      )}
                      <p className="font-semibold pt-0.5" style={{ color: phaseStyle.text }}>
                        {phaseStyle.shortLabel}
                      </p>
                    </div>
                  )
                }}
              />
              <Area
                type="monotone"
                dataKey={config.key}
                stroke={config.stroke}
                strokeWidth={2.5}
                fill={`url(#grad-${activeMetric})`}
                connectNulls
                dot={{ r: readings.length <= 35 ? 3 : 0, fill: config.stroke }}
                activeDot={{ r: 5, strokeWidth: 2, stroke: '#FFFFFF' }}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      )}
    </section>
  )
}
