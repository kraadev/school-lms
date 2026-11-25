<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue'
import type { LiveKitParticipantInfo } from '~/composables/useLiveKit'
import type { Track } from 'livekit-client'

const props = defineProps<{
  screenShare: {
    participant: LiveKitParticipantInfo
    track: Track
  }
}>()

const videoEl = ref<HTMLVideoElement | null>(null)

const attachTrack = () => {
  if (props.screenShare?.track && videoEl.value) {
    props.screenShare.track.attach(videoEl.value)
  }
}

onMounted(() => {
  attachTrack()
})

watch(
  () => props.screenShare?.track,
  async () => {
    await nextTick()
    attachTrack()
  }
)
</script>

<template>
  <div class="relative w-full h-full flex flex-col justify-center items-center overflow-hidden rounded-2xl bg-black border border-slate-800 shadow-2xl">
    <video
      ref="videoEl"
      autoplay
      playsinline
      class="w-full h-full object-contain max-h-[calc(100vh-220px)]"
    />

    <!-- Presenter Overlay Banner -->
    <div class="absolute top-3 left-3 z-10 flex items-center gap-2 px-3 py-1.5 rounded-xl bg-slate-950/80 backdrop-blur border border-slate-700/60 text-xs text-slate-200 shadow-md">
      <span class="w-2 h-2 rounded-full bg-emerald-400 animate-ping" />
      <span class="font-medium text-emerald-300">{{ screenShare.participant.name }}</span>
      <span class="text-slate-400">sedang membagikan layar</span>
    </div>
  </div>
</template>
