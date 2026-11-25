<script setup lang="ts">
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
auth.restore()

const videoRef = ref<HTMLVideoElement | null>(null)
const showSkip = ref(false)
const done = ref(false)

const proceed = () => {
  if (done.value) return
  done.value = true
  navigateTo(auth.isAuthenticated ? '/dashboard' : '/login')
}

onMounted(() => {
  setTimeout(() => { showSkip.value = true }, 1500)
  setTimeout(proceed, 8000)
})
</script>

<template>
  <div class="intro-screen">
    <video
      ref="videoRef"
      class="intro-video"
      autoplay
      muted
      playsinline
      @ended="proceed"
      @error="proceed"
    >
      <source src="/video/intro.mp4" type="video/mp4" />
    </video>

    <div class="intro-overlay">
      <div class="intro-brand">
        <img src="/images/logo.png" alt="School Space" class="intro-logo" />
        <h1 class="intro-title">School Space</h1>
        <p class="intro-tagline">Ruang belajar yang cerdas dan terhubung</p>
      </div>
    </div>

    <button v-if="showSkip" class="intro-skip" @click="proceed">
      Lewati →
    </button>
  </div>
</template>

<style scoped>
.intro-screen {
  position: fixed;
  inset: 0;
  background: #0d1f17;
  overflow: hidden;
  z-index: 9999;
}
.intro-video {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  opacity: 0.85;
}
.intro-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(13,31,23,0.55) 0%, rgba(94,182,157,0.18) 100%);
  animation: fadeIn 0.8s ease forwards;
}
.intro-brand {
  text-align: center;
  animation: slideUp 0.9s cubic-bezier(0.22,1,0.36,1) 0.3s both;
}
.intro-logo {
  width: 72px;
  height: 72px;
  border-radius: 20px;
  object-fit: cover;
  margin-bottom: 18px;
  filter: drop-shadow(0 8px 24px rgba(94,182,157,0.5));
}
.intro-title {
  font-family: 'Poppins', sans-serif;
  font-size: 42px;
  font-weight: 800;
  color: white;
  margin: 0 0 10px;
  letter-spacing: -0.02em;
  text-shadow: 0 4px 20px rgba(0,0,0,0.4);
}
.intro-tagline {
  font-family: 'Manrope', sans-serif;
  font-size: 16px;
  color: rgba(255,255,255,0.82);
  margin: 0;
  font-weight: 500;
}
.intro-skip {
  position: absolute;
  bottom: 36px;
  right: 36px;
  padding: 10px 22px;
  border-radius: 999px;
  border: 1.5px solid rgba(255,255,255,0.4);
  background: rgba(255,255,255,0.12);
  backdrop-filter: blur(12px);
  color: white;
  font-family: 'Manrope', sans-serif;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
  animation: fadeIn 0.5s ease forwards;
}
.intro-skip:hover {
  background: rgba(255,255,255,0.22);
  border-color: rgba(255,255,255,0.7);
  transform: translateX(3px);
}
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
@keyframes slideUp { from { opacity: 0; transform: translateY(30px); } to { opacity: 1; transform: translateY(0); } }
</style>

