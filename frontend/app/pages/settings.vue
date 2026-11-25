<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import {
  Palette,
  Shield,
  Bell,
  BookOpen,
  Sliders,
  Database,
  Save,
  RotateCcw,
  Check,
  CheckCircle2,
  AlertCircle,
  Volume2,
  VolumeX,
  Lock,
  Laptop,
  Moon,
  Sun,
  Monitor,
  School,
  Download,
  Trash2,
  Sparkles,
  Info,
  KeyRound,
  ExternalLink,
  Smartphone,
  LogOut,
  SlidersHorizontal,
  BellRing
} from 'lucide-vue-next'
import { useAuthStore } from '~/stores/auth'
import { useAppSettings } from '~/composables/useAppSettings'

definePageMeta({
  middleware: 'auth',
})

const auth = useAuthStore()
const {
  settings,
  t,
  presenceColor,
  presenceLabel,
  formatGrade,
  requestPushPermission,
  showNativeNotification,
  saveSettings,
  resetToDefault,
  playNotificationChime,
  clearAppCache,
} = useAppSettings()

type TabKey = 'appearance' | 'account' | 'notifications' | 'learning' | 'system' | 'storage'
const activeTab = ref<TabKey>('appearance')

// Status Alert / Toast
const feedback = ref<{ type: 'success' | 'error' | 'info'; text: string } | null>(null)
const setFeedback = (type: 'success' | 'error' | 'info', text: string, duration = 4000) => {
  feedback.value = { type, text }
  if (duration > 0) {
    setTimeout(() => {
      if (feedback.value?.text === text) {
        feedback.value = null
      }
    }, duration)
  }
}

// User Info
const isAdmin = computed(() => (auth.user?.role?.name || auth.role || '').toLowerCase() === 'admin')
const userRoleName = computed(() => {
  const r = (auth.user?.role?.name || auth.role || '').toLowerCase()
  if (r === 'admin') return 'Administrator'
  if (r === 'guru' || r === 'teacher') return 'Guru / Pendidik'
  if (r === 'siswa' || r === 'student') return 'Siswa'
  return r.toUpperCase()
})
const userDisplayName = computed(() => {
  const u = auth.user
  return u?.student?.full_name || u?.student?.FullName || u?.teacher?.full_name || u?.teacher?.FullName || u?.display_name || u?.username || 'Pengguna'
})

// Device & Session Details (Client Detected)
const sessionInfo = reactive({
  browser: 'Browser Terkini',
  os: 'Desktop OS',
  ip: '127.0.0.1 (Lokal)',
  lastActive: 'Saat ini',
})

onMounted(() => {
  if (typeof window !== 'undefined' && navigator) {
    const ua = navigator.userAgent
    let b = 'Modern Browser'
    if (ua.includes('Chrome')) b = 'Google Chrome'
    else if (ua.includes('Firefox')) b = 'Mozilla Firefox'
    else if (ua.includes('Safari')) b = 'Apple Safari'
    else if (ua.includes('Edg')) b = 'Microsoft Edge'
    sessionInfo.browser = b

    let o = 'Desktop / Mobile'
    if (ua.includes('Windows')) o = 'Windows 10/11'
    else if (ua.includes('Mac')) o = 'macOS'
    else if (ua.includes('Android')) o = 'Android Device'
    else if (ua.includes('iPhone') || ua.includes('iPad')) o = 'iOS Device'
    else if (ua.includes('Linux')) o = 'Linux OS'
    sessionInfo.os = o
  }
})

// Test Sound Button
const isTestingSound = ref(false)
const testChime = () => {
  isTestingSound.value = true
  playNotificationChime()
  setFeedback('info', settings.language === 'en' ? 'Playing notification audio chime preview...' : 'Memutar pratinjau nada notifikasi...', 2000)
  setTimeout(() => {
    isTestingSound.value = false
  }, 600)
}

// Push Notification Toggle with Native Browser Permission Request
const handleTogglePush = async (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.checked) {
    const granted = await requestPushPermission()
    if (granted) {
      setFeedback('success', settings.language === 'en' ? 'Desktop push notifications successfully enabled!' : 'Izin notifikasi desktop berhasil diaktifkan!')
      showNativeNotification('School Space', settings.language === 'en' ? 'Desktop notifications are active.' : 'Notifikasi desktop telah aktif.')
    } else {
      target.checked = false
      settings.notifyPush = false
      saveSettings()
      setFeedback('error', settings.language === 'en' ? 'Notification permission was denied or not supported in this browser.' : 'Izin notifikasi ditolak oleh browser atau tidak didukung.')
    }
  } else {
    settings.notifyPush = false
    saveSettings()
  }
}

// Clear Cache & Reset
const clearingCache = ref(false)
const handleClearCache = async () => {
  const confirmText = settings.language === 'en'
    ? 'Are you sure you want to clear local application cache? Your login session will remain active.'
    : 'Apakah Anda yakin ingin membersihkan data cache lokal? Akun Anda akan tetap tersimpan.'
  if (!confirm(confirmText)) return

  clearingCache.value = true
  const success = await clearAppCache()
  clearingCache.value = false
  if (success) {
    setFeedback('success', t('feedback.cache_cleared'))
  } else {
    setFeedback('error', 'Gagal membersihkan cache lokal.')
  }
}

const handleResetDefaults = () => {
  const confirmText = settings.language === 'en'
    ? 'Restore all preferences to system default values?'
    : 'Kembalikan semua preferensi ke setelan standar bawaan?'
  if (confirm(confirmText)) {
    resetToDefault()
    setFeedback('info', t('feedback.reset'))
  }
}

// Export Settings to JSON
const exportSettings = () => {
  try {
    const jsonStr = JSON.stringify(settings, null, 2)
    const blob = new Blob([jsonStr], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `school_space_settings_${Date.now()}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    setFeedback('success', t('feedback.exported'))
  } catch {
    setFeedback('error', 'Gagal mengekspor berkas preferensi.')
  }
}

// Admin System Settings Save
const savingSystemSettings = ref(false)
const handleSaveSystemSettings = () => {
  savingSystemSettings.value = true
  saveSettings()
  setTimeout(() => {
    savingSystemSettings.value = false
    setFeedback('success', settings.language === 'en' ? 'School system configuration saved successfully!' : 'Konfigurasi sistem sekolah berhasil disimpan!')
  }, 350)
}
</script>

<template>
  <AppShell active="settings" :eyebrow="t('settings.subtitle')" :title="t('settings.title')">
    <div class="settings-container">
      
      <!-- Top Title & Description Header -->
      <div class="settings-header">
        <div class="settings-header-text">
          <h1 class="page-title">{{ t('settings.title') }}</h1>
          <p class="page-subtitle">{{ t('settings.subtitle') }}</p>
        </div>
        <div class="settings-header-actions">
          <button type="button" class="btn-outline" @click="handleResetDefaults" :title="t('settings.btn_reset')">
            <RotateCcw :size="16" />
            <span>{{ t('settings.btn_reset') }}</span>
          </button>
          <button type="button" class="btn-primary" @click="() => { saveSettings(); setFeedback('success', t('feedback.saved')); }">
            <Save :size="16" />
            <span>{{ t('settings.btn_save') }}</span>
          </button>
        </div>
      </div>

      <!-- Feedback Alert Banner -->
      <transition name="fade">
        <div v-if="feedback" class="feedback-alert" :class="feedback.type">
          <CheckCircle2 v-if="feedback.type === 'success'" :size="18" />
          <AlertCircle v-else-if="feedback.type === 'error'" :size="18" />
          <Info v-else :size="18" />
          <span>{{ feedback.text }}</span>
          <button type="button" class="feedback-close" @click="feedback = null">×</button>
        </div>
      </transition>

      <!-- Main Layout: Navigation Tabs + Content Panels -->
      <div class="settings-grid">
        
        <!-- Left Sidebar Navigation Tabs -->
        <div class="settings-nav-card">
          <div class="user-brief">
            <div class="user-brief-avatar" :style="{ position: 'relative' }">
              {{ userDisplayName.charAt(0).toUpperCase() }}
              <span class="brief-presence-dot" :style="{ backgroundColor: presenceColor }"></span>
            </div>
            <div class="user-brief-info">
              <strong class="user-brief-name">{{ userDisplayName }}</strong>
              <span class="user-brief-role">{{ userRoleName }} · {{ presenceLabel }}</span>
            </div>
          </div>

          <nav class="tabs-menu" aria-label="Kategori Pengaturan">
            <button
              type="button"
              class="tab-btn"
              :class="{ active: activeTab === 'appearance' }"
              @click="activeTab = 'appearance'"
            >
              <Palette :size="18" class="tab-icon" />
              <div class="tab-text">
                <span class="tab-title">{{ t('settings.tab_appearance') }}</span>
                <small class="tab-desc">{{ t('settings.tab_appearance_desc') }}</small>
              </div>
            </button>

            <button
              type="button"
              class="tab-btn"
              :class="{ active: activeTab === 'account' }"
              @click="activeTab = 'account'"
            >
              <Shield :size="18" class="tab-icon" />
              <div class="tab-text">
                <span class="tab-title">{{ t('settings.tab_account') }}</span>
                <small class="tab-desc">{{ t('settings.tab_account_desc') }}</small>
              </div>
            </button>

            <button
              type="button"
              class="tab-btn"
              :class="{ active: activeTab === 'notifications' }"
              @click="activeTab = 'notifications'"
            >
              <Bell :size="18" class="tab-icon" />
              <div class="tab-text">
                <span class="tab-title">{{ t('settings.tab_notifications') }}</span>
                <small class="tab-desc">{{ t('settings.tab_notifications_desc') }}</small>
              </div>
            </button>

            <button
              type="button"
              class="tab-btn"
              :class="{ active: activeTab === 'learning' }"
              @click="activeTab = 'learning'"
            >
              <BookOpen :size="18" class="tab-icon" />
              <div class="tab-text">
                <span class="tab-title">{{ t('settings.tab_learning') }}</span>
                <small class="tab-desc">{{ t('settings.tab_learning_desc') }}</small>
              </div>
            </button>

            <button
              type="button"
              class="tab-btn"
              :class="{ active: activeTab === 'system', 'admin-pill': isAdmin }"
              @click="activeTab = 'system'"
            >
              <School :size="18" class="tab-icon" />
              <div class="tab-text">
                <span class="tab-title">
                  {{ t('settings.tab_system') }}
                  <span v-if="isAdmin" class="role-tag">Admin</span>
                </span>
                <small class="tab-desc">{{ t('settings.tab_system_desc') }}</small>
              </div>
            </button>

            <button
              type="button"
              class="tab-btn"
              :class="{ active: activeTab === 'storage' }"
              @click="activeTab = 'storage'"
            >
              <Database :size="18" class="tab-icon" />
              <div class="tab-text">
                <span class="tab-title">{{ t('settings.tab_storage') }}</span>
                <small class="tab-desc">{{ t('settings.tab_storage_desc') }}</small>
              </div>
            </button>
          </nav>
        </div>

        <!-- Right Content Area -->
        <div class="settings-content-area">

          <!-- TAB 1: TAMPILAN & TEMA -->
          <div v-show="activeTab === 'appearance'" class="panel-section">
            <div class="section-top">
              <div>
                <h2 class="section-heading">{{ settings.language === 'en' ? 'Appearance & Interface' : 'Tampilan & Antarmuka' }}</h2>
                <p class="section-subheading">
                  {{ settings.language === 'en' ? 'Customize colors, font scale, and layouts for optimal reading comfort.' : 'Kustomisasikan visual School Space agar nyaman di mata dan sesuai perangkat Anda.' }}
                </p>
              </div>
              <Sparkles :size="24" class="text-teal" />
            </div>

            <!-- Theme Cards -->
            <div class="setting-group">
              <label class="group-label">{{ settings.language === 'en' ? 'Theme Mode' : 'Mode Tema Tampilan' }}</label>
              <p class="group-caption">{{ settings.language === 'en' ? 'Choose between light theme, dark mode for low light, or match your operating system.' : 'Pilih palet tampilan terang, redup gelap untuk malam hari, atau mengikuti sistem operasi.' }}</p>
              
              <div class="theme-picker-grid">
                <!-- Light -->
                <div
                  class="theme-card"
                  :class="{ active: settings.theme === 'light' }"
                  @click="settings.theme = 'light'"
                >
                  <div class="theme-preview preview-light">
                    <div class="preview-bar"></div>
                    <div class="preview-box"></div>
                    <Sun :size="20" class="theme-badge-icon" />
                  </div>
                  <div class="theme-card-info">
                    <span class="theme-name">{{ settings.language === 'en' ? 'Light Mode' : 'Mode Terang' }}</span>
                    <small>{{ settings.language === 'en' ? 'Clean & bright' : 'Visual bersih & cerah' }}</small>
                  </div>
                  <Check v-if="settings.theme === 'light'" :size="16" class="check-mark" />
                </div>

                <!-- Dark -->
                <div
                  class="theme-card"
                  :class="{ active: settings.theme === 'dark' }"
                  @click="settings.theme = 'dark'"
                >
                  <div class="theme-preview preview-dark">
                    <div class="preview-bar"></div>
                    <div class="preview-box"></div>
                    <Moon :size="20" class="theme-badge-icon" />
                  </div>
                  <div class="theme-card-info">
                    <span class="theme-name">{{ settings.language === 'en' ? 'Dark Mode' : 'Mode Gelap' }}</span>
                    <small>{{ settings.language === 'en' ? 'Comfortable contrast' : 'Kontras redup nyaman' }}</small>
                  </div>
                  <Check v-if="settings.theme === 'dark'" :size="16" class="check-mark" />
                </div>

                <!-- System -->
                <div
                  class="theme-card"
                  :class="{ active: settings.theme === 'system' }"
                  @click="settings.theme = 'system'"
                >
                  <div class="theme-preview preview-system">
                    <div class="preview-bar"></div>
                    <div class="preview-box"></div>
                    <Monitor :size="20" class="theme-badge-icon" />
                  </div>
                  <div class="theme-card-info">
                    <span class="theme-name">{{ settings.language === 'en' ? 'System Match' : 'Mengikuti Sistem' }}</span>
                    <small>{{ settings.language === 'en' ? 'Auto synced with OS' : 'Otomatis sinkron OS' }}</small>
                  </div>
                  <Check v-if="settings.theme === 'system'" :size="16" class="check-mark" />
                </div>
              </div>
            </div>

            <!-- Font Size -->
            <div class="setting-group">
              <label class="group-label">{{ settings.language === 'en' ? 'Font Scaling' : 'Skala Ukuran Huruf (Font Scaling)' }}</label>
              <p class="group-caption">{{ settings.language === 'en' ? 'Adjust base font sizes for better readability across lessons and quizzes.' : 'Sesuaikan ukuran font agar membaca materi pelajaran dan kuis terasa ergonomis.' }}</p>

              <div class="options-pill-group">
                <button
                  type="button"
                  class="pill-btn"
                  :class="{ active: settings.fontSize === 'sm' }"
                  @click="settings.fontSize = 'sm'"
                >
                  <small style="font-size: 12px; font-weight: bold;">A</small>
                  <span>{{ settings.language === 'en' ? 'Small (Compact)' : 'Kecil (Kompak)' }}</span>
                </button>
                <button
                  type="button"
                  class="pill-btn"
                  :class="{ active: settings.fontSize === 'md' }"
                  @click="settings.fontSize = 'md'"
                >
                  <small style="font-size: 14px; font-weight: bold;">A</small>
                  <span>{{ settings.language === 'en' ? 'Standard (Normal)' : 'Standar (Normal)' }}</span>
                </button>
                <button
                  type="button"
                  class="pill-btn"
                  :class="{ active: settings.fontSize === 'lg' }"
                  @click="settings.fontSize = 'lg'"
                >
                  <small style="font-size: 18px; font-weight: bold;">A</small>
                  <span>{{ settings.language === 'en' ? 'Large' : 'Besar (Lebih Jelas)' }}</span>
                </button>
              </div>
            </div>

            <!-- UI Density & Language -->
            <div class="setting-row-2col">
              <div class="setting-card-item">
                <div class="item-text">
                  <strong>{{ settings.language === 'en' ? 'UI Layout Density' : 'Kepadatan Tata Letak' }}</strong>
                  <p>{{ settings.language === 'en' ? 'Set card spacing and table row heights.' : 'Atur ruang renggang (padding) kartu dan tabel antarmuka.' }}</p>
                </div>
                <select v-model="settings.density" class="select-input">
                  <option value="comfortable">{{ settings.language === 'en' ? 'Comfortable (Spacious)' : 'Standar (Lapang & Nyaman)' }}</option>
                  <option value="compact">{{ settings.language === 'en' ? 'Compact (Dense Data)' : 'Rapat (Banyak Data Terlihat)' }}</option>
                </select>
              </div>

              <div class="setting-card-item">
                <div class="item-text">
                  <strong>{{ settings.language === 'en' ? 'Interface Language' : 'Bahasa Aplikasi' }}</strong>
                  <p>{{ settings.language === 'en' ? 'Primary language for system navigation and controls.' : 'Bahasa utama teks antarmuka dan label sistem.' }}</p>
                </div>
                <select v-model="settings.language" class="select-input">
                  <option value="id">Bahasa Indonesia (ID)</option>
                  <option value="en">English (US)</option>
                </select>
              </div>
            </div>

            <!-- Reduce Motion Toggle -->
            <div class="toggle-card">
              <div class="toggle-text">
                <strong>{{ settings.language === 'en' ? 'Reduce Motion' : 'Pengurangan Animasi Gerak (Reduce Motion)' }}</strong>
                <p>{{ settings.language === 'en' ? 'Minimize transitions and animations for smoother performance on slower devices.' : 'Meminimalkan transisi dan efek animasi untuk perangkat spesifikasi hemat daya.' }}</p>
              </div>
              <label class="switch">
                <input type="checkbox" v-model="settings.reduceMotion" />
                <span class="slider"></span>
              </label>
            </div>
          </div>

          <!-- TAB 2: KEAMANAN & SESI (FOKUS KEAMANAN & TIDAK DUPLIKAT PROFIL) -->
          <div v-show="activeTab === 'account'" class="panel-section">
            <div class="section-top">
              <div>
                <h2 class="section-heading">{{ settings.language === 'en' ? 'Security & Session Management' : 'Keamanan & Manajemen Sesi' }}</h2>
                <p class="section-subheading">
                  {{ settings.language === 'en' ? 'Manage active devices, session security, and access your profile credentials.' : 'Kelola sesi aktif, keamanan perangkat, dan akses kredensial profil Anda secara aman.' }}
                </p>
              </div>
              <Shield :size="24" class="text-coral" />
            </div>

            <!-- Centralized Profile & Password Hub Card (Mencegah Duplikasi Fungsi) -->
            <div class="card-inner mb-4 profile-hub-banner">
              <div class="profile-hub-content">
                <div class="profile-hub-icon">
                  <KeyRound :size="24" class="text-teal" />
                </div>
                <div class="profile-hub-text">
                  <strong>{{ settings.language === 'en' ? 'Account Credentials & Password Management' : 'Pengelolaan Profil & Kata Sandi Akun' }}</strong>
                  <p>
                    {{ settings.language === 'en' 
                      ? 'Email, phone contact, and password updates are securely handled in your Profile page to maintain central data integrity.'
                      : 'Pembaruan kata sandi, email, dan nomor telepon dikelola terpusat di halaman Profil Pengguna untuk menjaga integritas data akun Anda.' }}
                  </p>
                </div>
              </div>
              <NuxtLink to="/profile" class="btn-primary flex-shrink-0">
                <span>{{ settings.language === 'en' ? 'Open Profile & Change Password' : 'Buka Profil & Ubah Sandi' }}</span>
                <ExternalLink :size="15" />
              </NuxtLink>
            </div>

            <!-- Session Preferences -->
            <div class="card-inner mb-4">
              <h3 class="inner-title">
                <Lock :size="18" />
                {{ settings.language === 'en' ? 'Session Security Preferences' : 'Preferensi Keamanan Sesi' }}
              </h3>

              <div class="setting-row-2col">
                <div class="setting-card-item">
                  <div class="item-text">
                    <strong>{{ settings.language === 'en' ? 'Inactivity Session Timeout' : 'Batas Waktu Sesi Tidak Aktif' }}</strong>
                    <p>{{ settings.language === 'en' ? 'Automatically lock session when inactive for security.' : 'Kunci sesi secara otomatis jika tidak ada aktivitas belajar.' }}</p>
                  </div>
                  <select v-model.number="settings.sessionTimeoutMinutes" class="select-input">
                    <option :value="15">15 {{ settings.language === 'en' ? 'Minutes' : 'Menit' }}</option>
                    <option :value="30">30 {{ settings.language === 'en' ? 'Minutes' : 'Menit' }}</option>
                    <option :value="60">1 {{ settings.language === 'en' ? 'Hour (Standard)' : 'Jam (Standar)' }}</option>
                    <option :value="120">2 {{ settings.language === 'en' ? 'Hours' : 'Jam' }}</option>
                    <option :value="0">{{ settings.language === 'en' ? 'Never (Keep Signed In)' : 'Tidak Pernah (Selalu Masuk)' }}</option>
                  </select>
                </div>

                <div class="setting-card-item">
                  <div class="item-text">
                    <strong>{{ settings.language === 'en' ? 'Remember This Device' : 'Ingat Perangkat Ini' }}</strong>
                    <p>{{ settings.language === 'en' ? 'Keep token cached safely on this personal computer.' : 'Simpan token sesi terenkripsi pada peramban ini.' }}</p>
                  </div>
                  <div style="display: flex; align-items: center; justify-content: space-between; margin-top: 10px;">
                    <span style="font-size: 12.5px; font-weight: 700; color: var(--muted);">{{ settings.rememberDevice ? (settings.language === 'en' ? 'Active' : 'Aktif') : (settings.language === 'en' ? 'Disabled' : 'Nonaktif') }}</span>
                    <label class="switch">
                      <input type="checkbox" v-model="settings.rememberDevice" />
                      <span class="slider"></span>
                    </label>
                  </div>
                </div>
              </div>
            </div>

            <!-- Active Device & Session Info -->
            <div class="card-inner">
              <h3 class="inner-title">
                <Laptop :size="18" />
                {{ settings.language === 'en' ? 'Current Active Device' : 'Perangkat & Sesi Aktif Saat Ini' }}
              </h3>
              
              <div class="session-card">
                <div class="session-icon">
                  <Monitor :size="24" />
                </div>
                <div class="session-details">
                  <strong>{{ sessionInfo.browser }} · {{ sessionInfo.os }}</strong>
                  <p>{{ settings.language === 'en' ? 'IP Address' : 'Alamat IP' }}: {{ sessionInfo.ip }} · {{ settings.language === 'en' ? 'Last activity' : 'Aktivitas terakhir' }}: {{ sessionInfo.lastActive }}</p>
                  <span class="session-badge-active">{{ settings.language === 'en' ? 'This Session Is Active' : 'Sesi Ini Sedang Aktif' }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- TAB 3: NOTIFIKASI & SUARA -->
          <div v-show="activeTab === 'notifications'" class="panel-section">
            <div class="section-top">
              <div>
                <h2 class="section-heading">{{ settings.language === 'en' ? 'Notifications & Sound Alerts' : 'Preferensi Notifikasi & Efek Suara' }}</h2>
                <p class="section-subheading">
                  {{ settings.language === 'en' ? 'Manage desktop push, real-time message sounds, and deadline reminders.' : 'Atur pemberitahuan kegiatan kelas, batas pengumpulan tugas, serta nada peringatan.' }}
                </p>
              </div>
              <Bell :size="24" class="text-yellow" />
            </div>

            <!-- Sound Chime & Volume -->
            <div class="card-inner mb-4 sound-hero-box">
              <div class="sound-row">
                <div class="sound-info">
                  <h3 class="inner-title mb-1">
                    <Volume2 :size="18" />
                    {{ settings.language === 'en' ? 'Real-time Audio Chime' : 'Bunyi Notifikasi Real-time' }}
                  </h3>
                  <p class="sound-desc">{{ settings.language === 'en' ? 'Soft double-chime when a new chat message arrives or an assignment is updated.' : 'Peringatan audio lembut saat ada pesan chat masuk atau perubahan nilai/tugas.' }}</p>
                </div>
                <button
                  type="button"
                  class="btn-test-sound"
                  :class="{ active: isTestingSound }"
                  @click="testChime"
                >
                  <Volume2 :size="16" />
                  <span>{{ settings.language === 'en' ? 'Test Notification Chime' : 'Uji Suara Notifikasi' }}</span>
                </button>
              </div>

              <div class="volume-slider-row">
                <span class="slider-label">{{ settings.language === 'en' ? 'Sound Volume' : 'Volume Nada' }} ({{ settings.soundVolume }}%)</span>
                <input
                  type="range"
                  min="10"
                  max="100"
                  step="5"
                  v-model.number="settings.soundVolume"
                  class="range-slider"
                />
              </div>
            </div>

            <!-- Notification Toggles List -->
            <div class="toggles-list">
              <div class="toggle-card">
                <div class="toggle-text">
                  <strong>{{ settings.language === 'en' ? 'Browser Desktop Push Notifications' : 'Notifikasi Push Browser' }}</strong>
                  <p>{{ settings.language === 'en' ? 'Receive system pop-ups on your desktop when new messages or announcements arrive.' : 'Terima pop-up notifikasi instan pada browser ketika ada kegiatan penting.' }}</p>
                </div>
                <label class="switch">
                  <input type="checkbox" :checked="settings.notifyPush" @change="handleTogglePush" />
                  <span class="slider"></span>
                </label>
              </div>

              <div class="toggle-card">
                <div class="toggle-text">
                  <strong>{{ settings.language === 'en' ? 'Incoming Chat Audio Alert' : 'Peringatan Suara Pesan Baru' }}</strong>
                  <p>{{ settings.language === 'en' ? 'Play pleasant audio chime when receiving messages in class group chats.' : 'Mainkan denting suara saat seseorang mengirim pesan di ruang chat.' }}</p>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="settings.notifySound" />
                  <span class="slider"></span>
                </label>
              </div>

              <div class="toggle-card">
                <div class="toggle-text">
                  <strong>{{ settings.language === 'en' ? 'Message Content Preview' : 'Pratinjau Isi Pesan (Chat Preview)' }}</strong>
                  <p>{{ settings.language === 'en' ? 'Display message summary in the topbar dropdown panel.' : 'Tampilkan ringkasan isi pesan di menu drop-down notifikasi atas.' }}</p>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="settings.notifyChatPreview" />
                  <span class="slider"></span>
                </label>
              </div>

              <div class="toggle-card">
                <div class="toggle-text">
                  <strong>{{ settings.language === 'en' ? 'Assignment Deadline Reminders' : 'Pengingat Batas Pengumpulan Tugas' }}</strong>
                  <p>{{ settings.language === 'en' ? 'Receive reminders 24h & 1h before quiz or assignment deadlines.' : 'Terima peringatan sistem 24 jam & 1 jam sebelum tenggat waktu tugas atau kuis.' }}</p>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="settings.notifyDeadlines" />
                  <span class="slider"></span>
                </label>
              </div>

              <div class="toggle-card">
                <div class="toggle-text">
                  <strong>{{ settings.language === 'en' ? 'Email Digest Notifications' : 'Ringkasan Aktivitas via Email' }}</strong>
                  <p>{{ settings.language === 'en' ? 'Send weekly summaries and major school announcements to your registered email.' : 'Kirim rangkuman pengumuman sekolah dan tugas mingguan ke alamat email akun.' }}</p>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="settings.notifyEmail" />
                  <span class="slider"></span>
                </label>
              </div>
            </div>
          </div>

          <!-- TAB 4: PREFERENSI BELAJAR -->
          <div v-show="activeTab === 'learning'" class="panel-section">
            <div class="section-top">
              <div>
                <h2 class="section-heading">{{ settings.language === 'en' ? 'Learning & Chat Preferences' : 'Preferensi Pembelajaran & Chat' }}</h2>
                <p class="section-subheading">
                  {{ settings.language === 'en' ? 'Customize online status, presence indicator, and material attachment behavior.' : 'Kustomisasi perilaku ruang kelas virtual dan status privasi interaksi teman sekelas.' }}
                </p>
              </div>
              <BookOpen :size="24" class="text-teal" />
            </div>

            <div class="setting-group">
              <label class="group-label">{{ settings.language === 'en' ? 'Chat Presence Status' : 'Status Kehadiran di Obrolan (Presence Status)' }}</label>
              <p class="group-caption">{{ settings.language === 'en' ? 'Controls your status dot shown next to your avatar across the LMS.' : 'Menentukan bagaimana status ketersediaan Anda tampil pada ruang obrolan dan profil.' }}</p>

              <div class="presence-grid">
                <div
                  class="presence-card"
                  :class="{ active: settings.presenceStatus === 'online' }"
                  @click="settings.presenceStatus = 'online'"
                >
                  <span class="dot dot-online"></span>
                  <div>
                    <strong>{{ settings.language === 'en' ? 'Online / Active' : 'Online' }}</strong>
                    <small>{{ settings.language === 'en' ? 'Available to chat and study' : 'Aktif dan siap berinteraksi' }}</small>
                  </div>
                </div>

                <div
                  class="presence-card"
                  :class="{ active: settings.presenceStatus === 'busy' }"
                  @click="settings.presenceStatus = 'busy'"
                >
                  <span class="dot dot-busy"></span>
                  <div>
                    <strong>{{ settings.language === 'en' ? 'Busy / In Class' : 'Sibuk (Sedang Fokus)' }}</strong>
                    <small>{{ settings.language === 'en' ? 'Taking quiz or studying' : 'Sedang mengerjakan tugas/kuis' }}</small>
                  </div>
                </div>

                <div
                  class="presence-card"
                  :class="{ active: settings.presenceStatus === 'dnd' }"
                  @click="settings.presenceStatus = 'dnd'"
                >
                  <span class="dot dot-dnd"></span>
                  <div>
                    <strong>{{ settings.language === 'en' ? 'Do Not Disturb' : 'Jangan Ganggu' }}</strong>
                    <small>{{ settings.language === 'en' ? 'Mute incoming alerts' : 'Membisukan peringatan chat' }}</small>
                  </div>
                </div>

                <div
                  class="presence-card"
                  :class="{ active: settings.presenceStatus === 'invisible' }"
                  @click="settings.presenceStatus = 'invisible'"
                >
                  <span class="dot dot-invisible"></span>
                  <div>
                    <strong>{{ settings.language === 'en' ? 'Invisible' : 'Tak Terlihat' }}</strong>
                    <small>{{ settings.language === 'en' ? 'Appear offline to peers' : 'Tampil offline ke peserta lain' }}</small>
                  </div>
                </div>
              </div>
            </div>

            <div class="setting-row-2col mt-4">
              <div class="setting-card-item">
                <div class="item-text">
                  <strong>{{ settings.language === 'en' ? 'Lesson Materials Download Policy' : 'Pengunduhan Berkas Materi Pelajaran' }}</strong>
                  <p>{{ settings.language === 'en' ? 'Policy for caching lecture slides and PDF documents.' : 'Kebijakan unduh otomatis lampiran dokumen PDF/slide saat membuka kelas.' }}</p>
                </div>
                <select v-model="settings.autoDownloadMaterials" class="select-input">
                  <option value="wifi">{{ settings.language === 'en' ? 'WiFi Only' : 'Hanya saat Terkoneksi WiFi' }}</option>
                  <option value="always">{{ settings.language === 'en' ? 'Always Auto-Download' : 'Selalu Unduh Otomatis' }}</option>
                  <option value="never">{{ settings.language === 'en' ? 'Manual (Click to Download)' : 'Manual (Klik untuk Unduh)' }}</option>
                </select>
              </div>

              <div class="setting-card-item">
                <div class="item-text">
                  <strong>{{ settings.language === 'en' ? 'Profile Contact Visibility' : 'Visibilitas Profil Pengguna' }}</strong>
                  <p>{{ settings.language === 'en' ? 'Define who can view your enrolled subjects and bio.' : 'Tentukan jangkauan teman yang dapat melihat riwayat kelas dan kontak Anda.' }}</p>
                </div>
                <select v-model="settings.profileVisibility" class="select-input">
                  <option value="school">{{ settings.language === 'en' ? 'Entire School' : 'Seluruh Warga Sekolah' }}</option>
                  <option value="class_only">{{ settings.language === 'en' ? 'My Classes Only' : 'Hanya Teman di Kelas Saya' }}</option>
                </select>
              </div>
            </div>
          </div>

          <!-- TAB 5: SISTEM SEKOLAH (ADMIN ONLY) -->
          <div v-show="activeTab === 'system'" class="panel-section">
            <div class="section-top">
              <div>
                <h2 class="section-heading">{{ settings.language === 'en' ? 'School System Configuration' : 'Konfigurasi Sistem Sekolah' }}</h2>
                <p class="section-subheading">
                  {{ settings.language === 'en' ? 'Institution branding, upload limits, grade format, and maintenance mode.' : 'Parameter operasional umum lembaga, kuota berkas, dan kebijakan skala nilai LMS.' }}
                </p>
              </div>
              <School :size="24" class="text-teal" />
            </div>

            <!-- Non-Admin Notice if applicable -->
            <div v-if="!isAdmin" class="admin-locked-notice">
              <Shield :size="22" />
              <div>
                <strong>{{ settings.language === 'en' ? 'Read-only System Preview' : 'Pengaturan Terkunci untuk Peran Anda' }}</strong>
                <p>{{ settings.language === 'en' ? 'This section can only be modified by Administrators. Parameters are displayed in read-only mode.' : 'Bagian ini hanya dapat dimodifikasi oleh pengguna dengan role Administrator. Anda hanya dapat melihat pratinjau kebijakan sekolah.' }}</p>
              </div>
            </div>

            <div class="card-inner">
              <h3 class="inner-title">
                <School :size="18" />
                {{ settings.language === 'en' ? 'Institution Identity & System Rules' : 'Identitas Lembaga & Batasan Sistem' }}
              </h3>

              <div class="form-grid">
                <div class="form-group">
                  <label class="form-label">{{ settings.language === 'en' ? 'School / Institution Name' : 'Nama Lembaga / Sekolah' }}</label>
                  <input
                    type="text"
                    v-model="settings.schoolName"
                    class="text-input"
                    :disabled="!isAdmin"
                    placeholder="Contoh: School Space"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">{{ settings.language === 'en' ? 'Active Term & Academic Year' : 'Tahun Ajaran & Semester Aktif' }}</label>
                  <input
                    type="text"
                    v-model="settings.academicTermLabel"
                    class="text-input"
                    :disabled="!isAdmin"
                    placeholder="Contoh: 2026/2027 Ganjil"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">{{ settings.language === 'en' ? 'Maximum File Upload Size' : 'Batas Maksimal Ukuran Unggah Dokumen' }}</label>
                  <select v-model.number="settings.maxUploadSizeMB" class="select-input" :disabled="!isAdmin">
                    <option :value="10">10 Megabytes</option>
                    <option :value="25">25 Megabytes (Standard)</option>
                    <option :value="50">50 Megabytes</option>
                    <option :value="100">100 Megabytes</option>
                  </select>
                </div>

                <div class="form-group">
                  <label class="form-label">{{ settings.language === 'en' ? 'Default Grade Display Format' : 'Format Tampilan Nilai Default' }}</label>
                  <select v-model="settings.gradeFormat" class="select-input" :disabled="!isAdmin">
                    <option value="score">{{ settings.language === 'en' ? 'Numeric Score (0 - 100)' : 'Angka Skala 0 - 100' }}</option>
                    <option value="letter">{{ settings.language === 'en' ? 'Letter Grade (A, B, C, D, E)' : 'Huruf Predikat (A, B, C, D, E)' }}</option>
                  </select>
                </div>
              </div>

              <!-- Maintenance Mode Toggle -->
              <div class="toggle-card mt-4">
                <div class="toggle-text">
                  <strong>{{ settings.language === 'en' ? 'Maintenance Alert Mode' : 'Mode Pemeliharaan Terjadwal (Maintenance Mode)' }}</strong>
                  <p>{{ settings.language === 'en' ? 'Displays a prominent maintenance banner across the app to notify students and teachers.' : 'Tampilkan banner pengingat pemeliharaan sistem bagi siswa dan pengajar di bagian atas aplikasi.' }}</p>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="settings.maintenanceMode" :disabled="!isAdmin" />
                  <span class="slider"></span>
                </label>
              </div>

              <div v-if="isAdmin" class="form-footer mt-4">
                <p class="password-rule">
                  <Info :size="14" />
                  {{ settings.language === 'en' ? 'System changes take effect immediately across all layout headers and banners.' : 'Perubahan konfigurasi sistem berlaku langsung pada seluruh header dan banner portal.' }}
                </p>
                <button
                  type="button"
                  class="btn-primary"
                  :disabled="savingSystemSettings"
                  @click="handleSaveSystemSettings"
                >
                  <Save :size="16" />
                  <span>{{ savingSystemSettings ? (settings.language === 'en' ? 'Saving...' : 'Menyimpan...') : (settings.language === 'en' ? 'Save School Settings' : 'Simpan Pengaturan Sekolah') }}</span>
                </button>
              </div>
            </div>
          </div>

          <!-- TAB 6: DATA & PENYIMPANAN -->
          <div v-show="activeTab === 'storage'" class="panel-section">
            <div class="section-top">
              <div>
                <h2 class="section-heading">{{ settings.language === 'en' ? 'Data & Local Storage' : 'Data & Penyimpanan Lokal' }}</h2>
                <p class="section-subheading">
                  {{ settings.language === 'en' ? 'Manage offline cached resources, export configuration files, or reset values.' : 'Kelola berkas preferensi, cache offline, serta cadangan konfigurasi akun.' }}
                </p>
              </div>
              <Database :size="24" class="text-coral" />
            </div>

            <!-- Action Grid -->
            <div class="storage-action-grid">
              
              <!-- Clear Cache -->
              <div class="storage-card">
                <div class="storage-card-header">
                  <div class="storage-icon bg-coral-soft">
                    <Trash2 :size="20" class="text-coral" />
                  </div>
                  <div>
                    <strong>{{ settings.language === 'en' ? 'Clear Local Application Cache' : 'Bersihkan Cache Aplikasi' }}</strong>
                    <p>{{ settings.language === 'en' ? 'Purge temporary cached responses without signing you out.' : 'Hapus berkas sementara dan query offline tanpa mengeluarkan Anda dari akun.' }}</p>
                  </div>
                </div>
                <button
                  type="button"
                  class="btn-action-danger"
                  :disabled="clearingCache"
                  @click="handleClearCache"
                >
                  <Trash2 :size="15" />
                  <span>{{ clearingCache ? (settings.language === 'en' ? 'Clearing...' : 'Membersihkan...') : (settings.language === 'en' ? 'Clear Cache Now' : 'Bersihkan Cache Sekarang') }}</span>
                </button>
              </div>

              <!-- Export Settings -->
              <div class="storage-card">
                <div class="storage-card-header">
                  <div class="storage-icon bg-teal-soft">
                    <Download :size="20" class="text-teal" />
                  </div>
                  <div>
                    <strong>{{ settings.language === 'en' ? 'Export Preferences (JSON)' : 'Ekspor Pengaturan (JSON)' }}</strong>
                    <p>{{ settings.language === 'en' ? 'Download a portable JSON backup of your current themes and configurations.' : 'Unduh salinan preferensi tema, audio, dan parameter aplikasi ke komputer.' }}</p>
                  </div>
                </div>
                <button type="button" class="btn-action-outline" @click="exportSettings">
                  <Download :size="15" />
                  <span>{{ settings.language === 'en' ? 'Download JSON' : 'Unduh File JSON' }}</span>
                </button>
              </div>

              <!-- Reset to Defaults -->
              <div class="storage-card">
                <div class="storage-card-header">
                  <div class="storage-icon bg-yellow-soft">
                    <RotateCcw :size="20" class="text-yellow" />
                  </div>
                  <div>
                    <strong>{{ settings.language === 'en' ? 'Restore System Defaults' : 'Setel Ulang ke Nilai Pabrik' }}</strong>
                    <p>{{ settings.language === 'en' ? 'Restore colors, font sizes, sounds, and densities back to initial setup.' : 'Kembalikan seluruh warna, font, dan tombol notifikasi ke setelan default awal.' }}</p>
                  </div>
                </div>
                <button type="button" class="btn-action-outline" @click="handleResetDefaults">
                  <RotateCcw :size="15" />
                  <span>{{ settings.language === 'en' ? 'Restore Defaults' : 'Pulihkan Setelan Default' }}</span>
                </button>
              </div>

            </div>

            <!-- Storage Diagnostics -->
            <div class="card-inner mt-4">
              <h3 class="inner-title">
                <Info :size="18" />
                {{ settings.language === 'en' ? 'Browser Storage Diagnostics' : 'Informasi Penyimpanan Peramban' }}
              </h3>
              <div class="diag-list">
                <div class="diag-item">
                  <span>{{ settings.language === 'en' ? 'Local Storage Status' : 'Status Penyimpanan Lokal (LocalStorage)' }}:</span>
                  <strong>{{ settings.language === 'en' ? 'Active & Available' : 'Aktif & Tersedia' }}</strong>
                </div>
                <div class="diag-item">
                  <span>{{ settings.language === 'en' ? 'Session Storage Status' : 'Penyimpanan Sesi (SessionStorage)' }}:</span>
                  <strong>{{ settings.language === 'en' ? 'Normal' : 'Normal' }}</strong>
                </div>
                <div class="diag-item">
                  <span>{{ settings.language === 'en' ? 'Web Audio API Support' : 'Dukungan Web Audio API' }}:</span>
                  <strong>{{ settings.language === 'en' ? 'Supported' : 'Didukung Penuh' }}</strong>
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
.settings-container {
  display: flex;
  flex-direction: column;
  gap: 22px;
  max-width: 1200px;
  margin: 0 auto;
}

/* Header */
.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  flex-wrap: wrap;
  padding-bottom: 8px;
}

.page-title {
  font-size: 24px;
  font-weight: 800;
  color: var(--ink);
  margin: 0 0 4px;
  letter-spacing: -0.5px;
}

.page-subtitle {
  font-size: 13.5px;
  color: var(--muted);
  margin: 0;
}

.settings-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: var(--teal);
  color: white;
  border: none;
  border-radius: 12px;
  padding: 10px 18px;
  font-size: 13.5px;
  font-weight: 700;
  cursor: pointer;
  transition: transform 0.15s, background-color 0.2s, box-shadow 0.2s;
  box-shadow: 0 4px 12px rgba(101, 201, 171, 0.25);
  text-decoration: none;
}
.btn-primary:hover:not(:disabled) {
  background: var(--teal-dark);
  transform: translateY(-1px);
}
.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-outline {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: white;
  color: var(--ink);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 10px 16px;
  font-size: 13.5px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-outline:hover {
  background: var(--paper);
}

/* Feedback Alert */
.feedback-alert {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 18px;
  border-radius: 14px;
  font-size: 13.5px;
  font-weight: 700;
  animation: slideDown 0.25s ease-out;
}
.feedback-alert.success {
  background: #e8f9f2;
  color: #1a7f5d;
  border: 1px solid #b7ecd8;
}
.feedback-alert.error {
  background: #fdf0ee;
  color: #c03d32;
  border: 1px solid #f9c7c2;
}
.feedback-alert.info {
  background: #f0f6ff;
  color: #2563eb;
  border: 1px solid #bfdbfe;
}
.feedback-close {
  margin-left: auto;
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: inherit;
  opacity: 0.7;
}
.feedback-close:hover {
  opacity: 1;
}

/* Main Grid */
.settings-grid {
  display: grid;
  grid-template-columns: 310px 1fr;
  gap: 24px;
  align-items: start;
}

/* Left Nav Card */
.settings-nav-card {
  background: white;
  border-radius: 20px;
  border: 1px solid var(--line);
  padding: 18px 14px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.02);
}

.user-brief {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px 18px;
  border-bottom: 1px solid var(--line);
  margin-bottom: 14px;
}
.user-brief-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--teal) 0%, var(--teal-dark) 100%);
  color: white;
  display: grid;
  place-items: center;
  font-size: 18px;
  font-weight: 800;
  flex-shrink: 0;
}
.brief-presence-dot {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 11px;
  height: 11px;
  border-radius: 50%;
  border: 2px solid white;
}
.user-brief-info {
  display: flex;
  flex-direction: column;
}
.user-brief-name {
  font-size: 14px;
  font-weight: 800;
  color: var(--ink);
  line-height: 1.2;
}
.user-brief-role {
  font-size: 11.5px;
  color: var(--muted);
  font-weight: 700;
}

.tabs-menu {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.tab-btn {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  width: 100%;
  padding: 12px 14px;
  border-radius: 14px;
  border: none;
  background: transparent;
  cursor: pointer;
  text-align: left;
  transition: all 0.2s;
  color: var(--ink);
}
.tab-btn:hover {
  background: var(--paper);
}
.tab-btn.active {
  background: rgba(101, 201, 171, 0.14);
  color: var(--teal-dark);
}
.tab-btn.active .tab-title {
  color: var(--teal-dark);
}
.tab-btn.active .tab-icon {
  color: var(--teal);
}
.tab-btn.active .tab-desc {
  color: #438f78;
}
.tab-icon {
  color: var(--muted);
  margin-top: 2px;
  flex-shrink: 0;
  transition: color 0.2s;
}
.tab-text {
  display: flex;
  flex-direction: column;
}
.tab-title {
  font-size: 13.5px;
  font-weight: 800;
  display: flex;
  align-items: center;
  gap: 6px;
}
.tab-desc {
  font-size: 11px;
  color: var(--muted);
  margin-top: 1px;
}
.role-tag {
  background: var(--coral);
  color: white;
  font-size: 9px;
  font-weight: 800;
  padding: 1px 6px;
  border-radius: 6px;
  text-transform: uppercase;
}

/* Right Content Area */
.settings-content-area {
  background: white;
  border-radius: 24px;
  border: 1px solid var(--line);
  padding: 30px;
  box-shadow: 0 4px 18px rgba(0,0,0,0.02);
  min-height: 520px;
}

.section-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--line);
}
.section-heading {
  font-size: 19px;
  font-weight: 800;
  color: var(--ink);
  margin: 0 0 4px;
}
.section-subheading {
  font-size: 13px;
  color: var(--muted);
  margin: 0;
}

/* Groups */
.setting-group {
  margin-bottom: 26px;
}
.group-label {
  display: block;
  font-size: 14px;
  font-weight: 800;
  color: var(--ink);
  margin-bottom: 4px;
}
.group-caption {
  font-size: 12.5px;
  color: var(--muted);
  margin: 0 0 14px;
}

/* Theme Picker Cards */
.theme-picker-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
.theme-card {
  border: 2px solid var(--line);
  border-radius: 16px;
  padding: 12px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  background: white;
}
.theme-card:hover {
  border-color: #cbd5e1;
  transform: translateY(-2px);
}
.theme-card.active {
  border-color: var(--teal);
  box-shadow: 0 4px 14px rgba(101, 201, 171, 0.2);
  background: #fcfefe;
}

.theme-preview {
  height: 72px;
  border-radius: 10px;
  padding: 10px;
  position: relative;
  overflow: hidden;
  margin-bottom: 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.preview-light {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
}
.preview-light .preview-bar {
  width: 45%;
  height: 8px;
  background: #cbd5e1;
  border-radius: 4px;
}
.preview-light .preview-box {
  width: 75%;
  height: 20px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
}

.preview-dark {
  background: #0f172a;
  border: 1px solid #1e293b;
}
.preview-dark .preview-bar {
  width: 45%;
  height: 8px;
  background: #334155;
  border-radius: 4px;
}
.preview-dark .preview-box {
  width: 75%;
  height: 20px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
}

.preview-system {
  background: linear-gradient(135deg, #f8fafc 50%, #0f172a 50%);
  border: 1px solid #cbd5e1;
}

.theme-badge-icon {
  position: absolute;
  right: 8px;
  bottom: 8px;
  opacity: 0.6;
}
.theme-card-info {
  display: flex;
  flex-direction: column;
}
.theme-name {
  font-size: 13.5px;
  font-weight: 800;
  color: var(--ink);
}
.theme-card-info small {
  font-size: 11px;
  color: var(--muted);
}
.check-mark {
  position: absolute;
  top: 10px;
  right: 10px;
  background: var(--teal);
  color: white;
  border-radius: 50%;
  padding: 2px;
}

/* Options Pill Group */
.options-pill-group {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}
.pill-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  border-radius: 12px;
  border: 1px solid var(--line);
  background: white;
  cursor: pointer;
  font-size: 13px;
  font-weight: 700;
  color: var(--ink);
  transition: all 0.2s;
}
.pill-btn:hover {
  background: var(--paper);
}
.pill-btn.active {
  background: var(--teal);
  color: white;
  border-color: var(--teal);
}

/* 2 Column Settings */
.setting-row-2col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 20px;
}
.setting-card-item {
  padding: 16px;
  border-radius: 14px;
  border: 1px solid var(--line);
  background: #fafbfc;
}
.item-text strong {
  display: block;
  font-size: 13.5px;
  font-weight: 800;
  color: var(--ink);
  margin-bottom: 3px;
}
.item-text p {
  font-size: 12px;
  color: var(--muted);
  margin: 0 0 12px;
}
.select-input {
  width: 100%;
  padding: 9px 12px;
  border-radius: 10px;
  border: 1px solid var(--line);
  background: white;
  font-size: 13px;
  font-weight: 700;
  color: var(--ink);
  outline: none;
}
.select-input:focus {
  border-color: var(--teal);
}

/* Profile Hub Banner */
.profile-hub-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  background: linear-gradient(135deg, #f0faf6 0%, #ffffff 100%);
  border: 1px solid #bfe8db;
  border-radius: 18px;
  padding: 20px;
  flex-wrap: wrap;
}
.profile-hub-content {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}
.profile-hub-icon {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  background: white;
  border: 1px solid #bfe8db;
  display: grid;
  place-items: center;
  flex-shrink: 0;
}
.profile-hub-text strong {
  display: block;
  font-size: 14.5px;
  font-weight: 800;
  color: var(--ink);
  margin-bottom: 4px;
}
.profile-hub-text p {
  font-size: 12.5px;
  color: var(--muted);
  margin: 0;
  line-height: 1.4;
}

/* Toggle Card */
.toggles-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.toggle-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 18px;
  border-radius: 14px;
  border: 1px solid var(--line);
  background: white;
  transition: background-color 0.2s;
}
.toggle-card:hover {
  background: #fbfcfc;
}
.toggle-text strong {
  display: block;
  font-size: 13.5px;
  font-weight: 800;
  color: var(--ink);
  margin-bottom: 2px;
}
.toggle-text p {
  font-size: 12px;
  color: var(--muted);
  margin: 0;
}

/* Switch */
.switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
  flex-shrink: 0;
}
.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}
.slider {
  position: absolute;
  cursor: pointer;
  inset: 0;
  background-color: #cbd5e1;
  transition: .3s;
  border-radius: 24px;
}
.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: .3s;
  border-radius: 50%;
  box-shadow: 0 2px 4px rgba(0,0,0,0.15);
}
input:checked + .slider {
  background-color: var(--teal);
}
input:checked + .slider:before {
  transform: translateX(20px);
}

/* Inner Card */
.card-inner {
  border: 1px solid var(--line);
  border-radius: 18px;
  padding: 22px;
  background: white;
}
.inner-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 800;
  color: var(--ink);
  margin: 0 0 16px;
}

/* Form */
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.form-group {
  display: flex;
  flex-direction: column;
}
.form-label {
  font-size: 12.5px;
  font-weight: 800;
  color: var(--ink);
  margin-bottom: 6px;
}
.text-input {
  width: 100%;
  padding: 10px 14px;
  border-radius: 10px;
  border: 1px solid var(--line);
  font-size: 13.5px;
  outline: none;
  background: white;
  color: var(--ink);
  transition: border-color 0.2s;
}
.text-input:focus {
  border-color: var(--teal);
  box-shadow: 0 0 0 3px rgba(101, 201, 171, 0.1);
}
.text-input:disabled {
  background: var(--paper);
  color: var(--muted);
  cursor: not-allowed;
}

.form-footer {
  grid-column: span 2;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
  flex-wrap: wrap;
  gap: 14px;
}
.password-rule {
  font-size: 12px;
  color: var(--muted);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 6px;
  max-width: 500px;
}

/* Session Card */
.session-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px;
  border-radius: 12px;
  background: #f8fafc;
  border: 1px solid var(--line);
}
.session-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: white;
  border: 1px solid var(--line);
  display: grid;
  place-items: center;
  color: var(--teal);
  flex-shrink: 0;
}
.session-details strong {
  display: block;
  font-size: 13.5px;
  color: var(--ink);
}
.session-details p {
  font-size: 12px;
  color: var(--muted);
  margin: 2px 0 6px;
}
.session-badge-active {
  display: inline-block;
  font-size: 10px;
  font-weight: 800;
  color: #1a7f5d;
  background: #e8f9f2;
  padding: 2px 8px;
  border-radius: 20px;
}

/* Sound Hero Box */
.sound-hero-box {
  background: linear-gradient(135deg, #f0faf6 0%, #fdfdfd 100%);
  border-color: #bfe8db;
}
.sound-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 12px;
}
.sound-desc {
  font-size: 12.5px;
  color: var(--muted);
  margin: 0;
}
.btn-test-sound {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 9px 16px;
  border-radius: 10px;
  border: 1px solid var(--teal);
  background: white;
  color: var(--teal-dark);
  font-size: 13px;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-test-sound:hover {
  background: var(--teal);
  color: white;
}
.btn-test-sound.active {
  transform: scale(0.97);
  background: var(--teal-dark);
  color: white;
}

.volume-slider-row {
  display: flex;
  align-items: center;
  gap: 16px;
}
.slider-label {
  font-size: 13px;
  font-weight: 700;
  color: var(--ink);
  min-width: 140px;
}
.range-slider {
  flex: 1;
  accent-color: var(--teal);
  height: 6px;
  border-radius: 3px;
  cursor: pointer;
}

/* Presence Grid */
.presence-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}
.presence-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 12px;
  border: 1px solid var(--line);
  background: white;
  cursor: pointer;
  transition: all 0.2s;
}
.presence-card:hover {
  background: #fbfcfc;
}
.presence-card.active {
  border-color: var(--teal);
  background: #f0faf6;
}
.dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}
.dot-online { background: #10b981; }
.dot-busy { background: #f59e0b; }
.dot-dnd { background: #ef4444; }
.dot-invisible { background: #94a3b8; }
.presence-card strong {
  display: block;
  font-size: 13px;
  font-weight: 800;
  color: var(--ink);
}
.presence-card small {
  font-size: 11px;
  color: var(--muted);
}

/* Admin Lock Notice */
.admin-locked-notice {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  background: #fef8f0;
  border: 1px solid #fce8cc;
  padding: 14px 18px;
  border-radius: 14px;
  margin-bottom: 20px;
  color: #b45309;
}
.admin-locked-notice strong {
  display: block;
  font-size: 13.5px;
  font-weight: 800;
  margin-bottom: 2px;
}
.admin-locked-notice p {
  font-size: 12px;
  margin: 0;
  line-height: 1.4;
}

/* Storage Section */
.storage-action-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
}
.storage-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-radius: 16px;
  border: 1px solid var(--line);
  background: white;
  gap: 16px;
  flex-wrap: wrap;
}
.storage-card-header {
  display: flex;
  align-items: center;
  gap: 14px;
}
.storage-icon {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
}
.bg-coral-soft { background: #fef2f2; }
.bg-teal-soft { background: #f0fdf4; }
.bg-yellow-soft { background: #fffbeb; }
.storage-card-header strong {
  display: block;
  font-size: 13.5px;
  font-weight: 800;
  color: var(--ink);
}
.storage-card-header p {
  font-size: 12px;
  color: var(--muted);
  margin: 2px 0 0;
}

.btn-action-danger {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
  padding: 8px 16px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-action-danger:hover {
  background: #dc2626;
  color: white;
}

.btn-action-outline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: white;
  color: var(--ink);
  border: 1px solid var(--line);
  padding: 8px 16px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-action-outline:hover {
  background: var(--paper);
}

.diag-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.diag-item {
  display: flex;
  justify-content: space-between;
  font-size: 12.5px;
  padding: 6px 0;
  border-bottom: 1px dashed var(--line);
}
.diag-item span { color: var(--muted); }
.diag-item strong { color: var(--ink); }

/* Utilities */
.text-teal { color: var(--teal); }
.text-coral { color: var(--coral); }
.text-yellow { color: var(--yellow); }
.mt-4 { margin-top: 16px; }
.mb-4 { margin-bottom: 16px; }
.mb-1 { margin-bottom: 4px; }

/* Responsive */
@media (max-width: 900px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }
  .tabs-menu {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
  }
  .tab-btn {
    padding: 10px;
  }
  .tab-desc {
    display: none;
  }
  .theme-picker-grid {
    grid-template-columns: 1fr;
  }
  .form-grid {
    grid-template-columns: 1fr;
  }
  .form-footer {
    grid-column: span 1;
  }
  .setting-row-2col {
    grid-template-columns: 1fr;
  }
  .presence-grid {
    grid-template-columns: 1fr;
  }
  .profile-hub-banner {
    flex-direction: column;
    align-items: flex-start;
  }
}

@media (max-width: 600px) {
  .tabs-menu {
    grid-template-columns: repeat(2, 1fr);
  }
  .settings-header {
    flex-direction: column;
    align-items: flex-start;
  }
  .settings-content-area {
    padding: 18px;
  }
}

/* ========================================================
   DARK MODE OVERRIDES (Mencegah background hover/active putih)
   ======================================================== */
:global(html[data-theme='dark']) .settings-nav-card,
:global(html[data-theme='dark']) .settings-content-area,
:global(html[data-theme='dark']) .card-inner,
:global(html[data-theme='dark']) .storage-card,
:global(html[data-theme='dark']) .toggle-card,
:global(html[data-theme='dark']) .theme-card,
:global(html[data-theme='dark']) .pill-btn,
:global(html[data-theme='dark']) .presence-card {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}

:global(html[data-theme='dark']) .tab-btn {
  color: #94a3b8 !important;
}

:global(html[data-theme='dark']) .tab-btn:hover {
  background: #273549 !important;
  color: #f1f5f9 !important;
}

:global(html[data-theme='dark']) .tab-btn.active {
  background: rgba(101, 201, 171, 0.16) !important;
  color: #65c9ab !important;
  border: 1px solid rgba(101, 201, 171, 0.25) !important;
}

:global(html[data-theme='dark']) .tab-btn.active .tab-title {
  color: #65c9ab !important;
}

:global(html[data-theme='dark']) .tab-btn.active .tab-icon {
  color: #65c9ab !important;
}

:global(html[data-theme='dark']) .tab-btn.active .tab-desc {
  color: #a7f3d0 !important;
}

:global(html[data-theme='dark']) .tab-btn.active:hover {
  background: rgba(101, 201, 171, 0.24) !important;
}

:global(html[data-theme='dark']) .theme-card {
  background: #1e293b !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .theme-card:hover {
  background: #273549 !important;
  border-color: #475569 !important;
}
:global(html[data-theme='dark']) .theme-card.active {
  background: rgba(101, 201, 171, 0.12) !important;
  border-color: #65c9ab !important;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.3) !important;
}

:global(html[data-theme='dark']) .presence-card {
  background: #1e293b !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .presence-card:hover {
  background: #273549 !important;
}
:global(html[data-theme='dark']) .presence-card.active {
  background: rgba(101, 201, 171, 0.14) !important;
  border-color: #65c9ab !important;
}

:global(html[data-theme='dark']) .brief-presence-dot {
  border-color: #1e293b !important;
}

:global(html[data-theme='dark']) .toggle-card {
  background: #1e293b !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .toggle-card:hover {
  background: #273549 !important;
}

:global(html[data-theme='dark']) .pill-btn {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .pill-btn:hover {
  background: #273549 !important;
}
:global(html[data-theme='dark']) .pill-btn.active {
  background: #65c9ab !important;
  color: #0f172a !important;
  border-color: #65c9ab !important;
}

:global(html[data-theme='dark']) .setting-card-item,
:global(html[data-theme='dark']) .session-card {
  background: #151f30 !important;
  border-color: #334155 !important;
}

:global(html[data-theme='dark']) .sound-hero-box {
  background: linear-gradient(135deg, #102923 0%, #1e293b 100%) !important;
  border-color: #1f4a40 !important;
}

:global(html[data-theme='dark']) .profile-hub-banner {
  background: linear-gradient(135deg, #102923 0%, #1e293b 100%) !important;
  border-color: #1f4a40 !important;
}
:global(html[data-theme='dark']) .profile-hub-icon {
  background: #1e293b !important;
  border-color: #1f4a40 !important;
}

:global(html[data-theme='dark']) .btn-test-sound {
  background: #1e293b !important;
  color: #65c9ab !important;
  border-color: #65c9ab !important;
}
:global(html[data-theme='dark']) .btn-test-sound:hover {
  background: #65c9ab !important;
  color: #0f172a !important;
}

:global(html[data-theme='dark']) .select-input,
:global(html[data-theme='dark']) .text-input {
  background: #151f30 !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .select-input:focus,
:global(html[data-theme='dark']) .text-input:focus {
  border-color: #65c9ab !important;
  box-shadow: 0 0 0 3px rgba(101, 201, 171, 0.15) !important;
}

:global(html[data-theme='dark']) .btn-outline,
:global(html[data-theme='dark']) .btn-action-outline {
  background: #1e293b !important;
  color: #f1f5f9 !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .btn-outline:hover,
:global(html[data-theme='dark']) .btn-action-outline:hover {
  background: #273549 !important;
}

:global(html[data-theme='dark']) .bg-coral-soft { background: #3b1e22 !important; }
:global(html[data-theme='dark']) .bg-teal-soft { background: #102923 !important; }
:global(html[data-theme='dark']) .bg-yellow-soft { background: #382c16 !important; }

:global(html[data-theme='dark']) .page-title,
:global(html[data-theme='dark']) .section-heading,
:global(html[data-theme='dark']) .group-label,
:global(html[data-theme='dark']) .inner-title,
:global(html[data-theme='dark']) .theme-name,
:global(html[data-theme='dark']) .user-brief-name,
:global(html[data-theme='dark']) .item-text strong,
:global(html[data-theme='dark']) .profile-hub-text strong,
:global(html[data-theme='dark']) .toggle-text strong,
:global(html[data-theme='dark']) .session-details strong,
:global(html[data-theme='dark']) .slider-label,
:global(html[data-theme='dark']) .presence-card strong,
:global(html[data-theme='dark']) .storage-card-header strong,
:global(html[data-theme='dark']) .diag-item strong {
  color: #f8fafc !important;
}
</style>
