import { defineStore } from 'pinia'

export interface ChatUser {
  id: string
  username: string
  email: string
  display_name?: string
  full_name?: string
  role?: {
    name: string
  }
}

export interface ChatParticipant {
  id: string
  conversation_id: string
  user_id: string
  joined_at: string
  user?: ChatUser
}

export interface ChatConversation {
  id: string
  type: string
  name?: string | null
  owner_id?: string
  class_id?: string
  participants: ChatParticipant[]
  created_at: string
  updated_at: string
}

export interface ChatContact {
  id: string
  username: string
  email?: string | null
  display_name?: string
  full_name?: string
  student?: { full_name?: string }
  teacher?: { full_name?: string }
  role?: { name: string }
  Role?: { Name: string }
}

export interface ChatMessage {
  id: string
  conversation_id: string
  sender_id: string
  message_type: string
  content: string
  created_at: string
  updated_at?: string
  sender?: ChatUser
}

export interface TypingState {
  userId: string
  isTyping: boolean
}

export const useChatStore = defineStore('chat', () => {
  const conversations = ref<ChatConversation[]>([])
  const messages = ref<Record<string, ChatMessage[]>>({})

  const unreadCounts = ref<Record<string, number>>({})

  const onlineUsers = ref<Record<string, boolean>>({})

  const typingUsers = ref<Record<string, string[]>>({})

  const activeConversationId = ref<string | null>(null)

  const loadingConversations = ref(false)
  const loadingMessages = ref(false)

  const setConversations = (
    data: ChatConversation[],
  ) => {
    conversations.value = data
  }

  const setMessages = (
    conversationId: string,
    data: ChatMessage[],
  ) => {
    messages.value[conversationId] = data
  }

  const addMessage = (
    message: ChatMessage,
  ) => {
    const conversationId =
      message.conversation_id

    if (!messages.value[conversationId]) {
      messages.value[conversationId] = []
    }

    const exists =
      messages.value[conversationId].some(
        item => item.id === message.id,
      )

    if (exists) {
      return
    }

    messages.value[conversationId].push(message)

    // Update conversation timestamp
    const conversation =
      conversations.value.find(
        item => item.id === conversationId,
      )

    if (conversation) {
      conversation.updated_at =
        message.created_at
    }
  }

  const setUnreadCount = (
    conversationId: string,
    count: number,
  ) => {
    unreadCounts.value[conversationId] = count
  }

  const incrementUnread = (
    conversationId: string,
  ) => {
    unreadCounts.value[conversationId] =
      (unreadCounts.value[conversationId] || 0) + 1
  }

  const clearUnread = (
    conversationId: string,
  ) => {
    unreadCounts.value[conversationId] = 0
  }

  const setOnline = (
    userId: string,
    online: boolean,
  ) => {
    onlineUsers.value[userId] = online
  }

  const isOnline = (
    userId: string,
  ) => {
    return !!onlineUsers.value[userId]
  }

  const setTyping = (
    conversationId: string,
    userId: string,
    isTyping: boolean,
  ) => {
    if (!typingUsers.value[conversationId]) {
      typingUsers.value[conversationId] = []
    }

    const users =
      typingUsers.value[conversationId]

    if (isTyping) {
      if (!users.includes(userId)) {
        users.push(userId)
      }
    } else {
      typingUsers.value[conversationId] =
        users.filter(
          id => id !== userId,
        )
    }
  }

  const getMessages = (
    conversationId: string,
  ) => {
    return messages.value[conversationId] || []
  }

  const getTypingUsers = (
    conversationId: string,
  ) => {
    return typingUsers.value[conversationId] || []
  }

  const setActiveConversation = (
    conversationId: string | null,
  ) => {
    activeConversationId.value =
      conversationId
  }

  return {
    conversations,
    messages,
    unreadCounts,
    onlineUsers,
    typingUsers,

    activeConversationId,

    loadingConversations,
    loadingMessages,

    setConversations,
    setMessages,
    addMessage,

    setUnreadCount,
    incrementUnread,
    clearUnread,

    setOnline,
    isOnline,

    setTyping,

    getMessages,
    getTypingUsers,

    setActiveConversation,
  }
})