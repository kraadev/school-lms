<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '~/composables/useApi'
import { useAppSettings } from '~/composables/useAppSettings'
import {
  Bell,
  BookOpen,
  ClipboardList,
  Clock3,
  Lightbulb,
  MessageCircle,
} from 'lucide-vue-next'

definePageMeta({ middleware: 'auth' })

const router = useRouter()
const { request } = useApi()
const { t, settings } = useAppSettings()
const notifications = ref<any[]>([])
const loading = ref(true)
const markingAll = ref(false)
const filter = ref<'all' | 'unread'>('all')

const notificationValue = (item: any, ...keys: string[]) => {
  for (const key of keys) {
    const value = item?.[key]
    if (value !== undefined && value !== null && value !== '') return value
  }
  return undefined
}

const notificationType = (item: any) =>
  String(notificationValue(item, 'type', 'Type', 'reference_type', 'ReferenceType') || '').toLowerCase()

const notificationTitle = (item: any) =>
  notificationValue(item, 'title', 'Title') || 'Pemberitahuan Baru'

const notificationMessage = (item: any) =>
  notificationValue(item, 'message', 'Message', 'content', 'Content') || 'Ada pembaruan aktivitas baru.'

const notificationIsRead = (item: any) =>
  Boolean(notificationValue(item, 'is_read', 'IsRead'))

const notificationId = (item: any) => notificationValue(item, 'id', 'ID')

const notificationCreatedAt = (item: any) =>
  notificationValue(item, 'created_at', 'CreatedAt')

const load = async () => {
  try {
    const response = await request<{ data?: any[] }>('/notifications')
    notifications.value = response.data || []
  } catch {
    notifications.value = []
  } finally {
    loading.value = false
  }
}

const unreadCount = computed(() => notifications.value.filter(n => !notificationIsRead(n)).length)

const filteredNotifications = computed(() => {
  if (filter.value === 'unread') {
    return notifications.value.filter(n => !notificationIsRead(n))
  }
  return notifications.value
})

const markAll = async () => {
  markingAll.value = true
  try {
    await request('/notifications/read-all', { method: 'PATCH' })
    for (const n of notifications.value) {
      if ('is_read' in n) n.is_read = true
      else n.IsRead = true
    }
  } catch {}
  finally {
    markingAll.value = false
  }
}

const markRead = async (item: any, event?: Event) => {
  if (event) event.stopPropagation()
  if (notificationIsRead(item)) return
  try {
    await request(`/notifications/${notificationId(item)}/read`, { method: 'PATCH' })
    if ('is_read' in item) item.is_read = true
    else item.IsRead = true
  } catch {}
}

const handleNotificationClick = async (item: any) => {
  if (!item.is_read) {
    await markRead(item)
  }

  const type = notificationType(item)
  if (type.includes('assignment') || type.includes('tugas')) {
    router.push('/learning?tab=assignments')
  } else if (type.includes('quiz') || type.includes('kuis')) {
    router.push('/learning?tab=quizzes')
  } else if (type.includes('material') || type.includes('materi')) {
    router.push('/learning?tab=materials')
  } else if (type.includes('attendance') || type.includes('presensi')) {
    router.push('/learning?tab=attendance')
  } else if (type.includes('chat') || type.includes('pesan')) {
    router.push('/chat')
  } else {
    router.push('/learning')
  }
}

const getIcon = (type?: string) => {
  const t = (type || '').toLowerCase()
  if (t.includes('assignment') || t.includes('tugas')) return ClipboardList
  if (t.includes('quiz') || t.includes('kuis')) return Lightbulb
  if (t.includes('material') || t.includes('materi')) return BookOpen
  if (t.includes('attendance') || t.includes('presensi')) return Clock3
  if (t.includes('chat') || t.includes('pesan')) return MessageCircle
  return Bell
}

const getIconBg = (type?: string) => {
  const t = (type || '').toLowerCase()
  if (t.includes('assignment') || t.includes('tugas')) return 'bg-assignment'
  if (t.includes('quiz') || t.includes('kuis')) return 'bg-quiz'
  if (t.includes('material') || t.includes('materi')) return 'bg-material'
  if (t.includes('attendance') || t.includes('presensi')) return 'bg-attendance'
  if (t.includes('chat') || t.includes('pesan')) return 'bg-chat'
  return 'bg-default'
}

const formatTime = (dateStr?: string) => {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    const now = new Date()
    const diffMs = now.getTime() - d.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return t('notif.just_now')
    if (diffMins < 60) return `${diffMins} ${t('notif.mins_ago')}`
    if (diffHours < 24) return `${diffHours} ${t('notif.hours_ago')}`
    if (diffDays === 1) return t('notif.yesterday')
    if (diffDays < 7) return `${diffDays} ${t('notif.days_ago')}`
    return d.toLocaleDateString(settings.language === 'en' ? 'en-US' : 'id-ID', { day: 'numeric', month: 'short' })
  } catch {
    return dateStr
  }
}

onMounted(load)
</script>

<template>
  <AppShell active="notifications" eyebrow="School Space" :title="t('notif.title')">
    <!-- Header Controls -->
    <div class="notif-header">
      <div class="filter-tabs">
        <button
          class="tab-btn"
          :class="{ active: filter === 'all' }"
          type="button"
          @click="filter = 'all'"
        >
          {{ t('notif.all') }}
          <span class="count-pill">{{ notifications.length }}</span>
        </button>
        <button
          class="tab-btn"
          :class="{ active: filter === 'unread' }"
          type="button"
          @click="filter = 'unread'"
        >
          {{ t('notif.unread') }}
          <span v-if="unreadCount > 0" class="count-pill unread-pill">{{ unreadCount }}</span>
        </button>
      </div>

      <button
        v-if="unreadCount > 0"
        class="mark-all-btn"
        type="button"
        :disabled="markingAll"
        @click="markAll"
      >
        {{ markingAll ? t('notif.marking') : t('notif.mark_all') }}
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <span class="loader" /> {{ t('notif.loading') }}
    </div>

    <!-- Notification List -->
    <section v-else class="notif-list">
      <article
        v-for="item in filteredNotifications"
        :key="notificationId(item)"
        class="notif-card"
        :class="{ unread: !notificationIsRead(item) }"
        @click="handleNotificationClick(item)"
      >
        <div class="notif-icon-col">
          <div class="icon-avatar" :class="getIconBg(notificationType(item))">
            <component :is="getIcon(notificationType(item))" :size="20" :stroke-width="2" />
          </div>
        </div>

        <div class="notif-content-col">
          <div class="notif-meta-row">
            <h4 class="notif-title">{{ notificationTitle(item) }}</h4>
            <span class="notif-time">{{ formatTime(notificationCreatedAt(item)) }}</span>
          </div>
          <p class="notif-message">{{ notificationMessage(item) }}</p>
          <div class="notif-footer-row">
            <span class="action-hint">{{ notificationType(item).replaceAll('_', ' ') || 'sistem' }}</span>
            <button
              v-if="!notificationIsRead(item)"
              class="quick-read-btn"
              type="button"
              :title="settings.language === 'en' ? 'Mark as read' : 'Tandai dibaca'"
              @click.stop="markRead(item, $event)"
            >
              {{ settings.language === 'en' ? 'Mark read' : 'Tandai dibaca' }}
            </button>
          </div>
        </div>

        <span v-if="!item.is_read" class="unread-glow-dot" />
      </article>

      <div v-if="!filteredNotifications.length" class="empty-state">
        <div class="empty-icon"><Bell :size="28" :stroke-width="2" /></div>
        <h3>{{ settings.language === 'en' ? 'No notifications' : 'Tidak ada notifikasi' }}</h3>
        <p>
          {{ filter === 'unread' ? (settings.language === 'en' ? 'All notifications have been read.' : 'Semua notifikasi sudah kamu baca.') : (settings.language === 'en' ? 'No new notifications for you right now.' : 'Belum ada notifikasi baru untukmu saat ini.') }}
        </p>
      </div>
    </section>
  </AppShell>
</template>

<style scoped>
.notif-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 14px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.filter-tabs {
  display: flex;
  gap: 8px;
  background: white;
  padding: 4px;
  border-radius: 999px;
  border: 1px solid #e5e7eb;
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: none;
  background: transparent;
  padding: 8px 16px;
  border-radius: 999px;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab-btn:hover {
  color: #111827;
}

.tab-btn.active {
  background: #111827;
  color: white;
}

.count-pill {
  font-size: 11px;
  font-weight: 700;
  background: #f3f4f6;
  color: #4b5563;
  padding: 2px 7px;
  border-radius: 999px;
}

.tab-btn.active .count-pill {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.unread-pill {
  background: #ff5a5f !important;
  color: white !important;
}

.mark-all-btn {
  background: white;
  border: 1px solid #e5e7eb;
  padding: 8px 16px;
  border-radius: 12px;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  color: #4b5563;
  cursor: pointer;
  transition: all 0.2s ease;
}

.mark-all-btn:hover:not(:disabled) {
  background: #f9fafb;
  color: #111827;
}

.notif-list {
  display: grid;
  gap: 12px;
}

.notif-card {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 18px 22px;
  background: white;
  border-radius: 20px;
  border: 1px solid #edf0f2;
  box-shadow: 0 2px 10px rgba(15, 23, 42, 0.03);
  cursor: pointer;
  transition: all 0.2s ease;
}

.notif-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.06);
  border-color: #d1d5db;
}

.notif-card.unread {
  background: #ffffff;
  border-left: 4px solid #ff5a5f;
}

.notif-icon-col {
  flex-shrink: 0;
}

.icon-avatar {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.bg-assignment {
  background: #fee2e2;
}

.bg-quiz {
  background: #fef3c7;
}

.bg-material {
  background: #e0f2fe;
}

.bg-attendance {
  background: #ecfdf5;
}

.bg-chat {
  background: #ede9fe;
}

.bg-default {
  background: #f3f4f6;
}

.notif-content-col {
  flex: 1;
  min-width: 0;
}

.notif-meta-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 4px;
}

.notif-title {
  font-size: 15px;
  font-weight: 700;
  color: #111827;
  margin: 0;
}

.notif-card.unread .notif-title {
  color: #0f172a;
}

.notif-time {
  font-size: 12px;
  color: #9ca3af;
  white-space: nowrap;
}

.notif-message {
  font-size: 13px;
  color: #4b5563;
  line-height: 1.45;
  margin: 0 0 10px;
}

.notif-footer-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.action-hint {
  font-size: 12px;
  font-weight: 600;
  color: #5eb69d;
  transition: color 0.2s;
}

.notif-card:hover .action-hint {
  color: #4ea58c;
}

.quick-read-btn {
  background: #f3f4f6;
  border: none;
  padding: 4px 10px;
  border-radius: 6px;
  font-family: inherit;
  font-size: 11px;
  font-weight: 600;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.15s ease;
}

.quick-read-btn:hover {
  background: #e5e7eb;
  color: #111827;
}

.unread-glow-dot {
  position: absolute;
  top: 18px;
  right: 18px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ff5a5f;
  box-shadow: 0 0 0 3px rgba(255, 90, 95, 0.2);
}

.empty-state {
  padding: 50px 20px;
  text-align: center;
  background: white;
  border-radius: 20px;
  border: 1px solid #edf0f2;
}

.empty-icon {
  font-size: 40px;
  margin-bottom: 12px;
  opacity: 0.8;
}

.empty-state h3 {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 6px;
}

.empty-state p {
  font-size: 13px;
  color: #6b7280;
  margin: 0;
}

@media (max-width: 600px) {
  .notif-card {
    padding: 14px 16px;
  }
  .notif-meta-row {
    flex-direction: column;
    align-items: flex-start;
  }
  .unread-glow-dot {
    top: 14px;
    right: 14px;
  }
}
</style>
