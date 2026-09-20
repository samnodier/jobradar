<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
interface Notice {
  id: string
  job_id: string
  title: string
  company_name: string
  match_score: number
  read_at: string | null
}
const notices = ref<Notice[]>([])
const open = ref(false)
const error = ref('')
const router = useRouter()
const unread = computed(() => notices.value.filter((n) => !n.read_at).length)
let timer: ReturnType<typeof setInterval> | undefined
let controller: AbortController | undefined
async function load() {
  controller?.abort()
  controller = new AbortController()
  try {
    const r = await fetch('/api/notifications', {
      credentials: 'include',
      signal: controller.signal,
    })
    if (!r.ok) throw new Error('Could not load notifications')
    notices.value = await r.json()
    error.value = ''
  } catch (e) {
    if (e instanceof Error && e.name !== 'AbortError') error.value = e.message
  }
}
function toggle() {
  open.value = !open.value
  if (open.value) load()
}
async function visit(n: Notice) {
  try {
    const r = await fetch(`/api/notifications/${n.id}/read`, {
      method: 'PATCH',
      credentials: 'include',
    })
    if (!r.ok) throw new Error('Could not mark notification read')
    n.read_at = new Date().toISOString()
  } catch {
    error.value = 'Could not mark notification read'
  }
  open.value = false
  router.push({ path: '/jobs', query: { job: n.job_id } })
}
onMounted(() => {
  load()
  timer = setInterval(() => {
    if (!document.hidden) load()
  }, 30000)
})
onUnmounted(() => {
  clearInterval(timer)
  controller?.abort()
})
</script>
<template>
  <div class="relative">
    <button
      class="button button-secondary"
      :aria-expanded="open"
      aria-controls="match-notifications"
      @click="toggle"
    >
      Notifications <span v-if="unread">({{ unread }})</span>
    </button>
    <div
      v-if="open"
      id="match-notifications"
      class="absolute right-0 top-full mt-2 w-80 max-h-96 overflow-y-auto whitespace-normal bg-white border shadow-lg p-4 space-y-3"
    >
      <h2 class="font-semibold">Recent matching jobs</h2>
      <p class="text-xs text-gray-500">Latest 50 matches above your profile threshold.</p>
      <p v-if="error" role="alert">{{ error }}</p>
      <p v-if="!notices.length" class="text-sm">
        No matching notifications yet. Check your profile and threshold.
      </p>
      <button
        v-for="n in notices"
        :key="n.id"
        class="block text-left w-full border-b py-2 text-sm"
        :class="{ 'font-semibold': !n.read_at }"
        @click="visit(n)"
      >
        {{ n.title }} · {{ n.match_score }}%<span class="block text-xs text-gray-500">{{
          n.company_name
        }}</span>
      </button>
    </div>
  </div>
</template>
