import { ref, computed } from 'vue'
import { useApi } from '~/composables/useApi'
import { useAuthStore } from '~/stores/auth'

export interface MeetingData {
  id: string
  host_id: string
  class_subject_id?: string
  title: string
  description?: string
  type: 'voice' | 'video' | 'meeting'
  room_name?: string
  status: 'scheduled' | 'active' | 'live' | 'ended' | 'cancelled'
  approval_required: boolean
  max_participants: number
  scheduled_at?: string
  started_at?: string
  ended_at?: string
  created_at: string
  updated_at: string
  host?: {
    id: string
    username: string
    role?: { name: string }
    teacher?: { full_name: string }
    student?: { full_name: string }
  }
  class_subject?: {
    class?: { name?: string }
    subject?: { name?: string }
  }
  participants?: Array<{
    id: string
    meeting_id: string
    user_id: string
    status: 'requested' | 'approved' | 'rejected' | 'kicked'
    mic_enabled: boolean
    camera_enabled: boolean
    screen_share_enabled: boolean
    joined_at?: string
    left_at?: string
    user?: {
      id: string
      username: string
      role?: { name: string }
      teacher?: { full_name: string }
      student?: { full_name: string }
    }
  }>
}

export interface JoinTokenData {
  token: string
  url: string
  room_name: string
  identity: string
}

export function useMeeting() {
  const { request } = useApi()
  const auth = useAuthStore()

  const meeting = ref<MeetingData | null>(null)
  const loading = ref<boolean>(false)
  const error = ref<string>('')
  const joinData = ref<JoinTokenData | null>(null)
  const pendingRequests = ref<any[]>([])

  const isHost = computed(() => {
    if (!meeting.value || !auth.user) return false
    return meeting.value.host_id === auth.user.id || auth.role === 'guru'
  })

  const myParticipant = computed(() => {
    if (!meeting.value?.participants || !auth.user) return null
    return meeting.value.participants.find((p) => p.user_id === auth.user?.id)
  })

  const isApproved = computed(() => {
    if (isHost.value) return true
    if (!meeting.value?.approval_required) return true
    return myParticipant.value?.status === 'approved'
  })

  const isPendingApproval = computed(() => {
    if (isHost.value) return false
    return meeting.value?.approval_required && myParticipant.value?.status === 'requested'
  })

  const isKicked = computed(() => {
    return myParticipant.value?.status === 'kicked'
  })

  const isRejected = computed(() => {
    return myParticipant.value?.status === 'rejected'
  })

  const loadMeeting = async (meetingId: string) => {
    loading.value = true
    error.value = ''
    try {
      const res = await request<{ data: MeetingData }>(`/meetings/${encodeURIComponent(meetingId)}`)
      meeting.value = res.data
      return res.data
    } catch (e: any) {
      error.value = e?.data?.message || e?.message || 'Gagal memuat informasi meeting'
      throw e
    } finally {
      loading.value = false
    }
  }

  const loadRequests = async (meetingId: string) => {
    if (auth.role !== 'guru') return []
    try {
      const res = await request<{ data: any[] }>(`/meetings/${encodeURIComponent(meetingId)}/requests`)
      pendingRequests.value = res.data || []
      return res.data
    } catch {
      pendingRequests.value = []
      return []
    }
  }

  const requestAccess = async (meetingId: string) => {
    try {
      await request(`/meetings/${encodeURIComponent(meetingId)}/request`, { method: 'POST' })
      await loadMeeting(meetingId)
    } catch (e: any) {
      error.value = e?.data?.message || e?.message || 'Gagal mengajukan akses meeting'
      throw e
    }
  }

  const getJoinToken = async (meetingId: string): Promise<JoinTokenData> => {
    error.value = ''
    try {
      const res = await request<{ data: JoinTokenData }>(`/meetings/${encodeURIComponent(meetingId)}/join`, {
        method: 'POST',
      })
      joinData.value = res.data
      return res.data
    } catch (e: any) {
      const msg = e?.data?.message || e?.message || 'Gagal memperoleh izin masuk meeting'
      error.value = msg
      throw new Error(msg)
    }
  }

  const startMeeting = async (meetingId: string) => {
    try {
      const res = await request<{ data: MeetingData }>(`/meetings/${encodeURIComponent(meetingId)}/start`, {
        method: 'POST',
      })
      meeting.value = res.data
      return res.data
    } catch (e: any) {
      error.value = e?.data?.message || e?.message || 'Gagal memulai meeting'
      throw e
    }
  }

  const endMeeting = async (meetingId: string) => {
    try {
      const res = await request<{ data: MeetingData }>(`/meetings/${encodeURIComponent(meetingId)}/end`, {
        method: 'POST',
      })
      meeting.value = res.data
      return res.data
    } catch (e: any) {
      error.value = e?.data?.message || e?.message || 'Gagal mengakhiri meeting'
      throw e
    }
  }

  const leaveMeeting = async (meetingId: string) => {
    try {
      await request(`/meetings/${encodeURIComponent(meetingId)}/leave`, {
        method: 'POST',
      })
    } catch (e) {
      console.warn('Error recording leave meeting:', e)
    }
  }

  const decideParticipant = async (meetingId: string, userId: string, status: 'approved' | 'rejected') => {
    try {
      await request(`/meetings/${encodeURIComponent(meetingId)}/participants/${encodeURIComponent(userId)}`, {
        method: 'PATCH',
        body: { status },
      })
      await Promise.all([loadMeeting(meetingId), loadRequests(meetingId)])
    } catch (e: any) {
      error.value = e?.data?.message || e?.message || 'Gagal memperbarui status peserta'
      throw e
    }
  }

  const decideAllParticipants = async (meetingId: string, status: 'approved' | 'rejected') => {
    try {
      await request(`/meetings/${encodeURIComponent(meetingId)}/requests/decide-all`, {
        method: 'PATCH',
        body: { status },
      })
      await Promise.all([loadMeeting(meetingId), loadRequests(meetingId)])
    } catch (e: any) {
      // Fallback: iterate over pendingRequests if bulk endpoint fails
      const currentList = [...pendingRequests.value]
      if (currentList.length > 0) {
        await Promise.all(
          currentList.map((r) =>
            request(`/meetings/${encodeURIComponent(meetingId)}/participants/${encodeURIComponent(r.user_id || r.UserID)}`, {
              method: 'PATCH',
              body: { status },
            }).catch(() => {})
          )
        )
        await Promise.all([loadMeeting(meetingId), loadRequests(meetingId)])
        return
      }
      error.value = e?.data?.message || e?.message || 'Gagal memproses semua pengajuan peserta'
      throw e
    }
  }

  const setParticipantPermission = async (
    meetingId: string,
    userId: string,
    permissions: {
      mic_enabled?: boolean
      camera_enabled?: boolean
      screen_share_enabled?: boolean
    }
  ) => {
    try {
      await request(`/meetings/${encodeURIComponent(meetingId)}/participants/${encodeURIComponent(userId)}/permission`, {
        method: 'PATCH',
        body: permissions,
      })
      await loadMeeting(meetingId)
    } catch (e: any) {
      error.value = e?.data?.message || e?.message || 'Gagal mengubah izin peserta'
      throw e
    }
  }

  const kickParticipant = async (meetingId: string, userId: string) => {
    try {
      await request(`/meetings/${encodeURIComponent(meetingId)}/participants/${encodeURIComponent(userId)}/kick`, {
        method: 'POST',
      })
      await loadMeeting(meetingId)
    } catch (e: any) {
      error.value = e?.data?.message || e?.message || 'Gagal mengeluarkan peserta'
      throw e
    }
  }

  const muteParticipant = async (meetingId: string, userId: string, trackSid?: string) => {
    try {
      await request(`/meetings/${encodeURIComponent(meetingId)}/participants/${encodeURIComponent(userId)}/mute`, {
        method: 'POST',
        body: { track_sid: trackSid },
      })
    } catch (e: any) {
      error.value = e?.data?.message || e?.message || 'Gagal memute peserta'
      throw e
    }
  }

  return {
    meeting,
    loading,
    error,
    joinData,
    pendingRequests,
    isHost,
    myParticipant,
    isApproved,
    isPendingApproval,
    isKicked,
    isRejected,
    loadMeeting,
    loadRequests,
    requestAccess,
    getJoinToken,
    startMeeting,
    endMeeting,
    leaveMeeting,
    decideParticipant,
    decideAllParticipants,
    setParticipantPermission,
    kickParticipant,
    muteParticipant,
  }
}
