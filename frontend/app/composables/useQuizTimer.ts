import { ref, computed, onMounted, onUnmounted } from 'vue'

export function useQuizTimer(deadlineAt: string) {
  const now = ref(Date.now())
  let timer: ReturnType<typeof setInterval> | null = null

  const remainingMs = computed(() => {
    const deadline = new Date(deadlineAt).getTime()

    return Math.max(
      0,
      deadline - now.value
    )
  })

  const remainingSeconds = computed(() => {
    return Math.floor(
      remainingMs.value / 1000
    )
  })

  const minutes = computed(() => {
    return Math.floor(
      remainingSeconds.value / 60
    )
  })

  const seconds = computed(() => {
    return remainingSeconds.value % 60
  })

  const isExpired = computed(() => {
    return remainingMs.value <= 0
  })

  const formattedTime = computed(() => {
    const minute = String(minutes.value).padStart(2, '0')
    const second = String(seconds.value).padStart(2, '0')

    return `${minute}:${second}`
  })

  onMounted(() => {

    timer = setInterval(() => {
      now.value = Date.now()
    }, 1000)

  })

  onUnmounted(() => {

    if (timer) {
      clearInterval(timer)
      timer = null
    }

  })

  return {
    remainingMs,
    remainingSeconds,
    minutes,
    seconds,
    formattedTime,
    isExpired,
  }
}