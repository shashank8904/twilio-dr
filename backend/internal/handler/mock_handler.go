package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"twilio-go/internal/telephony"
)

// MockTelephonyHandler exposes read-only endpoints for inspecting the
// in-memory mock Twilio state. These routes are ONLY registered in
// development; they would be removed or gated in production.
type MockTelephonyHandler struct {
	provider *telephony.MockTelephonyProvider
}

func NewMockTelephonyHandler(provider *telephony.MockTelephonyProvider) *MockTelephonyHandler {
	return &MockTelephonyHandler{provider: provider}
}

// participantView is the JSON shape returned for a single participant.
type participantView struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	CallSid string `json:"callSid,omitempty"`
	AgentID string `json:"agentId,omitempty"`
	Status  string `json:"status"`
}

// conferenceView is the JSON shape returned for a single conference.
type conferenceView struct {
	Name         string            `json:"name"`
	Status       string            `json:"status"`
	Participants []participantView `json:"participants"`
}

func toConferenceView(c telephony.Conference) conferenceView {
	parts := make([]participantView, 0, len(c.Participants))
	for _, p := range c.Participants {
		parts = append(parts, participantView{
			ID:      p.ID,
			Type:    p.Type,
			CallSid: p.CallSid,
			AgentID: p.AgentID,
			Status:  p.Status,
		})
	}
	return conferenceView{
		Name:         c.Name,
		Status:       c.Status,
		Participants: parts,
	}
}

// ListConferences handles GET /mock/telephony/conferences.
// Returns all conferences currently held in the mock provider.
func (h *MockTelephonyHandler) ListConferences(c *gin.Context) {
	raw := h.provider.Conferences()

	views := make([]conferenceView, 0, len(raw))
	for _, conf := range raw {
		views = append(views, toConferenceView(conf))
	}

	c.JSON(http.StatusOK, gin.H{"conferences": views})
}

// GetConference handles GET /mock/telephony/conferences/:name.
// Returns a single conference by name.
func (h *MockTelephonyHandler) GetConference(c *gin.Context) {
	name := c.Param("name")

	conf, err := h.provider.GetConference(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conference not found"})
		return
	}

	c.JSON(http.StatusOK, toConferenceView(*conf))
}
