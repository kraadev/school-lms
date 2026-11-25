<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { Expand, GraduationCap, Minimize2, Users, CheckCheck, X } from 'lucide-vue-next'
import type { LiveKitParticipantInfo } from '~/composables/useLiveKit'
import type { MeetingData } from '~/composables/useMeeting'
import ParticipantGrid from '~/components/meeting/ParticipantGrid.vue'
import ScreenShare from '~/components/meeting/ScreenShare.vue'
import ParticipantList from '~/components/meeting/ParticipantList.vue'
import MeetingControls from '~/components/meeting/MeetingControls.vue'
import DeviceSettings from '~/components/meeting/DeviceSettings.vue'

const props = defineProps<{
  meeting: MeetingData
  participants: LiveKitParticipantInfo[]
  currentScreenShare: any
  pendingRequests?: any[]
  isHost?: boolean
  isMicEnabled: boolean
  isCameraEnabled: boolean
  isScreenSharing: boolean
  connectionState: string
  connectionStatusText: string
  audioInputs: MediaDeviceInfo[]
  videoInputs: MediaDeviceInfo[]
  audioOutputs: MediaDeviceInfo[]
  selectedAudioInputId: string
  selectedVideoInputId: string
}>()

const emit = defineEmits<{
  (e: 'toggleMic'): void
  (e: 'toggleCamera'): void
  (e: 'toggleScreenShare'): void
  (e: 'leave'): void
  (e: 'endMeeting'): void
  (e: 'muteParticipant', identity: string): void
  (e: 'kickParticipant', identity: string): void
  (e: 'decideRequest', userId: string, status: 'approved' | 'rejected'): void
  (e: 'decideAllRequests', status: 'approved' | 'rejected'): void
  (e: 'switchCamera', deviceId: string): void
  (e: 'switchMicrophone', deviceId: string): void
}>()

const isParticipantListOpen = ref(false)
const isSettingsOpen = ref(false)
const isFullscreen = ref(false)
const bannerDismissed = ref(false)

const myParticipant = computed(() => props.participants.find((p) => p.isLocal))

watch(
  () => props.pendingRequests?.length,
  (newLen, oldLen) => {
    if (newLen && newLen > (oldLen || 0)) {
      bannerDismissed.value = false
    }
  },
)

const pendingCount = computed(() => props.pendingRequests?.length || 0)

const firstPendingStudent = computed(() => {
  if (!props.pendingRequests || props.pendingRequests.length === 0) return null
  const req = props.pendingRequests[0]
  const user = req.user || req.User || {}
  const name =
    user.student?.full_name ||
    user.student?.FullName ||
    user.teacher?.full_name ||
    user.teacher?.FullName ||
    user.name ||
    user.Name ||
    user.username ||
    'Siswa'

  const student = user.student || user.Student || {}
  const classes = student.classes || student.Classes || []
  let className = ''
  if (Array.isArray(classes) && classes.length > 0) {
    className = classes.map((c: any) => c.name || c.Name).filter(Boolean).join(', ')
  }
  return { name, className }
})

// Duration timer
const durationSeconds = ref(0)
let timerInterval: any = null

onMounted(() => {
  timerInterval = setInterval(() => {
    durationSeconds.value++
  }, 1000)
})

onBeforeUnmount(() => {
  if (timerInterval) clearInterval(timerInterval)
})

const formattedDuration = computed(() => {
  const m = Math.floor(durationSeconds.value / 60)
  const s = durationSeconds.value % 60
  return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
})

const relationValue = (item: any, ...paths: string[]) => {
  for (const path of paths) {
    const value = path.split('.').reduce((current, key) => current?.[key], item)
    if (value !== undefined && value !== null && value !== '') return value
  }
  return ''
}

const meetingContext = computed(() => {
  const meeting = props.meeting as any
  const className = relationValue(
    meeting,
    'class_subject.class.name',
    'class_subject.class.Name',
    'class_subject.Class.name',
    'class_subject.Class.Name',
    'ClassSubject.class.name',
    'ClassSubject.class.Name',
    'ClassSubject.Class.name',
    'ClassSubject.Class.Name',
    'class_name',
    'ClassName',
  ) || 'Kelas'
  const subjectName = relationValue(
    meeting,
    'class_subject.subject.name',
    'class_subject.subject.Name',
    'class_subject.Subject.name',
    'class_subject.Subject.Name',
    'ClassSubject.subject.name',
    'ClassSubject.subject.Name',
    'ClassSubject.Subject.name',
    'ClassSubject.Subject.Name',
    'subject_name',
    'SubjectName',
  ) || 'Mapel'

  return `${className} - ${subjectName}`
})

const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen().catch(() => {})
    isFullscreen.value = true
  } else {
    document.exitFullscreen().catch(() => {})
    isFullscreen.value = false
  }
}
</script>

<template>
  <div class="relative flex flex-col w-full h-screen bg-slate-950 text-slate-100 overflow-hidden font-sans select-none">
    <!-- Top Bar -->
    <header class="relative z-20 flex items-center justify-between gap-4 px-4 py-3.5 sm:px-6 bg-slate-900/95 backdrop-blur border-b border-slate-800/80 shrink-0">
      <div class="flex items-center gap-3.5 min-w-0">
        <div class="w-11 h-11 rounded-2xl bg-slate-800/90 border border-slate-700/70 p-1 flex items-center justify-center shadow-md shrink-0 overflow-hidden">
          <img src="/images/logo.png" alt="School Space" class="w-full h-full object-contain" />
        </div>
        <div class="min-w-0 space-y-0.5">
          <div class="flex flex-wrap items-center gap-2">
            <h1
              class="max-w-[42vw] text-base sm:text-lg font-bold truncate tracking-tight !text-white"
              style="color: #ffffff !important;"
              :title="meeting.title"
            >
              {{ meeting.title || 'Ruang Meeting' }}
            </h1>
            <span
              class="px-2.5 py-0.5 rounded-full text-[10px] font-bold capitalize tracking-wide shrink-0"
              :class="[
                meeting.status === 'live' || meeting.status === 'active'
                  ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
                  : 'bg-amber-500/20 text-amber-400 border border-amber-500/30',
              ]"
            >
              {{ meeting.type }} · {{ meeting.status }}
            </span>
          </div>
          <p class="text-xs text-slate-300 truncate" style="color: #cbd5e1 !important;">
            {{ meetingContext }}
          </p>
        </div>
      </div>

      <!-- Right Status & Fullscreen -->
      <div class="flex items-center gap-2 sm:gap-3 shrink-0">
        <!-- Timer -->
        <div class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-xl bg-slate-800/80 border border-slate-700/60 text-xs font-mono text-emerald-400">
          <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
          <span>{{ formattedDuration }}</span>
        </div>

        <!-- Connection Badge -->
        <span
          class="hidden md:inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-full text-xs font-semibold border"
          :class="[
            connectionState === 'connected'
              ? 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30'
              : 'bg-amber-500/10 text-amber-300 border-amber-500/30',
          ]"
        >
          <span
            class="w-1.5 h-1.5 rounded-full"
            :class="connectionState === 'connected' ? 'bg-emerald-400' : 'bg-amber-400 animate-ping'"
          />
          {{ connectionStatusText }}
        </span>

        <!-- Fullscreen Button -->
        <button
          type="button"
          class="p-2.5 text-slate-300 hover:text-white hover:bg-slate-800 rounded-xl transition-colors text-xs border border-slate-700/60"
          :title="isFullscreen ? 'Keluar Fullscreen' : 'Fullscreen'"
          @click="toggleFullscreen"
        >
          <Minimize2 v-if="isFullscreen" :size="16" />
          <Expand v-else :size="16" />
        </button>
      </div>
    </header>

    <!-- Floating Host Admission Alert Banner -->
    <transition
      enter-active-class="transition-all duration-300 ease-out"
      enter-from-class="opacity-0 -translate-y-3"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition-all duration-200 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-3"
    >
      <div
        v-if="isHost && pendingCount > 0 && !bannerDismissed"
        class="relative z-20 mx-4 mt-2 px-4 py-3 rounded-2xl bg-amber-500/15 border border-amber-500/30 backdrop-blur-md shadow-xl flex flex-wrap items-center justify-between gap-3 text-xs"
      >
        <div class="flex items-center gap-3 min-w-0">
          <div class="w-8 h-8 rounded-xl bg-amber-500/20 text-amber-300 flex items-center justify-center shrink-0">
            <Users :size="16" />
          </div>
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <span class="font-bold text-amber-200">
                {{ pendingCount }} Siswa Menunggu Izin Masuk
              </span>
              <span v-if="firstPendingStudent" class="hidden sm:inline text-slate-300 font-medium truncate">
                · {{ firstPendingStudent.name }}
                <span v-if="firstPendingStudent.className" class="text-amber-400 font-normal">({{ firstPendingStudent.className }})</span>
                <span v-if="pendingCount > 1" class="text-slate-400 font-normal"> dan {{ pendingCount - 1 }} lainnya</span>
              </span>
            </div>
            <p class="text-[11px] text-slate-400">
              Anda dapat mengizinkan semua siswa sekaligus atau meninjaunya di daftar peserta.
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2 shrink-0 ml-auto">
          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-xs shadow-md transition-all active:scale-95"
            @click="emit('decideAllRequests', 'approved')"
          >
            <CheckCheck :size="14" />
            Izinkan Semua ({{ pendingCount }})
          </button>
          <button
            type="button"
            class="px-3 py-1.5 rounded-xl bg-slate-800/90 hover:bg-slate-700 text-slate-200 font-medium text-xs border border-slate-700 transition-colors"
            @click="isParticipantListOpen = true; bannerDismissed = true"
          >
            Lihat Detail
          </button>
          <button
            type="button"
            class="p-1.5 text-slate-400 hover:text-slate-200 hover:bg-slate-800 rounded-lg transition-colors"
            title="Tutup pemberitahuan"
            @click="bannerDismissed = true"
          >
            <X :size="14" />
          </button>
        </div>
      </div>
    </transition>

    <!-- Main Stage Area -->
    <main class="relative flex-1 flex overflow-hidden min-h-0">
      <!-- Media Stage -->
      <div class="flex-1 flex flex-col p-2 sm:p-4 overflow-hidden min-w-0">
        <!-- Screen share presentation layout -->
        <div v-if="currentScreenShare" class="flex-1 flex flex-col gap-3 h-full overflow-hidden">
          <div class="flex-1 min-h-0">
            <ScreenShare :screen-share="currentScreenShare" />
          </div>
          <div class="shrink-0 h-[140px]">
            <ParticipantGrid
              :participants="participants"
              :is-host="isHost"
              :can-control="isHost"
              :is-screen-sharing-active="true"
              @mute="emit('muteParticipant', $event)"
              @kick="emit('kickParticipant', $event)"
            />
          </div>
        </div>

        <!-- Normal Participant Grid -->
        <div v-else class="flex-1 h-full overflow-hidden">
          <ParticipantGrid
            :participants="participants"
            :is-host="isHost"
            :can-control="isHost"
            @mute="emit('muteParticipant', $event)"
            @kick="emit('kickParticipant', $event)"
          />
        </div>
      </div>

      <!-- Slide-in Participant List Drawer -->
      <transition
        enter-active-class="transition-transform duration-300 ease-out"
        enter-from-class="translate-x-full"
        enter-to-class="translate-x-0"
        leave-active-class="transition-transform duration-200 ease-in"
        leave-from-class="translate-x-0"
        leave-to-class="translate-x-full"
      >
        <div
          v-if="isParticipantListOpen"
          class="absolute inset-y-0 right-0 z-30 sm:relative shrink-0 w-full sm:w-80 h-full"
        >
          <ParticipantList
            :participants="participants"
            :pending-requests="pendingRequests"
            :is-host="isHost"
            :can-control="isHost"
            :max-participants="meeting.max_participants"
            :default-class="meetingContext"
            @close="isParticipantListOpen = false"
            @mute="emit('muteParticipant', $event)"
            @kick="emit('kickParticipant', $event)"
            @decide-request="(userId, status) => emit('decideRequest', userId, status)"
            @decide-all-requests="(status) => emit('decideAllRequests', status)"
          />
        </div>
      </transition>
    </main>

    <!-- Bottom Controls Floating Bar -->
    <footer class="relative z-20 shrink-0 bg-slate-950/80 backdrop-blur-md">
      <MeetingControls
        :is-mic-enabled="isMicEnabled"
        :is-camera-enabled="isCameraEnabled"
        :is-screen-sharing="isScreenSharing"
        :can-screen-share="isHost || (myParticipant?.screen_share_enabled !== false)"
        :participant-count="participants.length"
        :pending-requests-count="pendingCount"
        :is-participant-list-open="isParticipantListOpen"
        :is-host="isHost"
        :is-voice-only="meeting.type === 'voice'"
        @toggle-mic="emit('toggleMic')"
        @toggle-camera="emit('toggleCamera')"
        @toggle-screen-share="emit('toggleScreenShare')"
        @toggle-participant-list="isParticipantListOpen = !isParticipantListOpen"
        @open-settings="isSettingsOpen = true"
        @leave="emit('leave')"
        @end-meeting="emit('endMeeting')"
      />
    </footer>

    <!-- Settings Modal -->
    <DeviceSettings
      v-if="isSettingsOpen"
      :audio-inputs="audioInputs"
      :video-inputs="videoInputs"
      :audio-outputs="audioOutputs"
      :selected-audio-input-id="selectedAudioInputId"
      :selected-video-input-id="selectedVideoInputId"
      @close="isSettingsOpen = false"
      @switch-camera="emit('switchCamera', $event)"
      @switch-microphone="emit('switchMicrophone', $event)"
    />
  </div>
</template>

<style scoped>
h1,
h2,
h3,
h4,
h5,
h6 {
  color: #ffffff !important;
}
</style>
