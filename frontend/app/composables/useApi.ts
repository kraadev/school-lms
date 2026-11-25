import { useAuthStore } from '../stores/auth'

export interface ApiResponse<T> {
  success?: boolean
  message?: string
  data?: T
  user?: T
}

export function getResolvedApiBase(configuredBase?: string): string {
  // When running in the browser, always use a relative path so that API requests
  // go through the Nuxt HTTPS server's devProxy (which forwards to the backend).
  // This avoids Mixed Content errors (HTTPS page → HTTP backend) and means the
  // same code works for both localhost and LAN/mobile device access.
  if (import.meta.client && typeof window !== 'undefined') {
    // If configuredBase is already a relative path (starts with /), use it directly.
    if (configuredBase && configuredBase.startsWith('/')) {
      return configuredBase
    }
    // Otherwise fall back to /api (works with Nuxt devProxy)
    return '/api'
  }
  // Server-side (SSR): use the absolute backend URL directly
  return configuredBase || 'http://localhost:8080/api'
}

export function useApi() {
  const config = useRuntimeConfig()
  const auth = useAuthStore()

  const request = async <T>(path: string, options: Parameters<typeof $fetch<T>>[1] = {}) => {
    const headers = new Headers((options as RequestInit).headers)
    if (auth.token) headers.set('Authorization', `Bearer ${auth.token}`)

    const apiBase = getResolvedApiBase(config.public.apiBase as string)
    return await $fetch<T>(`${apiBase}${path}`, {
      ...options,
      headers,
    })
  }

  return { request }
}
