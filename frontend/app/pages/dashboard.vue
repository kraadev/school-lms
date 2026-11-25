<script setup lang="ts">
import { 
  FileText, 
  PenTool, 
  Book, 
  Puzzle, 
  Library, 
  Bell, 
  Users, 
  GraduationCap, 
  BookOpen, 
  MessageSquare,
  Activity
} from 'lucide-vue-next'

import { useAppSettings } from '~/composables/useAppSettings'

definePageMeta({ middleware: 'auth' })

const auth = useAuthStore()
const { request } = useApi()
const { t, settings } = useAppSettings()
const loading = ref(true)
const errorMessage = ref('')
const data = ref<Record<string, any[] | Record<string, any>>>({})

const roleLabel = computed(() => t('role.' + (auth.role || '')) || auth.role)
const isTeacher = computed(() => auth.role === 'guru')
const isStudent = computed(() => auth.role === 'siswa')

const loadDashboard = async () => {
  loading.value = true
  errorMessage.value = ''
  const paths: Record<string, string[]> = {
    admin: ['/admin/dashboard', '/admin/users', '/admin/classes', '/admin/subjects'],
    guru: ['/guru/dashboard', '/teacher/materials', '/teacher/assignments', '/teacher/quizzes', '/teacher/attendance-sessions'],
    siswa: ['/siswa/dashboard', '/student/materials', '/student/assignments', '/student/quizzes', '/student/attendance-sessions'],
  }
  const keys = ['overview', ...(paths[auth.role] || []).slice(1).map(path => path.split('/').pop() || 'items')]
  const results = await Promise.allSettled((paths[auth.role] || []).map(path => request<{ data?: any }>(path)))
  results.forEach((result, index) => {
    if (result.status === 'fulfilled') data.value[keys[index]] = result.value.data || result.value
  })
  const notifications = await request<{ data?: { unread_count?: number } }>('/notifications/unread-count').catch(() => null)
  if (notifications?.data) data.value.notifications = notifications.data
  if (!results.some(result => result.status === 'fulfilled')) errorMessage.value = 'Data belum dapat dimuat. Pastikan backend sedang berjalan.'
  loading.value = false
}

const list = (key: string) => Array.isArray(data.value[key]) ? data.value[key] as any[] : []
const count = (key: string) => list(key).length
const dateLabel = (value: string) => value ? new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short' }).format(new Date(value)) : '-'

onMounted(loadDashboard)
</script>

<template>
  <AppShell active="dashboard" :eyebrow="`${t('dashboard.workspace')} ${roleLabel}`" :title="`${t('dashboard.greeting')} ${auth.user?.username}.`">
      <div v-if="loading" class="loading-state"><span class="loader" /> {{ t('dashboard.loading') }}</div>
      <template v-else>
        <div v-if="errorMessage" class="alert">{{ errorMessage }}</div>
        
        <div class="hero-wrapper">
          <div class="hero-section">
            <div class="hero-content">
              <h1>{{ t('dashboard.hero_title') }}</h1>
              <div class="hero-actions-row">
                <button class="primary-btn" @click="navigateTo('/learning')">{{ t('dashboard.start_learning') }}</button>
                <div class="hero-stats-group">
                  <div class="hero-stat-pill">
                    <span class="pill-icon coral"><Users :size="15" /></span>
                    <div>
                      <strong>{{ isTeacher ? count('materials') : count('materials') }}+</strong>
                      <small>{{ isTeacher ? t('dashboard.stat_materials_created') : t('dashboard.stat_materials_ready') }}</small>
                    </div>
                  </div>
                  <div class="hero-stat-pill">
                    <span class="pill-icon gold"><GraduationCap :size="15" /></span>
                    <div>
                      <strong>{{ isTeacher ? count('assignments') : count('assignments') }}+</strong>
                      <small>{{ t('dashboard.stat_assignments_active') }}</small>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="hero-image-overlay">
            <img src="/images/hero_illustration.png" alt="Online Learning" />
          </div>
        </div>

        <section class="stat-grid">
          <article class="stat-card">
            <span class="stat-icon stat-icon-blue">
              <FileText :size="24" color="white" v-if="isTeacher || isStudent" />
              <Users :size="24" v-else />
            </span>
            <div>
              <strong>{{ isTeacher ? count('materials') : isStudent ? count('materials') : count('users') }}</strong>
              <span>{{ isTeacher ? t('dashboard.stat_materials_created') : isStudent ? t('dashboard.stat_materials_ready') : t('dashboard.stat_total_users') }}</span>
            </div>
          </article>
          <article class="stat-card">
            <span class="stat-icon stat-icon-coral">
              <PenTool :size="24" color="white" v-if="isTeacher || isStudent" />
              <Book :size="24" v-else />
            </span>
            <div>
              <strong>{{ isTeacher || isStudent ? count('assignments') : count('classes') }}</strong>
              <span>{{ isTeacher || isStudent ? t('dashboard.stat_assignments_active') : t('dashboard.stat_classes_enrolled') }}</span>
            </div>
          </article>
          <article class="stat-card">
            <span class="stat-icon stat-icon-gold">
              <Puzzle :size="24" color="white" v-if="isTeacher || isStudent" />
              <Library :size="24" v-else />
            </span>
            <div>
              <strong>{{ isTeacher || isStudent ? count('quizzes') : count('subjects') }}</strong>
              <span>{{ isTeacher || isStudent ? t('dashboard.stat_quizzes_ready') : t('dashboard.stat_subjects') }}</span>
            </div>
          </article>
          <article class="stat-card">
            <span class="stat-icon stat-icon-green">
              <Bell :size="24" color="white" />
            </span>
            <div>
              <strong>{{ data.notifications?.unread_count || 0 }}</strong>
              <span>{{ t('nav.notifications') }}</span>
            </div>
          </article>
        </section>

        <section class="dashboard-grid">
          <article class="panel wide-panel">
            <div class="panel-heading">
              <div>
                <h3>{{ isStudent ? (settings.language === 'en' ? 'Recent Assignments' : 'Tugas terbaru') : isTeacher ? (settings.language === 'en' ? 'Recent Materials' : 'Materi terbaru') : (settings.language === 'en' ? 'Recent Users' : 'Pengguna terbaru') }}</h3>
              </div>
              <NuxtLink class="text-link" to="/learning">{{ settings.language === 'en' ? 'View all' : 'Lihat semua' }}</NuxtLink>
            </div>
            <div v-if="!list(isStudent || isTeacher ? (isStudent ? 'assignments' : 'materials') : 'users').length" class="empty-state">
              {{ settings.language === 'en' ? 'No data available to display.' : 'Belum ada data untuk ditampilkan.' }}
            </div>
            <div class="course-list">
              <div v-for="(item, index) in list(isStudent || isTeacher ? (isStudent ? 'assignments' : 'materials') : 'users').slice(0, 5)" :key="item.id || index" class="course-item">
                <div class="course-icon">
                  <Activity :size="20" style="color: var(--muted); margin: 14px;" />
                </div>
                <div class="course-info">
                  <strong>{{ item.title || item.username || (settings.language === 'en' ? 'New Activity' : 'Aktivitas baru') }}</strong>
                  <p>{{ item.description || item.email || (item.deadline ? `${settings.language === 'en' ? 'Due' : 'Tenggat'} ${dateLabel(item.deadline)}` : (settings.language === 'en' ? 'Recently updated' : 'Diperbarui baru-baru ini')) }}</p>
                </div>
                <span class="row-date">{{ dateLabel(item.created_at || item.deadline) }}</span>
              </div>
            </div>
          </article>
          
          <article class="panel side-panel">
            <div class="panel-heading">
              <div>
                <h3>{{ settings.language === 'en' ? 'Quick Access' : 'Akses Cepat' }}</h3>
              </div>
            </div>
            <div class="quick-list">
              <NuxtLink v-if="auth.role === 'admin'" to="/admin">
                <span class="icon"><Users :size="20" /></span>
                <div>
                  <strong>{{ t('nav.admin') }}</strong>
                  <small>{{ settings.language === 'en' ? 'Members & class structure' : 'Anggota dan struktur kelas' }}</small>
                </div>
              </NuxtLink>
              <NuxtLink to="/academic">
                <span class="icon"><GraduationCap :size="20" /></span>
                <div>
                  <strong>{{ auth.role === 'admin' ? t('nav.academic') : t('nav.academic_year') }}</strong>
                  <small>{{ auth.role === 'admin' ? (settings.language === 'en' ? 'Academic years, classes & calendar' : 'Tahun ajaran, kelas & timeline') : (settings.language === 'en' ? 'Academic calendar & schedule' : 'Kalender akademik & agenda sekolah') }}</small>
                </div>
              </NuxtLink>
              <NuxtLink to="/learning">
                <span class="icon"><BookOpen :size="20" /></span>
                <div>
                  <strong>{{ auth.role === 'guru' ? t('nav.learning_teacher') : t('nav.learning_student') }}</strong>
                  <small>{{ settings.language === 'en' ? 'Materials, assignments, quizzes, attendance' : 'Materi, tugas, kuis, presensi' }}</small>
                </div>
              </NuxtLink>
              <NuxtLink to="/chat">
                <span class="icon"><MessageSquare :size="20" /></span>
                <div>
                  <strong>{{ t('nav.chat') }}</strong>
                  <small>{{ settings.language === 'en' ? 'Connect with community' : 'Terhubung dengan komunitas' }}</small>
                </div>
              </NuxtLink>
            </div>
          </article>
        </section>
      </template>
  </AppShell>
</template>

