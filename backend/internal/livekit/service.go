package livekit

import (
	"context"
	"fmt"
	"strings"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/livekit/protocol/livekit"
)

type Service struct {
	cfg        Config
	roomClient *lksdk.RoomServiceClient
}

func NewService(cfg Config) *Service {
	// Convert ws:// or wss:// to http:// or https:// for RoomServiceClient RPC calls
	httpURL := cfg.URL
	if strings.HasPrefix(httpURL, "ws://") {
		httpURL = "http://" + strings.TrimPrefix(httpURL, "ws://")
	} else if strings.HasPrefix(httpURL, "wss://") {
		httpURL = "https://" + strings.TrimPrefix(httpURL, "wss://")
	}

	var roomClient *lksdk.RoomServiceClient
	if cfg.APIKey != "" && cfg.APISecret != "" {
		roomClient = lksdk.NewRoomServiceClient(httpURL, cfg.APIKey, cfg.APISecret)
	}

	return &Service{
		cfg:        cfg,
		roomClient: roomClient,
	}
}

func (s *Service) GetURL() string {
	return s.cfg.URL
}

func (s *Service) GenerateToken(opts TokenOptions) (string, error) {
	return GenerateAccessToken(s.cfg, opts)
}

// DeleteRoom ends a meeting room and disconnects all participants.
func (s *Service) DeleteRoom(ctx context.Context, roomName string) error {
	if s.roomClient == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.roomClient.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: roomName,
	})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to delete livekit room: %w", err)
	}
	return nil
}

// RemoveParticipant removes / kicks a participant from the LiveKit room.
func (s *Service) RemoveParticipant(ctx context.Context, roomName, identity string) error {
	if s.roomClient == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.roomClient.RemoveParticipant(ctx, &livekit.RoomParticipantIdentity{
		Room:     roomName,
		Identity: identity,
	})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to remove participant from livekit: %w", err)
	}
	return nil
}

// MuteParticipant mutes a specific track of a participant in LiveKit room.
func (s *Service) MuteParticipantTrack(ctx context.Context, roomName, identity, trackSid string, muted bool) error {
	if s.roomClient == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.roomClient.MutePublishedTrack(ctx, &livekit.MuteRoomTrackRequest{
		Room:     roomName,
		Identity: identity,
		TrackSid: trackSid,
		Muted:    muted,
	})
	return err
}

// UpdateParticipantPermissions updates participant's publish/subscribe capabilities.
func (s *Service) UpdateParticipantPermissions(ctx context.Context, roomName, identity string, canPublish, canPublishData, canSubscribe bool, canPublishSources []string) error {
	if s.roomClient == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sources := make([]livekit.TrackSource, 0)
	for _, src := range canPublishSources {
		switch strings.ToLower(src) {
		case "microphone", "mic", "audio":
			sources = append(sources, livekit.TrackSource_MICROPHONE)
		case "camera", "cam", "video":
			sources = append(sources, livekit.TrackSource_CAMERA)
		case "screen_share", "screenshare":
			sources = append(sources, livekit.TrackSource_SCREEN_SHARE)
		case "screen_share_audio":
			sources = append(sources, livekit.TrackSource_SCREEN_SHARE_AUDIO)
		}
	}

	_, err := s.roomClient.UpdateParticipant(ctx, &livekit.UpdateParticipantRequest{
		Room:     roomName,
		Identity: identity,
		Permission: &livekit.ParticipantPermission{
			CanPublish:        canPublish,
			CanPublishData:    canPublishData,
			CanSubscribe:      canSubscribe,
			CanPublishSources: sources,
		},
	})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to update participant permissions: %w", err)
	}
	return nil
}

// ListParticipants returns the list of current active participants from LiveKit.
func (s *Service) ListParticipants(ctx context.Context, roomName string) ([]*livekit.ParticipantInfo, error) {
	if s.roomClient == nil {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := s.roomClient.ListParticipants(ctx, &livekit.ListParticipantsRequest{
		Room: roomName,
	})
	if err != nil {
		return nil, err
	}
	return resp.Participants, nil
}
