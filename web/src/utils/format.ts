import type { PhaseCode } from '../types'

export interface PhaseStyle {
  accent: string
  bgSoft: string
  border: string
  text: string
  shortLabel: string
}

export const PHASE_STYLES: Record<PhaseCode, PhaseStyle> = {
  sin_datos: {
    accent: '#64748B',
    bgSoft: '#F1F5F9',
    border: '#CBD5E1',
    text: '#334155',
    shortLabel: 'Sin lecturas',
  },
  mesofila_inicial: {
    accent: '#2563EB',
    bgSoft: '#EFF6FF',
    border: '#BFDBFE',
    text: '#1D4ED8',
    shortLabel: 'Mesófila inicial',
  },
  termofila: {
    accent: '#D95F02',
    bgSoft: '#FFF7ED',
    border: '#FED7AA',
    text: '#C2410C',
    shortLabel: 'Termófila',
  },
  enfriamiento: {
    accent: '#B47B00',
    bgSoft: '#FEFCE8',
    border: '#FDE68A',
    text: '#92400E',
    shortLabel: 'Enfriamiento',
  },
  maduracion: {
    accent: '#15803D',
    bgSoft: '#F0FDF4',
    border: '#BBF7D0',
    text: '#166534',
    shortLabel: 'Maduración',
  },
}

export function formatShortDateTime(iso: string | null | undefined): string {
  if (!iso) return 'Sin registro'
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return iso
  return new Intl.DateTimeFormat('es-CO', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

export function formatChartAxisTime(iso: string): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat('es-CO', {
    day: '2-digit',
    month: 'short',
  }).format(date)
}

export function formatRelativeTime(iso: string | null | undefined): string {
  if (!iso) return 'Sin lecturas'
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  const diffMin = Math.floor((Date.now() - date.getTime()) / 60000)
  if (diffMin < 1) return 'Hace instantes'
  if (diffMin < 60) return `Hace ${diffMin} min`
  const diffHours = Math.floor(diffMin / 60)
  if (diffHours < 24) return `Hace ${diffHours} h`
  const diffDays = Math.floor(diffHours / 24)
  return `Hace ${diffDays} d`
}
