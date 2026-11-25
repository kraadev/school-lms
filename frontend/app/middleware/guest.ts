import { useAuthStore } from '../stores/auth'

export default defineNuxtRouteMiddleware(() => {
  if (!import.meta.client) return

  const auth = useAuthStore()
  auth.restore()

  if (auth.isAuthenticated) {
    return navigateTo('/dashboard')
  }
})
