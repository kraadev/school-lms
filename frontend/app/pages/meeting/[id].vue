<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  AlertTriangle,
  ArrowLeft,
  CameraOff,
  Check,
  CheckCheck,
  Clock3,
  GraduationCap,
  Hand,
  LockKeyhole,
  Mic,
  MicOff,
  Play,
  RefreshCw,
  Users,
  Video,
  X,
} from 'lucide-vue-next'

// Helper: sleep for ms milliseconds
const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms))
import { useAuthStore } from '~/stores/auth'
import { useMeeting } from '~/composables/useMeeting'
import { useLiveKit } from '~/composables/useLiveKit'
import MeetingRoom from '~/components/meeting/MeetingRoom.vue'

definePageMeta({ middleware: 'auth' })

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const meetingId = computed(() => String(route.params.id || ''))

const relationValue = (item: any, ...paths: string[]) => {
  for (const path of paths) {
    const value = path.split('.').reduce((current, key) => current?.[key], item)
    if (value !== undefined && value !== null && value !== '') return value
  }
  return ''
}

const meetingContext = (value: any) => {
  const meeting = value?.value || value
  const className = relationValue(
    meeting,
    'class.name',
    'class.Name',
    'Class.name',
    'Class.Name',
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
    'subject.name',
    'subject.Name',
    'Subject.name',
    'Subject.Name',
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
}

const {
  meeting,
  loading: meetingLoading,
  error: meetingError,
  pendingRequests,
  isHost,
  myParticipant,
  isApproved,
  isPendingApproval,
  isKicked,
  isRejected,
  loadMeeting,
  loadRequests,
  requestAccess,
  getJoinToken,
  startMeeting,
  endMeeting,
  leaveMeeting: apiLeaveMeeting,
  decideParticipant,
  decideAllParticipants,
  kickParticipant: apiKickParticipant,
  muteParticipant: apiMuteParticipant,
} = useMeeting()

const liveKit = useLiveKit()

const inMeeting = ref(false)
const joining = ref(false)
const prejoinMic = ref(true)
const prejoinCamera = ref(true)
const localPreviewEl = ref<HTMLVideoElement | null>(null)
const previewStream = ref<MediaStream | null>(null)

// Toast notification for camera/mic errors inside meeting
const mediaErrorMsg = ref('')
const mediaErrorTimer = ref<any>(null)

const showMediaError = (msg: string) => {
  mediaErrorMsg.value = msg
  if (mediaErrorTimer.value) clearTimeout(mediaErrorTimer.value)
  mediaErrorTimer.value = setTimeout(() => {
    mediaErrorMsg.value = ''
  }, 6000)
}

const startPrejoinPreview = async () => {
  if (typeof navigator === 'undefined' || !navigator.mediaDevices?.getUserMedia) return
  try {
    const stream = await navigator.mediaDevices.getUserMedia({
      video: prejoinCamera.value,
      audio: prejoinMic.value,
    })
    previewStream.value = stream
    if (localPreviewEl.value) {
      localPreviewEl.value.srcObject = stream
    }
  } catch (e: any) {
    // Camera might be in use by another browser tab/window — silently disable preview
    console.warn('Prejoin preview error (camera may be in use):', e)
    if (e?.name === 'NotReadableError' || e?.name === 'NotAllowedError' || e?.name === 'AbortError') {
      prejoinCamera.value = false
    }
  }
}

const stopPrejoinPreview = () => {
  if (previewStream.value) {
    previewStream.value.getTracks().forEach((t) => t.stop())
    previewStream.value = null
  }
  if (localPreviewEl.value) {
    localPreviewEl.value.srcObject = null
  }
}

const togglePrejoinMic = () => {
  prejoinMic.value = !prejoinMic.value
  if (previewStream.value) {
    previewStream.value.getAudioTracks().forEach((t) => (t.enabled = prejoinMic.value))
  }
}

const togglePrejoinCamera = () => {
  prejoinCamera.value = !prejoinCamera.value
  if (previewStream.value) {
    previewStream.value.getVideoTracks().forEach((t) => (t.enabled = prejoinCamera.value))
  }
}

const handleStartAndJoin = async () => {
  try {
    joining.value = true
    await startMeeting(meetingId.value)
    await handleJoin()
  } catch (e: any) {
    meetingError.value = e?.message || 'Gagal memulai meeting'
  } finally {
    joining.value = false
  }
}

const handleJoin = async () => {
  if (joining.value) return
  joining.value = true

  // IMPORTANT: Stop the prejoin preview and wait for OS/browser to fully release the camera
  // before LiveKit SDK tries to grab it. Without this delay, the second browser on the same
  // device may fail to access the camera because the track hasn't been fully released yet.
  stopPrejoinPreview()
  await sleep(350)

  try {
    const joinData = await getJoinToken(meetingId.value)
    // When frontend is served over HTTPS, route LiveKit through the same origin (wss://<host>)
    // so Nuxt's proxy forwards to LiveKit without triggering browser Mixed Content blocking.
    let lkUrl = joinData.url
    if (typeof window !== 'undefined' && window.location.protocol === 'https:') {
      lkUrl = `wss://${window.location.host}`
    }

    await liveKit.connect(lkUrl, joinData.token, {
      audio: prejoinMic.value,
      video: prejoinCamera.value && meeting.value?.type !== 'voice',
      publishOnJoin: true,
    })
    inMeeting.value = true
    if (isHost.value) {
      void loadRequests(meetingId.value)
    }
  } catch (e: any) {
    meetingError.value = e?.message || 'Gagal masuk ke ruang meeting'
    void startPrejoinPreview()
  } finally {
    joining.value = false
  }
}

const retryMeeting = async () => {
  meetingError.value = ''
  try {
    await loadMeeting(meetingId.value)
    await handleJoin()
  } catch (e: any) {
    meetingError.value = e?.message || 'Gagal masuk ke ruang meeting'
  }
}

// Retry enabling camera inside meeting (e.g. after it was unavailable on join)
const handleRetryCamera = async () => {
  try {
    await liveKit.setCamera(true)
    if (!liveKit.isCameraEnabled.value) {
      showMediaError('Kamera masih tidak bisa diakses. Pastikan tidak ada aplikasi lain yang menggunakannya.')
    }
  } catch (e: any) {
    const msg = e?.name === 'NotReadableError'
      ? 'Kamera sedang digunakan aplikasi/tab lain. Tutup tab lain yang mengakses kamera lalu coba lagi.'
      : e?.name === 'NotAllowedError'
      ? 'Izin kamera ditolak. Buka pengaturan browser dan izinkan akses kamera.'
      : 'Gagal mengaktifkan kamera: ' + (e?.message || '')
    showMediaError(msg)
  }
}

const handleLeave = async () => {
  await liveKit.disconnect()
  await apiLeaveMeeting(meetingId.value)
  inMeeting.value = false
  await router.push('/learning?tab=meetings')
}

const handleEndMeeting = async () => {
  if (!confirm('Apakah Anda yakin ingin mengakhiri meeting ini untuk semua peserta?')) return
  await liveKit.disconnect()
  await endMeeting(meetingId.value)
  inMeeting.value = false
  await router.push('/learning?tab=meetings')
}

const handleMuteParticipant = async (identity: string) => {
  await apiMuteParticipant(meetingId.value, identity)
}

const handleKickParticipant = async (identity: string) => {
  if (!confirm('Keluarkan peserta ini dari meeting?')) return
  await apiKickParticipant(meetingId.value, identity)
}

const handleDecideRequest = async (userId: string, status: 'approved' | 'rejected') => {
  await decideParticipant(meetingId.value, userId, status)
}

const handleDecideAllRequests = async (status: 'approved' | 'rejected') => {
  await decideAllParticipants(meetingId.value, status)
}

const getStudentName = (req: any) => {
  return (
    req.user?.student?.full_name ||
    req.user?.student?.FullName ||
    req.user?.teacher?.full_name ||
    req.user?.teacher?.FullName ||
    req.user?.name ||
    req.user?.Name ||
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
  return meetingContext(meeting) || 'Siswa'
}

const getStudentNIS = (req: any) => {
  const student = req.user?.student || req.user?.Student
  return student?.nis || student?.NIS || ''
}

// Periodic polling for admission requests when teacher (isHost)
let requestPollTimer: any = null

const startRequestPolling = () => {
  if (requestPollTimer) return
  requestPollTimer = setInterval(() => {
    if (isHost.value && meetingId.value) {
      void loadRequests(meetingId.value)
    }
  }, 3500)
}

const stopRequestPolling = () => {
  if (requestPollTimer) {
    clearInterval(requestPollTimer)
    requestPollTimer = null
  }
}

watch(
  isHost,
  (val) => {
    if (val) {
      startRequestPolling()
    } else {
      stopRequestPolling()
    }
  },
  { immediate: true },
)

onMounted(async () => {
  try {
    await loadMeeting(meetingId.value)
    await liveKit.loadDevices()
    if (!inMeeting.value) {
      void startPrejoinPreview()
    }
    if (isHost.value) {
      void loadRequests(meetingId.value)
      startRequestPolling()
    }
  } catch (e) {
    console.error('Error loading meeting page:', e)
  }
})

onBeforeUnmount(() => {
  stopRequestPolling()
  stopPrejoinPreview()
  void liveKit.disconnect()
})
</script>

<template>
  <div class="min-h-screen bg-[#f3f6f4] text-slate-900 flex flex-col font-sans">
    <!-- Active In-Meeting Room -->
    <div v-if="inMeeting && meeting" class="relative min-h-screen bg-slate-950 text-slate-100">
      <MeetingRoom
        :meeting="meeting"
        :participants="liveKit.participantsList.value"
        :current-screen-share="liveKit.currentScreenShare.value"
        :pending-requests="pendingRequests"
        :is-host="isHost"
        :is-mic-enabled="liveKit.isMicEnabled.value"
        :is-camera-enabled="liveKit.isCameraEnabled.value"
        :is-screen-sharing="liveKit.isScreenSharing.value"
        :connection-state="liveKit.connectionState.value"
        :connection-status-text="liveKit.connectionStatusText.value"
        :audio-inputs="liveKit.audioInputDevices.value"
        :video-inputs="liveKit.videoInputDevices.value"
        :audio-outputs="liveKit.audioOutputDevices.value"
        :selected-audio-input-id="liveKit.selectedAudioInputId.value"
        :selected-video-input-id="liveKit.selectedVideoInputId.value"
        @toggle-mic="liveKit.toggleMicrophone"
        @toggle-camera="liveKit.toggleCamera"
        @toggle-screen-share="liveKit.toggleScreenShare"
        @leave="handleLeave"
        @end-meeting="handleEndMeeting"
        @mute-participant="handleMuteParticipant"
        @kick-participant="handleKickParticipant"
        @decide-request="handleDecideRequest"
        @decide-all-requests="handleDecideAllRequests"
        @switch-camera="liveKit.switchCamera"
        @switch-microphone="liveKit.switchMicrophone"
      />

      <!-- Floating camera error toast + retry button -->
      <transition
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="opacity-0 translate-y-4"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-200 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 translate-y-4"
      >
        <div
          v-if="mediaErrorMsg || (!liveKit.isCameraEnabled.value && meeting.type !== 'voice')"
          class="absolute bottom-24 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 px-4 py-3 rounded-2xl bg-slate-900/95 backdrop-blur border border-amber-800/60 shadow-2xl text-xs max-w-sm w-[90%]"
        >
          <CameraOff :size="18" class="text-amber-400 shrink-0" />
          <span class="text-amber-200 flex-1">
            {{ mediaErrorMsg || 'Kamera tidak aktif. Klik Aktifkan untuk mencoba lagi.' }}
          </span>
          <button
            type="button"
            class="shrink-0 px-3 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-semibold transition-colors whitespace-nowrap"
            @click="handleRetryCamera"
          >
            Aktifkan
          </button>
          <button
            v-if="mediaErrorMsg"
            type="button"
            class="shrink-0 text-slate-400 hover:text-slate-200 transition-colors"
            @click="mediaErrorMsg = ''"
          >
            <X :size="16" />
          </button>
        </div>
      </transition>
    </div>

    <!-- Pre-join Lobby / Status Screen -->
    <div v-else class="relative flex-1 flex flex-col justify-center items-center p-4 sm:p-8 select-none overflow-hidden">
      <div class="absolute -right-24 -top-24 h-64 w-64 rounded-full bg-emerald-200/40 blur-3xl" />
      <div class="absolute -left-32 bottom-0 h-72 w-72 rounded-full bg-teal-100/60 blur-3xl" />

      <!-- Back Link -->
      <div class="relative w-full max-w-4xl mb-5">
        <NuxtLink
          to="/learning?tab=meetings"
          class="inline-flex items-center gap-2 text-xs font-semibold text-slate-500 hover:text-emerald-700 transition-colors"
        >
          <ArrowLeft :size="15" />
          Kembali ke Daftar Meeting
        </NuxtLink>
      </div>

      <!-- Loading State -->
      <div v-if="meetingLoading" class="relative flex flex-col items-center justify-center p-12 text-slate-500 gap-3">
        <div class="w-9 h-9 rounded-full border-2 border-emerald-500 border-t-transparent animate-spin" />
        <p class="text-sm">Menyiapkan ruang kelas...</p>
      </div>

      <!-- Error State -->
      <div
        v-else-if="meetingError"
        class="relative w-full max-w-lg p-7 rounded-3xl bg-white border border-rose-100 text-center space-y-4 shadow-[0_20px_60px_rgba(52,72,62,0.12)]"
      >
        <div class="w-12 h-12 rounded-2xl bg-rose-50 text-rose-500 flex items-center justify-center mx-auto">
          <AlertTriangle :size="24" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-900">Tidak Dapat Mengakses Meeting</h2>
          <p class="text-xs text-rose-600 mt-1">{{ meetingError }}</p>
        </div>
        <div class="flex justify-center gap-3">
          <button
            type="button"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-semibold text-xs transition-colors"
            @click="retryMeeting"
          >
            <RefreshCw :size="14" />
            Coba Lagi
          </button>
          <NuxtLink
            to="/learning?tab=meetings"
            class="px-4 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-semibold text-xs transition-colors"
          >
            Kembali
          </NuxtLink>
        </div>
      </div>

      <!-- Kicked / Rejected / Ended States -->
      <div
        v-else-if="isKicked || isRejected || meeting?.status === 'ended' || meeting?.status === 'cancelled'"
        class="relative w-full max-w-md p-7 rounded-3xl bg-white border border-amber-100 text-center space-y-4 shadow-[0_20px_60px_rgba(52,72,62,0.12)]"
      >
        <div class="w-12 h-12 rounded-2xl bg-amber-50 text-amber-500 flex items-center justify-center mx-auto">
          <LockKeyhole :size="24" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-900">
            {{
              isKicked
                ? 'Akses Dikeluarkan'
                : isRejected
                ? 'Pengajuan Ditolak'
                : 'Meeting Telah Berakhir'
            }}
          </h2>
          <p class="text-xs text-slate-500 mt-1">
            {{
              isKicked
                ? 'Anda telah dikeluarkan dari ruang meeting ini oleh pengajar.'
                : isRejected
                ? 'Permintaan akses Anda untuk ruang meeting ini tidak disetujui.'
                : 'Ruang meeting kelas ini sudah ditutup oleh pengajar.'
            }}
          </p>
        </div>
        <NuxtLink
          to="/learning?tab=meetings"
          class="inline-flex px-5 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-semibold text-xs transition-colors shadow-md"
        >
          Kembali ke Dashboard
        </NuxtLink>
      </div>

      <!-- Student Approval Pending State -->
      <div
        v-else-if="isPendingApproval"
        class="relative w-full max-w-md p-7 rounded-3xl bg-white border border-amber-100 text-center space-y-4 shadow-[0_20px_60px_rgba(52,72,62,0.12)]"
      >
        <div class="w-12 h-12 rounded-2xl bg-amber-50 text-amber-500 flex items-center justify-center mx-auto">
          <Clock3 :size="24" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-900">Menunggu Persetujuan Guru</h2>
          <p class="text-xs text-amber-700 mt-1">
            Permintaan akses Anda telah dikirimkan ke guru pengampu. Halaman ini akan dapat dimasuki setelah guru memberikan izin.
          </p>
        </div>
        <div class="flex justify-center gap-3">
          <button
            type="button"
            class="px-4 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-semibold text-xs transition-colors"
            @click="loadMeeting(meetingId)"
          >
            Periksa Status
          </button>
          <NuxtLink
            to="/learning?tab=meetings"
            class="px-4 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-semibold text-xs transition-colors"
          >
            Kembali
          </NuxtLink>
        </div>
      </div>

      <!-- Student Needs to Request Access -->
      <div
        v-else-if="!isApproved && meeting?.approval_required"
        class="relative w-full max-w-md p-7 rounded-3xl bg-white border border-indigo-100 text-center space-y-4 shadow-[0_20px_60px_rgba(52,72,62,0.12)]"
      >
        <div class="w-12 h-12 rounded-2xl bg-indigo-50 text-indigo-500 flex items-center justify-center mx-auto">
          <Hand :size="24" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-900">Izin Masuk Diperlukan</h2>
          <p class="text-xs text-slate-500 mt-1">
            Meeting kelas ini memerlukan izin persetujuan dari guru sebelum Anda dapat bergabung.
          </p>
        </div>
        <button
          type="button"
          class="w-full inline-flex items-center justify-center gap-2 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-semibold text-xs transition-colors shadow-md"
          @click="requestAccess(meetingId)"
        >
          <Hand :size="15" />
          Ajukan Izin Masuk
        </button>
      </div>

      <!-- Ready to Join / Pre-join Lobby -->
      <div
        v-else-if="meeting"
        class="relative w-full max-w-4xl bg-white border border-slate-200 rounded-3xl p-5 sm:p-7 shadow-[0_20px_60px_rgba(52,72,62,0.12)] space-y-6"
      >
        <div class="flex items-start justify-between gap-4 border-b border-slate-100 pb-5">
          <div>
            <span class="px-2.5 py-1 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 uppercase tracking-wider">
              {{ meeting.type }} · {{ meeting.status }}
            </span>
            <h1 class="text-lg sm:text-xl font-bold text-slate-900 mt-2">
              {{ meeting.title }}
            </h1>
            <p class="text-xs text-slate-500 mt-1">
              {{ meetingContext(meeting) }}
            </p>
          </div>
          <div class="text-right">
            <span class="text-xs font-semibold text-slate-700 block">Kapasitas</span>
            <span class="text-xs text-slate-500">{{ meeting.max_participants }} Peserta</span>
          </div>
        </div>

        <!-- Teacher Pending Requests Card in Lobby -->
        <div
          v-if="isHost && pendingRequests && pendingRequests.length > 0"
          class="p-4 sm:p-5 rounded-2xl bg-amber-50/80 border border-amber-200/90 space-y-3 shadow-xs"
        >
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="flex items-center gap-2.5">
              <div class="w-8 h-8 rounded-xl bg-amber-500/20 text-amber-800 flex items-center justify-center shrink-0">
                <Users :size="16" />
              </div>
              <div>
                <h4 class="text-xs sm:text-sm font-bold text-amber-950">
                  Pengajuan Izin Masuk ({{ pendingRequests.length }} Siswa)
                </h4>
                <p class="text-[11px] text-amber-700">
                  Siswa berikut telah mengajukan izin masuk dan menunggu persetujuan Anda.
                </p>
              </div>
            </div>

            <div class="flex items-center gap-2">
              <button
                type="button"
                class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs transition-all shadow-sm active:scale-95"
                @click="handleDecideAllRequests('approved')"
              >
                <CheckCheck :size="14" />
                Izinkan Semua ({{ pendingRequests.length }})
              </button>
              <button
                type="button"
                class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-rose-50 hover:bg-rose-100 text-rose-700 font-semibold text-xs border border-rose-200 transition-colors"
                @click="handleDecideAllRequests('rejected')"
              >
                <X :size="14" />
                Tolak Semua
              </button>
            </div>
          </div>

          <!-- Pending Students List -->
          <div class="space-y-2 max-h-48 overflow-y-auto pr-1">
            <div
              v-for="req in pendingRequests"
              :key="req.id"
              class="flex items-center justify-between gap-3 p-2.5 rounded-xl bg-white border border-amber-200/70 shadow-xs"
            >
              <div class="flex items-center gap-2.5 min-w-0">
                <div class="w-8 h-8 rounded-full bg-emerald-100 text-emerald-800 flex items-center justify-center text-xs font-bold shrink-0">
                  {{ (getStudentName(req)[0] || 'S').toUpperCase() }}
                </div>
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-1.5 sm:gap-2">
                    <span class="text-xs font-bold text-slate-800 truncate">
                      {{ getStudentName(req) }}
                    </span>
                    <span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-[10px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200 shrink-0">
                      <GraduationCap :size="11" />
                      {{ getStudentClass(req) }}
                    </span>
                  </div>
                  <span v-if="getStudentNIS(req)" class="text-[10px] text-slate-400 block">
                    NIS: {{ getStudentNIS(req) }}
                  </span>
                </div>
              </div>

              <div class="flex items-center gap-1.5 shrink-0">
                <button
                  type="button"
                  class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white font-semibold text-xs transition-colors"
                  @click="handleDecideRequest(req.user_id, 'approved')"
                >
                  <Check :size="13" />
                  Izinkan
                </button>
                <button
                  type="button"
                  class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-rose-50 hover:bg-rose-100 text-rose-700 font-semibold text-xs border border-rose-200 transition-colors"
                  @click="handleDecideRequest(req.user_id, 'rejected')"
                >
                  <X :size="13" />
                  Tolak
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Camera Preview & Pre-join Toggles -->
        <div class="flex flex-col sm:flex-row gap-4 items-center">
          <div class="relative w-full sm:w-[23rem] h-52 rounded-2xl bg-slate-950 border border-slate-200 overflow-hidden flex items-center justify-center shrink-0 shadow-inner">
            <video
              v-show="prejoinCamera && meeting.type !== 'voice'"
              ref="localPreviewEl"
              autoplay
              playsinline
              muted
              class="w-full h-full object-cover -scale-x-100"
            />
            <div
              v-show="!prejoinCamera || meeting.type === 'voice'"
              class="flex flex-col items-center justify-center text-center p-4"
            >
              <div class="w-12 h-12 rounded-2xl bg-slate-800 text-slate-400 flex items-center justify-center mb-2">
                <CameraOff :size="22" />
              </div>
              <span class="text-xs text-slate-400">Kamera dinonaktifkan</span>
            </div>

            <!-- Prejoin Controls Overlay -->
            <div class="absolute bottom-2 inset-x-2 flex justify-center gap-2">
              <button
                type="button"
                class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold backdrop-blur transition-colors"
                :class="prejoinMic ? 'bg-slate-800/80 text-emerald-300' : 'bg-rose-600/90 text-white'"
                @click="togglePrejoinMic"
              >
                <Mic v-if="prejoinMic" :size="14" />
                <MicOff v-else :size="14" />
                {{ prejoinMic ? 'Mic Nyala' : 'Mic Mati' }}
              </button>
              <button
                v-if="meeting.type !== 'voice'"
                type="button"
                class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold backdrop-blur transition-colors"
                :class="prejoinCamera ? 'bg-slate-800/80 text-emerald-300' : 'bg-rose-600/90 text-white'"
                @click="togglePrejoinCamera"
              >
                <Video v-if="prejoinCamera" :size="14" />
                <CameraOff v-else :size="14" />
                {{ prejoinCamera ? 'Kamera Nyala' : 'Kamera Mati' }}
              </button>
            </div>
          </div>

          <!-- Join Action Area -->
          <div class="flex-1 flex flex-col justify-center space-y-4 w-full">
            <div>
              <h3 class="text-sm font-bold text-slate-900">Siap Bergabung?</h3>
              <p class="text-xs text-slate-500 mt-1 leading-relaxed">
                {{
                  meeting.status === 'scheduled' && !isHost
                    ? 'Meeting dijadwalkan oleh guru. Tunggu guru memulai meeting.'
                    : 'Kamera dan mikrofon Anda akan terhubung langsung ke media server LiveKit.'
                }}
              </p>
            </div>

            <!-- Join Actions -->
            <div class="space-y-2">
              <!-- If Teacher and meeting is scheduled -->
              <button
                v-if="isHost && meeting.status === 'scheduled'"
                type="button"
                class="w-full py-3 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs shadow-lg shadow-emerald-900/15 transition-all flex items-center justify-center gap-2"
                :disabled="joining"
                @click="handleStartAndJoin"
              >
                <RefreshCw v-if="joining" :size="14" class="animate-spin" />
                <Play v-else :size="14" />
                <span>{{ joining ? 'Menyiapkan...' : 'Mulai & Gabung Meeting' }}</span>
              </button>

              <!-- Join meeting button -->
              <button
                v-else
                type="button"
                class="w-full py-3 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs shadow-lg shadow-emerald-900/15 transition-all flex items-center justify-center gap-2 disabled:opacity-50"
                :disabled="joining || (meeting.status === 'scheduled' && !isHost)"
                @click="handleJoin"
              >
                <RefreshCw v-if="joining" :size="14" class="animate-spin" />
                <Video v-else :size="14" />
                <span>{{ joining ? 'Menghubungkan...' : 'Masuk Ruang Meeting' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
