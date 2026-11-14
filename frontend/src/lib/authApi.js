import { request } from './http'

export const authApi = {
  register({ name, email, password }) {
    return request('/api/auth/register', {
      method: 'POST',
      auth: false,
      body: { name, email, password },
    })
  },

  login({ email, password }) {
    return request('/api/auth/login', {
      method: 'POST',
      auth: false,
      body: { email, password },
    })
  },

  me(signal) {
    return request('/api/auth/me', { signal })
  },
}
