import { getToken } from '../auth/tokenStorage'

const BASE_URL = (import.meta.env.VITE_API_URL || '').replace(/\/+$/, '')

export class ApiError extends Error {
  constructor(message, { status = 0, code = 'unknown_error', fields } = {}) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.fields = fields
  }
}

export async function request(path, { method = 'GET', body, auth = true, signal } = {}) {
  const headers = { Accept: 'application/json' }

  if (body !== undefined) {
    headers['Content-Type'] = 'application/json'
  }

  if (auth) {
    const token = getToken()
    if (token) headers.Authorization = `Bearer ${token}`
  }

  let response
  try {
    response = await fetch(`${BASE_URL}${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      signal,
    })
  } catch (error) {
    if (error.name === 'AbortError') throw error
    throw new ApiError('Tidak dapat terhubung ke server. Cek koneksi atau apakah backend sudah berjalan.', {
      code: 'network_error',
    })
  }

  const contentType = response.headers.get('content-type') || ''
  const data = contentType.includes('application/json') ? await response.json().catch(() => null) : null

  if (!response.ok) {
    const apiError = data?.error
    throw new ApiError(apiError?.message || `Request gagal (HTTP ${response.status})`, {
      status: response.status,
      code: apiError?.code || 'http_error',
      fields: apiError?.fields,
    })
  }

  return data
}
