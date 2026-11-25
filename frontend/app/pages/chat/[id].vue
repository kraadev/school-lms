<script setup lang="ts">
import { nextTick, onMounted, onUnmounted } from 'vue'
import { ArrowLeft, Paperclip, Smile, Send, Pencil, Trash2, CheckCheck } from 'lucide-vue-next'

import { useChatStore } from '~/stores/chat'
import { useChat } from '~/composables/useChat'
import { useChatWebSocket } from '~/composables/useChatWebSocket'
import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'
import { useAppSettings } from '~/composables/useAppSettings'

definePageMeta({
  middleware: 'auth',
})

const route = useRoute()
const conversationId = route.params.id as string

const validConversationId = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(conversationId)

const chatStore = useChatStore()
const auth = useAuthStore()
const { request } = useApi()
const { t, settings } = useAppSettings()

const {
  getMessages,
  sendMessage: sendMessageHttp,
  updateConversation,
  deleteConversation,
} = useChat()

const {
  connected,
  connect,
  sendMessage,
  sendTyping,
  markAsRead,
} = useChatWebSocket()

const input = ref('')
const loading = ref(true)
const sending = ref(false)
const errorMessage = ref('')
const editName = ref('')
const editingRoom = ref(false)

const messagesContainer = ref<HTMLElement | null>(null)
const typingTimeout = ref<ReturnType<typeof setTimeout> | null>(null)

const messages = computed(() => chatStore.getMessages(conversationId))

const conversation = computed(() => chatStore.conversations.find(item => item.id === conversationId))

const senderName = (message: any) => {
  const sender = message.sender || {}
  if ((sender.role?.name || sender.Role?.Name) === 'admin') return 'Admin'
  return String(
    sender.display_name || sender.DisplayName || sender.student?.full_name ||
    sender.Student?.FullName || sender.teacher?.full_name || sender.Teacher?.FullName ||
    sender.username || sender.Username || t('common.user'),
  )
}

const typingUsers = computed(() => chatStore.getTypingUsers(conversationId))

const loadMessages = async () => {
  if (!validConversationId) {
    errorMessage.value = 'ID percakapan tidak valid. Silakan pilih chat dari daftar.'
    loading.value = false
    return
  }
  try {
    loading.value = true
    const response = await getMessages(conversationId)
    chatStore.setMessages(conversationId, response.data)
  } catch (error) {
    console.error('Gagal memuat messages:', error)
  } finally {
    loading.value = false
  }
}

const scrollToBottom = async () => {
  await nextTick()
  if (!messagesContainer.value) return
  messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
}

const submitMessage = () => {
  const content = input.value.trim()
  if (!content) return

  sending.value = true
  errorMessage.value = ''
  if (auth.role === 'admin') {
    sendMessageHttpFallback(content)
  } else if (connected.value) {
    const success = sendMessage(conversationId, content)
    if (success) {
      input.value = ''
      sendTyping(conversationId, false)
      sending.value = false
    } else {
      sendMessageHttpFallback(content)
    }
  } else {
    sendMessageHttpFallback(content)
  }
}

const sendMessageHttpFallback = async (content: string) => {
  try {
    const response = await sendMessageHttp(conversationId, content)
    chatStore.addMessage(response.data)
    input.value = ''
  } catch (error: any) {
    errorMessage.value = error?.data?.message || 'Pesan belum dapat dikirim.'
  } finally {
    sending.value = false
  }
}

const handleTyping = () => {
  sendTyping(conversationId, true)
  if (typingTimeout.value) clearTimeout(typingTimeout.value)
  typingTimeout.value = setTimeout(() => {
    sendTyping(conversationId, false)
  }, 2000)
}

const markVisibleMessagesAsRead = () => {
  const ids = messages.value
    .filter(message => message.sender_id !== getCurrentUserId())
    .map(message => message.id)

  if (ids.length === 0) return

  markAsRead(conversationId, ids)
  chatStore.clearUnread(conversationId)
}

const getCurrentUserId = () => {
  const user = localStorage.getItem('user')
  if (!user) return ''
  try {
    const parsed = JSON.parse(user)
    return parsed.id || ''
  } catch {
    return ''
  }
}

const canEditRoom = computed(() => auth.role === 'guru' && conversation.value?.type === 'class' && conversation.value?.owner_id === auth.user?.id)
const canDeleteRoom = computed(() => auth.role === 'guru' && conversation.value?.type === 'class' && conversation.value?.owner_id === auth.user?.id)
const canSendMessage = computed(() =>
  conversation.value?.type === 'private'
    ? auth.role === 'admin'
    : conversation.value?.type === 'class' && (auth.role === 'guru' || auth.role === 'siswa'),
)
const switchConversation = () => navigateTo('/chat')
const closeConversation = () => navigateTo('/chat')

const saveRoomName = async () => {
  try {
    await updateConversation(conversationId, editName.value.trim())
    if (conversation.value) conversation.value.name = editName.value.trim()
    editingRoom.value = false
  } catch (error: any) {
    errorMessage.value = error?.data?.message || 'Ruang chat belum dapat diperbarui.'
  }
}

const removeRoom = async () => {
  if (!confirm('Hapus ruang chat ini?')) return
  try {
    await deleteConversation(conversationId)
    await navigateTo('/chat')
  } catch (error: any) {
    errorMessage.value = error?.data?.message || 'Ruang chat belum dapat dihapus.'
  }
}

watch(messages, async () => { await scrollToBottom() }, { deep: true })

onMounted(async () => {
  const session = await request<{ user?: any }>('/auth/me').catch(() => null)
  if (session?.user) auth.updateUser(session.user)
  if (!validConversationId) {
    await navigateTo('/chat')
    return
  }
  chatStore.setActiveConversation(conversationId)
  await loadMessages()
  connect()
  await scrollToBottom()
  markVisibleMessagesAsRead()
})

onUnmounted(() => {
  chatStore.setActiveConversation(null)
  if (typingTimeout.value) clearTimeout(typingTimeout.value)
})
</script>

<template>
  <AppShell active="chat" eyebrow="School Space" :title="t('nav.chat')">
  <div class="min-h-screen bg-[#f0f2f5] p-4 md:p-6 chat-page-bg">
    <div class="mx-auto flex h-[calc(100vh-3rem)] max-w-7xl overflow-hidden rounded-2xl bg-white shadow-lg chat-container-box">

      <!-- MAIN CHAT AREA -->
      <main class="flex flex-1 flex-col bg-white chat-main-area">

        <!-- HEADER -->
        <header class="flex h-[72px] shrink-0 items-center gap-4 border-b border-gray-100 px-5 chat-topbar">
          <!-- Back button (mobile) -->
          <button
            type="button"
            class="flex h-9 w-9 items-center justify-center rounded-full text-gray-500 transition hover:bg-gray-100 lg:hidden chat-icon-btn"
            @click="navigateTo('/chat')"
          >
            <ArrowLeft :size="18" />
          </button>

          <!-- Avatar -->
          <div class="relative flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-gray-100 font-semibold text-gray-600 chat-avatar-circle">
            {{ (conversation?.name || 'C').charAt(0).toUpperCase() }}
            <span class="absolute bottom-0 right-0 h-3 w-3 rounded-full border-2 border-white bg-green-500 chat-avatar-dot"></span>
          </div>

          <!-- Info -->
          <div class="min-w-0 flex-1">
            <h1 class="truncate text-sm font-bold text-gray-900 chat-title-text">
              {{ conversation?.name || (settings.language === 'en' ? 'Conversation' : 'Percakapan') }}
            </h1>
            <div class="flex items-center gap-2">
              <span class="h-2 w-2 rounded-full" :class="connected ? 'bg-green-500' : 'bg-gray-400'" />
              <span class="text-xs text-gray-500 chat-subtext">{{ connected ? t('common.online') : t('common.offline') }}</span>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex items-center gap-1">
            <button
              v-if="canEditRoom"
              class="flex h-9 w-9 items-center justify-center rounded-full text-gray-400 transition hover:bg-gray-100 hover:text-gray-600 chat-icon-btn"
              @click="editName = conversation?.name || ''; editingRoom = !editingRoom"
            >
              <Pencil :size="16" />
            </button>
            <button
              v-if="canDeleteRoom"
              class="flex h-9 w-9 items-center justify-center rounded-full text-red-400 transition hover:bg-red-50 hover:text-red-600"
              @click="removeRoom"
            >
              <Trash2 :size="16" />
            </button>
            <button type="button" class="rounded-lg border border-gray-200 px-3 py-2 text-xs font-medium text-gray-600 hover:bg-gray-50 chat-action-btn" @click="switchConversation">{{ t('chat.switch_chat') }}</button>
            <button type="button" class="rounded-lg bg-[#e97863] px-3 py-2 text-xs font-semibold text-white hover:bg-[#d96552]" @click="closeConversation">{{ t('common.close') }}</button>
          </div>
        </header>

        <!-- Edit Room Form -->
        <form v-if="editingRoom" class="border-b border-gray-100 bg-white p-3 chat-topbar" @submit.prevent="saveRoomName">
          <div class="mx-auto flex max-w-4xl gap-2">
            <input v-model="editName" class="flex-1 rounded-xl border border-gray-200 bg-gray-50 px-4 py-2 text-sm outline-none transition focus:border-[#e97863] focus:bg-white focus:ring-1 focus:ring-[#e97863] chat-search-input" required />
            <button type="submit" class="rounded-xl bg-[#e97863] px-4 py-2 text-xs font-semibold text-white transition hover:bg-[#d96552]">{{ t('common.save') }}</button>
            <button type="button" class="rounded-xl border border-gray-200 px-4 py-2 text-xs font-medium text-gray-600 transition hover:bg-gray-50 chat-action-btn" @click="editingRoom = false">{{ t('common.cancel') }}</button>
          </div>
        </form>

        <!-- MESSAGES -->
        <main ref="messagesContainer" class="flex-1 overflow-y-auto bg-[#fafafa] px-5 py-6 chat-main-area">
          <p v-if="errorMessage" class="mx-auto mb-3 max-w-4xl rounded-xl bg-red-50 px-4 py-3 text-xs text-red-700">{{ errorMessage }}</p>

          <div v-if="loading" class="flex h-full items-center justify-center text-sm text-gray-500 chat-subtext">
            {{ t('chat.loading_messages') }}
          </div>

          <div v-else class="mx-auto max-w-4xl space-y-4">
            <div
              v-for="message in messages"
              :key="message.id"
              class="flex"
              :class="message.sender_id === getCurrentUserId() ? 'justify-end' : 'justify-start'"
            >
              <div class="flex max-w-[75%] flex-col" :class="message.sender_id === getCurrentUserId() ? 'items-end' : 'items-start'">
                <!-- Sender name -->
                <p class="mb-1 px-2 text-[10px] font-semibold" :class="message.sender_id === getCurrentUserId() ? 'text-[#e97863]' : 'text-gray-500 chat-subtext'">
                  {{ senderName(message) }}
                </p>

                <!-- Bubble -->
                <div
                  class="rounded-2xl px-4 py-2.5 shadow-sm"
                  :class="
                    message.sender_id === getCurrentUserId()
                      ? 'rounded-br-md bg-[#e97863] text-white chat-bubble-outgoing'
                      : 'rounded-bl-md bg-white text-gray-900 ring-1 ring-gray-100 chat-bubble-incoming'
                  "
                >
                  <p class="whitespace-pre-wrap break-words text-sm leading-relaxed">{{ message.content }}</p>
                </div>

                <!-- Time + status -->
                <div class="mt-1 flex items-center gap-1 px-2">
                  <p class="text-[10px] text-gray-400">
                    {{ new Date(message.created_at).toLocaleTimeString(settings.language === 'en' ? 'en-US' : 'id-ID', { hour: '2-digit', minute: '2-digit' }) }}
                  </p>
                  <CheckCheck v-if="message.sender_id === getCurrentUserId()" :size="12" class="text-[#e97863]" />
                </div>
              </div>
            </div>

            <!-- Typing Indicator -->
            <div v-if="typingUsers.length" class="flex justify-start">
              <div class="flex items-center gap-2 rounded-2xl rounded-bl-md bg-white px-4 py-3 shadow-sm ring-1 ring-gray-100 chat-bubble-typing">
                <div class="flex gap-1">
                  <span class="h-2 w-2 animate-bounce rounded-full bg-gray-400" style="animation-delay: 0ms"></span>
                  <span class="h-2 w-2 animate-bounce rounded-full bg-gray-400" style="animation-delay: 150ms"></span>
                  <span class="h-2 w-2 animate-bounce rounded-full bg-gray-400" style="animation-delay: 300ms"></span>
                </div>
                <span class="text-xs text-gray-500 chat-subtext">{{ t('chat.typing') }}</span>
              </div>
            </div>
          </div>
        </main>

        <!-- INPUT -->
        <footer v-if="canSendMessage" class="shrink-0 border-t border-gray-100 bg-white p-4 chat-footer-bar">
          <form class="mx-auto flex max-w-4xl items-end gap-3" @submit.prevent="submitMessage">
            <div class="flex flex-1 items-end gap-2 rounded-2xl border border-gray-200 bg-[#f7f8fa] px-3 py-2 transition focus-within:border-[#e97863] focus-within:bg-white focus-within:ring-1 focus-within:ring-[#e97863] chat-input-wrapper">
              <button type="button" class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-gray-400 transition hover:bg-gray-200 hover:text-gray-600 chat-icon-btn">
                <Paperclip :size="16" />
              </button>
              <textarea
                v-model="input"
                rows="1"
                :placeholder="t('chat.input_placeholder')"
                class="max-h-32 min-h-[36px] flex-1 resize-none bg-transparent py-1.5 text-sm text-gray-700 outline-none placeholder:text-gray-400 chat-input-textarea"
                @input="handleTyping"
                @keydown.enter.exact.prevent="submitMessage"
              />
              <button type="button" class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-gray-400 transition hover:bg-gray-200 hover:text-gray-600 chat-icon-btn">
                <Smile :size="16" />
              </button>
            </div>

            <button
              type="submit"
              :disabled="!input.trim() || !connected || sending"
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-[#e97863] text-white transition hover:bg-[#d96552] disabled:cursor-not-allowed disabled:opacity-50"
            >
              <Send :size="18" />
            </button>
          </form>
          <p class="mx-auto mt-2 max-w-4xl text-center text-[10px] text-gray-400">
            {{ t('chat.press_enter') }}
          </p>
        </footer>

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
:global(html[data-theme='dark']) .chat-main-area {
  background-color: #0f172a !important;
}
:global(html[data-theme='dark']) .chat-topbar {
  background-color: #1e293b !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .chat-title-text {
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .chat-subtext {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .chat-icon-btn:hover {
  background-color: #273549 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .chat-avatar-circle {
  background-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .chat-avatar-dot {
  border-color: #1e293b !important;
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
:global(html[data-theme='dark']) .chat-search-input {
  background-color: #0f172a !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .chat-bubble-incoming {
  background-color: #1e293b !important;
  color: #f1f5f9 !important;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3) !important;
  --tw-ring-color: #334155 !important;
}
:global(html[data-theme='dark']) .chat-bubble-incoming p {
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .chat-bubble-typing {
  background-color: #1e293b !important;
  --tw-ring-color: #334155 !important;
}
:global(html[data-theme='dark']) .chat-footer-bar {
  background-color: #1e293b !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .chat-input-wrapper {
  background-color: #0f172a !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .chat-input-textarea {
  color: #f1f5f9 !important;
}
</style>