<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  Camera,
  CameraOff,
  Check,
  CheckCheck,
  GraduationCap,
  Mic,
  MicOff,
  MonitorUp,
  UserRoundX,
  Users,
  X,
} from 'lucide-vue-next'
import type { LiveKitParticipantInfo } from '~/composables/useLiveKit'

const props = defineProps<{
  participants: LiveKitParticipantInfo[]
  pendingRequests?: any[]
  isHost?: boolean
  canControl?: boolean
  maxParticipants?: number
  defaultClass?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'mute', identity: string): void
  (e: 'kick', identity: string): void
  (e: 'decideRequest', userId: string, status: 'approved' | 'rejected'): void
  (e: 'decideAllRequests', status: 'approved' | 'rejected'): void
  (e: 'togglePermission', userId: string, permissions: { mic?: boolean; camera?: boolean; screen?: boolean }): void
}>()

const searchQuery = ref('')

const filteredParticipants = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return props.participants
  return props.participants.filter((p) => (p.name || '').toLowerCase().includes(q))
})

const getInitials = (name: string) => {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase())
    .join('') || '?'
}

const getStudentName = (req: any) => {
  return (
    req.user?.student?.full_name ||
    req.user?.student?.FullName ||
    req.user?.teacher?.full_name ||
    req.user?.teacher?.FullName ||
    req.user?.display_name ||
    req.user?.username ||
    req.user?.Username ||
    'Siswa'
  )
}

const getStudentClass = (req: any) => {
  const student = req.user?.student || req.user?.Student
  const classes = student?.classes || student?.Classes || []
  if (Array.isArray(classes) && classes.length > 0) {
    return classes.map((c: any) => c.name || c.Name).filter(Boolean).join(', ')
  }
  return props.defaultClass || 'Siswa'
}

const getStudentNIS = (req: any) => {
  const student = req.user?.student || req.user?.Student
  return student?.nis || student?.NIS || ''
}
</script>

<template>
  <aside class="flex flex-col w-full sm:w-80 h-full bg-slate-900 border-l border-slate-800 text-slate-100 select-none shadow-2xl">
    <!-- Header -->
    <div class="flex items-center justify-between p-4 border-b border-slate-800">
      <div class="flex items-center gap-2">
        <h2 class="text-sm font-semibold text-slate-100">Peserta</h2>
        <span class="px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
          {{ participants.length }} / {{ maxParticipants || 50 }}
        </span>
      </div>
      <button
        type="button"
        class="p-1.5 text-slate-400 hover:text-slate-100 hover:bg-slate-800 rounded-lg transition-colors"
        @click="emit('close')"
      >
        <X :size="16" />
      </button>
    </div>

    <!-- Search input -->
    <div class="p-3 border-b border-slate-800/60">
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Cari peserta..."
        class="w-full px-3 py-1.5 rounded-xl bg-slate-950/70 border border-slate-800 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-emerald-500"
      />
    </div>

    <!-- Scrollable content -->
    <div class="flex-1 overflow-y-auto p-3 space-y-4">
      <!-- Pending Join Requests (Host / Guru Only) -->
      <div v-if="isHost && pendingRequests?.length" class="space-y-3 p-3.5 rounded-2xl bg-amber-950/40 border border-amber-500/30 shadow-lg">
        <div class="flex items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <span class="w-2 h-2 rounded-full bg-amber-400 animate-ping"></span>
            <h3 class="text-xs font-bold text-amber-300">Menunggu Izin Masuk</h3>
            <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-500/20 text-amber-300 border border-amber-500/30">
              {{ pendingRequests.length }}
            </span>
          </div>
        </div>

        <!-- Bulk Action Buttons: Izinkan Semua / Tolak Semua -->
        <div class="flex items-center gap-2 pt-1 border-t border-amber-800/40">
          <button
            type="button"
            class="flex-1 inline-flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-xs transition-colors shadow-sm cursor-pointer"
            @click="emit('decideAllRequests', 'approved')"
          >
            <CheckCheck :size="14" />
            Izinkan Semua ({{ pendingRequests.length }})
          </button>
          <button
            type="button"
            class="inline-flex items-center justify-center gap-1 px-2.5 py-1.5 rounded-xl bg-slate-800 hover:bg-rose-950/80 hover:text-rose-300 text-slate-300 border border-slate-700/80 font-medium text-xs transition-colors cursor-pointer"
            title="Tolak Semua Pengajuan"
            @click="emit('decideAllRequests', 'rejected')"
          >
            <X :size="13" />
            Tolak Semua
          </button>
        </div>

        <!-- Individual Requests List -->
        <div class="space-y-2 max-h-60 overflow-y-auto pr-1">
          <div
            v-for="req in pendingRequests"
            :key="req.id || req.ID"
            class="flex flex-col gap-2 p-2.5 rounded-xl bg-slate-900/95 border border-slate-800 shadow-sm"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="min-w-0">
                <div class="flex items-center gap-1.5">
                  <span class="font-bold text-xs text-slate-100 truncate">
                    {{ getStudentName(req) }}
                  </span>
                </div>
                <div class="flex flex-wrap items-center gap-1.5 mt-1">
                  <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[10px] font-semibold bg-emerald-500/15 text-emerald-300 border border-emerald-500/20">
                    <GraduationCap :size="10" />
                    {{ getStudentClass(req) }}
                  </span>
                  <span v-if="getStudentNIS(req)" class="text-[10px] text-slate-400">
                    NIS: {{ getStudentNIS(req) }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Individual Action Buttons -->
            <div class="flex items-center justify-end gap-1.5 pt-1 border-t border-slate-800/80">
              <button
                type="button"
                class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white font-semibold text-[11px] transition-colors cursor-pointer"
                @click="emit('decideRequest', req.user_id || req.UserID, 'approved')"
              >
                <Check :size="12" />
                Izinkan
              </button>
              <button
                type="button"
                class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-rose-600/80 hover:bg-rose-500 text-white font-medium text-[11px] transition-colors cursor-pointer"
                @click="emit('decideRequest', req.user_id || req.UserID, 'rejected')"
              >
                <X :size="12" />
                Tolak
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Active Participants List -->
      <div class="space-y-1.5">
        <div
          v-for="p in filteredParticipants"
          :key="p.identity"
          class="group flex items-center justify-between gap-2 p-2.5 rounded-xl bg-slate-950/40 hover:bg-slate-800/60 border border-slate-800/40 transition-colors"
        >
          <div class="flex items-center gap-2.5 min-w-0">
            <div
              class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold text-white shrink-0"
              :class="[
                p.role === 'guru'
                  ? 'bg-emerald-600 ring-2 ring-emerald-500/20'
                  : 'bg-indigo-600 ring-2 ring-indigo-500/20',
              ]"
            >
              {{ getInitials(p.name) }}
            </div>
            <div class="min-w-0">
              <div class="flex items-center gap-1.5">
                <span class="text-xs font-medium text-slate-200 truncate">
                  {{ p.name }}
                </span>
                <span
                  v-if="p.isLocal"
                  class="px-1.5 py-0.2 rounded text-[10px] font-medium bg-slate-800 text-emerald-300"
                >
                  Anda
                </span>
              </div>
              <span class="text-[10px] text-slate-400 capitalize">
                {{ p.role === 'guru' ? 'Guru' : p.role === 'admin' ? 'Admin' : 'Siswa' }}
              </span>
            </div>
          </div>

          <!-- Status & Actions -->
          <div class="flex items-center gap-1.5 shrink-0">
            <!-- Audio Icon -->
            <span
              class="text-xs"
              :class="p.isAudioEnabled ? 'text-emerald-400' : 'text-slate-500'"
              :title="p.isAudioEnabled ? 'Mic Aktif' : 'Mic Mati'"
            >
              <Mic v-if="p.isAudioEnabled" :size="14" />
              <MicOff v-else :size="14" />
            </span>

            <!-- Video Icon -->
            <span
              class="text-xs"
              :class="p.isVideoEnabled ? 'text-emerald-400' : 'text-slate-500'"
              :title="p.isVideoEnabled ? 'Kamera Aktif' : 'Kamera Mati'"
            >
              <Camera v-if="p.isVideoEnabled" :size="14" />
              <CameraOff v-else :size="14" />
            </span>

            <!-- Screen Share Icon -->
            <span
              v-if="p.isScreenShareEnabled"
              class="text-xs text-sky-400"
              title="Membagikan Layar"
            >
              <MonitorUp :size="14" />
            </span>

            <!-- Action buttons for Teacher -->
            <div v-if="canControl && !p.isLocal" class="opacity-0 group-hover:opacity-100 flex items-center gap-1 pl-1 transition-opacity">
              <button
                type="button"
                class="p-1 text-slate-400 hover:text-amber-400 rounded hover:bg-slate-700 text-xs"
                title="Mute Peserta"
                @click="emit('mute', p.identity)"
              >
                <MicOff :size="14" />
              </button>
              <button
                type="button"
                class="p-1 text-slate-400 hover:text-rose-400 rounded hover:bg-slate-700 text-xs"
                title="Keluarkan Peserta"
                @click="emit('kick', p.identity)"
              >
                <UserRoundX :size="14" />
              </button>
            </div>
          </div>
        </div>

        <p v-if="!filteredParticipants.length" class="text-center text-xs text-slate-500 py-6">
          Tidak ada peserta ditemukan.
        </p>
      </div>
    </div>
  </aside>
</template>
