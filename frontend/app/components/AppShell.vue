<script setup lang="ts">
import { 
  LayoutDashboard, 
  Users, 
  GraduationCap, 
  BookOpen, 
  MessageSquare, 
  CircleHelp,
  Bell, 
  Search, 
  Menu, 
  Mail, 
  LogOut,
  Settings,
  AlertTriangle,
  X
} from 'lucide-vue-next'
import { useChat } from '~/composables/useChat'
import { useAppSettings } from '~/composables/useAppSettings'

const props = withDefaults(defineProps<{
  active?: 'dashboard' | 'admin' | 'academic' | 'learning' | 'chat' | 'help-center' | 'notifications' | 'settings' | 'profile'
  eyebrow?: string
  title?: string
}>(), {
  active: 'dashboard',
  eyebrow: 'School Space',
  title: '',
})

const auth = useAuthStore()
const { request } = useApi()
const { settings, t, presenceColor } = useAppSettings()
const { getConversations, getMessages, getUnreadCounts } = useChat()
const notificationUnreadCount = ref(0)
const chatUnreadCount = ref(0)
const chatPreviewOpen = ref(false)
const chatPreviews = ref<any[]>([])
const chatPreviewLoading = ref(false)
const roleLabel = computed(() => t('role.' + (auth.role || '')) || auth.role)
const userDisplayName = computed(() => {
  const user = auth.user
  return user?.student?.full_name || user?.student?.FullName || user?.teacher?.full_name || user?.teacher?.FullName || user?.display_name || user?.username || 'Pengguna'
})
const userInitial = computed(() => userDisplayName.value.trim().charAt(0).toUpperCase() || 'U')
const logout = async () => { auth.clear(); await navigateTo('/login') }
const chatSenderName = (message: any) => {
  const sender = message?.sender || {}
  if ((sender.role?.name || sender.Role?.Name) === 'admin') return 'Admin'
  return sender.display_name || sender.DisplayName || sender.student?.full_name ||
    sender.Student?.FullName || sender.teacher?.full_name || sender.Teacher?.FullName ||
    sender.username || sender.Username || 'Pengguna'
}
const chatConversationName = (conversation: any) => {
  if (conversation?.name) return conversation.name
  const participant = (conversation?.participants || []).find((item: any) => item.user)
  return chatSenderName({ sender: participant?.user })
}
const loadChatPreview = async () => {
  chatPreviewLoading.value = true
  try {
    const [conversationResponse, unreadResponse] = await Promise.all([
      getConversations(),
      getUnreadCounts(),
    ])
    const unread = unreadResponse.data || {}
    chatUnreadCount.value = Object.values(unread).reduce((total, count) => total + Number(count || 0), 0)
    const unreadConversations = conversationResponse.data.filter((conversation: any) => Number(unread[conversation.id] || 0) > 0)
    chatPreviews.value = (await Promise.all(unreadConversations.map(async (conversation: any) => {
      const response = await getMessages(conversation.id)
      const message = response.data[response.data.length - 1]
      return message ? {
        id: conversation.id,
        name: chatConversationName(conversation),
        sender: chatSenderName(message),
        content: message.content,
        createdAt: message.created_at,
        unread: Number(unread[conversation.id] || 0),
      } : null
    }))).filter(Boolean)
  } catch {
    chatPreviews.value = []
  } finally {
    chatPreviewLoading.value = false
  }
}
const toggleChatPreview = async () => {
  chatPreviewOpen.value = !chatPreviewOpen.value
  if (chatPreviewOpen.value) await loadChatPreview()
}
const openPreviewConversation = (conversationId: string) => {
  chatPreviewOpen.value = false
  navigateTo(`/chat/${conversationId}`)
}
onMounted(async () => {
  const response = await request<{ data?: { unread_count?: number } }>('/notifications/unread-count').catch(() => null)
  notificationUnreadCount.value = response?.data?.unread_count || 0
  await loadChatPreview()
})

const mobileMenuOpen = ref(false)
const toggleMobileMenu = () => { mobileMenuOpen.value = !mobileMenuOpen.value }
</script>

<template>
  <div class="app-frame">
    <!-- Mobile Overlay -->
    <div class="mobile-overlay" :class="{ active: mobileMenuOpen }" @click="toggleMobileMenu"></div>
    
    <aside class="sidebar" :class="{ open: mobileMenuOpen }">
      <NuxtLink class="brand" to="/dashboard">
        <img src="/images/logo.png" alt="School Space Logo" class="brand-logo" />
        <span>{{ settings.schoolName || 'School Space' }}</span>
      </NuxtLink>
      
      <div class="sidebar-role">{{ roleLabel }}</div>
      
      <nav class="main-nav" aria-label="Navigasi utama">
        <NuxtLink class="nav-item" :class="{ active: props.active === 'dashboard' }" to="/dashboard">
          <LayoutDashboard :size="18" class="icon-nav" /> {{ t('nav.dashboard') }}
        </NuxtLink>
        <template v-if="auth.role === 'admin'">
          <NuxtLink class="nav-item" :class="{ active: props.active === 'admin' }" to="/admin">
            <Users :size="18" class="icon-nav" /> {{ t('nav.admin') }}
          </NuxtLink>
        </template>
        <NuxtLink class="nav-item" :class="{ active: props.active === 'academic' }" to="/academic">
          <GraduationCap :size="18" class="icon-nav" /> {{ auth.role === 'admin' ? t('nav.academic') : t('nav.academic_year') }}
        </NuxtLink>
        <NuxtLink class="nav-item" :class="{ active: props.active === 'learning' }" to="/learning">
          <BookOpen :size="18" class="icon-nav" /> {{ auth.role === 'guru' ? t('nav.learning_teacher') : t('nav.learning_student') }}
        </NuxtLink>
        <NuxtLink class="nav-item" :class="{ active: props.active === 'chat' }" to="/chat">
          <MessageSquare :size="18" class="icon-nav" /> {{ t('nav.chat') }}
        </NuxtLink>
        <NuxtLink class="nav-item" :class="{ active: props.active === 'help-center' }" to="/help-center">
          <CircleHelp :size="18" class="icon-nav" /> {{ t('nav.help_center') }}
        </NuxtLink>
      </nav>

      <div class="join-course-card">
        <img src="/images/sidebar_illustration_remove.png" alt="Student Studying" class="join-course-img" />
        <button class="join-btn" @click="logout">{{ t('nav.logout') }} <LogOut :size="14" style="display:inline; margin-left:4px;"/></button>
      </div>
    </aside>

    <div class="content-wrapper">
      <div class="content-area">
        <header class="topbar">
          <button class="mobile-toggle" @click="toggleMobileMenu">
            <Menu :size="20" />
          </button>
          
          <div class="search-bar">
            <Search :size="16" class="icon" />
            <input type="text" :placeholder="t('search.placeholder')" />
          </div>

          <div class="top-actions">
            <button class="icon-btn chat-preview-trigger" type="button" :title="t('nav.messages')" @click="toggleChatPreview">
              <Mail :size="20" />
              <span class="badge" v-if="chatUnreadCount > 0"></span>
            </button>
            <div v-if="chatPreviewOpen" class="chat-preview-panel">
              <div class="chat-preview-header">
                <strong>{{ t('nav.messages') }}</strong>
                <button type="button" aria-label="Tutup preview" @click="chatPreviewOpen = false"><X :size="15" /></button>
              </div>
              <div v-if="chatPreviewLoading" class="chat-preview-empty">{{ t('chat.loading') }}</div>
              <div v-else-if="!chatPreviews.length" class="chat-preview-empty">{{ t('chat.no_new_messages') }}</div>
              <button
                v-for="preview in chatPreviews"
                :key="preview.id"
                type="button"
                class="chat-preview-item"
                @click="openPreviewConversation(preview.id)"
              >
                <span class="chat-preview-avatar">{{ preview.sender.charAt(0).toUpperCase() }}</span>
                <span class="chat-preview-copy">
                  <strong>{{ preview.name }}</strong>
                  <small>{{ preview.sender }} · {{ settings.notifyChatPreview ? preview.content : '•••••' }}</small>
                </span>
                <span class="chat-preview-count">{{ preview.unread }}</span>
              </button>
            </div>
            <NuxtLink class="icon-btn" to="/notifications" :title="t('nav.notifications')">
              <Bell :size="20" />
              <span class="badge" v-if="notificationUnreadCount > 0"></span>
            </NuxtLink>
            <NuxtLink class="icon-btn" :class="{ active: props.active === 'settings' }" to="/settings" :title="t('nav.settings')">
              <Settings :size="20" />
            </NuxtLink>
            <NuxtLink class="user-profile" to="/profile" :title="t('nav.profile')">
              <div class="avatar" :style="{ position: 'relative' }">
                <span class="avatar-initial">{{ userInitial }}</span>
                <span class="presence-badge" :style="{ backgroundColor: presenceColor }" :title="settings.presenceStatus"></span>
              </div>
              <div class="user-info">
                <span class="user-name">{{ userDisplayName }}</span>
                <span class="user-handle">@{{ auth.user?.username?.toLowerCase() }}</span>
              </div>
            </NuxtLink>
          </div>
        </header>

        <!-- Scheduled Maintenance Mode Banner -->
        <div v-if="settings.maintenanceMode" class="maintenance-alert-banner">
          <AlertTriangle :size="18" class="maintenance-icon" />
          <span>{{ t('banner.maintenance') }}</span>
        </div>

        <!-- Dynamic Content -->
        <slot />
      </div>
    </div>
  </div>
</template>

<style scoped>
.icon-nav { color: #a4b1bc; }
.nav-item:hover .icon-nav, .nav-item.active .icon-nav { color: var(--teal); }
.top-actions { position: relative; }
.icon-btn.active { color: var(--teal); background: rgba(101, 201, 171, 0.12); border-radius: 8px; padding: 4px; }
.chat-preview-trigger { position: relative; border: 0; cursor: pointer; }
.top-actions .badge {
  position: absolute;
  top: 1px;
  right: 0;
  width: 8px;
  height: 8px;
  border: 2px solid var(--paper);
  border-radius: 50%;
  background: var(--coral);
}
.chat-preview-panel {
  position: absolute;
  z-index: 30;
  top: 46px;
  right: 54px;
  width: min(360px, calc(100vw - 32px));
  overflow: hidden;
  border: 1px solid #e5ebe7;
  border-radius: 12px;
  background: white;
  box-shadow: 0 14px 35px rgba(34, 52, 43, 0.16);
}
.chat-preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 13px 15px;
  border-bottom: 1px solid #edf1ee;
  color: var(--ink);
  font-size: 12px;
}
.chat-preview-header button {
  display: inline-flex;
  border: 0;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
}
.chat-preview-empty {
  padding: 22px 15px;
  color: var(--muted);
  font-size: 11px;
  text-align: center;
}
.chat-preview-item {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 10px;
  padding: 11px 13px;
  border: 0;
  border-bottom: 1px solid #f0f3f1;
  background: white;
  cursor: pointer;
  text-align: left;
}
.chat-preview-item:hover { background: #f6fbf7; }
.chat-preview-avatar {
  display: inline-flex;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #e5f4e8;
  color: #34704b;
  font-size: 12px;
  font-weight: 800;
}
.chat-preview-copy { min-width: 0; flex: 1; }
.chat-preview-copy strong,
.chat-preview-copy small { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.chat-preview-copy strong { color: var(--ink); font-size: 11px; }
.chat-preview-copy small { margin-top: 3px; color: var(--muted); font-size: 10px; }
.chat-preview-count {
  display: inline-flex;
  min-width: 18px;
  height: 18px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: var(--coral);
  color: white;
  font-size: 10px;
  font-weight: 800;
}

.join-course-card {
  position: relative;
  overflow: visible;
  margin-top: 115px;
  padding-top: 72px;
}

.join-course-img {
  position: absolute;
  top: -101px;          /* jauh lebih kecil dari -140px sebelumnya */
  left: 50%;
  width: 100%;          /* selebar card, biar nutup sisi kiri-kanan */
  max-width: 125px;      /* sesuaikan sama lebar card aslinya */
  height: auto;
  transform: translateX(-50%);  /* rotasi dihapus dulu biar rapi */
  filter: drop-shadow(0 10px 18px rgba(0, 0, 0, 0.2));
  pointer-events: none;
  z-index: 2;
}

.join-btn {
  position: relative;
  z-index: 1;
}

.maintenance-alert-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  background: #fffbeb;
  border: 1px solid #fef3c7;
  color: #b45309;
  padding: 12px 18px;
  border-radius: 14px;
  margin-bottom: 22px;
  font-size: 13.5px;
  font-weight: 700;
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.08);
}
.maintenance-icon {
  color: #d97706;
  flex-shrink: 0;
}

.presence-badge {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  border: 2px solid white;
  transition: background-color 0.25s ease;
}
</style>

