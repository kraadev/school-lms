import { ref, computed, shallowRef, onBeforeUnmount } from 'vue'
import {
  Room,
  RoomEvent,
  Track,
  Participant,
  RemoteParticipant,
  LocalParticipant,
  ConnectionState,
  type RemoteTrackPublication,
  type LocalTrackPublication,
  type TrackPublication,
} from 'livekit-client'


export interface LiveKitParticipantInfo {
  identity: string
  name: string
  isLocal: boolean
  isSpeaking: boolean
  audioLevel: number
  isAudioEnabled: boolean
  isVideoEnabled: boolean
  isScreenShareEnabled: boolean
  role?: string
  cameraPublication?: TrackPublication
  microphonePublication?: TrackPublication
  screenSharePublication?: TrackPublication
  participant: Participant
}

export function useLiveKit() {
  const room = shallowRef<Room | null>(null)
  const connectionState = ref<ConnectionState>(ConnectionState.Disconnected)
  const isConnected = computed(() => connectionState.value === ConnectionState.Connected)
  const isConnecting = computed(() => connectionState.value === ConnectionState.Connecting)
  const isReconnecting = computed(() => connectionState.value === ConnectionState.Reconnecting)

  const error = ref<string>('')
  const connectionStatusText = ref<string>('Belum terhubung')

  // Local device & media state
  const isMicEnabled = ref<boolean>(true)
  const isCameraEnabled = ref<boolean>(true)
  const isScreenSharing = ref<boolean>(false)

  // Devices
  const audioInputDevices = ref<MediaDeviceInfo[]>([])
  const videoInputDevices = ref<MediaDeviceInfo[]>([])
  const audioOutputDevices = ref<MediaDeviceInfo[]>([])
  const selectedAudioInputId = ref<string>('')
  const selectedVideoInputId = ref<string>('')

  // Participants & Speakers
  const remoteParticipants = ref<RemoteParticipant[]>([])
  const activeSpeakers = ref<Participant[]>([])

  // Refresh device list
  const loadDevices = async () => {
    if (typeof navigator === 'undefined' || !navigator.mediaDevices?.enumerateDevices) return
    try {
      const devices = await navigator.mediaDevices.enumerateDevices()
      audioInputDevices.value = devices.filter((d) => d.kind === 'audioinput')
      videoInputDevices.value = devices.filter((d) => d.kind === 'videoinput')
      audioOutputDevices.value = devices.filter((d) => d.kind === 'audiooutput')
    } catch (e) {
      console.warn('Could not enumerate devices:', e)
    }
  }

  const getParticipantInfo = (p: Participant): LiveKitParticipantInfo => {
    const isLocal = p instanceof LocalParticipant
    let role = 'siswa'
    try {
      if (p.metadata) {
        const meta = JSON.parse(p.metadata)
        if (meta.role) role = meta.role
      }
    } catch {}

    const camPub = p.getTrackPublication(Track.Source.Camera)
    const micPub = p.getTrackPublication(Track.Source.Microphone)
    const screenPub = p.getTrackPublication(Track.Source.ScreenShare)

    // For local participant, isSubscribed is not applicable — check only isMuted + track exists
    // For remote participant, also verify isSubscribed to ensure the track is actually received
    let isVideoEnabled: boolean
    let isAudioEnabled: boolean
    let isScreenShareEnabled: boolean

    if (isLocal) {
      isVideoEnabled = !!camPub?.track && !camPub.isMuted
      isAudioEnabled = !!micPub?.track && !micPub.isMuted
      isScreenShareEnabled = !!screenPub?.track && !screenPub.isMuted
    } else {
      isVideoEnabled = !!camPub && !camPub.isMuted && camPub.isSubscribed
      isAudioEnabled = !!micPub && !micPub.isMuted && micPub.isSubscribed
      isScreenShareEnabled = !!screenPub && !screenPub.isMuted && screenPub.isSubscribed
    }

    return {
      identity: p.identity,
      name: p.name || p.identity,
      isLocal,
      isSpeaking: p.isSpeaking,
      audioLevel: p.audioLevel,
      isAudioEnabled,
      isVideoEnabled,
      isScreenShareEnabled,
      role,
      cameraPublication: camPub,
      microphonePublication: micPub,
      screenSharePublication: screenPub,
      participant: p,
    }
  }

  const participantsList = computed<LiveKitParticipantInfo[]>(() => {
    // Access participantUpdateTick to force reactivity on local participant state changes
    void participantUpdateTick.value
    const list: LiveKitParticipantInfo[] = []
    if (room.value?.localParticipant) {
      list.push(getParticipantInfo(room.value.localParticipant))
    }
    for (const rp of remoteParticipants.value) {
      list.push(getParticipantInfo(rp))
    }
    return list
  })

  // Screen share track present in room
  const currentScreenShare = computed(() => {
    for (const info of participantsList.value) {
      if (info.screenSharePublication?.track) {
        return {
          participant: info,
          track: info.screenSharePublication.track,
        }
      }
    }
    return null
  })

  // Counter to force recompute of participantsList (including local participant state changes)
  const participantUpdateTick = ref(0)

  const updateRemoteParticipants = () => {
    if (!room.value) {
      remoteParticipants.value = []
      participantUpdateTick.value++
      return
    }
    remoteParticipants.value = Array.from(room.value.remoteParticipants.values())
    participantUpdateTick.value++
  }

  const connect = async (
    liveKitUrl: string,
    token: string,
    options?: {
      audio?: boolean
      video?: boolean
      publishOnJoin?: boolean
    }
  ) => {
    error.value = ''
    connectionStatusText.value = 'Menghubungkan ke server media...'
    connectionState.value = ConnectionState.Connecting

    try {
      // Disconnect existing room if any
      if (room.value) {
        await disconnect()
      }

      const lkRoom = new Room({
        adaptiveStream: true,
        dynacast: true,
        videoCaptureDefaults: {
          resolution: { width: 1280, height: 720, frameRate: 30 },
        },
      })

      // Setup event listeners
      lkRoom.on(RoomEvent.ConnectionStateChanged, (state: ConnectionState) => {
        connectionState.value = state
        switch (state) {
          case ConnectionState.Connected:
            connectionStatusText.value = 'Terhubung'
            break
          case ConnectionState.Connecting:
            connectionStatusText.value = 'Menghubungkan...'
            break
          case ConnectionState.Reconnecting:
            connectionStatusText.value = 'Menghubungkan kembali...'
            break
          case ConnectionState.Disconnected:
            connectionStatusText.value = 'Terputus'
            break
        }
      })

      lkRoom.on(RoomEvent.ParticipantConnected, () => {
        updateRemoteParticipants()
      })

      lkRoom.on(RoomEvent.ParticipantDisconnected, () => {
        updateRemoteParticipants()
      })

      lkRoom.on(RoomEvent.TrackSubscribed, () => {
        updateRemoteParticipants()
      })

      lkRoom.on(RoomEvent.TrackUnsubscribed, () => {
        updateRemoteParticipants()
      })

      lkRoom.on(RoomEvent.TrackMuted, () => {
        updateRemoteParticipants()
      })

      lkRoom.on(RoomEvent.TrackUnmuted, () => {
        updateRemoteParticipants()
      })

      // Handle local participant publishing / unpublishing tracks
      // (these are NOT covered by TrackSubscribed/TrackUnsubscribed)
      lkRoom.on(RoomEvent.LocalTrackPublished, () => {
        updateRemoteParticipants()
      })

      lkRoom.on(RoomEvent.LocalTrackUnpublished, () => {
        updateRemoteParticipants()
      })

      lkRoom.on(RoomEvent.ActiveSpeakersChanged, (speakers: Participant[]) => {
        activeSpeakers.value = [...speakers]
      })

      lkRoom.on(RoomEvent.Disconnected, () => {
        connectionState.value = ConnectionState.Disconnected
        updateRemoteParticipants()
      })

      room.value = lkRoom

      await lkRoom.connect(liveKitUrl, token)
      connectionState.value = lkRoom.state
      updateRemoteParticipants()
      await loadDevices()

      // Handle initial media publishing
      const shouldPublish = options?.publishOnJoin !== false
      if (shouldPublish) {
        const publishAudio = options?.audio !== false
        const publishVideo = options?.video !== false

        if (publishAudio) {
          try {
            await lkRoom.localParticipant.setMicrophoneEnabled(true)
            isMicEnabled.value = true
          } catch (e: any) {
            console.warn('Could not enable microphone on join:', e)
            isMicEnabled.value = false
          }
        } else {
          isMicEnabled.value = false
        }

        if (publishVideo) {
          try {
            await lkRoom.localParticipant.setCameraEnabled(true)
            isCameraEnabled.value = true
          } catch (e: any) {
            console.warn('Could not enable camera on join:', e)
            isCameraEnabled.value = false
          }
        } else {
          isCameraEnabled.value = false
        }
      }

      return lkRoom
    } catch (e: any) {
      error.value = e?.message || 'Gagal terhubung ke LiveKit Media Server'
      connectionState.value = ConnectionState.Disconnected
      connectionStatusText.value = 'Gagal terhubung'
      throw e
    }
  }

  const setMicrophone = async (enabled: boolean) => {
    if (!room.value?.localParticipant) return
    try {
      await room.value.localParticipant.setMicrophoneEnabled(enabled)
      isMicEnabled.value = enabled
    } catch (e: any) {
      error.value = 'Gagal mengubah status microphone: ' + (e?.message || '')
    }
  }

  const toggleMicrophone = async () => {
    await setMicrophone(!isMicEnabled.value)
  }

  const setCamera = async (enabled: boolean) => {
    if (!room.value?.localParticipant) return
    try {
      await room.value.localParticipant.setCameraEnabled(enabled)
      isCameraEnabled.value = enabled
    } catch (e: any) {
      isCameraEnabled.value = false
      error.value = 'Gagal mengubah status kamera: ' + (e?.message || '')
      throw e // Re-throw so callers can show specific error messages
    }
  }

  const toggleCamera = async () => {
    await setCamera(!isCameraEnabled.value)
  }

  const setScreenShare = async (enabled: boolean) => {
    if (!room.value?.localParticipant) return
    try {
      await room.value.localParticipant.setScreenShareEnabled(enabled)
      isScreenSharing.value = enabled
    } catch (e: any) {
      if (e?.name !== 'NotAllowedError') {
        error.value = 'Gagal membagikan layar: ' + (e?.message || '')
      }
      isScreenSharing.value = false
    }
  }

  const toggleScreenShare = async () => {
    await setScreenShare(!isScreenSharing.value)
  }

  const switchCamera = async (deviceId: string) => {
    if (!room.value) return
    try {
      await room.value.switchActiveDevice('videoinput', deviceId)
      selectedVideoInputId.value = deviceId
    } catch (e: any) {
      error.value = 'Gagal mengganti kamera: ' + (e?.message || '')
    }
  }

  const switchMicrophone = async (deviceId: string) => {
    if (!room.value) return
    try {
      await room.value.switchActiveDevice('audioinput', deviceId)
      selectedAudioInputId.value = deviceId
    } catch (e: any) {
      error.value = 'Gagal mengganti microphone: ' + (e?.message || '')
    }
  }

  const disconnect = async () => {
    if (room.value) {
      try {
        await room.value.disconnect()
      } catch (e) {
        console.warn('Error disconnecting room:', e)
      }
      room.value = null
    }
    remoteParticipants.value = []
    activeSpeakers.value = []
    connectionState.value = ConnectionState.Disconnected
    connectionStatusText.value = 'Terputus'
    isScreenSharing.value = false
  }

  onBeforeUnmount(() => {
    void disconnect()
  })

  return {
    room,
    connectionState,
    isConnected,
    isConnecting,
    isReconnecting,
    connectionStatusText,
    error,
    isMicEnabled,
    isCameraEnabled,
    isScreenSharing,
    audioInputDevices,
    videoInputDevices,
    audioOutputDevices,
    selectedAudioInputId,
    selectedVideoInputId,
    participantsList,
    remoteParticipants,
    activeSpeakers,
    currentScreenShare,
    connect,
    disconnect,
    setMicrophone,
    toggleMicrophone,
    setCamera,
    toggleCamera,
    setScreenShare,
    toggleScreenShare,
    switchCamera,
    switchMicrophone,
    loadDevices,
  }
}
