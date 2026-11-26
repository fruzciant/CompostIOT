import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  AlertTriangle,
  BookOpen,
  CheckCircle2,
  Download,
  Droplets,
  FlaskConical,
  Info,
  LayoutDashboard,
  Loader2,
  LogOut,
  MapPin,
  Pencil,
  Plus,
  PlusCircle,
  RefreshCw,
  Server,
  Settings,
  Sprout,
  Thermometer,
  Trash2,
  WifiOff,
} from 'lucide-react'
import type { ManualReadingInput, Pile, PileInput, Reading, ServerProfile, SessionState, User } from './types'
import {
  addManualReading,
  createPile,
  deletePile,
  exportPileCSV,
  fetchPiles,
  fetchReadings,
  getActiveServerId,
  getServers,
  logout,
  restoreSession,
  subscribeLiveEvents,
  updatePile,
} from './services/client'
import { LoginView } from './components/LoginView'
import { PhaseProgress } from './components/PhaseProgress'
import { ReadingsChart } from './components/ReadingsChart'
import { ManualReadingModal, PileModal } from './components/PileModal'
import { GuideView } from './components/GuideView'
import { AdminPanel } from './components/AdminPanel'
import { formatRelativeTime, formatShortDateTime, PHASE_STYLES } from './utils/format'

type NavTab = 'dashboard' | 'guide' | 'admin'

export function App() {
  const [booting, setBooting] = useState(true)
  const [servers, setServers] = useState<ServerProfile[]>([])
  const [selectedServerId, setSelectedServerId] = useState<string>('')
  const [session, setSession] = useState<SessionState | null>(null)

  const [navTab, setNavTab] = useState<NavTab>('dashboard')
  const [pileFilter, setPileFilter] = useState<'active' | 'archived'>('active')
  const [piles, setPiles] = useState<Pile[]>([])
  const [selectedPileId, setSelectedPileId] = useState<number | null>(null)
  const [readings, setReadings] = useState<Reading[]>([])

  const [loadingPiles, setLoadingPiles] = useState(false)
  const [loadingReadings, setLoadingReadings] = useState(false)
  const [syncing, setSyncing] = useState(false)
  const [errorBanner, setErrorBanner] = useState<string | null>(null)

  const [pileModalMode, setPileModalMode] = useState<'create' | 'edit' | null>(null)
  const [manualReadingOpen, setManualReadingOpen] = useState(false)

  // Arranque inicial: cargar servidores y restaurar sesión si existe token
  useEffect(() => {
    let mounted = true
    async function init() {
      const list = await getServers()
      if (!mounted) return
      setServers(list)
      const activeId = getActiveServerId() || list[0]?.id || ''
      setSelectedServerId(activeId)

      const target = list.find((s) => s.id === activeId)
      if (target) {
        const restored = await restoreSession(target)
        if (mounted && restored) {
          setSession({
            server: target,
            user: restored.user,
            offline: restored.offline,
            lastSyncAt: restored.lastSyncAt,
          })
        }
      }
      if (mounted) setBooting(false)
    }
    void init()
    return () => {
      mounted = false
    }
  }, [])

  const loadPiles = useCallback(
    async (silent = false) => {
      if (!session) return
      if (!silent) setLoadingPiles(true)
      else setSyncing(true)
      setErrorBanner(null)
      try {
        const res = await fetchPiles(session.server, 'all')
        setPiles(res.piles)
        setSession((prev) =>
          prev
            ? {
                ...prev,
                offline: res.offline,
                lastSyncAt: res.lastSyncAt ?? prev.lastSyncAt,
              }
            : null,
        )
      } catch (err) {
        const status = (err as { status?: number }).status
        if (status === 401) {
          await logout(session.server)
          setSession(null)
          return
        }
        setErrorBanner(err instanceof Error ? err.message : 'Error cargando pilas')
      } finally {
        setLoadingPiles(false)
        setSyncing(false)
      }
    },
    [session?.server.id],
  )

  useEffect(() => {
    if (!session) return
    void loadPiles(false)

    // Tiempo real instantáneo (SSE) en Web y Móvil + respaldo periódico para Wails/reconexión
    const unsubscribe = subscribeLiveEvents(session.server, () => {
      void loadPiles(true)
    })
    const timer = window.setInterval(() => {
      void loadPiles(true)
    }, 15_000)

    return () => {
      unsubscribe()
      window.clearInterval(timer)
    }
  }, [session?.server.id, loadPiles])

  const filteredPiles = useMemo(
    () => piles.filter((p) => p.status === pileFilter),
    [piles, pileFilter],
  )

  // Mantener una pila seleccionada automáticamente
  useEffect(() => {
    if (filteredPiles.length === 0) {
      setSelectedPileId(null)
      return
    }
    if (!filteredPiles.some((p) => p.id === selectedPileId)) {
      setSelectedPileId(filteredPiles[0].id)
    }
  }, [filteredPiles, selectedPileId])

  const selectedPile = useMemo(
    () => filteredPiles.find((p) => p.id === selectedPileId) ?? null,
    [filteredPiles, selectedPileId],
  )

  const loadReadings = useCallback(async () => {
    if (!session || !selectedPile) {
      setReadings([])
      return
    }
    setLoadingReadings(true)
    try {
      const res = await fetchReadings(session.server, selectedPile.id, 300)
      setReadings(res.readings)
    } catch {
      // Mantiene últimas lecturas si falla puntualmente
    } finally {
      setLoadingReadings(false)
    }
  }, [session?.server.id, selectedPile?.id, selectedPile?.updated_at])

  useEffect(() => {
    void loadReadings()
  }, [loadReadings])

  const handleAuthenticated = (server: ServerProfile, user: User, lastSyncAt: string | null) => {
    setSession({ server, user, offline: false, lastSyncAt })
    setNavTab('dashboard')
  }

  const handleLogout = async () => {
    if (!session) return
    await logout(session.server)
    setSession(null)
    setPiles([])
    setReadings([])
    setSelectedPileId(null)
  }

  const handleSavePile = async (input: PileInput) => {
    if (!session) return
    if (pileModalMode === 'edit' && selectedPile) {
      await updatePile(session.server, selectedPile.id, input)
    } else {
      const created = await createPile(session.server, input)
      setPileFilter('active')
      setSelectedPileId(created.id)
    }
    await loadPiles(true)
  }

  const handleDeletePile = async (pile: Pile) => {
    if (!session) return
    const ok = window.confirm(`¿Eliminar "${pile.name}" y todo su historial de lecturas? Esta acción no se puede deshacer.`)
    if (!ok) return
    try {
      await deletePile(session.server, pile.id)
      await loadPiles(true)
    } catch (err) {
      setErrorBanner(err instanceof Error ? err.message : 'No se pudo eliminar la pila')
    }
  }

  const handleManualReading = async (input: ManualReadingInput) => {
    if (!session || !selectedPile) return
    await addManualReading(session.server, selectedPile.id, input)
    await loadPiles(true)
    await loadReadings()
  }

  const handleExportCSV = async () => {
    if (!session || !selectedPile) return
    try {
      await exportPileCSV(session.server, selectedPile.id, selectedPile.name)
    } catch (err) {
      setErrorBanner(err instanceof Error ? err.message : 'No se pudo exportar CSV')
    }
  }

  if (booting) {
    return (
      <div className="min-h-dvh flex items-center justify-center bg-[#F8FAF9]">
        <div className="flex items-center gap-3 text-[#465C50]">
          <Loader2 className="h-6 w-6 animate-spin text-[#15613B]" />
          <span className="text-base font-medium">Cargando CompostIOT...</span>
        </div>
      </div>
    )
  }

  if (!session) {
    return (
      <LoginView
        servers={servers}
        selectedServerId={selectedServerId}
        onServersChanged={(next, id) => {
          setServers(next)
          setSelectedServerId(id)
        }}
        onAuthenticated={handleAuthenticated}
      />
    )
  }

  const isAdmin = session.user.role === 'admin'

  return (
    <div className="min-h-dvh flex flex-col bg-[#F8FAF9] text-[#12261C] pb-20 lg:pb-0">
      {/* Barra superior limpia de una sola línea */}
      <header className="sticky top-0 z-30 bg-white/95 backdrop-blur-xs border-b border-[#DCE5DF] safe-top">
        <div className="mx-auto max-w-[1440px] px-4 sm:px-6 h-16 flex items-center justify-between gap-4">
          <div className="flex items-center gap-6 min-w-0">
            <div className="flex items-center gap-2.5 shrink-0">
              <div className="h-9 w-9 rounded-xl bg-[#15613B] text-white flex items-center justify-center">
                <Sprout className="h-5 w-5" />
              </div>
              <span className="text-xl font-extrabold tracking-tight text-[#12261C]">
                CompostIOT
              </span>
            </div>

            {/* Navegación principal escritorio */}
            <nav aria-label="Principal" className="hidden lg:flex items-center gap-1">
              <button
                type="button"
                onClick={() => setNavTab('dashboard')}
                className={`px-3.5 py-2 rounded-lg text-sm font-semibold inline-flex items-center gap-2 transition-[background-color,color] duration-150 ${
                  navTab === 'dashboard'
                    ? 'bg-[#EDF4F0] text-[#15613B]'
                    : 'text-[#465C50] hover:text-[#12261C] hover:bg-[#F8FAF9]'
                }`}
              >
                <LayoutDashboard className="h-4 w-4" />
                Tablero
              </button>
              <button
                type="button"
                onClick={() => setNavTab('guide')}
                className={`px-3.5 py-2 rounded-lg text-sm font-semibold inline-flex items-center gap-2 transition-[background-color,color] duration-150 ${
                  navTab === 'guide'
                    ? 'bg-[#EDF4F0] text-[#15613B]'
                    : 'text-[#465C50] hover:text-[#12261C] hover:bg-[#F8FAF9]'
                }`}
              >
                <BookOpen className="h-4 w-4" />
                Guía de fases
              </button>
              {isAdmin && (
                <button
                  type="button"
                  onClick={() => setNavTab('admin')}
                  className={`px-3.5 py-2 rounded-lg text-sm font-semibold inline-flex items-center gap-2 transition-[background-color,color] duration-150 ${
                    navTab === 'admin'
                      ? 'bg-[#EDF4F0] text-[#15613B]'
                      : 'text-[#465C50] hover:text-[#12261C] hover:bg-[#F8FAF9]'
                  }`}
                >
                  <Settings className="h-4 w-4" />
                  Ajustes
                </button>
              )}
            </nav>
          </div>

          <div className="flex items-center gap-2 sm:gap-3 min-w-0">
            <div className="hidden md:flex items-center gap-2 text-sm text-[#465C50] truncate max-w-[230px]">
              <Server className="h-4 w-4 text-[#15613B] shrink-0" />
              <span className="truncate">{session.server.name}</span>
            </div>

            <button
              type="button"
              onClick={() => void loadPiles(true)}
              disabled={syncing}
              title="Actualizar lecturas"
              className="p-2 rounded-lg border border-[#DCE5DF] bg-white hover:bg-[#EDF4F0] text-[#12261C] transition-[background-color] duration-150"
            >
              <RefreshCw className={`h-4 w-4 ${syncing ? 'animate-spin text-[#15613B]' : ''}`} />
            </button>

            <div className="hidden sm:block text-right leading-tight">
              <p className="text-sm font-semibold text-[#12261C] truncate max-w-[160px]">
                {session.user.full_name}
              </p>
              <p className="text-sm text-[#465C50]">
                {isAdmin ? 'Instructor' : 'Aprendiz'}
              </p>
            </div>

            <button
              type="button"
              onClick={handleLogout}
              title="Cambiar servidor o salir"
              className="p-2 rounded-lg border border-[#DCE5DF] bg-white hover:bg-red-50 hover:border-red-200 text-[#465C50] hover:text-red-700 transition-[background-color,color,border-color] duration-150"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </header>

      {/* Banner de modo offline */}
      {session.offline && (
        <div className="bg-amber-50 border-b border-amber-200 px-4 py-2.5">
          <div className="mx-auto max-w-[1440px] flex flex-wrap items-center justify-between gap-2 text-sm text-amber-900">
            <div className="flex items-center gap-2">
              <WifiOff className="h-4 w-4 shrink-0" />
              <span>
                <strong>Modo sin conexión:</strong> mostrando datos del caché local
                {session.lastSyncAt ? ` (última sincronización: ${formatShortDateTime(session.lastSyncAt)})` : ''}.
              </span>
            </div>
            <button
              type="button"
              onClick={() => void loadPiles(true)}
              className="font-semibold underline hover:text-amber-950"
            >
              Reintentar conexión
            </button>
          </div>
        </div>
      )}

      {errorBanner && (
        <div className="bg-red-50 border-b border-red-200 px-4 py-2.5 text-sm text-red-800">
          <div className="mx-auto max-w-[1440px] flex items-center justify-between gap-3">
            <span>{errorBanner}</span>
            <button type="button" onClick={() => setErrorBanner(null)} className="font-semibold underline">
              Ocultar
            </button>
          </div>
        </div>
      )}

      {/* Contenido principal */}
      <main className="flex-1 mx-auto w-full max-w-[1440px] px-4 sm:px-6 py-6">
        {navTab === 'guide' && <GuideView />}
        {navTab === 'admin' && isAdmin && <AdminPanel server={session.server} currentUser={session.user} />}

        {navTab === 'dashboard' && (
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
            {/* Selector de pilas: sidebar en escritorio, tira deslizable en móvil/tablet */}
            <aside className="lg:col-span-4 xl:col-span-3 space-y-4">
              <div className="flex items-center justify-between gap-2">
                <h2 className="text-xl font-bold text-[#12261C]">Pilas</h2>
                {isAdmin && !session.offline && (
                  <button
                    type="button"
                    onClick={() => setPileModalMode('create')}
                    className="rounded-xl bg-[#15613B] hover:bg-[#0F4C2D] text-white px-3.5 py-2 text-sm font-semibold inline-flex items-center gap-1.5 transition-[background-color] duration-150"
                  >
                    <Plus className="h-4 w-4" />
                    Nueva pila
                  </button>
                )}
              </div>

              {/* Filtro Activas / Archivadas */}
              <div className="grid grid-cols-2 rounded-xl bg-[#EDF4F0] p-1 border border-[#DCE5DF]">
                {(
                  [
                    { id: 'active', label: `Activas (${piles.filter((p) => p.status === 'active').length})` },
                    { id: 'archived', label: `Archivadas (${piles.filter((p) => p.status === 'archived').length})` },
                  ] as const
                ).map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => setPileFilter(item.id)}
                    className={`rounded-lg py-1.5 text-sm font-semibold transition-[background-color,color] duration-150 ${
                      pileFilter === item.id ? 'bg-white text-[#12261C] shadow-xs' : 'text-[#465C50]'
                    }`}
                  >
                    {item.label}
                  </button>
                ))}
              </div>

              {loadingPiles ? (
                <div className="py-10 flex items-center justify-center gap-2 text-[#465C50]">
                  <Loader2 className="h-5 w-5 animate-spin text-[#15613B]" />
                  <span>Cargando pilas...</span>
                </div>
              ) : filteredPiles.length === 0 ? (
                <div className="rounded-2xl bg-white border border-dashed border-[#DCE5DF] p-6 text-center space-y-2">
                  <p className="text-base font-semibold text-[#12261C]">
                    No hay pilas {pileFilter === 'active' ? 'activas' : 'archivadas'}
                  </p>
                  <p className="text-sm text-[#465C50]">
                    {isAdmin
                      ? 'Crea tu primera pila con el botón "Nueva pila" para comenzar el seguimiento.'
                      : 'Espera a que tu instructor registre una pila de compostaje.'}
                  </p>
                </div>
              ) : (
                <div className="flex lg:flex-col gap-3 overflow-x-auto pb-1 lg:pb-0 snap-x">
                  {filteredPiles.map((pile) => {
                    const selected = pile.id === selectedPile?.id
                    const phaseStyle = PHASE_STYLES[pile.evaluation.phase.code]
                    return (
                      <button
                        key={pile.id}
                        type="button"
                        onClick={() => setSelectedPileId(pile.id)}
                        className={`snap-start text-left shrink-0 w-[265px] lg:w-full rounded-2xl border p-4 transition-[background-color,border-color,box-shadow] duration-150 ${
                          selected
                            ? 'bg-white border-[#15613B] shadow-[0_0_0_1px_#15613B]'
                            : 'bg-white border-[#DCE5DF] hover:border-[#9EB5A7]'
                        }`}
                      >
                        <div className="flex items-start justify-between gap-2">
                          <p className="text-base font-bold text-[#12261C] leading-snug line-clamp-2">
                            {pile.name}
                          </p>
                          <span
                            className="h-2.5 w-2.5 rounded-full shrink-0 mt-1.5"
                            style={{ backgroundColor: phaseStyle.accent }}
                            title={pile.evaluation.phase.label}
                          />
                        </div>

                        <p className="text-sm font-semibold mt-1" style={{ color: phaseStyle.text }}>
                          {phaseStyle.shortLabel}
                        </p>

                        <div className="mt-3 pt-2.5 border-t border-[#EDF4F0] flex items-center justify-between text-sm text-[#465C50] tabular">
                          <span>
                            {pile.latest_reading
                              ? `${pile.latest_reading.temperature_c.toFixed(1)} °C · ${pile.latest_reading.humidity_pct.toFixed(0)} %`
                              : 'Sin datos'}
                          </span>
                          <span>{pile.node_id}</span>
                        </div>
                      </button>
                    )
                  })}
                </div>
              )}
            </aside>

            {/* Panel de detalle de la pila seleccionada */}
            <div className="lg:col-span-8 xl:col-span-9 space-y-6 min-w-0">
              {!selectedPile ? (
                <div className="rounded-2xl bg-white border border-[#DCE5DF] p-10 text-center space-y-2">
                  <h2 className="text-xl font-bold text-[#12261C]">Selecciona una pila para ver sus datos</h2>
                  <p className="text-base text-[#465C50]">
                    Aquí verás temperatura, humedad, pH opcional, la fase actual del proceso y las curvas históricas.
                  </p>
                </div>
              ) : (
                <>
                  {/* Encabezado de la pila */}
                  <section className="bg-white border border-[#DCE5DF] rounded-2xl p-5 sm:p-6">
                    <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-4">
                      <div className="space-y-1.5 min-w-0">
                        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-[#12261C] leading-tight">
                          {selectedPile.name}
                        </h1>
                        <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-[#465C50]">
                          {selectedPile.location && (
                            <span className="inline-flex items-center gap-1">
                              <MapPin className="h-4 w-4 text-[#15613B]" />
                              {selectedPile.location}
                            </span>
                          )}
                          <span>Nodo: <strong className="text-[#12261C]">{selectedPile.node_id}</strong></span>
                          <span className="tabular">
                            Última lectura: <strong className="text-[#12261C]">{formatRelativeTime(selectedPile.latest_reading?.recorded_at)}</strong>
                          </span>
                        </div>
                        {selectedPile.substrate && (
                          <p className="text-base text-[#465C50] pt-1">
                            Sustrato: {selectedPile.substrate}
                          </p>
                        )}
                      </div>

                      <div className="flex flex-wrap items-center gap-2 shrink-0">
                        <button
                          type="button"
                          onClick={handleExportCSV}
                          disabled={session.offline}
                          className="rounded-xl border border-[#DCE5DF] bg-white hover:bg-[#EDF4F0] px-3.5 py-2 text-sm font-semibold text-[#12261C] inline-flex items-center gap-1.5 transition-[background-color] duration-150 disabled:opacity-50"
                        >
                          <Download className="h-4 w-4 text-[#15613B]" />
                          Exportar CSV
                        </button>

                        {isAdmin && !session.offline && (
                          <>
                            <button
                              type="button"
                              onClick={() => setManualReadingOpen(true)}
                              className="rounded-xl border border-[#DCE5DF] bg-white hover:bg-[#EDF4F0] px-3.5 py-2 text-sm font-semibold text-[#12261C] inline-flex items-center gap-1.5 transition-[background-color] duration-150"
                            >
                              <PlusCircle className="h-4 w-4 text-[#15613B]" />
                              Medición
                            </button>
                            <button
                              type="button"
                              onClick={() => setPileModalMode('edit')}
                              title="Editar pila"
                              className="p-2 rounded-xl border border-[#DCE5DF] bg-white hover:bg-[#EDF4F0] text-[#465C50] hover:text-[#12261C]"
                            >
                              <Pencil className="h-4 w-4" />
                            </button>
                            <button
                              type="button"
                              onClick={() => void handleDeletePile(selectedPile)}
                              title="Eliminar pila"
                              className="p-2 rounded-xl border border-[#DCE5DF] bg-white hover:bg-red-50 hover:border-red-200 text-[#465C50] hover:text-red-700"
                            >
                              <Trash2 className="h-4 w-4" />
                            </button>
                          </>
                        )}
                      </div>
                    </div>
                  </section>

                  {/* Tarjetas grandes de variables (legibles bajo luz de campo) */}
                  <section
                    className={`grid grid-cols-1 sm:grid-cols-2 ${
                      selectedPile.has_ph_sensor || selectedPile.latest_reading?.ph !== null
                        ? 'xl:grid-cols-3'
                        : ''
                    } gap-4`}
                  >
                    {/* Temperatura */}
                    <div className="bg-white border border-[#DCE5DF] rounded-2xl p-5 sm:p-6 flex flex-col justify-between gap-3">
                      <div className="flex items-center justify-between">
                        <span className="text-base font-semibold text-[#465C50]">Temperatura interna</span>
                        <Thermometer className="h-5 w-5 text-[#D95F02]" />
                      </div>
                      <div className="flex items-baseline gap-1.5 tabular">
                        <span className="text-4xl sm:text-5xl font-extrabold tracking-tight text-[#12261C]">
                          {selectedPile.latest_reading ? selectedPile.latest_reading.temperature_c.toFixed(1) : '--'}
                        </span>
                        <span className="text-xl font-semibold text-[#465C50]">°C</span>
                      </div>
                      <p className="text-sm text-[#465C50]">
                        Rango termófilo óptimo: 45 a 65 °C
                      </p>
                    </div>

                    {/* Humedad */}
                    <div className="bg-white border border-[#DCE5DF] rounded-2xl p-5 sm:p-6 flex flex-col justify-between gap-3">
                      <div className="flex items-center justify-between">
                        <span className="text-base font-semibold text-[#465C50]">Humedad del sustrato</span>
                        <Droplets className="h-5 w-5 text-[#0284C7]" />
                      </div>
                      <div className="flex items-baseline gap-1.5 tabular">
                        <span className="text-4xl sm:text-5xl font-extrabold tracking-tight text-[#12261C]">
                          {selectedPile.latest_reading ? selectedPile.latest_reading.humidity_pct.toFixed(1) : '--'}
                        </span>
                        <span className="text-xl font-semibold text-[#465C50]">%</span>
                      </div>
                      <p className="text-sm text-[#465C50]">
                        Rango saludable: 40 a 65 %
                      </p>
                    </div>

                    {/* pH (Opcional) */}
                    {(selectedPile.has_ph_sensor || selectedPile.latest_reading?.ph !== null) && (
                      <div className="bg-white border border-[#DCE5DF] rounded-2xl p-5 sm:p-6 flex flex-col justify-between gap-3 sm:col-span-2 xl:col-span-1">
                        <div className="flex items-center justify-between">
                          <span className="text-base font-semibold text-[#465C50]">pH (opcional)</span>
                          <FlaskConical className="h-5 w-5 text-[#15613B]" />
                        </div>
                        <div className="flex items-baseline gap-1.5 tabular">
                          <span className="text-4xl sm:text-5xl font-extrabold tracking-tight text-[#12261C]">
                            {selectedPile.latest_reading?.ph !== null && selectedPile.latest_reading?.ph !== undefined
                              ? selectedPile.latest_reading.ph.toFixed(2)
                              : '--'}
                          </span>
                          <span className="text-xl font-semibold text-[#465C50]">pH</span>
                        </div>
                        <p className="text-sm text-[#465C50]">
                          Rango equilibrado: 5.5 a 8.5
                        </p>
                      </div>
                    )}
                  </section>

                  {/* Fase visual del compostaje */}
                  <PhaseProgress evaluation={selectedPile.evaluation} />

                  {/* Alertas y recomendaciones prácticas */}
                  <section className="bg-white border border-[#DCE5DF] rounded-2xl p-5 sm:p-6 space-y-3">
                    <h3 className="text-xl font-bold text-[#12261C]">Diagnóstico y recomendaciones</h3>
                    <div className="space-y-2.5">
                      {selectedPile.evaluation.alerts.map((alert) => {
                        const styles = {
                          danger: 'bg-red-50 border-red-200 text-red-900',
                          warning: 'bg-amber-50 border-amber-200 text-amber-950',
                          info: 'bg-sky-50 border-sky-200 text-sky-950',
                          ok: 'bg-[#F0FDF4] border-[#BBF7D0] text-[#14532D]',
                        }[alert.severity]

                        const Icon =
                          alert.severity === 'ok'
                            ? CheckCircle2
                            : alert.severity === 'info'
                              ? Info
                              : AlertTriangle

                        return (
                          <div key={alert.code} className={`rounded-xl border px-4 py-3.5 flex gap-3 items-start ${styles}`}>
                            <Icon className="h-5 w-5 shrink-0 mt-0.5" />
                            <div className="space-y-0.5">
                              <p className="text-base font-bold">{alert.title}</p>
                              <p className="text-sm opacity-90 leading-relaxed">{alert.message}</p>
                            </div>
                          </div>
                        )
                      })}
                    </div>
                  </section>

                  {/* Gráfica histórica */}
                  {loadingReadings && readings.length === 0 ? (
                    <div className="bg-white border border-[#DCE5DF] rounded-2xl p-8 flex items-center justify-center gap-2 text-[#465C50]">
                      <Loader2 className="h-5 w-5 animate-spin text-[#15613B]" />
                      <span>Cargando histórico...</span>
                    </div>
                  ) : (
                    <ReadingsChart readings={readings} showPH={selectedPile.has_ph_sensor} />
                  )}

                  {/* Notas y últimas mediciones */}
                  <section className="bg-white border border-[#DCE5DF] rounded-2xl p-5 sm:p-6 space-y-4">
                    <div className="flex items-center justify-between gap-2">
                      <h3 className="text-xl font-bold text-[#12261C]">Últimas lecturas registradas</h3>
                      <span className="text-sm text-[#465C50] tabular">{readings.length} registros</span>
                    </div>

                    {selectedPile.notes && (
                      <p className="text-base text-[#465C50] bg-[#F8FAF9] rounded-xl px-4 py-3 border border-[#DCE5DF]">
                        <strong className="text-[#12261C]">Notas de la pila:</strong> {selectedPile.notes}
                      </p>
                    )}

                    {readings.length > 0 && (
                      <div className="overflow-x-auto">
                        <table className="w-full text-left border-collapse text-sm">
                          <thead>
                            <tr className="border-b border-[#DCE5DF] text-[#465C50]">
                              <th className="py-2.5 pr-4 font-semibold">Fecha y hora</th>
                              <th className="py-2.5 px-3 font-semibold">Temperatura</th>
                              <th className="py-2.5 px-3 font-semibold">Humedad</th>
                              {(selectedPile.has_ph_sensor || readings.some((r) => r.ph !== null)) && (
                                <th className="py-2.5 px-3 font-semibold">pH</th>
                              )}
                              <th className="py-2.5 pl-3 font-semibold">Fase</th>
                            </tr>
                          </thead>
                          <tbody className="divide-y divide-[#EDF4F0] tabular">
                            {readings
                              .slice(-8)
                              .reverse()
                              .map((r) => {
                                const ps = PHASE_STYLES[r.phase]
                                return (
                                  <tr key={r.id}>
                                    <td className="py-2.5 pr-4 text-[#12261C] whitespace-nowrap">
                                      {formatShortDateTime(r.recorded_at)}
                                    </td>
                                    <td className="py-2.5 px-3 font-semibold text-[#12261C]">
                                      {r.temperature_c.toFixed(1)} °C
                                    </td>
                                    <td className="py-2.5 px-3 text-[#12261C]">{r.humidity_pct.toFixed(1)} %</td>
                                    {(selectedPile.has_ph_sensor || readings.some((item) => item.ph !== null)) && (
                                      <td className="py-2.5 px-3 text-[#12261C]">
                                        {r.ph !== null ? r.ph.toFixed(2) : 'N/A'}
                                      </td>
                                    )}
                                    <td className="py-2.5 pl-3 font-semibold whitespace-nowrap" style={{ color: ps.text }}>
                                      {ps.shortLabel}
                                    </td>
                                  </tr>
                                )
                              })}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </section>
                </>
              )}
            </div>
          </div>
        )}
      </main>

      {/* Barra de navegación inferior en teléfono y tablet */}
      <nav
        aria-label="Navegación móvil"
        className="lg:hidden fixed bottom-0 inset-x-0 z-30 bg-white/95 backdrop-blur-xs border-t border-[#DCE5DF] px-2 pt-1.5 pb-1.5 safe-bottom"
      >
        <div className={`grid ${isAdmin ? 'grid-cols-3' : 'grid-cols-2'} max-w-lg mx-auto`}>
          <button
            type="button"
            onClick={() => setNavTab('dashboard')}
            className={`flex flex-col items-center justify-center py-1.5 rounded-xl text-sm font-semibold ${
              navTab === 'dashboard' ? 'text-[#15613B]' : 'text-[#465C50]'
            }`}
          >
            <LayoutDashboard className="h-5 w-5 mb-0.5" />
            Tablero
          </button>
          <button
            type="button"
            onClick={() => setNavTab('guide')}
            className={`flex flex-col items-center justify-center py-1.5 rounded-xl text-sm font-semibold ${
              navTab === 'guide' ? 'text-[#15613B]' : 'text-[#465C50]'
            }`}
          >
            <BookOpen className="h-5 w-5 mb-0.5" />
            Guía
          </button>
          {isAdmin && (
            <button
              type="button"
              onClick={() => setNavTab('admin')}
              className={`flex flex-col items-center justify-center py-1.5 rounded-xl text-sm font-semibold ${
                navTab === 'admin' ? 'text-[#15613B]' : 'text-[#465C50]'
              }`}
            >
              <Settings className="h-5 w-5 mb-0.5" />
              Ajustes
            </button>
          )}
        </div>
      </nav>

      {/* Modales */}
      {pileModalMode && (
        <PileModal
          pile={pileModalMode === 'edit' ? selectedPile : null}
          onClose={() => setPileModalMode(null)}
          onSave={handleSavePile}
        />
      )}

      {manualReadingOpen && selectedPile && (
        <ManualReadingModal
          pile={selectedPile}
          onClose={() => setManualReadingOpen(false)}
          onSubmit={handleManualReading}
        />
      )}
    </div>
  )
}

export default App
