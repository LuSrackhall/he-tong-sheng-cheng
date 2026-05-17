import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '../api'

interface User { id: number; username: string; role: string }

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<User | null>(null)
  const loading = ref(false)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function login(username: string, password: string) {
    loading.value = true
    try {
      const { data } = await authApi.login(username, password)
      token.value = data.token
      user.value = data.user
      localStorage.setItem('token', data.token)
      return true
    } finally {
      loading.value = false
    }
  }

  async function fetchMe() {
    try {
      const { data } = await authApi.me()
      user.value = data as any
    } catch {
      logout()
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  return { token, user, loading, isLoggedIn, isAdmin, login, fetchMe, logout }
})
