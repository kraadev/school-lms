import type {
  ChatConversation,
  ChatMessage,
  ChatContact,
} from '~/stores/chat'

export function useChat() {
  const config = useRuntimeConfig()

  const apiBase =
    config.public.apiBase as string

  const normalizeUser = (value: any): ChatContact | undefined => {
    if (!value) return undefined
    return {
      id: value.id || value.ID || '',
      username: value.username || value.Username || '',
      email: value.email ?? value.Email ?? null,
      display_name: value.display_name || value.DisplayName || '',
      full_name: value.full_name || value.FullName || '',
      student: value.student || value.Student,
      teacher: value.teacher || value.Teacher,
      role: value.role || value.Role,
    }
  }

  const normalizeConversation = (value: any): ChatConversation => ({
    ...value,
    id: value.id || value.ID || '',
    type: value.type || value.Type || '',
    name: value.name ?? value.Name ?? null,
    owner_id: value.owner_id || value.OwnerID || '',
    class_id: value.class_id || value.ClassID || '',
    participants: (value.participants || value.Participants || []).map((participant: any) => ({
      ...participant,
      id: participant.id || participant.ID || '',
      conversation_id: participant.conversation_id || participant.ConversationID || '',
      user_id: participant.user_id || participant.UserID || '',
      joined_at: participant.joined_at || participant.JoinedAt || '',
      user: normalizeUser(participant.user || participant.User),
    })),
    created_at: value.created_at || value.CreatedAt || '',
    updated_at: value.updated_at || value.UpdatedAt || '',
  })

  const normalizeMessage = (value: any): ChatMessage => ({
    ...value,
    id: value.id || value.ID || '',
    conversation_id: value.conversation_id || value.ConversationID || '',
    sender_id: value.sender_id || value.SenderID || '',
    message_type: value.message_type || value.MessageType || 'text',
    content: value.content || value.Content || '',
    created_at: value.created_at || value.CreatedAt || '',
    updated_at: value.updated_at || value.UpdatedAt,
    sender: normalizeUser(value.sender || value.Sender),
  })

  const getConversations = async () => {
    const response = await $fetch<{
      success: boolean
      data: any[]
    }>(
      `${apiBase}/chat/conversations`,
      {
        headers: {
          Authorization:
            `Bearer ${localStorage.getItem('token')}`,
        },
      },
    )
    return { ...response, data: (response.data || []).map(normalizeConversation) }
  }

  const getContacts = async () => {
    const response = await $fetch<{ success: boolean; data: any[] }>(`${apiBase}/chat/contacts`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    })
    return { ...response, data: (response.data || []).map(normalizeUser).filter((item): item is ChatContact => !!item?.id) }
  }

  const getContactsForClass = async (classId: string) => {
    const response = await $fetch<{ success: boolean; data: any[] }>(`${apiBase}/chat/contacts?class_id=${encodeURIComponent(classId)}`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    })
    return { ...response, data: (response.data || []).map(normalizeUser).filter((item): item is ChatContact => !!item?.id) }
  }

  const getTeacherClasses = async () => {
    return await $fetch<{ success: boolean; data: any[] }>(`${apiBase}/chat/classes`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    })
  }

  const createRoom = async (name: string, classId: string, participantIds: string[]) => {
    const response = await $fetch<{ success: boolean; data: any }>(`${apiBase}/chat/conversations`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
      body: { type: 'class', name, class_id: classId, participant_ids: participantIds },
    })
    return { ...response, data: normalizeConversation(response.data) }
  }

  const updateConversation = async (conversationId: string, name: string) => {
    return await $fetch<{ success: boolean; data: any }>(`${apiBase}/chat/conversations/${conversationId}`, {
      method: 'PUT', headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }, body: { name },
    })
  }

  const deleteConversation = async (conversationId: string) => {
    return await $fetch<{ success: boolean }>(`${apiBase}/chat/conversations/${conversationId}`, {
      method: 'DELETE', headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    })
  }

  const createConversation = async (participantId: string) => {
    const response = await $fetch<{ success: boolean; data: any }>(`${apiBase}/chat/conversations`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
      body: { type: 'private', participant_ids: [participantId] },
    })
    return { ...response, data: normalizeConversation(response.data) }
  }

  const getMessages = async (
    conversationId: string,
  ) => {
    const response = await $fetch<{
      success: boolean
      data: any[]
    }>(
      `${apiBase}/chat/conversations/${conversationId}/messages`,
      {
        headers: {
          Authorization:
            `Bearer ${localStorage.getItem('token')}`,
        },
      },
    )
    return { ...response, data: (response.data || []).map(normalizeMessage) }
  }

  const sendMessage = async (conversationId: string, content: string) => {
    const response = await $fetch<{ success: boolean; data: any }>(`${apiBase}/chat/conversations/${conversationId}/messages`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
      body: { message_type: 'text', content },
    })
    return { ...response, data: normalizeMessage(response.data) }
  }

  const getUnreadCounts = async () => {
    return await $fetch<{
      success: boolean
      data: Record<string, number>
    }>(
      `${apiBase}/chat/unread-counts`,
      {
        headers: {
          Authorization:
            `Bearer ${localStorage.getItem('token')}`,
        },
      },
    )
  }

  return {
    getConversations,
    getContacts,
    getContactsForClass,
    getTeacherClasses,
    createRoom,
    updateConversation,
    deleteConversation,
    createConversation,
    getMessages,
    sendMessage,
    getUnreadCounts,
  }
}