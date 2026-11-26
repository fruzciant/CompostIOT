import { useState, type FormEvent } from 'react'
import {
  CheckCircle2,
  Loader2,
  Plus,
  Server,
  Sprout,
  Trash2,
  Wifi,
  WifiOff,
} from 'lucide-react'
import type { HealthInfo, ServerProfile, User } from '../types'
import { checkHealth, deleteServer, isWailsDesktop, login, saveServer } from '../services/client'

interface LoginViewProps {
  servers: ServerProfile[]
  selectedServerId: string
  onServersChanged: (servers: ServerProfile[], selectedId: string) => void
  onAuthenticated: (server: ServerProfile, user: User, lastSyncAt: string | null) => void
}

export function LoginView({
  servers,
  selectedServerId,
  onServersChanged,
  onAuthenticated,
}: LoginViewProps) {
  const activeServer = servers.find((s) => s.id === selectedServerId) ?? servers[0]

  const [username, setUsername] = useState(activeServer?.username ?? '')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [addingServer, setAddingServer] = useState(false)
  const [newServerName, setNewServerName] = useState('')
  const [newServerUrl, setNewServerUrl] = useState('')
  const [probing, setProbing] = useState(false)
  const [probeResult, setProbeResult] = useState<{ ok: boolean; text: string; info?: HealthInfo } | null>(null)

  const handleSelectServer = (id: string) => {
    const found = servers.find((s) => s.id === id)
    onServersChanged(servers, id)
    setUsername(found?.username ?? '')
    setProbeResult(null)
    setError(null)
  }

  const handleProbe = async (urlToTest: string) => {
    setProbing(true)
    setProbeResult(null)
    try {
      const info = await checkHealth(urlToTest)
      setProbeResult({
        ok: true,
        text: `Conectado a ${info.server_name} (v${info.version})`,
        info,
      })
    } catch {
      setProbeResult({
        ok: false,
        text: 'No se pudo conectar. Verifica la URL o si el servidor está encendido.',
      })
    } finally {
      setProbing(false)
    }
  }

  const handleAddServer = async (e: FormEvent) => {
    e.preventDefault()
    if (!newServerUrl.trim()) return
    const created = await saveServer({
      name: newServerName.trim() || 'Servidor remoto',
      url: newServerUrl.trim(),
    })
    const updated = [...servers, created]
    onServersChanged(updated, created.id)
    setNewServerName('')
    setNewServerUrl('')
    setAddingServer(false)
    setProbeResult(null)
  }

  const handleDeleteServer = async (id: string) => {
    if (servers.length <= 1) return
    await deleteServer(id)
    const remaining = servers.filter((s) => s.id !== id)
    onServersChanged(remaining, remaining[0].id)
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!activeServer) return
    setSubmitting(true)
    setError(null)
    try {
      const res = await login(activeServer, username.trim(), password)
      onAuthenticated(activeServer, res.user, res.lastSyncAt)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No fue posible iniciar sesión')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="min-h-dvh flex flex-col justify-center bg-[#F8FAF9] px-4 py-8 sm:px-6">
      <div className="mx-auto w-full max-w-5xl grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Identidad y propósito */}
        <div className="lg:col-span-5 lg:pt-6 space-y-5">
          <div className="flex items-center gap-3">
            <div className="h-11 w-11 rounded-xl bg-[#15613B] text-white flex items-center justify-center shadow-sm">
              <Sprout className="h-6 w-6" />
            </div>
            <div>
              <span className="text-xl font-extrabold tracking-tight text-[#12261C] block leading-none">
                CompostIOT
              </span>
              <span className="text-sm text-[#465C50]">
                {isWailsDesktop() ? 'Cliente de escritorio SENA' : 'Monitoreo de compostaje SENA'}
              </span>
            </div>
          </div>

          <h1 className="text-3xl sm:text-4xl font-bold tracking-tight text-[#12261C] leading-[1.12]">
            Control claro y en tiempo real de tus pilas de compostaje
          </h1>

          <p className="text-base text-[#465C50] leading-relaxed max-w-[46ch]">
            Consulta temperatura, humedad, pH y la fase exacta del proceso desde el aula, la finca o tu computador local, sin tomar datos a mano.
          </p>

          <div className="border-t border-[#DCE5DF] pt-5 space-y-2.5 text-sm text-[#465C50]">
            <p className="font-semibold text-[#12261C]">Compatible con servidor local o en la nube</p>
            <p>
              Conéctate a un equipo en <span className="font-semibold text-[#12261C]">localhost</span> sin internet o a la instancia institucional de tu centro de formación.
            </p>
            {isWailsDesktop() && (
              <p className="text-[#15613B] font-medium">
                Tu sesión se guarda en el llavero del sistema con caché SQLite offline.
              </p>
            )}
          </div>
        </div>

        {/* Panel de conexión e inicio de sesión */}
        <div className="lg:col-span-7 bg-white border border-[#DCE5DF] rounded-2xl p-6 sm:p-8 shadow-[0_1px_3px_rgba(18,38,28,0.04)] space-y-6">
          {/* Selector de servidor */}
          <div className="space-y-3">
            <div className="flex items-center justify-between gap-2">
              <label htmlFor="server-select" className="text-sm font-semibold text-[#12261C] flex items-center gap-2">
                <Server className="h-4 w-4 text-[#15613B]" />
                Servidor de destino
              </label>
              <button
                type="button"
                onClick={() => setAddingServer((v) => !v)}
                className="text-sm font-semibold text-[#15613B] hover:text-[#0F4C2D] inline-flex items-center gap-1 transition-[color] duration-150"
              >
                <Plus className="h-4 w-4" />
                {addingServer ? 'Cancelar' : 'Agregar servidor'}
              </button>
            </div>

            {!addingServer ? (
              <div className="space-y-2">
                <div className="flex gap-2">
                  <select
                    id="server-select"
                    value={activeServer?.id ?? ''}
                    onChange={(e) => handleSelectServer(e.target.value)}
                    className="flex-1 rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-sm font-medium text-[#12261C] focus:outline-none focus:border-[#15613B]"
                  >
                    {servers.map((srv) => (
                      <option key={srv.id} value={srv.id}>
                        {srv.name} ({srv.url})
                      </option>
                    ))}
                  </select>

                  <button
                    type="button"
                    onClick={() => activeServer && handleProbe(activeServer.url)}
                    disabled={probing}
                    title="Probar conexión"
                    className="px-3.5 py-2.5 rounded-xl border border-[#DCE5DF] bg-white hover:bg-[#EDF4F0] text-sm font-semibold text-[#12261C] inline-flex items-center gap-1.5 transition-[background-color,border-color] duration-150 disabled:opacity-60"
                  >
                    {probing ? <Loader2 className="h-4 w-4 animate-spin" /> : <Wifi className="h-4 w-4 text-[#15613B]" />}
                    <span className="hidden sm:inline">Probar</span>
                  </button>

                  {servers.length > 1 && activeServer && (
                    <button
                      type="button"
                      onClick={() => handleDeleteServer(activeServer.id)}
                      title="Quitar servidor"
                      className="p-2.5 rounded-xl border border-[#DCE5DF] bg-white hover:bg-red-50 hover:border-red-200 text-[#465C50] hover:text-red-700 transition-[background-color,color,border-color] duration-150"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  )}
                </div>

                {probeResult && (
                  <div
                    className={`flex items-center gap-2 rounded-xl px-3.5 py-2.5 text-sm ${
                      probeResult.ok
                        ? 'bg-[#F0FDF4] text-[#166534] border border-[#BBF7D0]'
                        : 'bg-red-50 text-red-800 border border-red-200'
                    }`}
                  >
                    {probeResult.ok ? (
                      <CheckCircle2 className="h-4 w-4 shrink-0" />
                    ) : (
                      <WifiOff className="h-4 w-4 shrink-0" />
                    )}
                    <span>{probeResult.text}</span>
                  </div>
                )}
              </div>
            ) : (
              <form onSubmit={handleAddServer} className="rounded-xl bg-[#EDF4F0]/70 border border-[#DCE5DF] p-4 space-y-3">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div>
                    <label className="block text-sm font-medium text-[#12261C] mb-1">Nombre del servidor</label>
                    <input
                      type="text"
                      value={newServerName}
                      onChange={(e) => setNewServerName(e.target.value)}
                      placeholder="Ej. Nube SENA o Laboratorio"
                      className="w-full rounded-lg border border-[#DCE5DF] bg-white px-3 py-2 text-sm focus:outline-none focus:border-[#15613B]"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-[#12261C] mb-1">URL del backend</label>
                    <input
                      type="text"
                      required
                      value={newServerUrl}
                      onChange={(e) => setNewServerUrl(e.target.value)}
                      placeholder="http://192.168.1.20:8080"
                      className="w-full rounded-lg border border-[#DCE5DF] bg-white px-3 py-2 text-sm focus:outline-none focus:border-[#15613B]"
                    />
                  </div>
                </div>
                <div className="flex justify-end">
                  <button
                    type="submit"
                    className="rounded-lg bg-[#15613B] hover:bg-[#0F4C2D] text-white px-4 py-2 text-sm font-semibold transition-[background-color] duration-150"
                  >
                    Guardar servidor
                  </button>
                </div>
              </form>
            )}
          </div>

          <hr className="border-[#DCE5DF]" />

          {/* Formulario de credenciales */}
          <form onSubmit={handleSubmit} className="space-y-4">
            <h2 className="text-xl font-bold text-[#12261C]">Iniciar sesión</h2>

            <div>
              <label htmlFor="login-username" className="block text-sm font-semibold text-[#12261C] mb-1.5">
                Usuario
              </label>
              <input
                id="login-username"
                type="text"
                autoComplete="username"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="instructor o aprendiz"
                className="w-full rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-4 py-3 text-base text-[#12261C] focus:bg-white focus:outline-none focus:border-[#15613B]"
              />
            </div>

            <div>
              <label htmlFor="login-password" className="block text-sm font-semibold text-[#12261C] mb-1.5">
                Contraseña
              </label>
              <input
                id="login-password"
                type="password"
                autoComplete="current-password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Ingresa tu contraseña"
                className="w-full rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-4 py-3 text-base text-[#12261C] focus:bg-white focus:outline-none focus:border-[#15613B]"
              />
            </div>

            {error && (
              <div className="rounded-xl bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-800">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={submitting || !activeServer}
              className="w-full rounded-xl bg-[#15613B] hover:bg-[#0F4C2D] text-white py-3.5 px-5 text-base font-semibold flex items-center justify-center gap-2 transition-[background-color] duration-150 disabled:opacity-60"
            >
              {submitting ? (
                <>
                  <Loader2 className="h-5 w-5 animate-spin" />
                  Ingresando...
                </>
              ) : (
                'Entrar al tablero'
              )}
            </button>

            {activeServer?.url.includes('localhost') && (
              <p className="text-sm text-[#465C50] pt-1">
                Acceso de prueba local: <strong className="text-[#12261C]">instructor</strong> / <strong className="text-[#12261C]">compost2026</strong> (admin) o <strong className="text-[#12261C]">aprendiz</strong> / <strong className="text-[#12261C]">aprendiz2026</strong>.
              </p>
            )}
          </form>
        </div>
      </div>
    </div>
  )
}
