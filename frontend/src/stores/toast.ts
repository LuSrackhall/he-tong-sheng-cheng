import { defineStore } from 'pinia'
import { ref } from 'vue'

interface Toast {
  id: number
  type: 'success' | 'error' | 'info'
  message: string
}

let nextId = 0

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])

  function show(type: Toast['type'], message: string, duration = 3000) {
    const id = nextId++
    toasts.value.push({ id, type, message })
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, duration)
  }

  function success(message: string) { show('success', message) }
  function error(message: string) { show('error', message, 5000) }
  function info(message: string) { show('info', message) }

  return { toasts, show, success, error, info }
})
