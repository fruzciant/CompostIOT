import { Droplets, Flame, FlaskConical, Leaf, Sprout, Thermometer, ThermometerSnowflake } from 'lucide-react'
import { PHASE_STYLES } from '../utils/format'

export function GuideView() {
  const phases = [
    {
      code: 'mesofila_inicial' as const,
      title: 'Mesófila inicial (20 a 44 °C)',
      icon: Sprout,
      summary:
        'Al armar la pila, bacterias y hongos mesófilos comienzan a degradar azúcares y compuestos simples. La temperatura sube de forma constante durante los primeros días.',
      action:
        'Verifica que la humedad inicial esté entre 50 y 60 %. Si en 3 días no supera los 20 °C, revisa la relación Carbono/Nitrógeno o el tamaño mínimo de la pila.',
    },
    {
      code: 'termofila' as const,
      title: 'Termófila o de higienización (≥ 45 °C, óptimo 55 a 65 °C)',
      icon: Flame,
      summary:
        'Etapa clave de máxima actividad microbiana. El calor intenso elimina bacterias patógenas y semillas de malezas, liberando vapor visible al voltear.',
      action:
        'Voltea cuando supere 65 a 70 °C para evitar inhibir microorganismos benéficos y repón humedad si baja del 45 %.',
    },
    {
      code: 'enfriamiento' as const,
      title: 'Enfriamiento o mesófila II (30 a 44 °C tras pico térmico)',
      icon: ThermometerSnowflake,
      summary:
        'Una vez consumidos los compuestos más rápidos, la temperatura desciende gradualmente. Hongos y actinomicetos descomponen celulosa y lignina.',
      action:
        'Mantén aireación periódica con volteos más espaciados y evita que el sustrato se seque por debajo del 40 %.',
    },
    {
      code: 'maduracion' as const,
      title: 'Maduración (< 30 °C estable)',
      icon: Leaf,
      summary:
        'La temperatura se estabiliza cerca del ambiente, el olor recuerda a tierra de bosque y el abono alcanza estabilidad biológica para uso agrícola.',
      action:
        'Deja reposar de 2 a 4 semanas, tamiza si el uso productivo lo requiere y exporta la curva histórica en CSV para documentar la calidad del lote.',
    },
  ]

  return (
    <div className="space-y-8 max-w-4xl">
      <div className="space-y-2">
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-[#12261C]">
          Guía rápida para interpretar tu compostaje
        </h1>
        <p className="text-base text-[#465C50] leading-relaxed max-w-[65ch]">
          Esta referencia resume cómo CompostIOT determina automáticamente la fase de cada pila y qué acciones prácticas tomar en campo según temperatura, humedad y pH.
        </p>
      </div>

      {/* Etapas */}
      <section className="space-y-4">
        <h2 className="text-xl font-bold text-[#12261C]">Las 4 fases del ciclo aeróbico</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {phases.map((p) => {
            const style = PHASE_STYLES[p.code]
            const Icon = p.icon
            return (
              <article
                key={p.code}
                className="rounded-2xl border p-5 space-y-3 bg-white"
                style={{ borderColor: style.border }}
              >
                <div className="flex items-center gap-2.5">
                  <div
                    className="h-9 w-9 rounded-xl flex items-center justify-center shrink-0"
                    style={{ backgroundColor: style.bgSoft, color: style.accent }}
                  >
                    <Icon className="h-5 w-5" />
                  </div>
                  <h3 className="text-base font-bold text-[#12261C]">{p.title}</h3>
                </div>
                <p className="text-base text-[#465C50] leading-relaxed">{p.summary}</p>
                <div className="rounded-xl bg-[#F8FAF9] border border-[#DCE5DF] px-3.5 py-2.5 text-sm text-[#12261C]">
                  <strong>Qué hacer en campo:</strong> {p.action}
                </div>
              </article>
            )
          })}
        </div>
      </section>

      {/* Variables clave */}
      <section className="bg-white border border-[#DCE5DF] rounded-2xl p-6 space-y-5">
        <h2 className="text-xl font-bold text-[#12261C]">Rangos de control de variables</h2>
        <div className="divide-y divide-[#DCE5DF]">
          <div className="py-4 first:pt-0 flex gap-4 items-start">
            <Thermometer className="h-5 w-5 text-[#D95F02] shrink-0 mt-1" />
            <div className="space-y-1">
              <h3 className="text-base font-bold text-[#12261C]">Temperatura (obligatoria)</h3>
              <p className="text-base text-[#465C50] leading-relaxed">
                CompostIOT utiliza el promedio móvil de las últimas 3 lecturas para evitar falsos saltos por ruido. Si supera 70 °C se activa alerta roja de sobrecalentamiento para voltear de inmediato.
              </p>
            </div>
          </div>

          <div className="py-4 flex gap-4 items-start">
            <Droplets className="h-5 w-5 text-[#0284C7] shrink-0 mt-1" />
            <div className="space-y-1">
              <h3 className="text-base font-bold text-[#12261C]">Humedad (obligatoria)</h3>
              <p className="text-base text-[#465C50] leading-relaxed">
                El rango ideal se ubica entre 40 % y 65 %. Menos de 40 % frena la actividad microbiana (requiere riego); más de 65 % desplaza el oxígeno y provoca malos olores (agregar material seco estructurante).
              </p>
            </div>
          </div>

          <div className="py-4 last:pb-0 flex gap-4 items-start">
            <FlaskConical className="h-5 w-5 text-[#15613B] shrink-0 mt-1" />
            <div className="space-y-1">
              <h3 className="text-base font-bold text-[#12261C]">pH del sustrato (opcional)</h3>
              <p className="text-base text-[#465C50] leading-relaxed">
                Rango saludable entre 5.5 y 8.5. Al inicio puede bajar levemente por ácidos orgánicos y tiende a estabilizarse cerca de la neutralidad (6.8 a 7.5) en maduración.
              </p>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}
