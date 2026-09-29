package telephony

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// ── In-memory types ────────────────────────────────────────────────────────

// Conference represents a mock Twilio conference room.
type Conference struct {
	Name         string
	Status       string // "created" | "active" | "completed"
	Participants map[string]*Participant
}

// Participant represents a caller or agent inside a mock conference.
type Participant struct {
	ID      string
	Type    string // "caller" | "agent"
	CallSid string // set when Type == "caller"
	AgentID string // set when Type == "agent"
	Status  string // "joining" | "connected" | "disconnected"
}

// ── MockTelephonyProvider ──────────────────────────────────────────────────

// MockTelephonyProvider implements TelephonyProvider entirely in memory.
// It is safe for concurrent use via its internal mutex.
// It is ONLY intended for local development and testing.
type MockTelephonyProvider struct {
	mu          sync.Mutex
	conferences map[string]*Conference
}

// NewMockTelephonyProvider returns an initialised MockTelephonyProvider.
func NewMockTelephonyProvider() *MockTelephonyProvider {
	return &MockTelephonyProvider{
		conferences: make(map[string]*Conference),
	}
}

// Conferences returns a snapshot of the current conference state.
// This is a helper for the /mock/telephony/conferences debug endpoint only.
func (m *MockTelephonyProvider) Conferences() []Conference {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]Conference, 0, len(m.conferences))
	for _, c := range m.conferences {
		// Deep-copy so callers cannot mutate internal state.
		participants := make(map[string]*Participant, len(c.Participants))
		for k, p := range c.Participants {
			cp := *p
			participants[k] = &cp
		}
		out = append(out, Conference{
			Name:         c.Name,
			Status:       c.Status,
			Participants: participants,
		})
	}
	return out
}

// GetConference returns a snapshot of a single conference by name.
func (m *MockTelephonyProvider) GetConference(name string) (*Conference, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.conferences[name]
	if !ok {
		return nil, ErrConferenceNotFound
	}

	participants := make(map[string]*Participant, len(c.Participants))
	for k, p := range c.Participants {
		cp := *p
		participants[k] = &cp
	}
	return &Conference{
		Name:         c.Name,
		Status:       c.Status,
		Participants: participants,
	}, nil
}

// ── TelephonyProvider implementation ──────────────────────────────────────

// CreateConference creates a conference if it does not already exist.
// Calling it twice for the same name is idempotent and returns nil.
func (m *MockTelephonyProvider) CreateConference(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.conferences[name]; exists {
		log.Printf("MOCK TWILIO: conference already exists name=%s (idempotent)", name)
		return nil
	}

	m.conferences[name] = &Conference{
		Name:         name,
		Status:       "created",
		Participants: make(map[string]*Participant),
	}

	log.Printf("MOCK TWILIO: conference created name=%s", name)
	return nil
}

// JoinCaller connects the PSTN caller to the named conference.
func (m *MockTelephonyProvider) JoinCaller(_ context.Context, callSid string, conferenceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.conferences[conferenceName]
	if !ok {
		return fmt.Errorf("%w: %s", ErrConferenceNotFound, conferenceName)
	}

	participantID := fmt.Sprintf("caller-%s", callSid)
	c.Participants[participantID] = &Participant{
		ID:      participantID,
		Type:    "caller",
		CallSid: callSid,
		Status:  "connected",
	}

	if c.Status == "created" {
		c.Status = "active"
	}

	log.Printf("MOCK TWILIO: caller joined callSid=%s conference=%s", callSid, conferenceName)
	return nil
}

// JoinAgent connects an agent to the named conference.
func (m *MockTelephonyProvider) JoinAgent(_ context.Context, agentID string, conferenceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.conferences[conferenceName]
	if !ok {
		return fmt.Errorf("%w: %s", ErrConferenceNotFound, conferenceName)
	}

	participantID := fmt.Sprintf("agent-%s", agentID)
	c.Participants[participantID] = &Participant{
		ID:      participantID,
		Type:    "agent",
		AgentID: agentID,
		Status:  "connected",
	}

	if c.Status == "created" {
		c.Status = "active"
	}

	log.Printf("MOCK TWILIO: agent joined agentId=%s conference=%s", agentID, conferenceName)
	return nil
}

// RemoveParticipant marks a participant as disconnected across all conferences.
func (m *MockTelephonyProvider) RemoveParticipant(_ context.Context, participantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range m.conferences {
		if p, ok := c.Participants[participantID]; ok {
			p.Status = "disconnected"
			log.Printf("MOCK TWILIO: participant removed id=%s", participantID)
			return nil
		}
	}

	return fmt.Errorf("%w: %s", ErrParticipantNotFound, participantID)
}

// EndConference marks a conference as completed and disconnects all participants.
// Calling it on an already-completed conference is idempotent.
func (m *MockTelephonyProvider) EndConference(_ context.Context, conferenceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.conferences[conferenceName]
	if !ok {
		return fmt.Errorf("%w: %s", ErrConferenceNotFound, conferenceName)
	}

	if c.Status == "completed" {
		log.Printf("MOCK TWILIO: conference already ended name=%s (idempotent)", conferenceName)
		return nil
	}

	for _, p := range c.Participants {
		p.Status = "disconnected"
	}
	c.Status = "completed"

	log.Printf("MOCK TWILIO: conference ended name=%s", conferenceName)
	return nil
}
