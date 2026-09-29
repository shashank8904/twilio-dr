package telephony_test

import (
	"context"
	"errors"
	"testing"

	"twilio-go/internal/telephony"
)

// newProvider is a helper that returns a fresh MockTelephonyProvider.
func newProvider() *telephony.MockTelephonyProvider {
	return telephony.NewMockTelephonyProvider()
}

var ctx = context.Background()

// ── CreateConference ───────────────────────────────────────────────────────

func TestCreateConference(t *testing.T) {
	p := newProvider()

	if err := p.CreateConference(ctx, "dr-CA001"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, err := p.GetConference("dr-CA001")
	if err != nil {
		t.Fatalf("conference not found after create: %v", err)
	}
	if c.Status != "created" {
		t.Errorf("expected status 'created', got %q", c.Status)
	}
	if len(c.Participants) != 0 {
		t.Errorf("expected no participants, got %d", len(c.Participants))
	}
}

func TestCreateConferenceIdempotent(t *testing.T) {
	p := newProvider()

	// Create once.
	if err := p.CreateConference(ctx, "dr-CA001"); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	// Create again — must not error.
	if err := p.CreateConference(ctx, "dr-CA001"); err != nil {
		t.Fatalf("second create (idempotent) should not error, got: %v", err)
	}

	// State should be the same.
	c, _ := p.GetConference("dr-CA001")
	if c.Status != "created" {
		t.Errorf("expected status 'created' after idempotent create, got %q", c.Status)
	}
}

// ── JoinCaller ────────────────────────────────────────────────────────────

func TestJoinCaller(t *testing.T) {
	p := newProvider()
	_ = p.CreateConference(ctx, "dr-CA001")

	if err := p.JoinCaller(ctx, "CA001", "dr-CA001"); err != nil {
		t.Fatalf("JoinCaller failed: %v", err)
	}

	c, _ := p.GetConference("dr-CA001")
	if c.Status != "active" {
		t.Errorf("expected conference status 'active' after caller joined, got %q", c.Status)
	}

	participantID := "caller-CA001"
	part, ok := c.Participants[participantID]
	if !ok {
		t.Fatalf("participant %q not found", participantID)
	}
	if part.Status != "connected" {
		t.Errorf("expected participant status 'connected', got %q", part.Status)
	}
	if part.Type != "caller" {
		t.Errorf("expected participant type 'caller', got %q", part.Type)
	}
	if part.CallSid != "CA001" {
		t.Errorf("expected CallSid 'CA001', got %q", part.CallSid)
	}
}

func TestJoinCallerMissingConference(t *testing.T) {
	p := newProvider()

	err := p.JoinCaller(ctx, "CA001", "dr-MISSING")
	if err == nil {
		t.Fatal("expected error for missing conference, got nil")
	}
	if !errors.Is(err, telephony.ErrConferenceNotFound) {
		t.Errorf("expected ErrConferenceNotFound, got %v", err)
	}
}

// ── JoinAgent ─────────────────────────────────────────────────────────────

func TestJoinAgent(t *testing.T) {
	p := newProvider()
	_ = p.CreateConference(ctx, "dr-CA001")

	if err := p.JoinAgent(ctx, "agent-001", "dr-CA001"); err != nil {
		t.Fatalf("JoinAgent failed: %v", err)
	}

	c, _ := p.GetConference("dr-CA001")
	if c.Status != "active" {
		t.Errorf("expected conference status 'active' after agent joined, got %q", c.Status)
	}

	participantID := "agent-agent-001"
	part, ok := c.Participants[participantID]
	if !ok {
		t.Fatalf("participant %q not found", participantID)
	}
	if part.Status != "connected" {
		t.Errorf("expected participant status 'connected', got %q", part.Status)
	}
	if part.Type != "agent" {
		t.Errorf("expected participant type 'agent', got %q", part.Type)
	}
	if part.AgentID != "agent-001" {
		t.Errorf("expected AgentID 'agent-001', got %q", part.AgentID)
	}
}

func TestJoinAgentMissingConference(t *testing.T) {
	p := newProvider()

	err := p.JoinAgent(ctx, "agent-001", "dr-MISSING")
	if err == nil {
		t.Fatal("expected error for missing conference, got nil")
	}
	if !errors.Is(err, telephony.ErrConferenceNotFound) {
		t.Errorf("expected ErrConferenceNotFound, got %v", err)
	}
}

// ── Full join sequence ─────────────────────────────────────────────────────

func TestCallerAndAgentBothJoin(t *testing.T) {
	p := newProvider()
	_ = p.CreateConference(ctx, "dr-CA001")
	_ = p.JoinCaller(ctx, "CA001", "dr-CA001")
	_ = p.JoinAgent(ctx, "agent-001", "dr-CA001")

	c, _ := p.GetConference("dr-CA001")
	if len(c.Participants) != 2 {
		t.Errorf("expected 2 participants, got %d", len(c.Participants))
	}
	if c.Status != "active" {
		t.Errorf("expected status 'active', got %q", c.Status)
	}
}

// ── RemoveParticipant ──────────────────────────────────────────────────────

func TestRemoveParticipant(t *testing.T) {
	p := newProvider()
	_ = p.CreateConference(ctx, "dr-CA001")
	_ = p.JoinCaller(ctx, "CA001", "dr-CA001")

	if err := p.RemoveParticipant(ctx, "caller-CA001"); err != nil {
		t.Fatalf("RemoveParticipant failed: %v", err)
	}

	c, _ := p.GetConference("dr-CA001")
	part := c.Participants["caller-CA001"]
	if part.Status != "disconnected" {
		t.Errorf("expected status 'disconnected', got %q", part.Status)
	}
}

func TestRemoveParticipantNotFound(t *testing.T) {
	p := newProvider()

	err := p.RemoveParticipant(ctx, "caller-MISSING")
	if err == nil {
		t.Fatal("expected error for missing participant, got nil")
	}
	if !errors.Is(err, telephony.ErrParticipantNotFound) {
		t.Errorf("expected ErrParticipantNotFound, got %v", err)
	}
}

// ── EndConference ──────────────────────────────────────────────────────────

func TestEndConference(t *testing.T) {
	p := newProvider()
	_ = p.CreateConference(ctx, "dr-CA001")
	_ = p.JoinCaller(ctx, "CA001", "dr-CA001")
	_ = p.JoinAgent(ctx, "agent-001", "dr-CA001")

	if err := p.EndConference(ctx, "dr-CA001"); err != nil {
		t.Fatalf("EndConference failed: %v", err)
	}

	c, _ := p.GetConference("dr-CA001")
	if c.Status != "completed" {
		t.Errorf("expected conference status 'completed', got %q", c.Status)
	}
	for id, part := range c.Participants {
		if part.Status != "disconnected" {
			t.Errorf("expected participant %q to be 'disconnected', got %q", id, part.Status)
		}
	}
}

func TestEndConferenceIdempotent(t *testing.T) {
	p := newProvider()
	_ = p.CreateConference(ctx, "dr-CA001")

	if err := p.EndConference(ctx, "dr-CA001"); err != nil {
		t.Fatalf("first EndConference failed: %v", err)
	}
	// Second call must not error.
	if err := p.EndConference(ctx, "dr-CA001"); err != nil {
		t.Fatalf("second EndConference (idempotent) should not error, got: %v", err)
	}
}

func TestEndConferenceNotFound(t *testing.T) {
	p := newProvider()

	err := p.EndConference(ctx, "dr-MISSING")
	if err == nil {
		t.Fatal("expected error for missing conference, got nil")
	}
	if !errors.Is(err, telephony.ErrConferenceNotFound) {
		t.Errorf("expected ErrConferenceNotFound, got %v", err)
	}
}

// ── State transition ───────────────────────────────────────────────────────

func TestConferenceStateTransitions(t *testing.T) {
	p := newProvider()
	name := "dr-TRANS"

	// Initially does not exist.
	if _, err := p.GetConference(name); !errors.Is(err, telephony.ErrConferenceNotFound) {
		t.Fatalf("expected ErrConferenceNotFound before create, got %v", err)
	}

	// After create: "created".
	_ = p.CreateConference(ctx, name)
	c, _ := p.GetConference(name)
	if c.Status != "created" {
		t.Errorf("after create: expected 'created', got %q", c.Status)
	}

	// After JoinCaller: "active".
	_ = p.JoinCaller(ctx, "CA999", name)
	c, _ = p.GetConference(name)
	if c.Status != "active" {
		t.Errorf("after JoinCaller: expected 'active', got %q", c.Status)
	}

	// After EndConference: "completed".
	_ = p.EndConference(ctx, name)
	c, _ = p.GetConference(name)
	if c.Status != "completed" {
		t.Errorf("after EndConference: expected 'completed', got %q", c.Status)
	}
}
