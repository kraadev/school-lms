import { defineStore } from 'pinia'

export interface UserRole {
  id: string
  name: string
}

export interface AuthUser {
  id: string
  username: string
  display_name?: string
  full_name?: string
  email?: string | null
  is_active?: boolean
  role: UserRole
  student?: any
  teacher?: any
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref('')
  const user = ref<AuthUser | null>(null)
  const ready = ref(false)

  const normalizeUser = (value: any): AuthUser | null => {
    if (!value) return null
    const role = value.role || value.Role
    const student = value.student || value.Student
    const teacher = value.teacher || value.Teacher
    const fullName = student?.full_name || student?.FullName || teacher?.full_name || teacher?.FullName || ''
    const displayName = fullName || value.display_name || value.DisplayName || value.username || value.Username || ''

    return {
      id: value.id || value.ID || '',
      username: value.username || value.Username || '',
      display_name: displayName,
      full_name: fullName,
      email: value.email ?? value.Email ?? null,
      is_active: value.is_active ?? value.IsActive,
      role: {
        id: role?.id || role?.ID || '',
        name: role?.name || role?.Name || '',
      },
      student: student || null,
      teacher: teacher || null,
    }
  }

  const restore = () => {
    if (!import.meta.client || ready.value) return
    token.value = localStorage.getItem('token') || ''
    const savedUser = localStorage.getItem('user')
    if (savedUser) {
      try { user.value = normalizeUser(JSON.parse(savedUser)) } catch { clear() }
    }
    ready.value = true
  }

  const setSession = (nextToken: string, nextUser: AuthUser) => {
    token.value = nextToken
    user.value = normalizeUser(nextUser)
    if (import.meta.client) {
      localStorage.setItem('token', nextToken)
      localStorage.setItem('user', JSON.stringify(user.value))
    }
  }

    const updateUser = (nextUser: AuthUser) => {
      user.value = normalizeUser(nextUser)
      if (import.meta.client) localStorage.setItem('user', JSON.stringify(user.value))
    }

  const clear = () => {
    token.value = ''
    user.value = null
    if (import.meta.client) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    }
  }

  const role = computed(() => user.value?.role?.name || '')
  const isAuthenticated = computed(() => Boolean(token.value && user.value))

    return { token, user, ready, role, isAuthenticated, restore, setSession, updateUser, clear }
})
