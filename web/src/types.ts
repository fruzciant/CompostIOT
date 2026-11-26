export type UserRole = 'admin' | 'aprendiz'

export type PhaseCode =
  | 'sin_datos'
  | 'mesofila_inicial'
  | 'termofila'
  | 'enfriamiento'
  | 'maduracion'

export interface User {
  id: number
  username: string
  full_name: string
  role: UserRole
  active: boolean
  created_at: string
}

export interface PhaseInfo {
  code: PhaseCode
  label: string
  description: string
  color: string
  progress_pct: number
}

export interface Alert {
  code: string
  severity: 'ok' | 'info' | 'warning' | 'danger'
  title: string
  message: string
}

export interface Evaluation {
  phase: PhaseInfo
  ref_temperature_c: number | null
  has_reached_thermophilic: boolean
  peak_temperature_c: number | null
  days_elapsed: number
  alerts: Alert[]
}

export interface Reading {
  id: number
  pile_id: number
  temperature_c: number
  humidity_pct: number
  ph: number | null
  phase: PhaseCode
  power_source: 'mains' | 'battery'
  battery_v: number | null
  rssi: number | null
  seq: number | null
  recorded_at: string
}

export interface Pile {
  id: number
  name: string
  location: string
  substrate: string
  node_id: string
  has_ph_sensor: boolean
  status: 'active' | 'archived'
  notes: string
  started_at: string
  created_at: string
  updated_at: string
  latest_reading: Reading | null
  evaluation: Evaluation
}

export interface Device {
  id: number
  name: string
  key_prefix: string
  active: boolean
  last_seen_at: string | null
  created_at: string
}

export interface HealthInfo {
  status: string
  service: string
  version: string
  server_name: string
  server_time: string
}

export interface ServerProfile {
  id: string
  name: string
  url: string
  username?: string
  lastSyncAt?: string
}

export interface SessionState {
  server: ServerProfile
  user: User
  offline: boolean
  lastSyncAt: string | null
}

export interface PileInput {
  name: string
  location: string
  substrate: string
  node_id: string
  has_ph_sensor: boolean
  notes: string
  status?: 'active' | 'archived'
  started_at?: string
}

export interface ManualReadingInput {
  temperature_c: number
  humidity_pct: number
  ph: number | null
}
