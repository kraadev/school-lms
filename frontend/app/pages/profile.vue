<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import {
  User as UserIcon,
  Mail,
  Phone,
  Shield,
  Lock,
  CheckCircle2,
  AlertCircle,
  BookOpen,
  GraduationCap,
  Save,
  RotateCcw,
  Sparkles,
  Info
} from 'lucide-vue-next'
import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'

definePageMeta({
  middleware: 'auth',
})

const auth = useAuthStore()
const { request } = useApi()

const loading = ref(true)
const saving = ref(false)
const profileData = ref<any>(null)
const statusMessage = ref<{ type: 'success' | 'error'; text: string } | null>(null)

// Editable form state
const form = reactive({
  email: '',
  phone: '',
  new_password: '',
  confirm_password: '',
})

const roleName = computed(() => {
  const r = (profileData.value?.role?.name || auth.user?.role?.name || '').toLowerCase()
  if (r === 'admin') return 'Administrator'
  if (r === 'teacher') return 'Guru / Pengajar'
  if (r === 'student') return 'Siswa'
  return r.toUpperCase()
})

const isStudent = computed(() => {
  const r = (profileData.value?.role?.name || auth.user?.role?.name || '').toLowerCase()
  return r === 'student' || Boolean(profileData.value?.student)
})

const isTeacher = computed(() => {
  const r = (profileData.value?.role?.name || auth.user?.role?.name || '').toLowerCase()
  return r === 'teacher' || Boolean(profileData.value?.teacher)
})

const userInitials = computed(() => {
  const name = profileData.value?.student?.full_name || profileData.value?.teacher?.full_name || profileData.value?.display_name || profileData.value?.username || auth.user?.username || 'U'
  return name.trim().charAt(0).toUpperCase()
})

const enrolledClasses = computed(() => {
  return profileData.value?.student?.classes || []
})

const taughtSubjects = computed(() => {
  return profileData.value?.teacher?.class_subjects || []
})

const fetchProfile = async () => {
  loading.value = true
  statusMessage.value = null
  try {
    const res = await request<any>('/profile', { method: 'GET' })
    const data = res?.data || res
    profileData.value = data

    // Populate editable fields
    const s = data?.student
    const t = data?.teacher
    form.email = data?.email || ''
    form.phone = s?.phone || t?.phone || ''
    form.new_password = ''
    form.confirm_password = ''

    // Synchronize auth store display name and details
    auth.updateUser(data)
  } catch (err: any) {
    statusMessage.value = {
      type: 'error',
      text: err?.response?.data?.message || err?.message || 'Gagal memuat profil pengguna.',
    }
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  if (!profileData.value) return
  const data = profileData.value
  const s = data?.student
  const t = data?.teacher
  form.email = data?.email || ''
  form.phone = s?.phone || t?.phone || ''
  form.new_password = ''
  form.confirm_password = ''
  statusMessage.value = null
}

const handleSave = async () => {
  if (!form.email.trim()) {
    statusMessage.value = { type: 'error', text: 'Email wajib diisi.' }
    return
  }
  if (form.new_password && form.new_password.length < 6) {
    statusMessage.value = { type: 'error', text: 'Password baru minimal 6 karakter.' }
    return
  }
  if (form.new_password !== form.confirm_password) {
    statusMessage.value = { type: 'error', text: 'Konfirmasi password tidak cocok.' }
    return
  }

  saving.value = true
  statusMessage.value = null

  try {
    const payload: any = {
      email: form.email.trim(),
      phone: form.phone.trim(),
    }
    if (form.new_password) {
      payload.new_password = form.new_password
      payload.confirm_password = form.confirm_password
    }

    const res = await request<any>('/profile', {
      method: 'PUT',
      data: payload,
    })

    const updated = res?.data || res
    profileData.value = updated
    auth.updateUser(updated)

    statusMessage.value = {
      type: 'success',
      text: 'Profil berhasil diperbarui dengan aman!',
    }
  } catch (err: any) {
    statusMessage.value = {
      type: 'error',
      text: err?.response?.data?.message || err?.message || 'Gagal memperbarui profil.',
    }
  } finally {
    saving.value = false
  }
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleDateString('id-ID', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  } catch {
    return dateStr
  }
}

onMounted(() => {
  fetchProfile()
})
</script>

<template>
  <AppShell active="profile" eyebrow="Manajemen Akun" title="Profil Pengguna">
    <div class="profile-container">
      <!-- Loading State -->
      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <p>Memuat informasi akun Anda...</p>
      </div>

      <div v-else class="profile-layout">
        <!-- Hero Profile Banner -->
        <div class="profile-hero-card">
          <div class="hero-bg-pattern"></div>
          <div class="hero-content">
            <div class="hero-avatar">
              <span class="avatar-letter">{{ userInitials }}</span>
              <span class="online-indicator" title="Akun Aktif"></span>
            </div>

            <div class="hero-text">
              <div class="hero-name-row">
                <h2 class="hero-name">{{ profileData?.student?.full_name || profileData?.teacher?.full_name || profileData?.display_name || profileData?.username }}</h2>
                <span class="role-badge" :class="profileData?.role?.name">
                  <Shield :size="14" class="badge-icon" />
                  {{ roleName }}
                </span>
                <span class="status-badge" :class="{ active: profileData?.is_active }">
                  {{ profileData?.is_active ? 'Akun Aktif' : 'Nonaktif' }}
                </span>
              </div>
              <p class="hero-handle">@{{ profileData?.username }}</p>
              <div class="hero-meta">
                <span class="meta-item">
                  <Mail :size="14" />
                  {{ profileData?.email || 'Belum ada email' }}
                </span>
                <span v-if="form.phone" class="meta-item">
                  <Phone :size="14" />
                  {{ form.phone }}
                </span>
                <span v-if="profileData?.created_at" class="meta-item">
                  <Sparkles :size="14" />
                  Bergabung sejak {{ formatDate(profileData?.created_at) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Feedback Alert Toast -->
        <transition name="fade">
          <div v-if="statusMessage" class="alert-box" :class="statusMessage.type">
            <CheckCircle2 v-if="statusMessage.type === 'success'" :size="20" class="alert-icon" />
            <AlertCircle v-else :size="20" class="alert-icon" />
            <div class="alert-text">{{ statusMessage.text }}</div>
            <button class="alert-dismiss" @click="statusMessage = null">&times;</button>
          </div>
        </transition>

        <!-- Main Form & Info Grid -->
        <div class="profile-grid">
          <!-- Column 1: Editable Personal & Contact Info -->
          <div class="profile-card">
            <div class="card-header">
              <div class="header-icon-wrap edit-icon-wrap">
                <UserIcon :size="20" />
              </div>
              <div>
                <h3 class="card-title">Kontak & Keamanan</h3>
                <p class="card-subtitle">Hanya email, nomor telepon, dan password yang dapat diubah</p>
              </div>
            </div>

            <form @submit.prevent="handleSave" class="profile-form">
              <div class="form-group">
                <label for="email" class="form-label">
                  Alamat Email <span class="required">*</span>
                </label>
                <div class="input-with-icon">
                  <Mail :size="18" class="input-icon" />
                  <input
                    id="email"
                    v-model="form.email"
                    type="email"
                    class="form-input"
                    placeholder="contoh@sekolah.sch.id"
                    required
                  />
                </div>
                <small class="field-hint">Digunakan untuk menerima notifikasi dan pemulihan akun</small>
              </div>

              <div class="form-group">
                <label for="phone" class="form-label">Nomor Telepon / WhatsApp</label>
                <div class="input-with-icon">
                  <Phone :size="18" class="input-icon" />
                  <input
                    id="phone"
                    v-model="form.phone"
                    type="tel"
                    class="form-input"
                    placeholder="08xxxxxxxxxx"
                  />
                </div>
              </div>

              <div class="form-group">
                <label for="new_password" class="form-label">Password Baru</label>
                <div class="input-with-icon">
                  <Lock :size="18" class="input-icon" />
                  <input
                    id="new_password"
                    v-model="form.new_password"
                    type="password"
                    minlength="6"
                    class="form-input"
                    placeholder="Kosongkan jika tidak diubah"
                  />
                </div>
                <small class="field-hint">Minimal 6 karakter. Kosongkan jika password tetap.</small>
              </div>

              <div class="form-group">
                <label for="confirm_password" class="form-label">Konfirmasi Password Baru</label>
                <div class="input-with-icon">
                  <Lock :size="18" class="input-icon" />
                  <input
                    id="confirm_password"
                    v-model="form.confirm_password"
                    type="password"
                    minlength="6"
                    class="form-input"
                    placeholder="Ulangi password baru"
                  />
                </div>
              </div>

              <!-- Form Actions -->
              <div class="form-actions">
                <button
                  type="button"
                  class="btn-secondary"
                  :disabled="saving"
                  @click="resetForm"
                >
                  <RotateCcw :size="16" />
                  Reset
                </button>
                <button
                  type="submit"
                  class="btn-primary"
                  :disabled="saving"
                >
                  <span v-if="saving" class="btn-spinner"></span>
                  <Save v-else :size="16" />
                  {{ saving ? 'Menyimpan...' : 'Simpan Perubahan' }}
                </button>
              </div>
            </form>
          </div>

          <!-- Column 2: Account Credentials & Academic Info (Read-Only) -->
          <div class="profile-side-column">
            <!-- Security & Administrative Info Card (Strictly Read-Only) -->
            <div class="profile-card readonly-card">
              <div class="card-header">
                <div class="header-icon-wrap lock-icon-wrap">
                  <Lock :size="20" />
                </div>
                <div>
                  <h3 class="card-title">Kredensial & Akun</h3>
                  <p class="card-subtitle">Data resmi terdaftar di pangkalan data sekolah</p>
                </div>
              </div>

              <div class="security-notice">
                <Info :size="18" class="notice-icon" />
                <div class="notice-text">
                  <strong>Dikelola oleh Administrator</strong>
                  <p>
                    Username, role akses, status akun, dan nomor induk resmi dilindungi dan hanya dapat dikelola oleh administrator sekolah.
                  </p>
                </div>
              </div>

              <div class="readonly-grid">
                <div class="readonly-item">
                  <span class="ro-label">Username</span>
                  <div class="ro-value-wrapper">
                    <span class="ro-value font-mono">@{{ profileData?.username }}</span>
                    <span class="ro-lock" title="Hanya admin yang bisa mengubah username"><Lock :size="14" /></span>
                  </div>
                </div>

                <div class="readonly-item">
                  <span class="ro-label">Peran Pengguna (Role)</span>
                  <div class="ro-value-wrapper">
                    <span class="role-pill" :class="profileData?.role?.name">{{ roleName }}</span>
                    <span class="ro-lock"><Lock :size="14" /></span>
                  </div>
                </div>

                <div class="readonly-item">
                  <span class="ro-label">Status Akun</span>
                  <div class="ro-value-wrapper">
                    <span class="status-pill active">
                      <span class="dot"></span>
                      {{ profileData?.is_active ? 'Aktif' : 'Nonaktif' }}
                    </span>
                    <span class="ro-lock"><Lock :size="14" /></span>
                  </div>
                </div>

                <!-- NIS / NISN for Student -->
                <template v-if="isStudent">
                  <div class="readonly-item">
                    <span class="ro-label">Nomor Induk Siswa (NIS)</span>
                    <div class="ro-value-wrapper">
                      <span class="ro-value">{{ profileData?.student?.nis || '-' }}</span>
                      <span class="ro-lock"><Lock :size="14" /></span>
                    </div>
                  </div>

                  <div class="readonly-item">
                    <span class="ro-label">NISN</span>
                    <div class="ro-value-wrapper">
                      <span class="ro-value">{{ profileData?.student?.nisn || '-' }}</span>
                      <span class="ro-lock"><Lock :size="14" /></span>
                    </div>
                  </div>
                </template>

                <!-- NIP for Teacher -->
                <template v-if="isTeacher">
                  <div class="readonly-item">
                    <span class="ro-label">Nomor Induk Pegawai (NIP)</span>
                    <div class="ro-value-wrapper">
                      <span class="ro-value">{{ profileData?.teacher?.nip || '-' }}</span>
                      <span class="ro-lock"><Lock :size="14" /></span>
                    </div>
                  </div>
                </template>
              </div>
            </div>

            <!-- Academic Assignment Card (Read-Only) -->
            <div class="profile-card academic-card">
              <div class="card-header">
                <div class="header-icon-wrap academic-icon-wrap">
                  <GraduationCap v-if="isStudent" :size="20" />
                  <BookOpen v-else :size="20" />
                </div>
                <div>
                  <h3 class="card-title">{{ isStudent ? 'Kelas Terdaftar' : isTeacher ? 'Mata Pelajaran Diampu' : 'Hak Akses Sistem' }}</h3>
                  <p class="card-subtitle">
                    {{ isStudent ? 'Daftar kelas aktif tahun ajaran berjalan' : isTeacher ? 'Penugasan mengajar Anda' : 'Akses penuh administrasi sekolah' }}
                  </p>
                </div>
              </div>

              <!-- Student Classes List -->
              <div v-if="isStudent" class="academic-content">
                <div v-if="enrolledClasses.length === 0" class="empty-state-card">
                  <BookOpen :size="28" class="empty-icon" />
                  <p>Belum ada kelas yang terhubung dengan akun Anda.</p>
                </div>
                <div v-else class="classes-list">
                  <div v-for="cls in enrolledClasses" :key="cls.id" class="class-badge-item">
                    <div class="class-icon-pill">
                      <GraduationCap :size="16" />
                    </div>
                    <div class="class-info">
                      <span class="class-name">{{ cls.name }}</span>
                      <span v-if="cls.grade" class="class-grade">Tingkat {{ cls.grade }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Teacher Subjects List -->
              <div v-else-if="isTeacher" class="academic-content">
                <div v-if="taughtSubjects.length === 0" class="empty-state-card">
                  <BookOpen :size="28" class="empty-icon" />
                  <p>Belum ada penugasan mata pelajaran yang terdaftar.</p>
                </div>
                <div v-else class="classes-list">
                  <div v-for="item in taughtSubjects" :key="item.id" class="class-badge-item">
                    <div class="class-icon-pill teacher-pill">
                      <BookOpen :size="16" />
                    </div>
                    <div class="class-info">
                      <span class="class-name">{{ item.subject?.name || 'Mata Pelajaran' }}</span>
                      <span class="class-grade">Kelas {{ item.class?.name || '-' }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Admin Info -->
              <div v-else class="academic-content">
                <div class="admin-access-badge">
                  <Shield :size="20" />
                  <div>
                    <strong>Super Administrator</strong>
                    <p>Memiliki akses penuh untuk mengelola pengguna, kelas, kurikulum, dan konfigurasi sistem.</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppShell>
</template>

<style scoped>
.profile-container {
  max-width: 1180px;
  margin: 0 auto;
  padding: 8px 0 40px;
}

/* Loading State */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  color: var(--muted);
  font-size: 14px;
}
.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid #e0ebe3;
  border-top-color: var(--teal);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 16px;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Hero Profile Banner */
.profile-hero-card {
  position: relative;
  background: linear-gradient(135deg, #145942 0%, #1f7a5c 50%, #2e8b6b 100%);
  border-radius: 20px;
  padding: 28px 32px;
  color: white;
  margin-bottom: 24px;
  overflow: hidden;
  box-shadow: 0 10px 30px rgba(20, 89, 66, 0.18);
}
.hero-bg-pattern {
  position: absolute;
  top: -50%;
  right: -10%;
  width: 380px;
  height: 380px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.12) 0%, rgba(255, 255, 255, 0) 70%);
  pointer-events: none;
}
.hero-content {
  position: relative;
  display: flex;
  align-items: center;
  gap: 24px;
  z-index: 1;
}
.hero-avatar {
  position: relative;
  width: 76px;
  height: 76px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  border: 3px solid rgba(255, 255, 255, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  backdrop-filter: blur(4px);
}
.avatar-letter {
  font-size: 32px;
  font-weight: 800;
  color: white;
  text-shadow: 0 2px 4px rgba(0,0,0,0.15);
}
.online-indicator {
  position: absolute;
  bottom: 2px;
  right: 2px;
  width: 16px;
  height: 16px;
  background: #10b981;
  border: 2px solid white;
  border-radius: 50%;
}
.hero-text {
  flex: 1;
  min-width: 0;
}
.hero-name-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.hero-name {
  font-size: 24px;
  font-weight: 700;
  color: white !important;
  margin: 0;
  line-height: 1.2;
}
.role-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.3);
}
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  background: rgba(16, 185, 129, 0.25);
  color: #d1fae5;
  border: 1px solid rgba(16, 185, 129, 0.4);
}
.hero-handle {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.82);
  margin: 4px 0 10px;
  font-family: monospace;
}
.hero-meta {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}
.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.9);
}

/* Alert Toast */
.alert-box {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 18px;
  border-radius: 12px;
  margin-bottom: 24px;
  font-size: 13px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.04);
}
.alert-box.success {
  background: #ecfdf5;
  border: 1px solid #a7f3d0;
  color: #065f46;
}
.alert-box.error {
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #991b1b;
}
.alert-icon {
  flex-shrink: 0;
}
.alert-text {
  flex: 1;
  font-weight: 500;
}
.alert-dismiss {
  border: none;
  background: transparent;
  font-size: 20px;
  cursor: pointer;
  color: inherit;
  line-height: 1;
  padding: 0 4px;
}

/* Main Grid Layout */
.profile-grid {
  display: grid;
  grid-template-columns: 1.15fr 0.85fr;
  gap: 24px;
  align-items: start;
}
@media (max-width: 900px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }
}

/* Profile Card Styling */
.profile-card {
  background: white;
  border-radius: 16px;
  border: 1px solid #eef1ef;
  box-shadow: 0 2px 10px rgba(0,0,0,0.02);
  padding: 24px;
  margin-bottom: 24px;
}
.card-header {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f2f5f3;
}
.header-icon-wrap {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.edit-icon-wrap {
  background: #eef7f2;
  color: var(--teal);
}
.lock-icon-wrap {
  background: #eff6ff;
  color: #2563eb;
}
.academic-icon-wrap {
  background: #fef3c7;
  color: #d97706;
}
.card-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--ink);
  margin: 0;
}
.card-subtitle {
  font-size: 12px;
  color: var(--muted);
  margin: 2px 0 0;
}

/* Forms */
.profile-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.form-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink);
}
.required {
  color: #ef4444;
}
.input-with-icon {
  position: relative;
  display: flex;
  align-items: center;
}
.input-with-icon.icon-top {
  align-items: flex-start;
}
.input-with-icon.icon-top .input-icon {
  margin-top: 12px;
}
.input-icon {
  position: absolute;
  left: 14px;
  color: #94a3b8;
  pointer-events: none;
}
.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 10px 14px 10px 42px;
  border: 1px solid #dcdfe3;
  border-radius: 10px;
  font-size: 13.5px;
  color: var(--ink);
  background: #fafbfb;
  transition: all 0.2s;
  font-family: inherit;
}
.form-select {
  padding-left: 14px;
  cursor: pointer;
}
.form-textarea {
  resize: vertical;
}
.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  background: white;
  border-color: var(--teal);
  box-shadow: 0 0 0 3px rgba(20, 89, 66, 0.12);
}
.field-hint {
  font-size: 11px;
  color: #64748b;
  margin-top: 2px;
}

/* Buttons */
.form-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 10px;
  padding-top: 16px;
  border-top: 1px solid #f2f5f3;
}
.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: 10px;
  background: var(--teal);
  color: white;
  font-size: 13.5px;
  font-weight: 600;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-primary:hover:not(:disabled) {
  background: #0f4633;
  box-shadow: 0 4px 12px rgba(20, 89, 66, 0.25);
}
.btn-primary:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}
.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: 10px;
  background: #f1f5f3;
  color: var(--ink);
  font-size: 13.5px;
  font-weight: 600;
  border: 1px solid #e2e8e4;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-secondary:hover:not(:disabled) {
  background: #e5ece8;
}
.btn-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255,255,255,0.4);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

/* Security Notice */
.security-notice {
  display: flex;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 10px;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  color: #166534;
  font-size: 12px;
  margin-bottom: 18px;
}
.notice-icon {
  flex-shrink: 0;
  color: #15803d;
  margin-top: 1px;
}
.notice-text strong {
  display: block;
  font-size: 12.5px;
  margin-bottom: 2px;
}
.notice-text p {
  margin: 0;
  color: #1e3a2b;
  line-height: 1.4;
}

/* Readonly Grid */
.readonly-grid {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.readonly-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-radius: 10px;
  background: #f9fafb;
  border: 1px solid #f1f3f5;
}
.ro-label {
  font-size: 12.5px;
  font-weight: 500;
  color: #64748b;
}
.ro-value-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}
.ro-value {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink);
}
.ro-value.font-mono {
  font-family: monospace;
}
.ro-lock {
  color: #94a3b8;
  display: flex;
  align-items: center;
}
.role-pill {
  font-size: 11px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 999px;
  background: #dbeafe;
  color: #1e40af;
}
.role-pill.admin {
  background: #fee2e2;
  color: #991b1b;
}
.role-pill.teacher {
  background: #fef3c7;
  color: #92400e;
}
.role-pill.student {
  background: #dcfce7;
  color: #166534;
}
.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 999px;
  background: #dcfce7;
  color: #15803d;
}
.status-pill .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #16a34a;
}

/* Academic Card & Classes List */
.academic-content {
  margin-top: 10px;
}
.classes-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.class-badge-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
}
.class-icon-pill {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: #e0f2fe;
  color: #0284c7;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.class-icon-pill.teacher-pill {
  background: #fef3c7;
  color: #d97706;
}
.class-info {
  display: flex;
  flex-direction: column;
}
.class-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink);
}
.class-grade {
  font-size: 11px;
  color: #64748b;
}
.empty-state-card {
  text-align: center;
  padding: 30px 16px;
  color: var(--muted);
  font-size: 12.5px;
}
.empty-icon {
  color: #cbd5e1;
  margin-bottom: 8px;
}
.admin-access-badge {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px;
  border-radius: 12px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  color: #1e40af;
  font-size: 12.5px;
}
.admin-access-badge strong {
  display: block;
  margin-bottom: 2px;
}
.admin-access-badge p {
  margin: 0;
  font-size: 11.5px;
  color: #3b82f6;
}

/* Fade animation */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
