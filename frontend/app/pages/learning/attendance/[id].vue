<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '~/composables/useApi'
import { ArrowLeft, CalendarDays, Check, Clock3, Download, Search, X } from 'lucide-vue-next'

definePageMeta({ middleware: 'auth' })

const route = useRoute()
const { request } = useApi()
const sessionId = route.params.id as string

const session = ref<any>(null)
const records = ref<any[]>([])
const loading = ref(true)
const savingId = ref<string | null>(null)
const togglingState = ref(false)
const message = ref('')
const errorMessage = ref('')
const searchQuery = ref('')
const filterStatus = ref<'all' | 'present' | 'late' | 'excused' | 'absent' | 'not_checked_in'>('all')

const idOf = (item: any) => item?.id || item?.ID || ''
const studentIdOf = (item: any) => item?.student_id || item?.StudentID || item?.student?.id || ''
const nameOf = (item: any) => item?.student_name || item?.StudentName || item?.student?.full_name || item?.Student?.FullName || 'Siswa'
const nisOf = (item: any) => item?.nis || item?.NIS || item?.student?.nis || '-'
const statusOf = (item: any) => item?.status || item?.Status || 'not_checked_in'

const load = async () => {
  try {
    const [sessionResponse, recordResponse] = await Promise.all([
      request<{ data?: any }>(`/teacher/attendance-sessions/${sessionId}`),
      request<{ data?: any[] }>(`/teacher/attendance-sessions/${sessionId}/records`)
    ])
    session.value = sessionResponse.data
    records.value = recordResponse.data || []
  } catch (error: any) {
    errorMessage.value = error?.data?.message || 'Data presensi belum dapat dimuat.'
  } finally {
    loading.value = false
  }
}

const updateStatus = async (record: any, newStatus: string) => {
  const recId = idOf(record) || '00000000-0000-0000-0000-000000000000'
  const studId = studentIdOf(record)
  savingId.value = studId || recId
  try {
    const res = await request<{ data?: any }>(`/teacher/attendance-records/${recId}`, {
      method: 'PATCH',
      body: {
        status: newStatus,
        note: record.note || '',
        session_id: sessionId,
        student_id: studId
      }
    })
    record.status = newStatus
    if (res.data?.id) record.id = res.data.id
    if (res.data?.checked_at) record.checked_at = res.data.checked_at
    message.value = `Status presensi ${nameOf(record)} berhasil diubah.`
    setTimeout(() => { message.value = '' }, 3000)
  } catch (error: any) {
    errorMessage.value = error?.data?.message || 'Presensi belum dapat diperbarui.'
    setTimeout(() => { errorMessage.value = '' }, 4000)
  } finally {
    savingId.value = null
  }
}

const setSessionState = async (state: 'open' | 'close') => {
  togglingState.value = true
  try {
    await request(`/teacher/attendance-sessions/${sessionId}/${state}`, { method: 'POST' })
    message.value = state === 'open' ? 'Sesi presensi berhasil dibuka.' : 'Sesi presensi ditutup.'
    if (session.value) {
      session.value.status = state === 'open' ? 'open' : 'closed'
    }
    setTimeout(() => { message.value = '' }, 3000)
    await load()
  } catch (error: any) {
    errorMessage.value = error?.data?.message || 'Status sesi belum dapat diperbarui.'
    setTimeout(() => { errorMessage.value = '' }, 4000)
  } finally {
    togglingState.value = false
  }
}

const stats = computed(() => {
  const total = records.value.length
  let present = 0
  let late = 0
  let excused = 0
  let absent = 0
  let notCheckedIn = 0

  for (const r of records.value) {
    const s = statusOf(r)
    if (s === 'present') present++
    else if (s === 'late') late++
    else if (s === 'excused') excused++
    else if (s === 'absent') absent++
    else notCheckedIn++
  }

  const attended = present + late
  const rate = total > 0 ? Math.round((attended / total) * 100) : 0

  return { total, present, late, excused, absent, notCheckedIn, rate }
})

const filteredRecords = computed(() => {
  const q = searchQuery.value.toLowerCase().trim()
  return records.value.filter(r => {
    if (filterStatus.value !== 'all') {
      const s = statusOf(r)
      if (filterStatus.value === 'not_checked_in') {
        if (s !== 'not_checked_in') return false
      } else if (s !== filterStatus.value) {
        return false
      }
    }
    if (!q) return true
    const name = nameOf(r).toLowerCase()
    const nis = nisOf(r).toLowerCase()
    return name.includes(q) || nis.includes(q)
  })
})

const getInitials = (name: string) => {
  return name.split(' ').slice(0, 2).map(n => n.charAt(0).toUpperCase()).join('')
}

const formatDateTime = (val?: string) => {
  if (!val) return '-'
  try {
    const d = new Date(val)
    return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
  } catch {
    return val
  }
}

const formatDate = (val?: string) => {
  if (!val) return '-'
  try {
    const d = new Date(val)
    return d.toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
  } catch {
    return val
  }
}

const exportRecap = () => {
  const escapeCsv = (value: any) => `"${String(value ?? '').replaceAll('"', '""')}"`
  const rows = [
    ['No', 'NIS', 'Nama Siswa', 'Status', 'Waktu Presensi', 'Catatan'],
    ...filteredRecords.value.map((record, index) => [
      index + 1,
      nisOf(record),
      nameOf(record),
      statusOf(record) === 'not_checked_in' ? 'Belum Absen' : statusOf(record),
      record.checked_at ? new Date(record.checked_at).toLocaleString('id-ID') : '-',
      record.note || '-',
    ]),
  ]
  const csv = '\uFEFF' + rows.map(row => row.map(escapeCsv).join(';')).join('\r\n')
  const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8' }))
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `rekap-presensi-${session.value?.title || 'sesi'}.csv`
  anchor.click()
  URL.revokeObjectURL(url)
}

onMounted(load)
</script>

<template>
  <AppShell active="learning" eyebrow="School Space" :title="session?.title || session?.Title || 'Presensi Sesi'">
    <!-- Navigation Back -->
    <div class="header-nav">
      <NuxtLink class="back-link" to="/learning?tab=attendance">
        <span class="back-arrow">←</span> Kembali ke Presensi
      </NuxtLink>
    </div>

    <!-- Feedback Alerts -->
    <div v-if="errorMessage" class="alert error-alert">
      <span>{{ errorMessage }}</span>
      <button class="alert-close" type="button" @click="errorMessage = ''">×</button>
    </div>
    <div v-if="message" class="alert success-alert">
      <span><Check :size="15" /> {{ message }}</span>
      <button class="alert-close" type="button" @click="message = ''">×</button>
    </div>

    <div v-if="loading" class="loading-state">
      <span class="loader" /> Memuat data presensi sesi...
    </div>

    <template v-else>
      <!-- Session Detail Hero Card -->
      <section class="session-hero card-panel">
        <div class="hero-main">
          <div class="hero-tag-row">
            <span class="badge" :class="session?.status === 'open' ? 'badge-open' : 'badge-closed'">
              {{ session?.status === 'open' ? 'Sesi Dibuka' : 'Sesi Ditutup' }}
            </span>
            <span class="date-badge">
              <CalendarDays :size="14" /> {{ formatDate(session?.attendance_date || session?.AttendanceDate) }}
            </span>
          </div>
          <h2 class="session-title">{{ session?.title || session?.Title || 'Sesi Presensi' }}</h2>
          <p class="session-desc">{{ session?.description || session?.Description || 'Presensi kehadiran siswa.' }}</p>
          <div class="session-meta">
            <span v-if="session?.start_time"><Clock3 :size="14" /> Jam: {{ formatDateTime(session?.start_time) }} - {{ formatDateTime(session?.end_time) }}</span>
          </div>
        </div>

        <div class="hero-actions">
          <button
            v-if="session?.status === 'closed'"
            class="btn btn-primary"
            type="button"
            :disabled="togglingState"
            @click="setSessionState('open')"
          >
            {{ togglingState ? 'Memproses...' : 'Buka Sesi Presensi' }}
          </button>
          <button
            v-else
            class="btn btn-outline-danger"
            type="button"
            :disabled="togglingState"
            @click="setSessionState('close')"
          >
            {{ togglingState ? 'Memproses...' : 'Tutup Sesi Presensi' }}
          </button>
          <button class="btn btn-export" type="button" @click="exportRecap">
            <Download :size="15" /> Ekspor Rekap
          </button>
        </div>
      </section>

      <!-- KPI Statistics Cards -->
      <section class="stats-grid">
        <div class="stat-card rate-card">
          <div class="stat-val">{{ stats.rate }}%</div>
          <div class="stat-lbl">Kehadiran Kelas</div>
        </div>
        <div class="stat-card total-card">
          <div class="stat-val">{{ stats.total }}</div>
          <div class="stat-lbl">Total Siswa</div>
        </div>
        <div class="stat-card present-card">
          <div class="stat-val">{{ stats.present }}</div>
          <div class="stat-lbl">Hadir Tepat Waktu</div>
        </div>
        <div class="stat-card late-card">
          <div class="stat-val">{{ stats.late }}</div>
          <div class="stat-lbl">Terlambat</div>
        </div>
        <div class="stat-card excused-card">
          <div class="stat-val">{{ stats.excused }}</div>
          <div class="stat-lbl">Izin / Sakit</div>
        </div>
        <div class="stat-card absent-card">
          <div class="stat-val">{{ stats.notCheckedIn + stats.absent }}</div>
          <div class="stat-lbl">Alpa / Belum Absen</div>
        </div>
      </section>

      <!-- Search & Filter Controls -->
      <section class="filter-bar card-panel">
        <div class="search-wrap">
          <Search :size="16" class="search-icon" />
          <input
            v-model="searchQuery"
            type="search"
            class="search-input"
            placeholder="Cari nama atau NIS siswa..."
          />
        </div>
        <div class="filter-pills">
          <button
            class="pill-btn"
            :class="{ active: filterStatus === 'all' }"
            type="button"
            @click="filterStatus = 'all'"
          >
            Semua ({{ stats.total }})
          </button>
          <button
            class="pill-btn pill-present"
            :class="{ active: filterStatus === 'present' }"
            type="button"
            @click="filterStatus = 'present'"
          >
            Hadir ({{ stats.present }})
          </button>
          <button
            class="pill-btn pill-late"
            :class="{ active: filterStatus === 'late' }"
            type="button"
            @click="filterStatus = 'late'"
          >
            Terlambat ({{ stats.late }})
          </button>
          <button
            class="pill-btn pill-excused"
            :class="{ active: filterStatus === 'excused' }"
            type="button"
            @click="filterStatus = 'excused'"
          >
            Izin ({{ stats.excused }})
          </button>
          <button
            class="pill-btn pill-absent"
            :class="{ active: filterStatus === 'absent' }"
            type="button"
            @click="filterStatus = 'absent'"
          >
            Alpa ({{ stats.absent }})
          </button>
          <button
            class="pill-btn pill-not-checked"
            :class="{ active: filterStatus === 'not_checked_in' }"
            type="button"
            @click="filterStatus = 'not_checked_in'"
          >
            Belum Absen ({{ stats.notCheckedIn }})
          </button>
        </div>
      </section>

      <!-- Student Record Cards -->
      <section class="student-records-list">
        <article
          v-for="record in filteredRecords"
          :key="studentIdOf(record) || idOf(record)"
          class="record-card card-panel"
        >
          <div class="record-left">
            <div class="avatar-circle">
              {{ getInitials(nameOf(record)) }}
            </div>
            <div class="student-info">
              <div class="nis-row">
                <span class="nis-badge">NIS: {{ nisOf(record) }}</span>
                <span v-if="record.checked_at" class="checked-time">
                  <Clock3 :size="13" /> {{ formatDateTime(record.checked_at) }}
                </span>
                <span v-else class="not-checked-tag">
                  Belum check-in
                </span>
              </div>
              <h3 class="student-name">{{ nameOf(record) }}</h3>
              <p v-if="record.note" class="record-note">
                Catatan: {{ record.note }}
              </p>
            </div>
          </div>

          <div class="record-actions">
            <!-- Fast Status Buttons -->
            <div class="status-btn-group">
              <button
                type="button"
                class="status-choice-btn btn-choice-present"
                :class="{ active: statusOf(record) === 'present' }"
                :disabled="savingId === (studentIdOf(record) || idOf(record))"
                @click="updateStatus(record, 'present')"
              >
                Hadir
              </button>
              <button
                type="button"
                class="status-choice-btn btn-choice-late"
                :class="{ active: statusOf(record) === 'late' }"
                :disabled="savingId === (studentIdOf(record) || idOf(record))"
                @click="updateStatus(record, 'late')"
              >
                Terlambat
              </button>
              <button
                type="button"
                class="status-choice-btn btn-choice-excused"
                :class="{ active: statusOf(record) === 'excused' }"
                :disabled="savingId === (studentIdOf(record) || idOf(record))"
                @click="updateStatus(record, 'excused')"
              >
                Izin
              </button>
              <button
                type="button"
                class="status-choice-btn btn-choice-absent"
                :class="{ active: statusOf(record) === 'absent' }"
                :disabled="savingId === (studentIdOf(record) || idOf(record))"
                @click="updateStatus(record, 'absent')"
              >
                Alpa
              </button>
            </div>
          </div>
        </article>

        <div v-if="!filteredRecords.length" class="empty-state card-panel">
          <p>Tidak ada data siswa yang cocok dengan filter atau pencarian.</p>
        </div>
      </section>
    </template>
  </AppShell>
</template>

<style scoped>
.header-nav {
  margin-bottom: 16px;
}
.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: #4b5563;
  text-decoration: none;
  padding: 6px 14px;
  border-radius: 999px;
  background: white;
  border: 1px solid #e5e7eb;
  transition: all 0.2s ease;
}
.back-link:hover {
  background: #f3f4f6;
  color: #111827;
}
.back-arrow {
  font-size: 14px;
}

.card-panel {
  background: white;
  border: 1px solid #edf0f2;
  border-radius: 20px;
  box-shadow: 0 4px 14px rgba(15, 23, 42, 0.04);
}

.alert {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 18px;
  border-radius: 14px;
  margin-bottom: 16px;
  font-size: 13px;
  font-weight: 500;
}
.error-alert {
  background: #fee2e2;
  color: #991b1b;
  border: 1px solid #fca5a5;
}
.success-alert {
  background: #ecfdf5;
  color: #065f46;
  border: 1px solid #a7f3d0;
}
.alert-close {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: inherit;
  line-height: 1;
}

/* Hero card */
.session-hero {
  padding: 24px 28px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  margin-bottom: 20px;
}
.hero-tag-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}
.badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.02em;
}
.badge-open {
  background: #ecfdf5;
  color: #059669;
  border: 1px solid #a7f3d0;
}
.badge-closed {
  background: #f3f4f6;
  color: #6b7280;
  border: 1px solid #e5e7eb;
}
.date-badge {
  font-size: 12px;
  color: #6b7280;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.session-title {
  font-size: 22px;
  font-weight: 800;
  color: #111827;
  margin: 0 0 6px;
}
.session-desc {
  font-size: 13px;
  color: #6b7280;
  margin: 0 0 10px;
}
.session-meta {
  font-size: 12px;
  color: #4b5563;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.session-meta span { display: inline-flex; align-items: center; gap: 5px; }
.btn-export { color: #226045; background: #e8f6ed; border: 1px solid #b9e2c7; margin-top: 8px; }
.btn-export:hover { background: #d8f0df; }
.hero-actions {
  flex-shrink: 0;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  padding: 10px 20px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}
.btn-primary {
  background: #5eb69d;
  color: white;
  box-shadow: 0 4px 10px rgba(94, 182, 157, 0.3);
}
.btn-primary:hover:not(:disabled) {
  background: #4ea58c;
}
.btn-outline-danger {
  background: white;
  color: #e11d48;
  border: 1.5px solid #fecdd3;
}
.btn-outline-danger:hover:not(:disabled) {
  background: #fff1f2;
}

/* Stats grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
  gap: 12px;
  margin-bottom: 20px;
}
.stat-card {
  padding: 16px;
  border-radius: 16px;
  background: white;
  border: 1px solid #edf0f2;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.03);
  text-align: center;
}
.stat-val {
  font-size: 24px;
  font-weight: 800;
  line-height: 1.1;
  margin-bottom: 4px;
}
.stat-lbl {
  font-size: 11px;
  font-weight: 600;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}
.rate-card .stat-val { color: #5eb69d; }
.total-card .stat-val { color: #6366f1; }
.present-card .stat-val { color: #10b981; }
.late-card .stat-val { color: #f59e0b; }
.excused-card .stat-val { color: #3b82f6; }
.absent-card .stat-val { color: #f43f5e; }

/* Filter bar */
.filter-bar {
  padding: 16px 20px;
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.search-wrap {
  position: relative;
  width: 100%;
}
.search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 14px;
  opacity: 0.5;
}
.search-input {
  width: 100%;
  box-sizing: border-box;
  padding: 10px 16px 10px 38px;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  font-size: 13px;
  outline: none;
  transition: border-color 0.2s ease;
}
.search-input:focus {
  border-color: #5eb69d;
  box-shadow: 0 0 0 3px rgba(94, 182, 157, 0.15);
}
.filter-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.pill-btn {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 999px;
  padding: 6px 14px;
  font-size: 12px;
  font-weight: 600;
  color: #4b5563;
  cursor: pointer;
  transition: all 0.2s ease;
}
.pill-btn:hover {
  background: #f3f4f6;
}
.pill-btn.active {
  background: #111827;
  color: white;
  border-color: #111827;
}
.pill-present.active {
  background: #059669;
  border-color: #059669;
}
.pill-late.active {
  background: #d97706;
  border-color: #d97706;
}
.pill-excused.active {
  background: #2563eb;
  border-color: #2563eb;
}
.pill-absent.active {
  background: #e11d48;
  border-color: #e11d48;
}
.pill-not-checked.active {
  background: #4b5563;
  border-color: #4b5563;
}

/* Student records list */
.student-records-list {
  display: grid;
  gap: 12px;
}
.record-card {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}
.record-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.06);
}
.record-left {
  display: flex;
  align-items: center;
  gap: 14px;
}
.avatar-circle {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: linear-gradient(135deg, #ffd6d6, #ffe5c2);
  color: #b91c1c;
  font-weight: 800;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.nis-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 2px;
}
.nis-badge {
  font-size: 11px;
  font-weight: 600;
  color: #6b7280;
}
.checked-time {
  font-size: 11px;
  color: #059669;
  font-weight: 600;
  background: #ecfdf5;
  padding: 1px 6px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.not-checked-tag {
  font-size: 11px;
  color: #9ca3af;
  font-style: italic;
}
.student-name {
  font-size: 15px;
  font-weight: 700;
  color: #111827;
  margin: 0;
}
.record-note {
  font-size: 11px;
  color: #6b7280;
  margin: 2px 0 0;
}

/* Status choice buttons */
.status-btn-group {
  display: flex;
  gap: 6px;
}
.status-choice-btn {
  padding: 6px 12px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid #e5e7eb;
  background: #f9fafb;
  color: #4b5563;
  cursor: pointer;
  transition: all 0.15s ease;
}
.status-choice-btn:hover:not(:disabled) {
  background: #f3f4f6;
}
.btn-choice-present.active {
  background: #ecfdf5;
  color: #059669;
  border-color: #a7f3d0;
  font-weight: 700;
}
.btn-choice-late.active {
  background: #fef3c7;
  color: #d97706;
  border-color: #fde68a;
  font-weight: 700;
}
.btn-choice-excused.active {
  background: #eff6ff;
  color: #2563eb;
  border-color: #bfdbfe;
  font-weight: 700;
}
.btn-choice-absent.active {
  background: #fff1f2;
  color: #e11d48;
  border-color: #fecdd3;
  font-weight: 700;
}

.empty-state {
  padding: 40px;
  text-align: center;
  color: #6b7280;
  font-size: 13px;
}

@media (max-width: 768px) {
  .session-hero {
    flex-direction: column;
    align-items: flex-start;
  }
  .record-card {
    flex-direction: column;
    align-items: flex-start;
  }
  .hero-actions { width: 100%; display: flex; flex-wrap: wrap; gap: 8px; }
  .hero-actions .btn { flex: 1 1 auto; }
  .status-btn-group {
    width: 100%;
    flex-wrap: wrap;
  }
  .status-choice-btn {
    flex: 1 1 calc(50% - 6px);
    text-align: center;
  }
}

/* ================= DARK MODE STYLES ================= */
:global(html[data-theme='dark']) .student-name {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .record-card {
  background: #1e293b !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .status-choice-btn {
  background: #151f30 !important;
  border-color: #334155 !important;
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .status-choice-btn:hover:not(:disabled) {
  background: #273549 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .pill-filter {
  background: #151f30 !important;
  border-color: #334155 !important;
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .search-box input {
  background: #151f30 !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
</style>
