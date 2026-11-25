<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Settings, X } from 'lucide-vue-next'

const props = defineProps<{
  audioInputs: MediaDeviceInfo[]
  videoInputs: MediaDeviceInfo[]
  audioOutputs: MediaDeviceInfo[]
  selectedAudioInputId: string
  selectedVideoInputId: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'switchCamera', deviceId: string): void
  (e: 'switchMicrophone', deviceId: string): void
}>()

const previewVideoEl = ref<HTMLVideoElement | null>(null)
const previewStream = ref<MediaStream | null>(null)
const audioLevel = ref<number>(0)
let audioCtx: AudioContext | null = null
let analyser: AnalyserNode | null = null
let animFrame: number | null = null

const startPreview = async () => {
  try {
    const videoConstraints: MediaTrackConstraints = props.selectedVideoInputId
      ? { deviceId: { exact: props.selectedVideoInputId } }
      : true
    const audioConstraints: MediaTrackConstraints = props.selectedAudioInputId
      ? { deviceId: { exact: props.selectedAudioInputId } }
      : true

    const stream = await navigator.mediaDevices.getUserMedia({
      video: videoConstraints,
      audio: audioConstraints,
    })
    previewStream.value = stream

    if (previewVideoEl.value) {
      previewVideoEl.value.srcObject = stream
    }

    // Audio meter
    const audioTrack = stream.getAudioTracks()[0]
    if (audioTrack) {
      audioCtx = new (window.AudioContext || (window as any).webkitAudioContext)()
      const source = audioCtx.createMediaStreamSource(stream)
      analyser = audioCtx.createAnalyser()
      analyser.fftSize = 64
      source.connect(analyser)

      const dataArray = new Uint8Array(analyser.frequencyBinCount)
      const updateMeter = () => {
        if (!analyser) return
        analyser.getByteFrequencyData(dataArray)
        let sum = 0
        for (let i = 0; i < dataArray.length; i++) {
          sum += dataArray[i]
        }
        audioLevel.value = Math.min(100, Math.round((sum / dataArray.length / 255) * 150))
        animFrame = requestAnimationFrame(updateMeter)
      }
      updateMeter()
    }
  } catch (e) {
    console.warn('Could not start settings preview:', e)
  }
}

const stopPreview = () => {
  if (animFrame) cancelAnimationFrame(animFrame)
  if (previewStream.value) {
    previewStream.value.getTracks().forEach((t) => t.stop())
    previewStream.value = null
  }
  if (audioCtx) {
    audioCtx.close()
    audioCtx = null
  }
}

onMounted(() => {
  void startPreview()
})

onBeforeUnmount(() => {
  stopPreview()
})

const onCameraChange = (e: Event) => {
  const target = e.target as HTMLSelectElement
  emit('switchCamera', target.value)
  stopPreview()
  void startPreview()
}

const onMicrophoneChange = (e: Event) => {
  const target = e.target as HTMLSelectElement
  emit('switchMicrophone', target.value)
  stopPreview()
  void startPreview()
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm animate-fadeIn select-none">
    <div class="flex flex-col w-full max-w-lg bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl overflow-hidden">
      <!-- Modal Header -->
      <div class="flex items-center justify-between p-4 border-b border-slate-800">
        <h2 class="text-sm font-semibold text-slate-100 flex items-center gap-2">
          <Settings :size="16" />
          Pengaturan Perangkat
        </h2>
        <button
          type="button"
          class="p-1.5 text-slate-400 hover:text-slate-100 hover:bg-slate-800 rounded-lg transition-colors"
          @click="emit('close')"
        >
          <X :size="16" />
        </button>
      </div>

      <!-- Body -->
      <div class="p-5 space-y-4 max-h-[80vh] overflow-y-auto">
        <!-- Camera Preview -->
        <div>
          <label class="block text-xs font-medium text-slate-300 mb-1.5">Kamera</label>
          <div class="relative w-full h-44 rounded-xl bg-slate-950 border border-slate-800 overflow-hidden flex items-center justify-center mb-2">
            <video
              ref="previewVideoEl"
              autoplay
              playsinline
              muted
              class="w-full h-full object-cover -scale-x-100"
            />
            <div v-if="!previewStream" class="text-xs text-slate-500">
              Pratinjau kamera tidak aktif
            </div>
          </div>
          <select
            :value="selectedVideoInputId"
            class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-xs text-slate-200 focus:outline-none focus:border-emerald-500"
            @change="onCameraChange"
          >
            <option value="">Default Kamera</option>
            <option
              v-for="d in videoInputs"
              :key="d.deviceId"
              :value="d.deviceId"
            >
              {{ d.label || `Kamera ${d.deviceId.slice(0, 5)}` }}
            </option>
          </select>
        </div>

        <!-- Microphone Settings & Meter -->
        <div>
          <label class="block text-xs font-medium text-slate-300 mb-1.5">Microphone</label>
          <select
            :value="selectedAudioInputId"
            class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-xs text-slate-200 focus:outline-none focus:border-emerald-500 mb-2"
            @change="onMicrophoneChange"
          >
            <option value="">Default Microphone</option>
            <option
              v-for="d in audioInputs"
              :key="d.deviceId"
              :value="d.deviceId"
            >
              {{ d.label || `Microphone ${d.deviceId.slice(0, 5)}` }}
            </option>
          </select>

          <!-- Live Mic Volume Meter -->
          <div class="space-y-1">
            <div class="flex items-center justify-between text-[11px] text-slate-400">
              <span>Sensitivitas Suara</span>
              <span>{{ audioLevel }}%</span>
            </div>
            <div class="w-full h-2 rounded-full bg-slate-950 border border-slate-800 overflow-hidden">
              <div
                class="h-full bg-gradient-to-r from-emerald-500 to-teal-400 transition-all duration-75 rounded-full"
                :style="{ width: `${audioLevel}%` }"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end p-4 border-t border-slate-800 bg-slate-950/40">
        <button
          type="button"
          class="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-xs transition-colors shadow-sm"
          @click="emit('close')"
        >
          Selesai
        </button>
      </div>
    </div>
  </div>
</template>
