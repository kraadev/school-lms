<script setup lang="ts">
import { computed } from 'vue'
import { Users } from 'lucide-vue-next'
import type { LiveKitParticipantInfo } from '~/composables/useLiveKit'
import ParticipantTile from '~/components/meeting/ParticipantTile.vue'

const props = defineProps<{
  participants: LiveKitParticipantInfo[]
  isHost?: boolean
  canControl?: boolean
  isScreenSharingActive?: boolean
}>()

const emit = defineEmits<{
  (e: 'mute', identity: string): void
  (e: 'kick', identity: string): void
  (e: 'toggleScreenPermission', identity: string): void
}>()

const gridClass = computed(() => {
  const count = props.participants.length
  if (props.isScreenSharingActive) {
    return 'flex gap-3 overflow-x-auto pb-2 shrink-0 max-h-[160px]'
  }
  if (count <= 1) {
    return 'grid grid-cols-1 max-w-4xl mx-auto w-full h-full p-2 sm:p-4'
  }
  if (count === 2) {
    return 'grid grid-cols-1 md:grid-cols-2 gap-3 sm:gap-4 max-w-5xl mx-auto w-full h-full p-2 sm:p-4'
  }
  if (count <= 4) {
    return 'grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4 w-full h-full p-2 sm:p-4'
  }
  if (count <= 6) {
    return 'grid grid-cols-2 sm:grid-cols-3 gap-3 sm:gap-4 w-full h-full p-2 sm:p-4'
  }
  if (count <= 9) {
    return 'grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-3 gap-3 sm:gap-4 w-full h-full p-2 sm:p-4'
  }
  return 'grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3 sm:gap-4 w-full h-full p-2 sm:p-4'
})
</script>

<template>
  <div class="w-full h-full flex flex-col justify-center items-center overflow-y-auto">
    <div :class="gridClass">
      <div
        v-for="p in participants"
        :key="p.identity"
        class="w-full h-full transition-all duration-300"
        :class="[
          isScreenSharingActive ? 'w-[200px] h-[130px] shrink-0' : 'min-h-[180px]',
        ]"
      >
        <ParticipantTile
          :participant="p"
          :is-host="isHost"
          :can-control="canControl"
          @mute="emit('mute', $event)"
          @kick="emit('kick', $event)"
          @toggle-screen-permission="emit('toggleScreenPermission', $event)"
        />
      </div>

      <!-- Empty state if no one is in room -->
      <div
        v-if="!participants.length"
        class="flex flex-col items-center justify-center p-8 rounded-2xl bg-slate-900/60 border border-slate-800 text-center max-w-md mx-auto"
      >
        <div class="w-12 h-12 rounded-full bg-emerald-500/10 text-emerald-400 flex items-center justify-center text-xl mb-3">
          <Users :size="22" />
        </div>
        <h3 class="text-sm font-semibold text-slate-200 mb-1">Menunggu Peserta</h3>
        <p class="text-xs text-slate-400">
          Peserta yang bergabung dengan meeting ini akan muncul secara otomatis di sini.
        </p>
      </div>
    </div>
  </div>
</template>
