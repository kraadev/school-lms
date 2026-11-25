<script setup lang="ts">
import { computed } from 'vue'
import {
  DoorOpen,
  Mic,
  MicOff,
  MonitorUp,
  Settings,
  Users,
  Video,
  CameraOff,
  XCircle,
} from 'lucide-vue-next'

const props = defineProps<{
  isMicEnabled: boolean
  isCameraEnabled: boolean
  isScreenSharing: boolean
  canScreenShare?: boolean
  participantCount: number
  pendingRequestsCount?: number
  isParticipantListOpen: boolean
  isHost?: boolean
  isVoiceOnly?: boolean
}>()

const emit = defineEmits<{
  (e: 'toggleMic'): void
  (e: 'toggleCamera'): void
  (e: 'toggleScreenShare'): void
  (e: 'toggleParticipantList'): void
  (e: 'openSettings'): void
  (e: 'leave'): void
  (e: 'endMeeting'): void
}>()
</script>

<template>
  <div class="flex items-center justify-center p-3 select-none">
    <div class="flex items-center gap-2 sm:gap-3 px-3 sm:px-5 py-2.5 rounded-2xl bg-slate-900/90 backdrop-blur-md border border-slate-800 shadow-2xl">
      <!-- Microphone Button -->
      <button
        type="button"
        class="flex flex-col items-center justify-center w-11 h-11 sm:w-12 sm:h-12 rounded-xl transition-all duration-200"
        :class="[
          isMicEnabled
            ? 'bg-slate-800 hover:bg-slate-700 text-slate-100 border border-slate-700/60'
            : 'bg-rose-600/90 hover:bg-rose-500 text-white border border-rose-500 shadow-lg shadow-rose-900/30',
        ]"
        :title="isMicEnabled ? 'Matikan Microphone' : 'Nyalakan Microphone'"
        @click="emit('toggleMic')"
      >
        <Mic v-if="isMicEnabled" :size="18" />
        <MicOff v-else :size="18" />
        <span class="text-[9px] font-medium mt-0.5 leading-none">{{ isMicEnabled ? 'Mic' : 'Mute' }}</span>
      </button>

      <!-- Camera Button (if not voice only) -->
      <button
        v-if="!isVoiceOnly"
        type="button"
        class="flex flex-col items-center justify-center w-11 h-11 sm:w-12 sm:h-12 rounded-xl transition-all duration-200"
        :class="[
          isCameraEnabled
            ? 'bg-slate-800 hover:bg-slate-700 text-slate-100 border border-slate-700/60'
            : 'bg-rose-600/90 hover:bg-rose-500 text-white border border-rose-500 shadow-lg shadow-rose-900/30',
        ]"
        :title="isCameraEnabled ? 'Matikan Kamera' : 'Nyalakan Kamera'"
        @click="emit('toggleCamera')"
      >
        <Video v-if="isCameraEnabled" :size="18" />
        <CameraOff v-else :size="18" />
        <span class="text-[9px] font-medium mt-0.5 leading-none">{{ isCameraEnabled ? 'Kamera' : 'Off' }}</span>
      </button>

      <!-- Screen Share Button -->
      <button
        v-if="canScreenShare !== false && !isVoiceOnly"
        type="button"
        class="flex flex-col items-center justify-center w-11 h-11 sm:w-12 sm:h-12 rounded-xl transition-all duration-200"
        :class="[
          isScreenSharing
            ? 'bg-emerald-600 hover:bg-emerald-500 text-white border border-emerald-400 ring-2 ring-emerald-500/30 shadow-lg shadow-emerald-950/40'
            : 'bg-slate-800 hover:bg-slate-700 text-slate-100 border border-slate-700/60',
        ]"
        :title="isScreenSharing ? 'Hentikan Bagikan Layar' : 'Bagikan Layar'"
        @click="emit('toggleScreenShare')"
      >
        <MonitorUp :size="18" />
        <span class="text-[9px] font-medium mt-0.5 leading-none">Layar</span>
      </button>

      <div class="h-6 w-px bg-slate-800 my-auto" />

      <!-- Participants Toggle Button -->
      <button
        type="button"
        class="relative flex flex-col items-center justify-center w-11 h-11 sm:w-12 sm:h-12 rounded-xl transition-all duration-200"
        :class="[
          isParticipantListOpen
            ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/40'
            : (pendingRequestsCount && pendingRequestsCount > 0)
            ? 'bg-amber-500/20 text-amber-300 border border-amber-500/50 shadow-lg shadow-amber-950/40'
            : 'bg-slate-800 hover:bg-slate-700 text-slate-100 border border-slate-700/60',
        ]"
        :title="pendingRequestsCount && pendingRequestsCount > 0 ? `${pendingRequestsCount} siswa menunggu izin masuk` : 'Daftar Peserta'"
        @click="emit('toggleParticipantList')"
      >
        <Users :size="18" />
        <span class="text-[9px] font-medium mt-0.5 leading-none">Peserta</span>
        <span
          v-if="pendingRequestsCount && pendingRequestsCount > 0"
          class="absolute -top-1 -right-1 px-1.5 py-0.2 rounded-full text-[9px] font-bold bg-amber-500 text-white animate-pulse shadow-md ring-2 ring-slate-900"
        >
          ! {{ pendingRequestsCount }}
        </span>
        <span
          v-else
          class="absolute -top-1 -right-1 px-1.5 py-0.2 rounded-full text-[9px] font-bold bg-emerald-500 text-white shadow-sm"
        >
          {{ participantCount }}
        </span>
      </button>

      <!-- Device Settings Button -->
      <button
        type="button"
        class="flex flex-col items-center justify-center w-11 h-11 sm:w-12 sm:h-12 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-100 border border-slate-700/60 transition-all duration-200"
        title="Pengaturan Perangkat"
        @click="emit('openSettings')"
      >
        <Settings :size="18" />
        <span class="text-[9px] font-medium mt-0.5 leading-none">Setelan</span>
      </button>

      <div class="h-6 w-px bg-slate-800 my-auto" />

      <!-- Teacher End Meeting vs Student Leave Button -->
      <div class="flex items-center gap-1.5">
        <button
          v-if="isHost"
          type="button"
          class="flex items-center gap-1.5 px-3 sm:px-4 py-2 sm:py-2.5 rounded-xl bg-rose-600 hover:bg-rose-500 text-white font-semibold text-xs transition-colors shadow-lg shadow-rose-950/40"
          title="Akhiri Meeting untuk Semua Peserta"
          @click="emit('endMeeting')"
        >
          <XCircle :size="16" />
          <span class="hidden sm:inline">Akhiri</span>
        </button>

        <button
          type="button"
          class="flex items-center gap-1.5 px-3 sm:px-4 py-2 sm:py-2.5 rounded-xl bg-slate-800 hover:bg-rose-950/80 hover:text-rose-300 hover:border-rose-800 text-slate-200 font-semibold text-xs border border-slate-700/80 transition-all duration-200"
          title="Keluar dari Meeting"
          @click="emit('leave')"
        >
          <DoorOpen :size="16" />
          <span class="hidden sm:inline">Keluar</span>
        </button>
      </div>
    </div>
  </div>
</template>
