import { reactive, watch, computed } from 'vue'

export interface AppSettingsState {
  // Tampilan & Bahasa
  theme: 'light' | 'dark' | 'system'
  fontSize: 'sm' | 'md' | 'lg'
  density: 'comfortable' | 'compact'
  language: 'id' | 'en'
  reduceMotion: boolean

  // Notifikasi & Suara
  notifyEmail: boolean
  notifyPush: boolean
  notifyChatPreview: boolean
  notifySound: boolean
  notifyDeadlines: boolean
  soundVolume: number

  // Pembelajaran & Privasi
  autoDownloadMaterials: 'wifi' | 'always' | 'never'
  presenceStatus: 'online' | 'busy' | 'dnd' | 'invisible'
  profileVisibility: 'school' | 'class_only'

  // Keamanan & Sesi
  sessionTimeoutMinutes: number
  rememberDevice: boolean

  // Admin / Sistem Sekolah
  schoolName: string
  academicTermLabel: string
  maxUploadSizeMB: number
  gradeFormat: 'score' | 'letter'
  maintenanceMode: boolean
}

const STORAGE_KEY = 'school_lms_app_settings'

const defaultSettings: AppSettingsState = {
  theme: 'light',
  fontSize: 'md',
  density: 'comfortable',
  language: 'id',
  reduceMotion: false,

  notifyEmail: true,
  notifyPush: false,
  notifyChatPreview: true,
  notifySound: true,
  notifyDeadlines: true,
  soundVolume: 80,

  autoDownloadMaterials: 'wifi',
  presenceStatus: 'online',
  profileVisibility: 'school',

  sessionTimeoutMinutes: 60,
  rememberDevice: true,

  schoolName: 'School Space',
  academicTermLabel: '2026/2027 Ganjil',
  maxUploadSizeMB: 25,
  gradeFormat: 'score',
  maintenanceMode: false,
}

// Dictionary for true multi-language support (ID & EN)
const translations: Record<'id' | 'en', Record<string, string>> = {
  id: {
    // Navigasi
    'nav.dashboard': 'Dashboard',
    'nav.admin': 'Administrasi',
    'nav.academic': 'Akademik',
    'nav.academic_year': 'Tahun Ajaran',
    'nav.learning_teacher': 'Konten Mengajar',
    'nav.learning_student': 'Kelas Saya',
    'nav.chat': 'Grup Chat',
    'nav.help_center': 'Pusat Bantuan',
    'nav.settings': 'Pengaturan',
    'nav.logout': 'Keluar',
    'nav.notifications': 'Notifikasi',
    'nav.messages': 'Pesan masuk',
    'nav.profile': 'Buka Profil Saya',
    'search.placeholder': 'Cari sesuatu...',

    // Topbar & Banner
    'banner.maintenance': 'Pemberitahuan: Sistem sedang dalam mode pemeliharaan/evaluasi terjadwal. Beberapa fitur mungkin dibatasi sementara.',

    // Settings Header & Tabs
    'settings.title': 'Pusat Pengaturan',
    'settings.subtitle': 'Kelola preferensi antarmuka, notifikasi, keamanan sesi, dan sistem sekolah secara terpadu.',
    'settings.btn_save': 'Simpan Pengaturan',
    'settings.btn_reset': 'Reset Standar',
    'settings.tab_appearance': 'Tampilan & Tema',
    'settings.tab_appearance_desc': 'Mode gelap, ukuran font, antarmuka',
    'settings.tab_account': 'Keamanan & Sesi',
    'settings.tab_account_desc': 'Status akun & sesi login',
    'settings.tab_notifications': 'Notifikasi & Suara',
    'settings.tab_notifications_desc': 'Alert, push & bunyi chime',
    'settings.tab_learning': 'Preferensi Belajar',
    'settings.tab_learning_desc': 'Unduhan berkas & status obrolan',
    'settings.tab_system': 'Sistem Sekolah',
    'settings.tab_system_desc': 'Nama sekolah, kuota & format nilai',
    'settings.tab_storage': 'Data & Penyimpanan',
    'settings.tab_storage_desc': 'Cache lokal, ekspor & cadangan',

    // Role
    'role.admin': 'Administrator',
    'role.guru': 'Guru',
    'role.teacher': 'Guru',
    'role.siswa': 'Siswa',
    'role.student': 'Siswa',

    // Common
    'common.save': 'Simpan',
    'common.saving': 'Menyimpan...',
    'common.cancel': 'Batal',
    'common.delete': 'Hapus',
    'common.edit': 'Ubah',
    'common.back': 'Kembali',
    'common.close': 'Tutup',
    'common.loading': 'Memuat...',
    'common.all': 'Semua',
    'common.active': 'Aktif',
    'common.inactive': 'Nonaktif',
    'common.online': 'Online',
    'common.offline': 'Offline',
    'common.user': 'Pengguna',

    // Dashboard
    'dashboard.workspace': 'Ruang kerja',
    'dashboard.greeting': 'Halo,',
    'dashboard.loading': 'Menyiapkan ruang kerjamu...',
    'dashboard.hero_title': 'Pembelajaran Online yang dapat diakses dimana saja dengan mudah',
    'dashboard.start_learning': 'Mulai Belajar',
    'dashboard.stat_materials_created': 'Materi Dibuat',
    'dashboard.stat_materials_ready': 'Materi Siap',
    'dashboard.stat_assignments_active': 'Tugas Aktif',
    'dashboard.stat_quizzes_ready': 'Kuis Tersedia',
    'dashboard.stat_total_users': 'Total Pengguna',
    'dashboard.stat_classes_enrolled': 'Kelas Terdaftar',
    'dashboard.stat_subjects': 'Mata Pelajaran',
    'dashboard.recent_announcements': 'Pengumuman Terbaru',
    'dashboard.no_announcements': 'Belum ada pengumuman baru.',

    // Chat
    'chat.title_admin': 'Chat Sekolah',
    'chat.title_teacher': 'Chat Guru & Siswa',
    'chat.title_student': 'Chat dengan Guru',
    'chat.contact_help_student': 'Pilih guru atau admin untuk memulai percakapan.',
    'chat.contact_help_default': 'Pilih pengguna yang dapat diajak berkoordinasi.',
    'chat.connected': 'Terhubung',
    'chat.connecting': 'Menghubungkan...',
    'chat.search_placeholder': 'Cari percakapan...',
    'chat.filter_messages': 'Pesan',
    'chat.filter_unread': 'Belum dibaca',
    'chat.create_room': 'Buat ruang',
    'chat.create_room_title': 'Buat ruang chat',
    'chat.room_name_label': 'Nama ruang chat',
    'chat.room_name_placeholder': 'Contoh: Diskusi XI IPA 1',
    'chat.class_label': 'Kelas',
    'chat.select_class': 'Pilih kelas yang diajar',
    'chat.no_students': 'Belum ada siswa di kelas ini.',
    'chat.loading': 'Memuat percakapan...',
    'chat.loading_messages': 'Memuat pesan...',
    'chat.empty_conversations': 'Belum ada percakapan.',
    'chat.empty_contacts': 'Belum ada pengguna lain yang aktif.',
    'chat.click_to_open': 'Klik untuk membuka chat',
    'chat.select_conversation': 'Pilih percakapan',
    'chat.select_conversation_desc': 'Pilih percakapan di sebelah kiri untuk mulai mengobrol.',
    'chat.welcome_title': 'Selamat datang di Chat',
    'chat.welcome_desc': 'Pilih percakapan yang ada di sidebar atau buat percakapan baru untuk memulai.',
    'chat.typing': 'Sedang mengetik...',
    'chat.input_placeholder': 'Ketik pesan...',
    'chat.switch_chat': 'Ganti chat',
    'chat.press_enter': 'Tekan Enter untuk mengirim, Shift + Enter untuk baris baru.',
    'chat.no_new_messages': 'Tidak ada pesan baru.',

    // Notifications
    'notif.title': 'Pusat Notifikasi',
    'notif.all': 'Semua',
    'notif.unread': 'Belum Dibaca',
    'notif.mark_all': 'Tandai semua dibaca',
    'notif.marking': 'Menandai...',
    'notif.loading': 'Memuat notifikasi...',
    'notif.empty': 'Tidak ada notifikasi saat ini.',
    'notif.just_now': 'Baru saja',
    'notif.mins_ago': 'menit lalu',
    'notif.hours_ago': 'jam lalu',
    'notif.yesterday': 'Kemarin',
    'notif.days_ago': 'hari lalu',

    // Feedback
    'feedback.saved': 'Pengaturan berhasil diperbarui dan diterapkan!',
    'feedback.reset': 'Semua preferensi telah dikembalikan ke setelan bawaan.',
    'feedback.cache_cleared': 'Cache lokal berhasil dibersihkan!',
    'feedback.exported': 'Berkas konfigurasi pengaturan berhasil diunduh.',
  },
  en: {
    // Navigation
    'nav.dashboard': 'Dashboard',
    'nav.admin': 'Administration',
    'nav.academic': 'Academic',
    'nav.academic_year': 'Academic Year',
    'nav.learning_teacher': 'Teaching Content',
    'nav.learning_student': 'My Classes',
    'nav.chat': 'Group Chat',
    'nav.help_center': 'Help Center',
    'nav.settings': 'Settings',
    'nav.logout': 'Sign Out',
    'nav.notifications': 'Notifications',
    'nav.messages': 'Inbox messages',
    'nav.profile': 'View My Profile',
    'search.placeholder': 'Search anything...',

    // Topbar & Banner
    'banner.maintenance': 'Notice: The system is currently in scheduled maintenance mode. Certain features may be temporarily limited.',

    // Settings Header & Tabs
    'settings.title': 'Settings Center',
    'settings.subtitle': 'Manage interface preferences, notifications, session security, and school system settings.',
    'settings.btn_save': 'Save Settings',
    'settings.btn_reset': 'Reset Defaults',
    'settings.tab_appearance': 'Appearance & Theme',
    'settings.tab_appearance_desc': 'Dark mode, font scaling, UI density',
    'settings.tab_account': 'Security & Session',
    'settings.tab_account_desc': 'Account status & active login sessions',
    'settings.tab_notifications': 'Notifications & Sounds',
    'settings.tab_notifications_desc': 'Alerts, push & audio chime',
    'settings.tab_learning': 'Learning Preferences',
    'settings.tab_learning_desc': 'File downloads & chat presence status',
    'settings.tab_system': 'School System',
    'settings.tab_system_desc': 'School identity, quotas & grade format',
    'settings.tab_storage': 'Data & Storage',
    'settings.tab_storage_desc': 'Local cache, export & backup',

    // Role
    'role.admin': 'Administrator',
    'role.guru': 'Teacher',
    'role.teacher': 'Teacher',
    'role.siswa': 'Student',
    'role.student': 'Student',

    // Common
    'common.save': 'Save',
    'common.saving': 'Saving...',
    'common.cancel': 'Cancel',
    'common.delete': 'Delete',
    'common.edit': 'Edit',
    'common.back': 'Back',
    'common.close': 'Close',
    'common.loading': 'Loading...',
    'common.all': 'All',
    'common.active': 'Active',
    'common.inactive': 'Inactive',
    'common.online': 'Online',
    'common.offline': 'Offline',
    'common.user': 'User',

    // Dashboard
    'dashboard.workspace': 'Workspace',
    'dashboard.greeting': 'Hello,',
    'dashboard.loading': 'Preparing your workspace...',
    'dashboard.hero_title': 'Online Learning easily accessible anywhere and anytime',
    'dashboard.start_learning': 'Start Learning',
    'dashboard.stat_materials_created': 'Created Materials',
    'dashboard.stat_materials_ready': 'Available Materials',
    'dashboard.stat_assignments_active': 'Active Assignments',
    'dashboard.stat_quizzes_ready': 'Available Quizzes',
    'dashboard.stat_total_users': 'Total Users',
    'dashboard.stat_classes_enrolled': 'Enrolled Classes',
    'dashboard.stat_subjects': 'Subjects',
    'dashboard.recent_announcements': 'Recent Announcements',
    'dashboard.no_announcements': 'No new announcements available.',

    // Chat
    'chat.title_admin': 'School Chat',
    'chat.title_teacher': 'Teacher & Student Chat',
    'chat.title_student': 'Teacher & Peer Chat',
    'chat.contact_help_student': 'Select a teacher or admin to start a conversation.',
    'chat.contact_help_default': 'Select a user to collaborate with.',
    'chat.connected': 'Connected',
    'chat.connecting': 'Connecting...',
    'chat.search_placeholder': 'Search conversations...',
    'chat.filter_messages': 'Messages',
    'chat.filter_unread': 'Unread',
    'chat.create_room': 'Create room',
    'chat.create_room_title': 'Create chat room',
    'chat.room_name_label': 'Chat room name',
    'chat.room_name_placeholder': 'e.g., Grade 11 Science Discussion',
    'chat.class_label': 'Class',
    'chat.select_class': 'Select your teaching class',
    'chat.no_students': 'No students in this class yet.',
    'chat.loading': 'Loading conversations...',
    'chat.loading_messages': 'Loading messages...',
    'chat.empty_conversations': 'No conversations yet.',
    'chat.empty_contacts': 'No active contacts available.',
    'chat.click_to_open': 'Click to open chat',
    'chat.select_conversation': 'Select a conversation',
    'chat.select_conversation_desc': 'Choose a conversation on the left to start messaging.',
    'chat.welcome_title': 'Welcome to Chat',
    'chat.welcome_desc': 'Select an existing conversation or start a new one to begin.',
    'chat.typing': 'Typing...',
    'chat.input_placeholder': 'Type a message...',
    'chat.switch_chat': 'Switch chat',
    'chat.press_enter': 'Press Enter to send, Shift + Enter for a new line.',
    'chat.no_new_messages': 'No new messages.',

    // Notifications
    'notif.title': 'Notification Center',
    'notif.all': 'All',
    'notif.unread': 'Unread',
    'notif.mark_all': 'Mark all as read',
    'notif.marking': 'Marking...',
    'notif.loading': 'Loading notifications...',
    'notif.empty': 'No notifications at this time.',
    'notif.just_now': 'Just now',
    'notif.mins_ago': 'mins ago',
    'notif.hours_ago': 'hours ago',
    'notif.yesterday': 'Yesterday',
    'notif.days_ago': 'days ago',

    // Feedback
    'feedback.saved': 'Settings successfully updated and applied!',
    'feedback.reset': 'All preferences have been restored to defaults.',
    'feedback.cache_cleared': 'Local application cache successfully cleared!',
    'feedback.exported': 'Settings configuration file successfully downloaded.',
  }
}

const state = reactive<AppSettingsState>({ ...defaultSettings })
let isInitialized = false

export const useAppSettings = () => {
  const applyThemeToDOM = () => {
    if (typeof window === 'undefined' || typeof document === 'undefined') return

    let activeTheme = state.theme
    if (activeTheme === 'system') {
      const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
      activeTheme = prefersDark ? 'dark' : 'light'
    }

    document.documentElement.setAttribute('data-theme', activeTheme)
    document.documentElement.setAttribute('data-font-size', state.fontSize)
    document.documentElement.setAttribute('data-density', state.density)
    document.documentElement.setAttribute('lang', state.language)

    if (state.reduceMotion) {
      document.documentElement.setAttribute('data-reduce-motion', 'true')
    } else {
      document.documentElement.removeAttribute('data-reduce-motion')
    }
  }

  const loadSettings = () => {
    if (typeof window === 'undefined') return
    try {
      const saved = localStorage.getItem(STORAGE_KEY)
      if (saved) {
        const parsed = JSON.parse(saved)
        Object.assign(state, { ...defaultSettings, ...parsed })
      }
    } catch {
      // Ignore parse errors and use default
    }
    applyThemeToDOM()
  }

  const saveSettings = () => {
    if (typeof window === 'undefined') return
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
    } catch (e) {
      console.warn('Gagal menyimpan setelan ke localStorage:', e)
    }
    applyThemeToDOM()
  }

  const resetToDefault = () => {
    Object.assign(state, defaultSettings)
    saveSettings()
  }

  // Reactive translation function
  const t = (key: string, fallback?: string): string => {
    const lang = state.language || 'id'
    const dict = translations[lang] || translations.id
    return dict[key] || fallback || key
  }

  // Presence Status Helpers
  const presenceColor = computed(() => {
    switch (state.presenceStatus) {
      case 'online': return '#10b981' // emerald green
      case 'busy': return '#f59e0b' // amber
      case 'dnd': return '#ef4444' // red
      case 'invisible': return '#94a3b8' // slate gray
      default: return '#10b981'
    }
  })

  const presenceLabel = computed(() => {
    const isEn = state.language === 'en'
    switch (state.presenceStatus) {
      case 'online': return isEn ? 'Online' : 'Aktif'
      case 'busy': return isEn ? 'Busy / In Class' : 'Sedang Fokus / Sibuk'
      case 'dnd': return isEn ? 'Do Not Disturb' : 'Jangan Ganggu'
      case 'invisible': return isEn ? 'Invisible' : 'Tak Terlihat'
      default: return 'Online'
    }
  })

  // Format Grade Score / Letter
  const formatGrade = (score: number | string): string => {
    const num = Number(score)
    if (isNaN(num)) return String(score || '-')
    if (state.gradeFormat === 'letter') {
      if (num >= 85) return `A (${num})`
      if (num >= 75) return `B (${num})`
      if (num >= 60) return `C (${num})`
      if (num >= 50) return `D (${num})`
      return `E (${num})`
    }
    return String(num)
  }

  // Request native browser Push Notification permission
  const requestPushPermission = async (): Promise<boolean> => {
    if (typeof window === 'undefined' || !('Notification' in window)) return false
    try {
      const perm = await Notification.requestPermission()
      state.notifyPush = perm === 'granted'
      saveSettings()
      return perm === 'granted'
    } catch {
      return false
    }
  }

  // Show native desktop notification
  const showNativeNotification = (title: string, body?: string) => {
    if (typeof window === 'undefined' || !('Notification' in window)) return
    if (state.notifyPush && Notification.permission === 'granted') {
      try {
        new Notification(title, {
          body: body || '',
          icon: '/images/logo.png',
        })
      } catch (err) {
        console.warn('Native notification failed:', err)
      }
    }
  }

  // Web Audio chime generator (preview and real notification audio)
  const playNotificationChime = () => {
    if (typeof window === 'undefined') return
    try {
      const AudioContextClass = window.AudioContext || (window as any).webkitAudioContext
      if (!AudioContextClass) return

      const ctx = new AudioContextClass()
      const now = ctx.currentTime
      const volumeLevel = (state.soundVolume / 100) * 0.25

      // Pleasant double bell chime (E5 -> B5)
      const osc1 = ctx.createOscillator()
      const osc2 = ctx.createOscillator()
      const gain = ctx.createGain()

      osc1.type = 'sine'
      osc1.frequency.setValueAtTime(659.25, now) // E5
      osc1.frequency.exponentialRampToValueAtTime(880.0, now + 0.12) // A5

      osc2.type = 'triangle'
      osc2.frequency.setValueAtTime(1318.5, now + 0.1) // E6

      gain.gain.setValueAtTime(0, now)
      gain.gain.linearRampToValueAtTime(volumeLevel, now + 0.04)
      gain.gain.exponentialRampToValueAtTime(0.001, now + 0.55)

      osc1.connect(gain)
      osc2.connect(gain)
      gain.connect(ctx.destination)

      osc1.start(now)
      osc2.start(now + 0.08)
      osc1.stop(now + 0.55)
      osc2.stop(now + 0.55)

      setTimeout(() => {
        try { ctx.close() } catch {}
      }, 700)
    } catch (err) {
      console.warn('Tidak dapat memutar audio chime:', err)
    }
  }

  const clearAppCache = async () => {
    if (typeof window === 'undefined') return
    try {
      const token = localStorage.getItem('token')
      const user = localStorage.getItem('user')
      sessionStorage.clear()
      
      if (token) localStorage.setItem('token', token)
      if (user) localStorage.setItem('user', user)

      return true
    } catch {
      return false
    }
  }

  if (typeof window !== 'undefined' && !isInitialized) {
    isInitialized = true
    loadSettings()

    if (window.matchMedia) {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      mediaQuery.addEventListener('change', () => {
        if (state.theme === 'system') {
          applyThemeToDOM()
        }
      })
    }

    watch(
      () => ({ ...state }),
      () => {
        saveSettings()
      },
      { deep: true }
    )
  }

  return {
    settings: state,
    t,
    presenceColor,
    presenceLabel,
    formatGrade,
    requestPushPermission,
    showNativeNotification,
    applyThemeToDOM,
    loadSettings,
    saveSettings,
    resetToDefault,
    playNotificationChime,
    clearAppCache,
  }
}
