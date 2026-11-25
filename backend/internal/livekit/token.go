package livekit

import (
	"encoding/json"
	"time"

	"github.com/livekit/protocol/auth"
)

type TokenOptions struct {
	RoomName           string
	Identity           string
	DisplayName        string
	Role               string
	CanPublishAudio    bool
	CanPublishVideo    bool
	CanPublishData     bool
	CanScreenShare     bool
	IsRoomAdmin        bool
	ValidFor           time.Duration
}

type ParticipantMetadata struct {
	Role               string `json:"role"`
	DisplayName        string `json:"display_name"`
	CanScreenShare     bool   `json:"can_screen_share"`
	CanPublishAudio    bool   `json:"can_publish_audio"`
	CanPublishVideo    bool   `json:"can_publish_video"`
}

// GenerateAccessToken creates a secure, signed LiveKit access token.
func GenerateAccessToken(cfg Config, opts TokenOptions) (string, error) {
	at := auth.NewAccessToken(cfg.APIKey, cfg.APISecret)

	validFor := opts.ValidFor
	if validFor <= 0 {
		validFor = 6 * time.Hour
	}

	canPublish := opts.CanPublishAudio || opts.CanPublishVideo || opts.CanPublishData || opts.CanScreenShare
	canPublishSources := make([]string, 0)
	if opts.CanPublishAudio {
		canPublishSources = append(canPublishSources, "microphone")
	}
	if opts.CanPublishVideo {
		canPublishSources = append(canPublishSources, "camera")
	}
	if opts.CanScreenShare {
		canPublishSources = append(canPublishSources, "screen_share", "screen_share_audio")
	}

	canSubscribe := true
	canPublishBool := canPublish
	canPublishDataBool := opts.CanPublishData

	grant := &auth.VideoGrant{
		RoomJoin:             true,
		Room:                 opts.RoomName,
		CanPublish:           &canPublishBool,
		CanPublishData:       &canPublishDataBool,
		CanSubscribe:         &canSubscribe,
		CanPublishSources:    canPublishSources,
	}

	if opts.IsRoomAdmin {
		grant.RoomAdmin = true
		grant.RoomCreate = true
		grant.RoomList = true
	}

	metaObj := ParticipantMetadata{
		Role:            opts.Role,
		DisplayName:     opts.DisplayName,
		CanScreenShare:  opts.CanScreenShare,
		CanPublishAudio: opts.CanPublishAudio,
		CanPublishVideo: opts.CanPublishVideo,
	}
	metaBytes, _ := json.Marshal(metaObj)

	at.AddGrant(grant).
		SetIdentity(opts.Identity).
		SetName(opts.DisplayName).
		SetMetadata(string(metaBytes)).
		SetValidFor(validFor)

	return at.ToJWT()
}
