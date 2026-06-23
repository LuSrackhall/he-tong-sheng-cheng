import { onMounted, onUnmounted } from 'vue'

/**
 * 按 Escape 键时调用回调（关闭弹窗等）
 */
export function useEscapeKey(callback: () => void) {
  function handler(e: KeyboardEvent) {
    if (e.key === 'Escape') callback()
  }
  onMounted(() => document.addEventListener('keydown', handler))
  onUnmounted(() => document.removeEventListener('keydown', handler))
}
