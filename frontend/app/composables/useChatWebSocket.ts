import { ref, onUnmounted } from 'vue'
import { useChatStore } from '~/stores/chat'
import { useAppSettings } from '~/composables/useAppSettings'

export interface ChatWebSocketMessage {
  type: string
  data?: any

  conversation_id?: string
  user_id?: string

  is_online?: boolean
  is_typing?: boolean

  message_ids?: string[]

  notification_type?: string
  message_id?: string
}

export function useChatWebSocket() {
  const config = useRuntimeConfig()
  const chatStore = useChatStore()
  const { settings, playNotificationChime, showNativeNotification } = useAppSettings()

  const socket = ref<WebSocket | null>(null)
  const connected = ref(false)

  const messages = ref<
    ChatWebSocketMessage[]
  >([])

  const normalizeMessage = (value: any) => ({
    ...value,
    id: value.id || value.ID || '',
    conversation_id: value.conversation_id || value.ConversationID || '',
    sender_id: value.sender_id || value.SenderID || '',
    message_type: value.message_type || value.MessageType || 'text',
    content: value.content || value.Content || '',
    created_at: value.created_at || value.CreatedAt || '',
  })

  const connect = () => {
    if (
      socket.value?.readyState ===
      WebSocket.OPEN
    ) {
      return
    }

    const token =
      localStorage.getItem('token')

    if (!token) {
      console.error(
        'Token tidak ditemukan',
      )

      return
    }

    const apiBase = getResolvedApiBase(config.public.apiBase as string)

    let url: string
    if (apiBase.startsWith('/')) {
      const protocol = typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = typeof window !== 'undefined' ? window.location.host : 'localhost:3000'
      url = `${protocol}//${host}${apiBase}/ws/chat`
    } else {
      const wsBase = apiBase
        .replace(/^http:\/\//, 'ws://')
        .replace(/^https:\/\//, 'wss://')
      url = `${wsBase}/ws/chat`
    }

    socket.value = new WebSocket(
      url,
      [
        'access_token',
        token,
      ],
    )

    socket.value.onopen = () => {
      connected.value = true

      console.log(
        'WebSocket connected',
      )
    }

    socket.value.onmessage = (
      event,
    ) => {
      try {
        const data:
          ChatWebSocketMessage =
          JSON.parse(event.data)

        messages.value.push(data)

        handleEvent(data)
      } catch (error) {
        console.error(
          'WebSocket message invalid:',
          error,
        )
      }
    }

    socket.value.onclose = () => {
      connected.value = false

      console.log(
        'WebSocket disconnected',
      )
    }

    socket.value.onerror = error => {
      console.error(
        'WebSocket error:',
        error,
      )
    }
  }

  const handleEvent = (
    event: ChatWebSocketMessage,
  ) => {
    switch (event.type) {

      case 'message':
        if (event.data) {
          const normalized = normalizeMessage(event.data)
          chatStore.addMessage(normalized)
          if (settings.notifySound) {
            playNotificationChime()
          }
          if (settings.notifyPush) {
            showNativeNotification('Pesan Baru', normalized.content || 'Ada pesan baru masuk')
          }
        }
        break

      case 'typing':
        if (
          event.conversation_id &&
          event.user_id
        ) {
          chatStore.setTyping(
            event.conversation_id,
            event.user_id,
            !!event.is_typing,
          )
        }
        break

      case 'read':
        // Read receipt nanti kita
        // gunakan untuk UI "dibaca".
        break

      case 'presence':
        if (event.user_id) {
          chatStore.setOnline(
            event.user_id,
            !!event.is_online,
          )
        }
        break

      case 'notification':
        handleNotification(
          event,
        )
        break

      case 'error':
        console.error(
          'WebSocket server error:',
          event,
        )
        break

      default:
        console.log(
          'Unknown WebSocket event:',
          event,
        )
    }
  }

  const handleNotification = (
    event: ChatWebSocketMessage,
  ) => {
    if (
      event.notification_type !==
      'chat_message'
    ) {
      return
    }

    if (
      !event.conversation_id
    ) {
      return
    }

    if (
      chatStore.activeConversationId !==
      event.conversation_id
    ) {
      chatStore.incrementUnread(
        event.conversation_id,
      )
    }
  }

  const send = (
    data: Record<string, any>,
  ) => {
    if (
      !socket.value ||
      socket.value.readyState !==
        WebSocket.OPEN
    ) {
      console.error(
        'WebSocket belum terhubung',
      )

      return false
    }

    socket.value.send(
      JSON.stringify(data),
    )

    return true
  }

  const sendMessage = (
    conversationID: string,
    content: string,
  ) => {
    return send({
      type: 'message',
      conversation_id:
        conversationID,
      message_type: 'text',
      content,
    })
  }

  const sendTyping = (
    conversationID: string,
    isTyping: boolean,
  ) => {
    return send({
      type: 'typing',
      conversation_id:
        conversationID,
      is_typing: isTyping,
    })
  }

  const markAsRead = (
    conversationID: string,
    messageIDs: string[],
  ) => {
    return send({
      type: 'read',
      conversation_id:
        conversationID,
      message_ids: messageIDs,
    })
  }

  const disconnect = () => {
    if (socket.value) {
      socket.value.close()
      socket.value = null
    }

    connected.value = false
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    socket,
    connected,
    messages,

    connect,
    disconnect,
    send,

    sendMessage,
    sendTyping,
    markAsRead,
  }
}