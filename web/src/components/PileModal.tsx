import { useState, type FormEvent } from 'react'
import { Loader2, X } from 'lucide-react'
import type { ManualReadingInput, Pile, PileInput } from '../types'

interface PileModalProps {
  pile?: Pile | null
  onClose: () => void
  onSave: (input: PileInput) => Promise<void>
}

export function PileModal({ pile, onClose, onSave }: PileModalProps) {
  const [name, setName] = useState(pile?.name ?? '')
  const [location, setLocation] = useState(pile?.location ?? '')
  const [substrate, setSubstrate] = useState(pile?.substrate ?? '')
  const [nodeId, setNodeId] = useState(pile?.node_id ?? '')
  const [hasPH, setHasPH] = useState(pile?.has_ph_sensor ?? false)
  const [status, setStatus] = useState<'active' | 'archived'>(pile?.status ?? 'active')
  const [notes, setNotes] = useState(pile?.notes ?? '')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      await onSave({
        name: name.trim(),
        location: location.trim(),
        substrate: substrate.trim(),
        node_id: nodeId.trim().toUpperCase(),
        has_ph_sensor: hasPH,
        status,
        notes: notes.trim(),
      })
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo guardar la pila')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[#12261C]/45 p-4 backdrop-blur-[1px]">
      <div className="w-full max-w-lg rounded-2xl bg-white border border-[#DCE5DF] shadow-xl overflow-hidden">
        <div className="flex items-center justify-between px-6 py-4 border-b border-[#DCE5DF]">
          <h2 className="text-xl font-bold text-[#12261C]">
            {pile ? 'Editar pila de compostaje' : 'Agregar nueva pila'}
          </h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Cerrar"
            className="p-1.5 rounded-lg text-[#465C50] hover:bg-[#EDF4F0] hover:text-[#12261C] transition-[background-color,color] duration-150"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4 max-h-[82dvh] overflow-y-auto">
          <div>
            <label className="block text-sm font-semibold text-[#12261C] mb-1">Nombre de la pila</label>
            <input
              type="text"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Ej. Pila #4 Residuos de poda y café"
              className="w-full rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-base focus:bg-white focus:outline-none focus:border-[#15613B]"
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-semibold text-[#12261C] mb-1">ID del nodo LoRa</label>
              <input
                type="text"
                required
                value={nodeId}
                onChange={(e) => setNodeId(e.target.value.toUpperCase())}
                placeholder="NODO-04"
                className="w-full rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-base uppercase focus:bg-white focus:outline-none focus:border-[#15613B]"
              />
            </div>

            <div>
              <label className="block text-sm font-semibold text-[#12261C] mb-1">Ubicación</label>
              <input
                type="text"
                value={location}
                onChange={(e) => setLocation(e.target.value)}
                placeholder="Ej. Invernadero Sur"
                className="w-full rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-base focus:bg-white focus:outline-none focus:border-[#15613B]"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-semibold text-[#12261C] mb-1">Mezcla o sustrato</label>
            <input
              type="text"
              value={substrate}
              onChange={(e) => setSubstrate(e.target.value)}
              placeholder="Ej. Restos vegetales + hojarasca + estiércol"
              className="w-full rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-base focus:bg-white focus:outline-none focus:border-[#15613B]"
            />
          </div>

          <div className="flex items-center justify-between gap-4 rounded-xl bg-[#F8FAF9] border border-[#DCE5DF] px-4 py-3">
            <div>
              <p className="text-sm font-semibold text-[#12261C]">Sensor de pH instalado</p>
              <p className="text-sm text-[#465C50]">Variable opcional. Actívalo si esta pila tiene sonda de pH.</p>
            </div>
            <input
              type="checkbox"
              checked={hasPH}
              onChange={(e) => setHasPH(e.target.checked)}
              className="h-5 w-5 accent-[#15613B] rounded cursor-pointer"
            />
          </div>

          {pile && (
            <div>
              <label className="block text-sm font-semibold text-[#12261C] mb-1">Estado de la pila</label>
              <select
                value={status}
                onChange={(e) => setStatus(e.target.value as 'active' | 'archived')}
                className="w-full rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-base focus:bg-white focus:outline-none focus:border-[#15613B]"
              >
                <option value="active">Activa (recibiendo mediciones)</option>
                <option value="archived">Archivada (ciclo finalizado)</option>
              </select>
            </div>
          )}

          <div>
            <label className="block text-sm font-semibold text-[#12261C] mb-1">Notas de campo</label>
            <textarea
              rows={3}
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Observaciones iniciales, frecuencia de volteo o ficha de aprendices..."
              className="w-full rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-base focus:bg-white focus:outline-none focus:border-[#15613B]"
            />
          </div>

          {error && (
            <div className="rounded-xl bg-red-50 border border-red-200 px-4 py-2.5 text-sm text-red-800">
              {error}
            </div>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-xl border border-[#DCE5DF] px-4 py-2.5 text-sm font-semibold text-[#465C50] hover:bg-[#EDF4F0] transition-[background-color] duration-150"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={saving}
              className="rounded-xl bg-[#15613B] hover:bg-[#0F4C2D] text-white px-5 py-2.5 text-sm font-semibold inline-flex items-center gap-2 transition-[background-color] duration-150 disabled:opacity-60"
            >
              {saving && <Loader2 className="h-4 w-4 animate-spin" />}
              {pile ? 'Guardar cambios' : 'Crear pila'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

interface ManualReadingModalProps {
  pile: Pile
  onClose: () => void
  onSubmit: (input: ManualReadingInput) => Promise<void>
}

export function ManualReadingModal({ pile, onClose, onSubmit }: ManualReadingModalProps) {
  const [temp, setTemp] = useState(pile.latest_reading?.temperature_c.toFixed(1) ?? '48.0')
  const [hum, setHum] = useState(pile.latest_reading?.humidity_pct.toFixed(1) ?? '55.0')
  const [ph, setPh] = useState(pile.latest_reading?.ph?.toFixed(2) ?? '')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const tempVal = Number.parseFloat(temp)
      const humVal = Number.parseFloat(hum)
      const phVal = ph.trim() === '' ? null : Number.parseFloat(ph)
      if (Number.isNaN(tempVal) || Number.isNaN(humVal)) {
        throw new Error('Ingresa valores numéricos válidos para temperatura y humedad')
      }
      if (phVal !== null && Number.isNaN(phVal)) {
        throw new Error('El valor de pH debe ser numérico o quedar vacío')
      }
      await onSubmit({ temperature_c: tempVal, humidity_pct: humVal, ph: phVal })
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error registrando medición')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[#12261C]/45 p-4 backdrop-blur-[1px]">
      <div className="w-full max-w-md rounded-2xl bg-white border border-[#DCE5DF] shadow-xl overflow-hidden">
        <div className="flex items-center justify-between px-6 py-4 border-b border-[#DCE5DF]">
          <div>
            <h2 className="text-xl font-bold text-[#12261C]">Registrar medición manual</h2>
            <p className="text-sm text-[#465C50]">{pile.name}</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Cerrar"
            className="p-1.5 rounded-lg text-[#465C50] hover:bg-[#EDF4F0]"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-semibold text-[#12261C] mb-1">Temperatura (°C)</label>
              <input
                type="number"
                step="0.1"
                min="-10"
                max="110"
                required
                value={temp}
                onChange={(e) => setTemp(e.target.value)}
                className="w-full rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-base tabular focus:bg-white focus:outline-none focus:border-[#15613B]"
              />
            </div>

            <div>
              <label className="block text-sm font-semibold text-[#12261C] mb-1">Humedad (%)</label>
              <input
                type="number"
                step="0.1"
                min="0"
                max="100"
                required
                value={hum}
                onChange={(e) => setHum(e.target.value)}
                className="w-full rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-base tabular focus:bg-white focus:outline-none focus:border-[#15613B]"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-semibold text-[#12261C] mb-1">pH (opcional)</label>
            <input
              type="number"
              step="0.01"
              min="0"
              max="14"
              value={ph}
              onChange={(e) => setPh(e.target.value)}
              placeholder="Dejar vacío si no se midió pH"
              className="w-full rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-base tabular focus:bg-white focus:outline-none focus:border-[#15613B]"
            />
          </div>

          {error && (
            <div className="rounded-xl bg-red-50 border border-red-200 px-4 py-2.5 text-sm text-red-800">
              {error}
            </div>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-xl border border-[#DCE5DF] px-4 py-2.5 text-sm font-semibold text-[#465C50] hover:bg-[#EDF4F0]"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={saving}
              className="rounded-xl bg-[#15613B] hover:bg-[#0F4C2D] text-white px-5 py-2.5 text-sm font-semibold inline-flex items-center gap-2 disabled:opacity-60"
            >
              {saving && <Loader2 className="h-4 w-4 animate-spin" />}
              Guardar lectura
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
