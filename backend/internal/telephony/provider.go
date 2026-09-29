package telephony

import (
	"context"
	"errors"
)

// Sentinel errors returned by TelephonyProvider implementations.
var (
	ErrConferenceNotFound      = errors.New("conference not found")
	ErrConferenceAlreadyExists = errors.New("conference already exists")
	ErrParticipantNotFound     = errors.New("participant not found")
)

// TelephonyProvider abstracts all telephony operations.
// The CallService depends on this interface; it never imports a concrete
// implementation, making it trivial to swap the mock for real Twilio later.
type TelephonyProvider interface {
	// CreateConference creates a new conference with the given name.
	// Implementations must be idempotent: creating the same conference twice
	// must not return an error.
	CreateConference(ctx context.Context, name string) error

	// JoinCaller connects the PSTN caller (identified by callSid) to the
	// named conference.
	JoinCaller(ctx context.Context, callSid string, conferenceName string) error

	// JoinAgent connects an agent (identified by agentID) to the named
	// conference.
	JoinAgent(ctx context.Context, agentID string, conferenceName string) error

	// RemoveParticipant disconnects a participant from any conference they
	// are currently in.
	RemoveParticipant(ctx context.Context, participantID string) error

	// EndConference marks the named conference as completed and disconnects
	// all participants. Implementations must be idempotent.
	EndConference(ctx context.Context, conferenceName string) error
}
