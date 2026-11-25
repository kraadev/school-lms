<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { useApi } from '../composables/useApi'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const { request } = useApi()
const username = ref('')
const password = ref('')
const loading = ref(false)
const errorMessage = ref('')

onMounted(() => {
  auth.restore()
  if (auth.isAuthenticated) navigateTo(String(route.query.redirect || '/dashboard'))
})

const submit = async () => {
  errorMessage.value = ''
  loading.value = true
  try {
    const response = await request<{ success: boolean; data: { token: string; user: any }; message?: string }>('/auth/login', {
      method: 'POST',
      body: { username: username.value, password: password.value },
    })
    auth.setSession(response.data.token, response.data.user)
    await navigateTo(String(route.query.redirect || '/dashboard'))
  } catch (error: any) {
    errorMessage.value = error?.data?.message || 'Username atau password belum benar.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="auth-page">
    <section class="auth-intro">
      <div class="brand-row">
        <img src="/images/logo.png" alt="School LMS logo" class="login-logo" />
        <span class="brand-name">School LMS</span>
      </div>

      <p class="eyebrow">School learning management system</p>
      <h1>Ruang belajar yang terasa lebih terarah.</h1>
      <p class="intro-copy">Kelola materi, tugas, kuis, presensi, dan percakapan kelas dari satu tempat.</p>

      <div class="feature-list">
        <div class="feature-pill">
          <span class="feature-icon">✓</span>
          Materi & tugas terintegrasi
        </div>
        <div class="feature-pill">
          <span class="feature-icon">✓</span>
          Presensi & komunikasi kelas
        </div>
      </div>

      <div class="mini-stat">
        <strong>24/7</strong>
        <span>akses pembelajaran siswa</span>
      </div>
    </section>

    <section class="auth-panel">
      <div class="auth-card">
        <p class="eyebrow eyebrow-card">Selamat datang kembali</p>
        <h2>Masuk ke akunmu</h2>
        <p class="muted">Gunakan akun sekolah untuk melanjutkan.</p>

        <form class="stack" @submit.prevent="submit">
          <label class="field">
            <span>Username</span>
            <input v-model="username" autocomplete="username" placeholder="contoh: andi" required />
          </label>

          <label class="field">
            <span>Password</span>
            <input v-model="password" autocomplete="current-password" type="password" placeholder="Masukkan password" required />
          </label>

          <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

          <button class="button button-primary button-wide" :disabled="loading" type="submit">
            {{ loading ? 'Memeriksa...' : 'Masuk ke dashboard' }}
           
          </button>
        </form>
      </div>
    </section>
  </main>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 1.15fr 0.85fr;
  background:
    radial-gradient(circle at top left, rgba(101, 201, 171, 0.18), transparent 30%),
    linear-gradient(135deg, #f6fbfa 0%, #f2f5f7 100%);
}

.auth-intro {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 72px clamp(32px, 5vw, 96px);
  position: relative;
}

.auth-intro::after {
  content: '';
  position: absolute;
  inset: 10% 8% 10% auto;
  width: min(32vw, 360px);
  border-radius: 28px;
  background: linear-gradient(135deg, rgba(101, 201, 171, 0.15), rgba(255, 90, 95, 0.08));
  filter: blur(8px);
  z-index: 0;
}

.brand-row,
.eyebrow,
.auth-card,
.feature-list,
.mini-stat {
  position: relative;
  z-index: 1;
}

.brand-row {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  width: fit-content;
  margin-bottom: 24px;
}

.login-logo {
  width: 46px;
  height: 46px;
  object-fit: contain;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid rgba(101, 201, 171, 0.2);
  box-shadow: 0 10px 22px rgba(101, 201, 171, 0.12);
}

.brand-name {
  font-size: 0.95rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--ink);
}

.eyebrow {
  margin: 0 0 18px;
  font-size: 0.76rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--teal-dark);
}

.eyebrow-card {
  color: var(--teal-dark);
}

.auth-intro h1 {
  position: relative;
  z-index: 1;
  margin: 0;
  max-width: 540px;
  font-size: clamp(2.5rem, 4vw, 4.25rem);
  line-height: 1.04;
  letter-spacing: -0.06em;
}

.intro-copy {
  position: relative;
  z-index: 1;
  max-width: 520px;
  margin: 18px 0 30px;
  color: #5d6975;
  font-size: 1.08rem;
  line-height: 1.7;
}

.feature-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 28px;
}

.feature-pill {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 12px 18px;
  border: 1px solid rgba(101, 201, 171, 0.18);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(6px);
  color: #364150;
  font-size: 0.92rem;
  font-weight: 700;
  box-shadow: 0 10px 30px rgba(59, 130, 246, 0.05);
}

.feature-icon {
  width: 24px;
  height: 24px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: var(--teal);
  color: white;
  font-size: 0.78rem;
  font-weight: 800;
}

.mini-stat {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  width: fit-content;
  padding: 14px 18px;
  border-radius: 18px;
  background: #ffffff;
  border: 1px solid var(--line);
  box-shadow: 0 16px 36px rgba(31, 41, 55, 0.08);
}

.mini-stat strong {
  display: block;
  font-size: 1.35rem;
  color: var(--ink);
  font-family: 'Poppins', sans-serif;
}

.mini-stat span {
  color: #66737e;
  font-size: 0.88rem;
  font-weight: 700;
}

.auth-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px;
}

.auth-card {
  width: min(100%, 460px);
  padding: clamp(26px, 3vw, 38px);
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(234, 237, 241, 0.95);
  box-shadow: 0 28px 70px rgba(15, 23, 42, 0.08);
  backdrop-filter: blur(12px);
}

.auth-card h2 {
  margin: 0;
  font-size: clamp(1.8rem, 2vw, 2.5rem);
  line-height: 1.1;
  letter-spacing: -0.04em;
}

.muted {
  margin: 12px 0 24px;
  color: #66737e;
  font-size: 0.98rem;
}

.stack {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  color: var(--ink);
  font-size: 0.88rem;
  font-weight: 700;
}

.field input {
  width: 100%;
  border: 1px solid var(--line);
  border-radius: 14px;
  background: #f7f9fa;
  padding: 14px 16px;
  font-size: 0.98rem;
  color: var(--ink);
  transition: border-color 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.field input::placeholder {
  color: #a1aeb9;
}

.field input:focus {
  outline: none;
  border-color: rgba(101, 201, 171, 0.7);
  background: white;
  box-shadow: 0 0 0 4px rgba(101, 201, 171, 0.12);
}

.form-error {
  margin: -4px 0 0;
  padding: 10px 12px;
  border-radius: 12px;
  background: #fff1f1;
  border: 1px solid rgba(255, 90, 95, 0.2);
  color: #c54f52;
  font-size: 0.86rem;
  font-weight: 700;
}

.button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border: none;
  border-radius: 14px;
  padding: 15px 20px;
  font-size: 0.96rem;
  font-weight: 800;
  cursor: pointer;
  transition: transform 0.18s ease, box-shadow 0.2s ease, opacity 0.2s ease;
}

.button:hover:not(:disabled) {
  transform: translateY(-1px);
}

.button-primary {
  background: linear-gradient(135deg, var(--teal) 0%, #53b398 100%);
  color: white;
  box-shadow: 0 16px 28px rgba(101, 201, 171, 0.24);
}

.button-primary:hover:not(:disabled) {
  box-shadow: 0 18px 32px rgba(101, 201, 171, 0.3);
}

.button-wide {
  width: 100%;
}

.cta-arrow {
  font-size: 1.1rem;
  line-height: 1;
}

@media (max-width: 980px) {
  .auth-page {
    grid-template-columns: 1fr;
  }

  .auth-intro {
    padding-bottom: 24px;
  }

  .auth-panel {
    padding-top: 0;
  }
}

@media (max-width: 640px) {
  .auth-intro {
    padding: 48px 22px 20px;
  }

  .auth-panel {
    padding: 0 18px 28px;
  }

  .auth-card {
    padding: 22px 18px;
    border-radius: 22px;
  }

  .feature-list {
    gap: 10px;
  }

  .feature-pill {
    width: 100%;
    justify-content: center;
  }
}
</style>
