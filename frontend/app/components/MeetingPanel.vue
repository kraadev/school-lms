<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  BellRing,
  Check,
  CheckCheck,
  CheckCircle2,
  Clock3,
  GraduationCap,
  MonitorPlay,
  Play,
  Plus,
  RefreshCw,
  Trash2,
  UserCheck,
  UserRoundX,
  Video,
  X,
} from 'lucide-vue-next'
import { useApi } from '~/composables/useApi'
import { useAuthStore } from '~/stores/auth'

const auth = useAuthStore()
const { request } = useApi()

const meetings = ref<any[]>([])
const pendingRequests = ref<Record<string, any[]>>({})
const classSubjects = ref<any[]>([]);

const title = ref('')
const type = ref('meeting')
const classSubjectId = ref('')
const approvalRequired = ref(true)
const maxParticipants = ref(50)
const loading = ref(false)
const message = ref('')
const messageType = ref<'success' | 'error'>('success')

const meetingId = (m: any) => m.id || m.ID
const relationValue = (item: any, ...paths: string[]) => {
  for (const path of paths) {
    const value = path.split('.').reduce((current, key) => current?.[key], item)
    if (value !== undefined && value !== null && value !== '') return value
  }
  return ''
}

const meetingContext = (meeting: any) => {
  const className = relationValue(
    meeting,
    'class.name',
    'class.Name',
    'Class.name',
    'Class.Name',
    'class_subject.class.name',
    'class_subject.class.Name',
    'class_subject.Class.name',
    'class_subject.Class.Name',
    'ClassSubject.class.name',
    'ClassSubject.class.Name',
    'ClassSubject.Class.name',
    'ClassSubject.Class.Name',
    'class_name',
    'ClassName',
  ) || 'Kelas'
  const subjectName = relationValue(
    meeting,
    'subject.name',
    'subject.Name',
    'Subject.name',
    'Subject.Name',
    'class_subject.subject.name',
    'class_subject.subject.Name',
    'class_subject.Subject.name',
    'class_subject.Subject.Name',
    'ClassSubject.subject.name',
    'ClassSubject.subject.Name',
    'ClassSubject.Subject.name',
    'ClassSubject.Subject.Name',
    'subject_name',
    'SubjectName',
  ) || 'Mapel'

  return `${className} - ${subjectName}`
}

const selectedClassSubject = computed(() =>
  classSubjects.value.find((item) => String(item.id || item.ID) === String(classSubjectId.value)),
)

const selectedClassName = computed(() =>
  relationValue(selectedClassSubject.value, 'class.name', 'Class.Name') || 'Kelas'
)

const selectedSubjectName = computed(() =>
  relationValue(selectedClassSubject.value, 'subject.name', 'Subject.Name') || 'Mata Pelajaran'
)

const meetingDescription = computed(() =>
  selectedClassSubject.value
    ? `Ruang pembelajaran ${selectedClassName.value} untuk mata pelajaran ${selectedSubjectName.value}.`
    : '',
)

const load = async () => {
  loading.value = true
  try {
    const res = await request<{ data?: any[] }>(auth.role === 'admin' ? '/admin/meetings' : '/meetings')
    meetings.value = res.data || []
  } catch (error: any) {
    message.value = error?.data?.message || 'Meeting belum dapat dimuat.'
    messageType.value = 'error'
  } finally {
    loading.value = false
    if (auth.role === 'guru') {
      const entries = await Promise.all(
        meetings.value.map(async (m) => {
          try {
            const reqs = await request<{ data?: any[] }>(`/meetings/${meetingId(m)}/requests`)
            return [meetingId(m), reqs.data || []] as const
          } catch {
            return [meetingId(m), []] as const
          }
        })
      )
      pendingRequests.value = Object.fromEntries(entries)
    }
  }
}

const loadClassSubjects = async () => {
  if (auth.role !== 'guru') return
  try {
    const res = await request<{ data?: any[] }>('/teacher/class-subjects')
    classSubjects.value = res.data || []
  } catch (error: any) {
    console.warn('Could not load class subjects:', error)
  }
}

const create = async () => {
  if (!title.value.trim() || !classSubjectId.value) return
  try {
    await request('/meetings', {
      method: 'POST',
      body: {
        title: title.value,
        description: meetingDescription.value,
        type: type.value,
        class_subject_id: classSubjectId.value,
        approval_required: approvalRequired.value,
        max_participants: Number(maxParticipants.value) || 50,
      },
    })
    title.value = ''
    classSubjectId.value = ''
    message.value = 'Ruang meeting kelas berhasil dibuat.'
    messageType.value = 'success'
    await load()
  } catch (error: any) {
    message.value = error?.data?.message || 'Meeting belum dapat dibuat.'
    messageType.value = 'error'
  }
}

const runAction = async (path: string, successText: string) => {
  try {
    await request(path, { method: 'POST' })
    message.value = successText
    messageType.value = 'success'
    await load()
  } catch (error: any) {
    message.value = error?.data?.message || 'Aksi meeting gagal.'
    messageType.value = 'error'
  }
}

const deleteMeeting = async (id: string) => {
  if (!confirm('Hapus meeting ini?')) return
  try {
    const endpoint = auth.role === 'admin' ? `/admin/meetings/${id}` : `/meetings/${id}`
    await request(endpoint, { method: 'DELETE' })
    message.value = 'Meeting berhasil dihapus.'
    messageType.value = 'success'
    await load()
  } catch (error: any) {
    message.value = error?.data?.message || 'Gagal menghapus meeting.'
    messageType.value = 'error'
  }
}

const decideRequest = async (meeting: any, participant: any, status: 'approved' | 'rejected') => {
  try {
    const pId = participant.user_id || participant.UserID
    await request(`/meetings/${meetingId(meeting)}/participants/${pId}`, {
      method: 'PATCH',
      body: { status },
    })
    message.value = status === 'approved' ? 'Siswa disetujui masuk meeting.' : 'Pengajuan siswa ditolak.'
    messageType.value = 'success'
    await load()
  } catch (error: any) {
    message.value = error?.data?.message || 'Keputusan peserta belum dapat disimpan.'
    messageType.value = 'error'
  }
}

const decideAllRequests = async (meeting: any, status: 'approved' | 'rejected') => {
  try {
    await request(`/meetings/${meetingId(meeting)}/requests/decide-all`, {
      method: 'PATCH',
      body: { status },
    })
    message.value = status === 'approved' ? 'Semua siswa disetujui masuk meeting.' : 'Semua pengajuan siswa ditolak.'
    messageType.value = 'success'
    await load()
  } catch (error: any) {
    // Fallback: decide individually if bulk endpoint not yet available
    const reqs = pendingRequests.value[meetingId(meeting)] || []
    if (reqs.length > 0) {
      await Promise.all(
        reqs.map((p: any) =>
          request(`/meetings/${meetingId(meeting)}/participants/${p.user_id || p.UserID}`, {
            method: 'PATCH',
            body: { status },
          }).catch(() => {})
        )
      )
      message.value = status === 'approved' ? 'Semua siswa disetujui masuk meeting.' : 'Semua pengajuan siswa ditolak.'
      messageType.value = 'success'
      await load()
      return
    }
    message.value = error?.data?.message || 'Keputusan peserta belum dapat disimpan.'
    messageType.value = 'error'
  }
}

const getStudentName = (p: any) => {
  const user = p?.user || p?.User || p
  const student = user?.student || user?.Student
  const teacher = user?.teacher || user?.Teacher
  return (
    student?.full_name ||
    student?.FullName ||
    teacher?.full_name ||
    teacher?.FullName ||
    user?.display_name ||
    user?.DisplayName ||
    user?.name ||
    user?.Name ||
    user?.username ||
    user?.Username ||
    'Siswa'
  )
}

const getStudentClass = (p: any, meetingItem?: any) => {
  const user = p?.user || p?.User || p
  const student = user?.student || user?.Student
  const classes = student?.classes || student?.Classes || []
  if (Array.isArray(classes) && classes.length > 0) {
    const names = classes.map((c: any) => c.name || c.Name).filter(Boolean)
    if (names.length > 0) return names.join(', ')
  }
  if (meetingItem) {
    const ctx = meetingContext(meetingItem)
    const split = ctx.split(' - ')
    if (split[0] && split[0].trim() && split[0].trim() !== 'Kelas') {
      return split[0].trim()
    }
  }
  return student?.nis || student?.NIS ? `NIS: ${student.nis || student.NIS}` : 'Siswa'
}

const getStudentNIS = (p: any) => {
  const user = p?.user || p?.User || p
  const student = user?.student || user?.Student
  return student?.nis || student?.NIS || ''
}

const requestAccess = async (meeting: any) => {
  try {
    await request(`/meetings/${meetingId(meeting)}/request`, { method: 'POST' })
    message.value = meeting.approval_required
      ? 'Pengajuan akses terkirim. Tunggu persetujuan guru.'
      : 'Akses diberikan. Silakan masuk ke ruang meeting.'
    messageType.value = 'success'
    await load()
  } catch (error: any) {
    message.value = error?.data?.message || 'Akses meeting belum dapat diproses.'
    messageType.value = 'error'
  }
}

const openMeeting = async (meeting: any) => {
  await navigateTo(`/meeting/${meetingId(meeting)}`)
}

onMounted(async () => {
  await Promise.all([load(), loadClassSubjects()])
})
</script>

<template>
  <div class="meeting-panel">
    <header class="meeting-header">
      <div>
        <div class="header-title-wrap">
          <span class="header-icon"><Video :size="18" /></span>
          <h2>Ruang Meeting & Kelas Real-Time</h2>
        </div>
        <p>
          Komunikasi audio, video, dan screen sharing bertenaga LiveKit Media Server.
        </p>
      </div>

      <button type="button" class="header-button" @click="load">
        <RefreshCw :size="15" />
        Refresh
      </button>
    </header>

    <div v-if="message" class="alert" :class="messageType === 'success' ? 'success' : 'error'">
      <span>{{ message }}</span>
      <button type="button" class="alert-close" @click="message = ''">
        <X :size="14" />
      </button>
    </div>

    <form v-if="auth.role === 'guru'" class="meeting-form" @submit.prevent="create">
      <div class="form-header">
        <Plus :size="15" />
        Buat Ruang Kelas Baru
      </div>

      <div class="form-grid">
        <input v-model="title" required placeholder="Judul / Topik Meeting..." aria-label="Judul atau topik meeting" />

        <select v-model="classSubjectId" required aria-label="Kelas dan mata pelajaran">
          <option value="">Pilih Kelas &amp; Mapel</option>
          <option v-for="item in classSubjects" :key="item.id || item.ID" :value="item.id || item.ID">
            {{ meetingContext(item) }}
          </option>
        </select>

        <div v-if="selectedClassSubject" class="selected-context">
          <span class="context-label">Terhubung ke</span>
          <strong>{{ selectedClassName }} · {{ selectedSubjectName }}</strong>
          <small>{{ meetingDescription }}</small>
        </div>

        <select v-model="type" aria-label="Tipe meeting">
          <option value="meeting">Video Meeting Lengkap</option>
          <option value="voice">Voice Room Saja</option>
        </select>

        <input v-model="maxParticipants" type="number" min="2" max="500" placeholder="Max Peserta (50)" aria-label="Maksimal peserta" />
      </div>

      <div class="form-footer">
        <label class="checkbox-wrap">
          <input v-model="approvalRequired" type="checkbox" />
          <span>Perlu persetujuan guru sebelum siswa masuk (Lobi)</span>
        </label>

        <button type="submit" class="primary-button">
          <Plus :size="15" />
          Buat Ruang Meeting
        </button>
      </div>
    </form>

    <section class="meeting-list-wrap">
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner" />
        <span>Memuat daftar meeting...</span>
      </div>

      <div v-else-if="meetings.length" class="meeting-grid">
        <article
          v-for="m in meetings"
          :key="meetingId(m)"
          class="meeting-card"
          :class="m.status === 'live' || m.status === 'active' ? 'is-live' : ''"
        >
          <div class="meeting-card-top">
            <span class="status-badge" :class="m.status === 'live' || m.status === 'active' ? 'status-live' : m.status === 'scheduled' ? 'status-scheduled' : 'status-idle'">
              {{ m.type }} · {{ m.status }}
            </span>
            <span class="cap-label">Max {{ m.max_participants || 50 }} peserta</span>
          </div>

          <h3 :title="m.title">{{ m.title }}</h3>

          <p class="meeting-meta">
            {{ meetingContext(m) }}
          </p>

          <p v-if="m.description" class="meeting-description">{{ m.description }}</p>

          <div
            v-if="auth.role === 'guru' && pendingRequests[meetingId(m)]?.length"
            class="request-box"
          >
            <div class="request-header">
              <div class="request-header-left">
                <span>Pengajuan Akses Siswa</span>
                <span class="request-count">{{ pendingRequests[meetingId(m)].length }}</span>
              </div>
              <div class="request-header-actions">
                <button
                  type="button"
                  class="mini-success bulk-btn"
                  title="Izinkan semua pengajuan"
                  @click="decideAllRequests(m, 'approved')"
                >
                  <CheckCheck :size="12" />
                  Izinkan Semua
                </button>
                <button
                  type="button"
                  class="mini-danger bulk-btn"
                  title="Tolak semua pengajuan"
                  @click="decideAllRequests(m, 'rejected')"
                >
                  <X :size="12" />
                  Tolak Semua
                </button>
              </div>
            </div>

            <div v-for="p in pendingRequests[meetingId(m)]" :key="p.id || p.ID" class="request-item">
              <div class="request-user-info">
                <div class="request-user-name-wrap">
                  <span class="request-user-name">{{ getStudentName(p) }}</span>
                  <span class="request-user-badge">
                    <GraduationCap :size="11" />
                    {{ getStudentClass(p, m) }}
                  </span>
                </div>
                <span v-if="getStudentNIS(p)" class="request-user-nis">NIS: {{ getStudentNIS(p) }}</span>
              </div>
              <div class="request-actions">
                <button type="button" class="mini-success" @click="decideRequest(m, p, 'approved')">
                  <Check :size="12" />
                  Izinkan
                </button>
                <button type="button" class="mini-danger" @click="decideRequest(m, p, 'rejected')">
                  <UserRoundX :size="12" />
                  Tolak
                </button>
              </div>
            </div>
          </div>

          <div class="meeting-actions">
            <div class="left-actions">
              <template v-if="auth.role === 'siswa'">
                <button
                  v-if="m.status === 'scheduled'"
                  type="button"
                  class="secondary-button"
                  @click="requestAccess(m)"
                >
                  {{ m.approval_required ? 'Ajukan Izin Masuk' : 'Daftar Meeting' }}
                </button>

                <button
                  v-if="m.status === 'live' || m.status === 'active'"
                  type="button"
                  class="primary-button compact"
                  @click="openMeeting(m)"
                >
                  <Play :size="14" />
                  Masuk Ruang
                </button>
              </template>

              <template v-if="auth.role === 'guru'">
                <button
                  v-if="m.status === 'scheduled'"
                  type="button"
                  class="primary-button compact"
                  @click="runAction(`/meetings/${meetingId(m)}/start`, 'Meeting telah dimulai.')"
                >
                  <MonitorPlay :size="14" />
                  Mulai
                </button>

                <button
                  v-if="m.status === 'live' || m.status === 'active'"
                  type="button"
                  class="primary-button compact"
                  @click="openMeeting(m)"
                >
                  <Video :size="14" />
                  Buka Ruang
                </button>

                <button
                  v-if="m.status === 'live' || m.status === 'active'"
                  type="button"
                  class="danger-button compact"
                  @click="runAction(`/meetings/${meetingId(m)}/end`, 'Meeting telah diakhiri.')"
                >
                  <Clock3 :size="14" />
                  Akhiri
                </button>
              </template>

              <template v-if="auth.role === 'admin'">
                <button
                  v-if="m.status === 'live' || m.status === 'active'"
                  type="button"
                  class="secondary-button"
                  @click="openMeeting(m)"
                >
                  <MonitorPlay :size="14" />
                  Pantau
                </button>
              </template>
            </div>

            <button
              v-if="auth.role === 'guru' || auth.role === 'admin'"
              type="button"
              class="trash-button"
              title="Hapus Meeting"
              @click="deleteMeeting(meetingId(m))"
            >
              <Trash2 :size="14" />
            </button>
          </div>
        </article>
      </div>

      <div v-else class="empty-state">
        <span class="empty-icon"><BellRing :size="28" /></span>
        <h3>Belum Ada Ruang Meeting</h3>
        <p>
          {{
            auth.role === 'guru'
              ? 'Buat ruang kelas baru untuk memulai pembelajaran realtime melalui video, audio, dan presentasi layar.'
              : 'Belum ada meeting yang dijadwalkan untuk kelas Anda saat ini.'
          }}
        </p>
      </div>
    </section>
  </div>
</template>

<style scoped>
.meeting-panel {
  display: grid;
  gap: 18px;
}

.meeting-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 24px 26px;
  border-radius: 24px;
  background: linear-gradient(135deg, #1e2b36 0%, #243a39 100%);
  border: 1px solid rgba(93, 112, 128, 0.2);
  box-shadow: 0 16px 30px rgba(17, 24, 39, 0.12);
  color: #f7faf9;
}

.header-title-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.header-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 12px;
  background: rgba(101, 201, 171, 0.18);
  border: 1px solid rgba(101, 201, 171, 0.32);
  color: white;
}

.meeting-header h2 {
  margin: 0;
  font-size: 1.08rem;
  font-weight: 800;
  color: white;
}

.meeting-header p {
  margin: 0;
  color: rgba(255, 255, 255, 0.72);
  font-size: 0.78rem;
  line-height: 1.6;
}

.header-button,
.primary-button,
.secondary-button,
.danger-button,
.mini-success,
.mini-danger,
.trash-button,
.alert-close {
  font: inherit;
}

.header-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 12px;
  padding: 10px 14px;
  background: rgba(255, 255, 255, 0.05);
  color: white;
  cursor: pointer;
  font-size: 0.76rem;
  font-weight: 700;
}

.alert {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 14px;
  font-size: 0.8rem;
  font-weight: 700;
}

.alert.success {
  background: #ecfdf5;
  border: 1px solid #a7f3d0;
  color: #047857;
}

.alert.error {
  background: #fff1f1;
  border: 1px solid #fecaca;
  color: #b42318;
}

.alert-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: inherit;
  cursor: pointer;
  padding: 4px;
}

.meeting-form {
  display: grid;
  gap: 16px;
  padding: 22px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid var(--line);
  box-shadow: 0 10px 28px rgba(20, 28, 38, 0.04);
}

.form-header {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--ink);
  font-size: 0.8rem;
  font-weight: 800;
  letter-spacing: 0.02em;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.form-grid input,
.form-grid select {
  width: 100%;
  padding: 12px 14px;
  border-radius: 12px;
  border: 1px solid #dfe7e1;
  background: white;
  color: var(--ink);
  font-size: 0.8rem;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.form-grid input:focus,
.form-grid select:focus {
  outline: none;
  border-color: rgba(101, 201, 171, 0.8);
  box-shadow: 0 0 0 4px rgba(101, 201, 171, 0.1);
}

.selected-context {
  display: grid;
  align-content: center;
  gap: 3px;
  grid-column: span 2;
  min-height: 46px;
  padding: 9px 12px;
  border: 1px solid #cfe8da;
  border-radius: 12px;
  background: #f1faf4;
  color: #245b3d;
}

.selected-context .context-label {
  color: #5b8a6d;
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.selected-context strong {
  font-size: 0.78rem;
}

.selected-context small {
  color: #5b8067;
  font-size: 0.68rem;
  line-height: 1.35;
}

.form-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  flex-wrap: wrap;
}

.checkbox-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #4d5f6c;
  font-size: 0.77rem;
  font-weight: 700;
}

.checkbox-wrap input {
  width: 16px;
  height: 16px;
  accent-color: var(--teal);
}

.primary-button,
.secondary-button,
.danger-button,
.mini-success,
.mini-danger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  border: none;
  border-radius: 12px;
  padding: 10px 14px;
  cursor: pointer;
  font-size: 0.77rem;
  font-weight: 800;
  transition: transform 0.15s ease, opacity 0.2s ease, box-shadow 0.2s ease;
}

.primary-button svg,
.secondary-button svg,
.danger-button svg,
.mini-success svg,
.mini-danger svg,
.header-button svg,
.alert-close svg,
.form-header svg,
.empty-icon svg,
.trash-button svg {
  flex-shrink: 0;
}

.primary-button:hover,
.secondary-button:hover,
.danger-button:hover,
.mini-success:hover,
.mini-danger:hover {
  transform: translateY(-1px);
}

.primary-button {
  background: linear-gradient(135deg, var(--teal) 0%, #53b398 100%);
  color: white;
  box-shadow: 0 12px 22px rgba(101, 201, 171, 0.22);
}

.secondary-button {
  background: #f4f8f6;
  color: var(--ink);
  border: 1px solid #dfe7e1;
}

.danger-button {
  background: #fff0ec;
  color: #b45242;
  border: 1px solid #f8d5cf;
}

.compact {
  padding: 9px 12px;
}

.meeting-list-wrap {
  display: grid;
}

.loading-state,
.empty-state {
  display: grid;
  place-items: center;
  min-height: 220px;
  padding: 32px 18px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.72);
  border: 1px solid var(--line);
  text-align: center;
  color: #65727d;
}

.loading-spinner {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: 3px solid rgba(101, 201, 171, 0.2);
  border-top-color: var(--teal);
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.meeting-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 16px;
}

.meeting-card {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 14px;
  padding: 18px 18px 16px;
  background: rgba(255, 255, 255, 0.95);
  border: 1px solid var(--line);
  border-radius: 22px;
  box-shadow: 0 8px 26px rgba(9, 19, 30, 0.05);
  transition: transform 0.18s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}

.meeting-card:hover {
  transform: translateY(-2px);
  border-color: rgba(101, 201, 171, 0.35);
  box-shadow: 0 14px 26px rgba(25, 41, 52, 0.07);
}

.meeting-card.is-live {
  border-color: rgba(16, 185, 129, 0.4);
  background: linear-gradient(180deg, rgba(236, 253, 245, 0.95), rgba(255, 255, 255, 0.96));
}

.meeting-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px 8px;
  border-radius: 999px;
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.status-live {
  background: rgba(16, 185, 129, 0.14);
  color: #047857;
  border: 1px solid rgba(16, 185, 129, 0.25);
}

.status-scheduled {
  background: rgba(245, 158, 11, 0.12);
  color: #b45309;
  border: 1px solid rgba(245, 158, 11, 0.22);
}

.status-idle {
  background: #eef2f4;
  color: #4a5964;
  border: 1px solid #e2e8ec;
}

.cap-label {
  color: #67737e;
  font-size: 0.68rem;
  font-weight: 700;
}

.meeting-card h3 {
  margin: 0;
  color: var(--ink);
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.35;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.meeting-meta,
.meeting-description {
  margin: 0;
  color: #5d6975;
  line-height: 1.6;
}

.meeting-meta {
  font-size: 0.74rem;
  font-weight: 700;
}

.meeting-description {
  font-size: 0.74rem;
}

.request-box {
  padding: 10px 12px;
  border-radius: 14px;
  background: rgba(245, 158, 11, 0.08);
  border: 1px solid rgba(245, 158, 11, 0.2);
}

.request-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 8px;
  color: #b45309;
  font-size: 0.7rem;
  font-weight: 800;
}

.request-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.request-header-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.bulk-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  font-size: 0.65rem;
  font-weight: 700;
  border-radius: 8px;
  cursor: pointer;
  border: none;
}

.request-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 22px;
  height: 22px;
  border-radius: 999px;
  background: rgba(245, 158, 11, 0.16);
}

.request-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: 8px;
  padding: 6px 8px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid rgba(245, 158, 11, 0.15);
  color: var(--ink);
  font-size: 0.72rem;
  font-weight: 700;
}

.request-user-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.request-user-name-wrap {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.request-user-name {
  font-weight: 700;
  color: var(--ink);
  font-size: 0.76rem;
}

.request-user-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 2px 6px;
  border-radius: 6px;
  background: rgba(16, 185, 129, 0.12);
  color: #065f46;
  font-size: 0.65rem;
  font-weight: 700;
}

.request-user-nis {
  font-size: 0.65rem;
  color: #64748b;
  font-weight: 500;
}

.request-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.mini-success,
.mini-danger {
  padding: 7px 8px;
  border-radius: 9px;
  font-size: 0.66rem;
}

.mini-success,
.mini-danger {
  gap: 5px;
}

.mini-success {
  background: #10b981;
  color: white;
}

.mini-danger {
  background: #ef4444;
  color: white;
}

.meeting-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding-top: 12px;
  border-top: 1px solid rgba(148, 163, 184, 0.18);
}

.left-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.trash-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 10px;
  background: #f8fafb;
  color: #475569;
  cursor: pointer;
}

.empty-state h3 {
  margin: 0;
  color: var(--ink);
  font-size: 1.02rem;
}

.empty-state p {
  max-width: 420px;
  margin: 0;
  color: #5d6975;
  font-size: 0.76rem;
  line-height: 1.7;
}

.empty-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--teal);
  margin-bottom: 8px;
}

@media (max-width: 980px) {
  .form-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .selected-context {
    grid-column: 1 / -1;
  }
}

@media (max-width: 640px) {
  .meeting-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .selected-context {
    grid-column: auto;
  }

  .form-footer {
    align-items: flex-start;
    flex-direction: column;
  }

  .meeting-actions {
    align-items: flex-start;
    flex-direction: column;
  }

  .left-actions {
    width: 100%;
  }

  .left-actions > * {
    flex: 1 1 auto;
  }
}
</style>
