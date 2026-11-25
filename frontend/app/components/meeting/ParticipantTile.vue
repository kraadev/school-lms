<script setup lang="ts">
import { ref, watch, onMounted, computed, nextTick } from 'vue'
import { Camera, CameraOff, Mic, MicOff, UserRoundX } from 'lucide-vue-next'
import type { LiveKitParticipantInfo } from '~/composables/useLiveKit'
import { Track } from 'livekit-client'

const props = defineProps<{
  participant: LiveKitParticipantInfo
  isHost?: boolean
  canControl?: boolean
}>()

const emit = defineEmits<{
  (e: 'mute', identity: string): void
  (e: 'kick', identity: string): void
  (e: 'toggleScreenPermission', identity: string): void
}>()

const videoEl = ref<HTMLVideoElement | null>(null)
const audioEl = ref<HTMLAudioElement | null>(null)

const initials = computed(() => {
  const name = props.participant.name || 'User'
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase())
    .join('') || '?'
})

const attachVideo = () => {
  if (!videoEl.value) return
  const camPub = props.participant.cameraPublication
  if (camPub?.track && videoEl.value) {
    camPub.track.attach(videoEl.value)
  } else if (videoEl.value) {
    videoEl.value.srcObject = null
  }
}

const attachAudio = () => {
  if (props.participant.isLocal) return // Don't play local audio to avoid feedback
  if (!audioEl.value) return
  const micPub = props.participant.microphonePublication
  if (micPub?.track && audioEl.value) {
    micPub.track.attach(audioEl.value)
  }
}

onMounted(() => {
  attachVideo()
  attachAudio()
})

watch(
  () => [
    props.participant.cameraPublication?.track,
    props.participant.cameraPublication?.isMuted,
    props.participant.isVideoEnabled,
  ],
  async () => {
    await nextTick()
    attachVideo()
  }
)

watch(
  () => [
    props.participant.microphonePublication?.track,
    props.participant.microphonePublication?.isMuted,
  ],
  async () => {
    await nextTick()
    attachAudio()
  }
)
</script>

<template>
  <div
    class="relative group flex flex-col justify-between overflow-hidden rounded-2xl bg-slate-900 border transition-all duration-200 shadow-lg min-h-[180px] h-full"
    :class="[
      participant.isSpeaking
        ? 'border-emerald-400 ring-2 ring-emerald-400/40 shadow-emerald-950/40'
        : 'border-slate-800/80 hover:border-slate-700/80',
    ]"
  >
    <!-- Background Video -->
    <div class="absolute inset-0 flex items-center justify-center bg-slate-950">
      <video
        v-show="participant.isVideoEnabled"
        ref="videoEl"
        autoplay
        playsinline
        :muted="participant.isLocal"
        class="w-full h-full object-cover"
        :class="{ '-scale-x-100': participant.isLocal }"
      />

      <!-- Hidden Audio element for remote participant -->
      <audio v-if="!participant.isLocal" ref="audioEl" autoplay />

      <!-- Avatar Fallback when video is off -->
      <div
        v-show="!participant.isVideoEnabled"
        class="flex flex-col items-center justify-center gap-3 p-4 text-center select-none"
      >
        <div
          class="w-16 h-16 sm:w-20 sm:h-20 rounded-full flex items-center justify-center text-xl sm:text-2xl font-bold font-heading text-white shadow-xl transition-transform duration-300 group-hover:scale-105"
          :class="[
            participant.role === 'guru'
              ? 'bg-gradient-to-br from-emerald-500 to-teal-700 ring-4 ring-emerald-500/20'
              : 'bg-gradient-to-br from-indigo-600 to-slate-800 ring-4 ring-indigo-500/20',
          ]"
        >
          {{ initials }}
        </div>
        <p class="text-xs sm:text-sm font-medium text-slate-300 max-w-[160px] truncate">
          {{ participant.name }}
        </p>
      </div>
    </div>

    <!-- Top Badges / Role & Controls -->
    <div class="relative z-10 flex items-center justify-between p-3">
      <div class="flex items-center gap-1.5">
        <span
          v-if="participant.role === 'guru'"
          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold bg-emerald-500/90 text-white backdrop-blur shadow-sm"
        >
          <span>🎓</span> Guru
        </span>
        <span
          v-else-if="participant.role === 'admin'"
          class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-amber-500/90 text-white backdrop-blur shadow-sm"
        >
          Admin
        </span>
        <span
          v-if="participant.isLocal"
          class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium bg-slate-800/80 text-emerald-300 backdrop-blur border border-slate-700/60"
        >
          Anda
        </span>
      </div>

      <!-- Action dropdown for Host/Teacher -->
      <div v-if="canControl && !participant.isLocal" class="opacity-0 group-hover:opacity-100 transition-opacity">
        <div class="flex items-center gap-1 bg-slate-900/90 backdrop-blur p-1 rounded-xl border border-slate-700/80 shadow-md">
          <button
            type="button"
            class="p-1.5 text-xs text-slate-300 hover:text-amber-400 hover:bg-slate-800 rounded-lg transition-colors"
            title="Mute Peserta"
            @click="emit('mute', participant.identity)"
          >
            <MicOff :size="13" />
          </button>
          <button
            type="button"
            class="p-1.5 text-xs text-slate-300 hover:text-rose-400 hover:bg-slate-800 rounded-lg transition-colors"
            title="Keluarkan dari Room"
            @click="emit('kick', participant.identity)"
          >
            <UserRoundX :size="13" />
          </button>
        </div>
      </div>
    </div>

    <!-- Bottom Name Tag & Audio Status -->
    <div class="relative z-10 flex items-center justify-between gap-2 p-2.5 m-2 rounded-xl bg-slate-950/75 backdrop-blur border border-slate-800/60 text-xs">
      <div class="flex items-center gap-2 min-w-0">
        <span class="font-medium text-slate-100 truncate text-[12px] sm:text-xs">
          {{ participant.name }}
        </span>
      </div>

      <div class="flex items-center gap-1.5 shrink-0">
        <!-- Speaking wave indicator -->
        <span
          v-if="participant.isSpeaking"
          class="inline-flex items-center px-1.5 py-0.5 rounded-md text-[10px] font-semibold bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 animate-pulse"
        >
          Bicara
        </span>

        <!-- Microphone Status -->
        <span
          class="flex items-center justify-center w-5 h-5 rounded-full text-[11px]"
          :class="[
            participant.isAudioEnabled
              ? 'bg-emerald-500/20 text-emerald-400'
              : 'bg-rose-500/20 text-rose-400',
          ]"
        >
          <Mic v-if="participant.isAudioEnabled" :size="13" />
          <MicOff v-else :size="13" />
        </span>

        <!-- Camera Status -->
        <span
          class="flex items-center justify-center w-5 h-5 rounded-full text-[11px]"
          :class="[
            participant.isVideoEnabled
              ? 'bg-emerald-500/20 text-emerald-400'
              : 'bg-slate-800 text-slate-400',
          ]"
        >
          <Camera v-if="participant.isVideoEnabled" :size="13" />
          <CameraOff v-else :size="13" />
        </span>
      </div>
    </div>
  </div>
</template>
