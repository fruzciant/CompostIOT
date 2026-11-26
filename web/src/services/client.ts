import type {
  Device,
  HealthInfo,
  ManualReadingInput,
  Pile,
  PileInput,
  Reading,
  ServerProfile,
  User,
} from '../types'

// Contrato opcional inyectado por el cliente de escritorio Go (Wails) en Fase 3.
interface WailsDesktopBridge {
  GetServers(): Promise<ServerProfile[]>
  SaveServer(profile: ServerProfile): Promise<ServerProfile>
  DeleteServer(id: string): Promise<void>
  CheckHealth(serverUrl: string): Promise<HealthInfo>
  Login(serverId: string, username: string, password: string): Promise<{ user: User; lastSyncAt: string }>
  RestoreSession(serverId: string): Promise<{ user: User; offline: boolean; lastSyncAt: string | null }>
  Logout(serverId: string): Promise<void>
  SyncServer(serverId: string): Promise<{ piles: Pile[]; lastSyncAt: string; offline: boolean }>
  GetPiles(serverId: string, status: string): Promise<{ piles: Pile[]; offline: boolean; lastSyncAt: string | null }>
  GetReadings(serverId: string, pileId: number, limit: number): Promise<{ readings: Reading[]; offline: boolean }>
  CreatePile(serverId: string, input: PileInput): Promise<Pile>
  UpdatePile(serverId: string, pileId: number, input: PileInput): Promise<Pile>
  DeletePile(serverId: string, pileId: number): Promise<void>
  AddManualReading(serverId: string, pileId: number, input: ManualReadingInput): Promise<Reading>
  ExportCSV(serverId: string, pileId: number): Promise<string>
  GetUsers(serverId: string): Promise<User[]>
  CreateUser(serverId: string, payload: { username: string; full_name: string; password: string; role: string }): Promise<User>
  DeleteUser(serverId: string, userId: number): Promise<void>
  GetDevices(serverId: string): Promise<Device[]>
  CreateDevice(serverId: string, name: string): Promise<{ device: Device; api_key: string }>
  DeleteDevice(serverId: string, deviceId: number): Promise<void>
}

declare global {
  interface Window {
    go?: {
      main?: {
        App?: WailsDesktopBridge
      }
    }
  }
}

const STORAGE_SERVERS = 'compostiot_servers_v1'
const STORAGE_ACTIVE_SERVER = 'compostiot_active_server_v1'
const STORAGE_TOKEN_PREFIX = 'compostiot_token_'
const STORAGE_USER_PREFIX = 'compostiot_user_'
const STORAGE_CACHE_PILES = 'compostiot_cache_piles_'
const STORAGE_CACHE_READINGS = 'compostiot_cache_readings_'
const STORAGE_LAST_SYNC = 'compostiot_last_sync_'

export function isWailsDesktop(): boolean {
  return Boolean(window.go?.main?.App)
}

function getWails(): WailsDesktopBridge | null {
  return window.go?.main?.App ?? null
}

export function normalizeServerUrl(url: string): string {
  const trimmed = url.trim().replace(/\/+$/, '')
  if (!trimmed) return 'http://localhost:8080'
  if (!/^https?:\/\//i.test(trimmed)) {
    return `http://${trimmed}`
  }
  return trimmed
}

function defaultServers(): ServerProfile[] {
  const origin =
    typeof window !== 'undefined' &&
    window.location.protocol.startsWith('http') &&
    window.location.port !== '5173'
      ? window.location.origin
      : 'http://localhost:8080'

  return [
    {
      id: 'local-default',
      name: 'Servidor local (Localhost)',
      url: origin,
    },
  ]
}

export async function getServers(): Promise<ServerProfile[]> {
  const wails = getWails()
  if (wails) return wails.GetServers()

  try {
    const raw = localStorage.getItem(STORAGE_SERVERS)
    if (!raw) {
      const defaults = defaultServers()
      localStorage.setItem(STORAGE_SERVERS, JSON.stringify(defaults))
      return defaults
    }
    const parsed = JSON.parse(raw) as ServerProfile[]
    return parsed.length > 0 ? parsed : defaultServers()
  } catch {
    return defaultServers()
  }
}

export async function saveServer(profile: Omit<ServerProfile, 'id'> & { id?: string }): Promise<ServerProfile> {
  const clean: ServerProfile = {
    id: profile.id || `srv_${Date.now().toString(36)}`,
    name: profile.name.trim() || 'Servidor CompostIOT',
    url: normalizeServerUrl(profile.url),
    username: profile.username,
    lastSyncAt: profile.lastSyncAt,
  }

  const wails = getWails()
  if (wails) return wails.SaveServer(clean)

  const list = await getServers()
  const idx = list.findIndex((s) => s.id === clean.id)
  if (idx >= 0) {
    list[idx] = clean
  } else {
    list.push(clean)
  }
  localStorage.setItem(STORAGE_SERVERS, JSON.stringify(list))
  return clean
}

export async function deleteServer(id: string): Promise<void> {
  const wails = getWails()
  if (wails) return wails.DeleteServer(id)

  const list = (await getServers()).filter((s) => s.id !== id)
  localStorage.setItem(STORAGE_SERVERS, JSON.stringify(list.length > 0 ? list : defaultServers()))
  localStorage.removeItem(STORAGE_TOKEN_PREFIX + id)
  localStorage.removeItem(STORAGE_USER_PREFIX + id)
  localStorage.removeItem(STORAGE_CACHE_PILES + id)
  localStorage.removeItem(STORAGE_LAST_SYNC + id)
}

export function getActiveServerId(): string | null {
  return localStorage.getItem(STORAGE_ACTIVE_SERVER)
}

export function setActiveServerId(id: string | null): void {
  if (id) {
    localStorage.setItem(STORAGE_ACTIVE_SERVER, id)
  } else {
    localStorage.removeItem(STORAGE_ACTIVE_SERVER)
  }
}

async function httpRequest<T>(
  server: ServerProfile,
  path: string,
  options: RequestInit = {},
  requireAuth = true,
): Promise<T> {
  const headers = new Headers(options.headers)
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  if (requireAuth) {
    const token = localStorage.getItem(STORAGE_TOKEN_PREFIX + server.id)
    if (token) {
      headers.set('Authorization', `Bearer ${token}`)
    }
  }

  const resp = await fetch(`${normalizeServerUrl(server.url)}${path}`, {
    ...options,
    headers,
  })

  if (resp.status === 204) {
    return undefined as T
  }

  const contentType = resp.headers.get('Content-Type') || ''
  if (!resp.ok) {
    let message = `Error del servidor (${resp.status})`
    if (contentType.includes('application/json')) {
      try {
        const errBody = await resp.json()
        if (errBody?.error?.message) message = errBody.error.message
      } catch {
        // ignore parse error
      }
    }
    const error = new Error(message) as Error & { status?: number }
    error.status = resp.status
    throw error
  }

  if (contentType.includes('text/csv')) {
    return (await resp.text()) as T
  }
  return (await resp.json()) as T
}

export async function checkHealth(serverUrl: string): Promise<HealthInfo> {
  const wails = getWails()
  if (wails) return wails.CheckHealth(normalizeServerUrl(serverUrl))

  const tempProfile: ServerProfile = { id: 'probe', name: 'probe', url: normalizeServerUrl(serverUrl) }
  return httpRequest<HealthInfo>(tempProfile, '/api/v1/health', { method: 'GET' }, false)
}

export async function login(
  server: ServerProfile,
  username: string,
  password: string,
): Promise<{ user: User; lastSyncAt: string | null }> {
  const wails = getWails()
  if (wails) {
    const res = await wails.Login(server.id, username, password)
    setActiveServerId(server.id)
    return res
  }

  const data = await httpRequest<{ access_token: string; user: User }>(
    server,
    '/api/v1/auth/login',
    {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    },
    false,
  )
  localStorage.setItem(STORAGE_TOKEN_PREFIX + server.id, data.access_token)
  localStorage.setItem(STORAGE_USER_PREFIX + server.id, JSON.stringify(data.user))
  setActiveServerId(server.id)
  await saveServer({ ...server, username: data.user.username })
  return { user: data.user, lastSyncAt: localStorage.getItem(STORAGE_LAST_SYNC + server.id) }
}

export async function restoreSession(
  server: ServerProfile,
): Promise<{ user: User; offline: boolean; lastSyncAt: string | null } | null> {
  const wails = getWails()
  if (wails) {
    try {
      return await wails.RestoreSession(server.id)
    } catch {
      return null
    }
  }

  const token = localStorage.getItem(STORAGE_TOKEN_PREFIX + server.id)
  if (!token) return null

  try {
    const user = await httpRequest<User>(server, '/api/v1/auth/me')
    localStorage.setItem(STORAGE_USER_PREFIX + server.id, JSON.stringify(user))
    return { user, offline: false, lastSyncAt: localStorage.getItem(STORAGE_LAST_SYNC + server.id) }
  } catch (err) {
    const status = (err as { status?: number }).status
    if (status === 401) {
      localStorage.removeItem(STORAGE_TOKEN_PREFIX + server.id)
      return null
    }
    const cachedUserRaw = localStorage.getItem(STORAGE_USER_PREFIX + server.id)
    if (cachedUserRaw) {
      return {
        user: JSON.parse(cachedUserRaw) as User,
        offline: true,
        lastSyncAt: localStorage.getItem(STORAGE_LAST_SYNC + server.id),
      }
    }
    return null
  }
}

export async function logout(server: ServerProfile): Promise<void> {
  const wails = getWails()
  if (wails) {
    await wails.Logout(server.id)
    setActiveServerId(null)
    return
  }
  localStorage.removeItem(STORAGE_TOKEN_PREFIX + server.id)
  setActiveServerId(null)
}

export async function fetchPiles(
  server: ServerProfile,
  status: 'active' | 'archived' | 'all' = 'all',
): Promise<{ piles: Pile[]; offline: boolean; lastSyncAt: string | null }> {
  const wails = getWails()
  if (wails) return wails.GetPiles(server.id, status)

  try {
    const piles = await httpRequest<Pile[]>(server, `/api/v1/piles?status=${status}`)
    const now = new Date().toISOString()
    localStorage.setItem(STORAGE_CACHE_PILES + server.id, JSON.stringify(piles))
    localStorage.setItem(STORAGE_LAST_SYNC + server.id, now)
    return { piles, offline: false, lastSyncAt: now }
  } catch (err) {
    const statusErr = (err as { status?: number }).status
    if (statusErr === 401) throw err
    const cached = localStorage.getItem(STORAGE_CACHE_PILES + server.id)
    if (cached) {
      return {
        piles: JSON.parse(cached) as Pile[],
        offline: true,
        lastSyncAt: localStorage.getItem(STORAGE_LAST_SYNC + server.id),
      }
    }
    throw err
  }
}

export async function fetchReadings(
  server: ServerProfile,
  pileId: number,
  limit = 300,
): Promise<{ readings: Reading[]; offline: boolean }> {
  const wails = getWails()
  if (wails) return wails.GetReadings(server.id, pileId, limit)

  const cacheKey = `${STORAGE_CACHE_READINGS}${server.id}_${pileId}`
  try {
    const readings = await httpRequest<Reading[]>(server, `/api/v1/piles/${pileId}/readings?limit=${limit}`)
    localStorage.setItem(cacheKey, JSON.stringify(readings))
    return { readings, offline: false }
  } catch (err) {
    const statusErr = (err as { status?: number }).status
    if (statusErr === 401) throw err
    const cached = localStorage.getItem(cacheKey)
    if (cached) {
      return { readings: JSON.parse(cached) as Reading[], offline: true }
    }
    throw err
  }
}

export async function createPile(server: ServerProfile, input: PileInput): Promise<Pile> {
  const wails = getWails()
  if (wails) return wails.CreatePile(server.id, input)
  return httpRequest<Pile>(server, '/api/v1/piles', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export async function updatePile(server: ServerProfile, pileId: number, input: PileInput): Promise<Pile> {
  const wails = getWails()
  if (wails) return wails.UpdatePile(server.id, pileId, input)
  return httpRequest<Pile>(server, `/api/v1/piles/${pileId}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
}

export async function deletePile(server: ServerProfile, pileId: number): Promise<void> {
  const wails = getWails()
  if (wails) return wails.DeletePile(server.id, pileId)
  await httpRequest<void>(server, `/api/v1/piles/${pileId}`, { method: 'DELETE' })
}

export async function addManualReading(
  server: ServerProfile,
  pileId: number,
  input: ManualReadingInput,
): Promise<Reading> {
  const wails = getWails()
  if (wails) return wails.AddManualReading(server.id, pileId, input)
  const res = await httpRequest<{ reading: Reading }>(server, `/api/v1/piles/${pileId}/readings`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
  return res.reading
}

export async function exportPileCSV(server: ServerProfile, pileId: number, pileName: string): Promise<void> {
  const wails = getWails()
  let csvText: string
  if (wails) {
    csvText = await wails.ExportCSV(server.id, pileId)
  } else {
    csvText = await httpRequest<string>(server, `/api/v1/piles/${pileId}/export?format=csv`)
  }
  const blob = new Blob([csvText], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  const safeName = pileName.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')
  link.download = `compostiot-${safeName || `pila-${pileId}`}.csv`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

export async function fetchUsers(server: ServerProfile): Promise<User[]> {
  const wails = getWails()
  if (wails) return wails.GetUsers(server.id)
  return httpRequest<User[]>(server, '/api/v1/users')
}

export async function createUser(
  server: ServerProfile,
  payload: { username: string; full_name: string; password: string; role: string },
): Promise<User> {
  const wails = getWails()
  if (wails) return wails.CreateUser(server.id, payload)
  return httpRequest<User>(server, '/api/v1/users', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export async function deleteUser(server: ServerProfile, userId: number): Promise<void> {
  const wails = getWails()
  if (wails) return wails.DeleteUser(server.id, userId)
  await httpRequest<void>(server, `/api/v1/users/${userId}`, { method: 'DELETE' })
}

export async function fetchDevices(server: ServerProfile): Promise<Device[]> {
  const wails = getWails()
  if (wails) return wails.GetDevices(server.id)
  return httpRequest<Device[]>(server, '/api/v1/devices')
}

export async function createDevice(
  server: ServerProfile,
  name: string,
): Promise<{ device: Device; api_key: string }> {
  const wails = getWails()
  if (wails) return wails.CreateDevice(server.id, name)
  return httpRequest<{ device: Device; api_key: string }>(server, '/api/v1/devices', {
    method: 'POST',
    body: JSON.stringify({ name }),
  })
}

export async function deleteDevice(server: ServerProfile, deviceId: number): Promise<void> {
  const wails = getWails()
  if (wails) return wails.DeleteDevice(server.id, deviceId)
  await httpRequest<void>(server, `/api/v1/devices/${deviceId}`, { method: 'DELETE' })
}
