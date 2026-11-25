<script setup lang="ts">
import { onMounted } from 'vue'
import { ArrowLeft, Search, Phone, Video, MoreVertical, Paperclip, Smile, Send, Plus, X } from 'lucide-vue-next'
import { useChatStore } from '~/stores/chat'
import { useChat } from '~/composables/useChat'
import { useChatWebSocket } from '~/composables/useChatWebSocket'
import type { ChatContact, ChatConversation } from '~/stores/chat'
import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'
import { useAppSettings } from '~/composables/useAppSettings'

definePageMeta({
  middleware: 'auth',
})

const chatStore = useChatStore()
const auth = useAuthStore()
const { request } = useApi()
const { t, settings } = useAppSettings()

const {
  getConversations,
  getContacts,
  getContactsForClass,
  getTeacherClasses,
  createRoom,
  getUnreadCounts,
  createConversation,
} = useChat()

const {
  connected,
  connect,
} = useChatWebSocket()

const loading = ref(true)
const contacts = ref<ChatContact[]>([])
const contactsOpen = ref(false)
const startingConversation = ref('')
const errorMessage = ref('')
const teacherClasses = ref<any[]>([])
const roomFormOpen = ref(false)
const roomSaving = ref(false)
const roomName = ref('')
const selectedClassId = ref('')
const classStudents = ref<ChatContact[]>([])
const selectedStudentIds = ref<string[]>([])
const searchQuery = ref('')
const activeFilter = ref<'messages' | 'unread'>('messages')

const chatTitle = computed(() => {
  if (auth.role === 'admin') return t('chat.title_admin')
  if (auth.role === 'guru') return t('chat.title_teacher')
  return t('chat.title_student')
})
const contactHelp = computed(() => {
  if (auth.role === 'siswa') return t('chat.contact_help_student')
  return t('chat.contact_help_default')
})

const filteredConversations = computed(() => {
  const source = activeFilter.value === 'unread'
    ? chatStore.conversations.filter(item => !!chatStore.unreadCounts[item.id])
    : chatStore.conversations
  if (!searchQuery.value) return source
  const q = searchQuery.value.toLowerCase()
  return source.filter(c =>
    conversationName(c).toLowerCase().includes(q)
  )
})
const unreadFor = (conversationId: string) => chatStore.unreadCounts[conversationId] || 0

const contactName = (contact: ChatContact | any) => String(
  contact?.display_name || contact?.DisplayName || contact?.student?.full_name ||
  contact?.Student?.FullName || contact?.teacher?.full_name || contact?.Teacher?.FullName ||
  contact?.username || contact?.Username || contact?.email || contact?.Email || t('common.user'),
)
const contactRole = (contact: any) => {
  const roleName = contact?.role?.name || contact?.Role?.Name
  if (!roleName) return t('common.user')
  return t('role.' + roleName.toLowerCase()) || roleName
}
const contactInitial = (contact: ChatContact | any) => contactName(contact).charAt(0).toUpperCase()
const conversationName = (conversation: ChatConversation) => {
  if (conversation.name) return conversation.name
  const participant = conversation.participants?.find(item => item.user?.username || (item.user as any)?.Username)
  return participant?.user?.username || (participant?.user as any)?.Username || (settings.language === 'en' ? 'Conversation' : 'Percakapan')
}

const loadChat = async () => {
  try {
    loading.value = true
    const session = await request<{ user?: any }>('/auth/me')
    if (session.user) auth.updateUser(session.user)

    const [
      conversationResponse,
      contactResponse,
      unreadResponse,
    ] = await Promise.all([
      getConversations(),
      getContacts(),
      getUnreadCounts(),
    ])

    chatStore.setConversations(conversationResponse.data)
    const currentUserId = auth.user?.id
    contacts.value = (contactResponse.data || []).filter((contact) => {
      const contactId = contact?.id || (contact as any)?.ID
      return contactId && contactId !== currentUserId
    })
    if (auth.role === 'guru') {
      const classResponse = await getTeacherClasses()
      teacherClasses.value = classResponse.data || []
    }

    for (const [conversationId, count] of Object.entries(unreadResponse.data || {})) {
      chatStore.setUnreadCount(conversationId, count)
    }
  } catch (error: any) {
    errorMessage.value = error?.data?.message || 'Chat belum dapat dimuat.'
    console.error('Gagal memuat chat:', error)
  } finally {
    loading.value = false
  }
}

const openRoomForm = () => {
  roomFormOpen.value = true
  roomName.value = ''
  selectedClassId.value = ''
  classStudents.value = []
  selectedStudentIds.value = []
}
const loadClassStudents = async () => {
  if (!selectedClassId.value) { classStudents.value = []; return }
  try {
    const response = await getContactsForClass(selectedClassId.value)
    classStudents.value = response.data
    selectedStudentIds.value = []
  } catch (error: any) { errorMessage.value = error?.data?.message || 'Daftar siswa belum dapat dimuat.' }
}
const saveRoom = async () => {
  if (!roomName.value.trim() || !selectedClassId.value) { errorMessage.value = 'Nama ruang dan kelas wajib dipilih.'; return }
  roomSaving.value = true
  errorMessage.value = ''
  try {
    const response = await createRoom(roomName.value.trim(), selectedClassId.value, selectedStudentIds.value)
    chatStore.setConversations([response.data, ...chatStore.conversations])
    roomFormOpen.value = false
    await navigateTo(`/chat/${response.data.id}`)
  } catch (error: any) { errorMessage.value = error?.data?.message || 'Ruang chat belum dapat dibuat.' }
  finally { roomSaving.value = false }
}

const startConversation = async (contact: ChatContact) => {
  const contactId = contact.id || (contact as any).ID
  if (!contactId) return
  startingConversation.value = contactId
  errorMessage.value = ''
  try {
    const response = await createConversation(contactId)
    const conversation = response.data
    if (!chatStore.conversations.some(item => item.id === conversation.id)) {
      chatStore.setConversations([conversation, ...chatStore.conversations])
    }
    contactsOpen.value = false
    await navigateTo(`/chat/${conversation.id}`)
  } catch (error: any) {
    errorMessage.value = error?.data?.message || 'Percakapan belum dapat dibuat.'
  } finally {
    startingConversation.value = ''
  }
}

const openConversation = (conversationId: string) => {
  navigateTo(`/chat/${conversationId}`)
}

onMounted(async () => {
  await loadChat()
  connect()
})
</script>

<template>
  <AppShell active="chat" eyebrow="School Space" :title="t('nav.chat')">
  <div class="min-h-screen bg-[#f0f2f5] p-4 md:p-6 chat-page-bg">
    <div class="mx-auto flex h-[calc(100vh-3rem)] max-w-7xl overflow-hidden rounded-2xl bg-white shadow-lg chat-container-box">
      
      <!-- SIDEBAR -->
      <aside class="flex w-[360px] shrink-0 flex-col border-r border-gray-100 bg-white chat-sidebar">
        
        <!-- Sidebar Header -->
        <div class="flex items-center justify-between px-5 pt-5 pb-3 border-b border-gray-100 chat-sidebar-header">
          <div class="flex items-center gap-2">
            <NuxtLink
              to="/dashboard"
              class="flex h-8 w-8 items-center justify-center rounded-full text-gray-400 transition hover:bg-gray-100 hover:text-gray-600 chat-icon-btn"
              :title="t('common.back')"
            >
              <ArrowLeft :size="18" />
            </NuxtLink>
            <h1 class="text-xl font-bold text-gray-900 chat-title-text">
              {{ chatTitle }}
            </h1>
          </div>
          <div class="flex items-center gap-1">
            <button
              v-if="auth.role === 'guru'"
              type="button"
              class="flex h-8 w-8 items-center justify-center rounded-full bg-[#e97863] text-white transition hover:bg-[#d96552]"
              :title="t('chat.create_room')"
              @click="openRoomForm"
            >
              <Plus :size="16" />
            </button>
            <button
              v-else-if="auth.role === 'admin'"
              type="button"
              class="flex h-8 w-8 items-center justify-center rounded-full bg-[#e97863] text-white transition hover:bg-[#d96552]"
              :title="t('chat.create_room')"
              @click="contactsOpen = !contactsOpen"
            >
              <Plus :size="16" />
            </button>
          </div>
        </div>

        <!-- Status Koneksi -->
        <div class="flex items-center gap-2 px-5 py-2.5">
          <span
            class="h-2 w-2 rounded-full"
            :class="connected ? 'bg-green-500' : 'bg-gray-400'"
          />
          <span class="text-xs text-gray-500 chat-subtext">
            {{ connected ? t('chat.connected') : t('chat.connecting') }}
          </span>
        </div>

        <!-- Search Bar -->
        <div class="px-5 pb-3 chat-search-bar">
          <div class="relative">
            <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('chat.search_placeholder')"
              class="w-full rounded-xl border border-gray-200 bg-gray-50 py-2 pl-9 pr-3 text-sm text-gray-700 outline-none transition focus:border-[#e97863] focus:bg-white focus:ring-1 focus:ring-[#e97863] chat-search-input"
            />
          </div>
        </div>

        <!-- Filter Tabs -->
        <div class="flex gap-2 px-5 pb-3 overflow-x-auto">
          <button type="button" class="whitespace-nowrap rounded-full px-3 py-1 text-xs font-medium chat-filter-btn" :class="activeFilter === 'messages' ? 'bg-[#e97863] text-white active' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'" @click="activeFilter = 'messages'">{{ t('chat.filter_messages') }}</button>
          <button type="button" class="whitespace-nowrap rounded-full px-3 py-1 text-xs font-medium chat-filter-btn" :class="activeFilter === 'unread' ? 'bg-[#e97863] text-white active' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'" @click="activeFilter = 'unread'">{{ t('chat.filter_unread') }}</button>
        </div>

        <!-- Form Buat Ruang (Guru) -->
        <form v-if="roomFormOpen" class="mx-5 mb-3 grid gap-3 rounded-xl border border-[#d8e9df] bg-[#f0f8f2] p-4 chat-room-form" @submit.prevent="saveRoom">
          <div class="flex items-center justify-between">
            <p class="text-xs font-semibold text-gray-700 chat-form-title">{{ t('chat.create_room_title') }}</p>
            <button type="button" aria-label="Tutup form ruang chat" @click="roomFormOpen = false" class="text-gray-400 hover:text-gray-600"><X :size="14" /></button>
          </div>
          <label class="grid gap-1 text-[11px] font-semibold text-gray-600 chat-form-label">
            {{ t('chat.room_name_label') }}
            <input v-model="roomName" class="w-full rounded-lg border border-gray-200 px-3 py-2 text-xs font-normal chat-form-input" :placeholder="t('chat.room_name_placeholder')" required />
          </label>
          <label class="grid gap-1 text-[11px] font-semibold text-gray-600 chat-form-label">
            {{ t('chat.class_label') }}
            <select v-model="selectedClassId" class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-normal chat-form-input" required @change="loadClassStudents">
            <option value="">{{ t('chat.select_class') }}</option>
            <option v-for="classItem in teacherClasses" :key="classItem.id || classItem.ID" :value="classItem.id || classItem.ID">{{ classItem.name || classItem.Name }}</option>
            </select>
          </label>
          <div v-if="selectedClassId" class="max-h-40 overflow-y-auto rounded-lg border border-gray-100 bg-white p-2 chat-form-inner-box">
            <label v-for="student in classStudents" :key="student.id" class="flex items-center gap-2 px-2 py-1 text-xs font-normal chat-form-label"><input v-model="selectedStudentIds" type="checkbox" :value="student.id" />{{ contactName(student) }}</label>
            <p v-if="!classStudents.length" class="p-2 text-xs text-gray-500 chat-subtext">{{ t('chat.no_students') }}</p>
          </div>
          <div class="flex flex-wrap justify-end gap-2 pt-1">
            <button type="button" class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs chat-action-btn" @click="roomFormOpen = false">{{ t('common.cancel') }}</button>
            <button type="submit" class="rounded-lg bg-[#e97863] px-3 py-2 text-xs font-semibold text-white disabled:opacity-50" :disabled="roomSaving">{{ roomSaving ? t('common.saving') : t('chat.create_room') }}</button>
          </div>
        </form>

        <!-- Form Kontak (Admin) -->
        <div v-if="contactsOpen" class="mx-5 mb-3 rounded-xl border border-[#d8e9df] bg-[#f0f8f2] p-4 chat-room-form">
          <div class="mb-2 flex items-center justify-between">
            <p class="text-xs font-semibold text-gray-700 chat-form-title">{{ contactHelp }}</p>
            <button type="button" @click="contactsOpen = false" class="text-gray-400 hover:text-gray-600"><X :size="14" /></button>
          </div>
          <div v-if="!contacts.length" class="text-xs text-gray-500 chat-subtext">{{ t('chat.empty_contacts') }}</div>
          <button
            v-for="contact in contacts"
            :key="contact.id"
            type="button"
            class="flex w-full items-center gap-3 rounded-lg px-2 py-2 text-left hover:bg-white disabled:opacity-50 chat-contact-btn"
            :disabled="!!startingConversation"
            @click="startConversation(contact)"
          >
            <span class="flex h-8 w-8 items-center justify-center rounded-full bg-white text-xs font-semibold text-gray-600 chat-avatar-circle">{{ contactInitial(contact) }}</span>
            <span class="min-w-0">
              <strong class="block truncate text-xs text-gray-900 chat-title-text">{{ contactName(contact) }}</strong>
              <small class="text-[10px] text-gray-500 chat-subtext">{{ contactRole(contact) }}</small>
            </span>
          </button>
        </div>

        <p v-if="errorMessage" class="mx-5 mb-3 rounded-lg border border-red-100 bg-red-50 px-3 py-2 text-xs text-red-700">{{ errorMessage }}</p>

        <!-- Daftar Percakapan -->
        <div class="flex-1 overflow-y-auto px-3">
          <div v-if="loading" class="p-5 text-center text-sm text-gray-500 chat-subtext">
            {{ t('chat.loading') }}
          </div>

          <button
            v-for="conversation in filteredConversations"
            :key="conversation.id"
            type="button"
            class="group mb-1 flex w-full items-center gap-3 rounded-xl px-3 py-3 text-left transition hover:bg-gray-50 chat-convo-item"
            :class="{ 'bg-[#e97863]/5': false }"
            @click="openConversation(conversation.id)"
          >
            <!-- Avatar -->
            <div class="relative flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-gray-100 font-semibold text-gray-600 chat-avatar-circle">
              {{ conversationName(conversation).charAt(0).toUpperCase() }}
              <span class="absolute bottom-0 right-0 h-3 w-3 rounded-full border-2 border-white bg-green-500 chat-avatar-dot"></span>
            </div>

            <!-- Info -->
            <div class="min-w-0 flex-1">
              <div class="flex items-center justify-between gap-2">
                <p class="truncate text-sm font-semibold text-gray-900 chat-convo-name">
                  {{ conversationName(conversation) }}
                </p>
                <span class="text-[10px] text-gray-400">2m</span>
              </div>
              <div class="mt-0.5 flex items-center justify-between gap-2">
                <p class="truncate text-xs text-gray-500 chat-convo-desc">
                  {{ t('chat.click_to_open') }}
                </p>
                <span
                  v-if="unreadFor(conversation.id)"
                  class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-[#3f9b68] text-[10px] font-bold leading-none text-white shadow-sm ring-2 ring-white"
                  :aria-label="`${unreadFor(conversation.id)} ${t('chat.filter_unread')}`"
                >
                  {{ unreadFor(conversation.id) > 9 ? '9+' : unreadFor(conversation.id) }}
                </span>
              </div>
            </div>
          </button>

          <div
            v-if="!loading && chatStore.conversations.length === 0"
            class="p-5 text-center text-sm text-gray-500 chat-subtext"
          >
            {{ t('chat.empty_conversations') }}
          </div>
        </div>
      </aside>

      <!-- MAIN CHAT AREA -->
      <main class="flex flex-1 flex-col bg-[#fafafa] chat-main-area">
        
        <!-- Chat Header (Empty State) -->
        <div class="flex items-center justify-between border-b border-gray-100 bg-white px-6 py-4 chat-topbar">
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 items-center justify-center rounded-full bg-gray-100 text-gray-400 chat-empty-icon-circle">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 10h8m-8 4h5m8-2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <h2 class="text-base font-semibold text-gray-900 chat-title-text">{{ t('chat.select_conversation') }}</h2>
              <p class="text-xs text-gray-500 chat-subtext">{{ t('chat.select_conversation_desc') }}</p>
            </div>
          </div>
        </div>

        <!-- Empty State Content -->
        <div class="flex flex-1 items-center justify-center">
          <div class="text-center px-4">
            <div class="mx-auto flex h-20 w-20 items-center justify-center rounded-full bg-gray-100 chat-empty-icon-circle">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-9 w-9 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 10h8m-8 4h5m8-2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h2 class="mt-4 text-lg font-semibold text-gray-900 chat-title-text">{{ t('chat.welcome_title') }}</h2>
            <p class="mt-1 text-sm text-gray-500 max-w-xs mx-auto chat-subtext">
              {{ t('chat.welcome_desc') }}
            </p>
          </div>
        </div>

      </main>

    </div>
  </div>
  </AppShell>
</template>

<style scoped>
:global(html[data-theme='dark']) .chat-page-bg {
  background-color: #0f172a !important;
}
:global(html[data-theme='dark']) .chat-container-box {
  background-color: #1e293b !important;
  border: 1px solid #334155 !important;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.45) !important;
}
:global(html[data-theme='dark']) .chat-sidebar {
  background-color: #1e293b !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .chat-sidebar-header {
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .chat-title-text,
:global(html[data-theme='dark']) .chat-convo-name {
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .chat-subtext,
:global(html[data-theme='dark']) .chat-convo-desc {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .chat-icon-btn:hover {
  background-color: #273549 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .chat-search-input {
  background-color: #0f172a !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .chat-filter-btn:not(.active) {
  background-color: #334155 !important;
  color: #cbd5e1 !important;
}
:global(html[data-theme='dark']) .chat-filter-btn:not(.active):hover {
  background-color: #475569 !important;
  color: #ffffff !important;
}
:global(html[data-theme='dark']) .chat-convo-item:hover {
  background-color: #273549 !important;
}
:global(html[data-theme='dark']) .chat-avatar-circle {
  background-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .chat-avatar-dot {
  border-color: #1e293b !important;
}
:global(html[data-theme='dark']) .chat-main-area {
  background-color: #0f172a !important;
}
:global(html[data-theme='dark']) .chat-topbar {
  background-color: #1e293b !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .chat-empty-icon-circle {
  background-color: #1e293b !important;
  color: #64748b !important;
}
:global(html[data-theme='dark']) .chat-room-form {
  background-color: #13221b !important;
  border-color: #1f4a40 !important;
}
:global(html[data-theme='dark']) .chat-form-title {
  color: #a7f3d0 !important;
}
:global(html[data-theme='dark']) .chat-form-label {
  color: #cbd5e1 !important;
}
:global(html[data-theme='dark']) .chat-form-input {
  background-color: #1e293b !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .chat-form-inner-box {
  background-color: #1e293b !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .chat-action-btn {
  background-color: #1e293b !important;
  border-color: #334155 !important;
  color: #cbd5e1 !important;
}
:global(html[data-theme='dark']) .chat-action-btn:hover {
  background-color: #273549 !important;
  color: #ffffff !important;
}
:global(html[data-theme='dark']) .chat-contact-btn:hover {
  background-color: #273549 !important;
}
</style>