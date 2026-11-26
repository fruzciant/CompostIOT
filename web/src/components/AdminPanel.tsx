import { useEffect, useState, type FormEvent } from 'react'
import { Check, Copy, KeyRound, Loader2, Plus, Radio, Trash2, Users } from 'lucide-react'
import type { Device, ServerProfile, User } from '../types'
import {
  createDevice,
  createUser,
  deleteDevice,
  deleteUser,
  fetchDevices,
  fetchUsers,
} from '../services/client'
import { formatShortDateTime } from '../utils/format'

interface AdminPanelProps {
  server: ServerProfile
  currentUser: User
}

export function AdminPanel({ server, currentUser }: AdminPanelProps) {
  const [devices, setDevices] = useState<Device[]>([])
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [newDeviceName, setNewDeviceName] = useState('')
  const [createdKey, setCreatedKey] = useState<{ name: string; key: string } | null>(null)
  const [copiedKey, setCopiedKey] = useState(false)

  const [newUsername, setNewUsername] = useState('')
  const [newFullName, setNewFullName] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [newRole, setNewRole] = useState<'aprendiz' | 'admin'>('aprendiz')

  const loadAll = async () => {
    setLoading(true)
    setError(null)
    try {
      const [devs, usrs] = await Promise.all([fetchDevices(server), fetchUsers(server)])
      setDevices(devs)
      setUsers(usrs)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error cargando administración (requiere conexión)')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadAll()
  }, [server.id])

  const handleCreateDevice = async (e: FormEvent) => {
    e.preventDefault()
    if (!newDeviceName.trim()) return
    try {
      const res = await createDevice(server, newDeviceName.trim())
      setDevices((prev) => [...prev, res.device])
      setCreatedKey({ name: res.device.name, key: res.api_key })
      setNewDeviceName('')
      setCopiedKey(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo crear el gateway')
    }
  }

  const handleDeleteDevice = async (id: number) => {
    try {
      await deleteDevice(server, id)
      setDevices((prev) => prev.filter((d) => d.id !== id))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo revocar el gateway')
    }
  }

  const handleCreateUser = async (e: FormEvent) => {
    e.preventDefault()
    try {
      const created = await createUser(server, {
        username: newUsername.trim(),
        full_name: newFullName.trim(),
        password: newPassword,
        role: newRole,
      })
      setUsers((prev) => [...prev, created])
      setNewUsername('')
      setNewFullName('')
      setNewPassword('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo crear el usuario')
    }
  }

  const handleDeleteUser = async (id: number) => {
    try {
      await deleteUser(server, id)
      setUsers((prev) => prev.filter((u) => u.id !== id))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo eliminar el usuario')
    }
  }

  const copyApiKey = async () => {
    if (!createdKey) return
    await navigator.clipboard.writeText(createdKey.key)
    setCopiedKey(true)
  }

  if (loading) {
    return (
      <div className="flex items-center gap-2 text-[#465C50] py-12">
        <Loader2 className="h-5 w-5 animate-spin text-[#15613B]" />
        <span>Cargando ajustes del servidor...</span>
      </div>
    )
  }

  return (
    <div className="space-y-8 max-w-5xl">
      <div className="space-y-1">
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-[#12261C]">
          Administración del sistema
        </h1>
        <p className="text-base text-[#465C50]">
          Gestiona las claves de enlace para los gateways TTGO LoRa32 y las cuentas de aprendices e instructores.
        </p>
      </div>

      {error && (
        <div className="rounded-xl bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-800">
          {error}
        </div>
      )}

      {/* Gateways LoRa */}
      <section className="bg-white border border-[#DCE5DF] rounded-2xl p-6 space-y-5">
        <div className="flex items-center gap-2.5">
          <Radio className="h-5 w-5 text-[#15613B]" />
          <h2 className="text-xl font-bold text-[#12261C]">Gateways receptores (ESP32 LoRa32)</h2>
        </div>

        <form onSubmit={handleCreateDevice} className="flex flex-col sm:flex-row gap-3">
          <input
            type="text"
            required
            value={newDeviceName}
            onChange={(e) => setNewDeviceName(e.target.value)}
            placeholder="Nombre del nuevo gateway (ej. TTGO Aula Ambiental)"
            className="flex-1 rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-4 py-2.5 text-base focus:bg-white focus:outline-none focus:border-[#15613B]"
          />
          <button
            type="submit"
            className="rounded-xl bg-[#15613B] hover:bg-[#0F4C2D] text-white px-5 py-2.5 text-sm font-semibold inline-flex items-center justify-center gap-2 transition-[background-color] duration-150"
          >
            <KeyRound className="h-4 w-4" />
            Generar clave
          </button>
        </form>

        {createdKey && (
          <div className="rounded-xl bg-[#F0FDF4] border border-[#BBF7D0] p-4 space-y-2">
            <p className="text-sm font-bold text-[#166534]">
              Clave generada para {createdKey.name} (cópiala ahora, solo se muestra una vez)
            </p>
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded-lg bg-white border border-[#BBF7D0] px-3 py-2 text-sm text-[#12261C] break-all">
                {createdKey.key}
              </code>
              <button
                type="button"
                onClick={copyApiKey}
                className="rounded-lg bg-[#15613B] text-white px-3.5 py-2 text-sm font-semibold inline-flex items-center gap-1.5 shrink-0"
              >
                {copiedKey ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                {copiedKey ? 'Copiada' : 'Copiar'}
              </button>
            </div>
          </div>
        )}

        <div className="divide-y divide-[#DCE5DF] border-t border-[#DCE5DF]">
          {devices.length === 0 ? (
            <p className="py-4 text-sm text-[#465C50]">Aún no hay gateways registrados.</p>
          ) : (
            devices.map((d) => (
              <div key={d.id} className="py-3.5 flex items-center justify-between gap-4">
                <div>
                  <p className="text-base font-semibold text-[#12261C]">{d.name}</p>
                  <p className="text-sm text-[#465C50] tabular">
                    Prefijo: {d.key_prefix}... · Última actividad: {formatShortDateTime(d.last_seen_at)}
                  </p>
                </div>
                <button
                  type="button"
                  onClick={() => handleDeleteDevice(d.id)}
                  className="p-2 rounded-lg text-[#465C50] hover:bg-red-50 hover:text-red-700 transition-[background-color,color] duration-150"
                  title="Revocar gateway"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            ))
          )}
        </div>
      </section>

      {/* Usuarios */}
      <section className="bg-white border border-[#DCE5DF] rounded-2xl p-6 space-y-5">
        <div className="flex items-center gap-2.5">
          <Users className="h-5 w-5 text-[#15613B]" />
          <h2 className="text-xl font-bold text-[#12261C]">Usuarios del sistema</h2>
        </div>

        <form onSubmit={handleCreateUser} className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
          <input
            type="text"
            required
            value={newUsername}
            onChange={(e) => setNewUsername(e.target.value)}
            placeholder="Usuario"
            className="rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-sm focus:bg-white focus:outline-none focus:border-[#15613B]"
          />
          <input
            type="text"
            required
            value={newFullName}
            onChange={(e) => setNewFullName(e.target.value)}
            placeholder="Nombre completo"
            className="rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-sm focus:bg-white focus:outline-none focus:border-[#15613B]"
          />
          <input
            type="password"
            required
            minLength={6}
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            placeholder="Contraseña (mín. 6)"
            className="rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-sm focus:bg-white focus:outline-none focus:border-[#15613B]"
          />
          <select
            value={newRole}
            onChange={(e) => setNewRole(e.target.value as 'aprendiz' | 'admin')}
            className="rounded-xl border border-[#DCE5DF] bg-[#F8FAF9] px-3.5 py-2.5 text-sm focus:bg-white focus:outline-none focus:border-[#15613B]"
          >
            <option value="aprendiz">Aprendiz</option>
            <option value="admin">Instructor (Admin)</option>
          </select>
          <button
            type="submit"
            className="rounded-xl bg-[#15613B] hover:bg-[#0F4C2D] text-white px-4 py-2.5 text-sm font-semibold inline-flex items-center justify-center gap-1.5"
          >
            <Plus className="h-4 w-4" />
            Crear
          </button>
        </form>

        <div className="divide-y divide-[#DCE5DF] border-t border-[#DCE5DF]">
          {users.map((u) => (
            <div key={u.id} className="py-3.5 flex items-center justify-between gap-4">
              <div>
                <p className="text-base font-semibold text-[#12261C]">
                  {u.full_name} <span className="text-sm font-normal text-[#465C50]">(@{u.username})</span>
                </p>
                <p className="text-sm text-[#465C50]">
                  Rol: {u.role === 'admin' ? 'Instructor / Administrador' : 'Aprendiz'}
                </p>
              </div>
              {u.id !== currentUser.id && (
                <button
                  type="button"
                  onClick={() => handleDeleteUser(u.id)}
                  className="p-2 rounded-lg text-[#465C50] hover:bg-red-50 hover:text-red-700 transition-[background-color,color] duration-150"
                  title="Eliminar usuario"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              )}
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}
