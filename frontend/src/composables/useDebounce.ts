let timer: ReturnType<typeof setTimeout>

export function useDebounce<F extends (...args: any[]) => void>(fn: F, delay: number): F {
  return ((...args: any[]) => {
    clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }) as F
}
