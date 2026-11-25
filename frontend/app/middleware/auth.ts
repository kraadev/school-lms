import { useAuthStore } from '../stores/auth'

export default defineNuxtRouteMiddleware((to) => {
  if (!import.meta.client) return

  const auth = useAuthStore()
  auth.restore()

  if (!auth.isAuthenticated) {
    return navigateTo({ path: '/login', query: { redirect: to.fullPath } })
  }
})
