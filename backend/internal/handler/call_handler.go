package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"twilio-go/internal/model"
	"twilio-go/internal/repository"
	"twilio-go/internal/service"
)

type CallHandler struct {
	service *service.CallService
}

func NewCallHandler(service *service.CallService) *CallHandler {
	return &CallHandler{
		service: service,
	}
}

func (h *CallHandler) Enqueue(c *gin.Context) {
	var call model.Call

	if err := c.ShouldBindJSON(&call); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if call.CallSid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "callSid is required",
		})
		return
	}

	if err := h.service.Enqueue(
		c.Request.Context(),
		&call,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, call)
}

func (h *CallHandler) Queue(c *gin.Context) {
	agentID := c.Query("agentId")

	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "agentId is required",
		})
		return
	}

	calls, err := h.service.GetQueue(
		c.Request.Context(),
		agentID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"calls": calls,
	})
}

// Accept handles POST /dr/accept.
// It claims a waiting call for an agent using a Firestore transaction,
// ensuring only one agent can successfully claim the same call.
func (h *CallHandler) Accept(c *gin.Context) {
	var req struct {
		CallSid string `json:"callSid"`
		AgentID string `json:"agentId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.CallSid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "callSid is required"})
		return
	}

	if req.AgentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId is required"})
		return
	}

	call, agent, err := h.service.AcceptCall(
		c.Request.Context(),
		req.CallSid,
		req.AgentID,
	)

	if err != nil {
		switch {
		case errors.Is(err, repository.ErrCallNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "call not found"})
		case errors.Is(err, repository.ErrAgentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		case errors.Is(err, repository.ErrCallNotWaiting):
			c.JSON(http.StatusConflict, gin.H{"error": "call is no longer available"})
		case errors.Is(err, repository.ErrAgentUnavailable),
			errors.Is(err, repository.ErrAgentAtCapacity):
			c.JSON(http.StatusConflict, gin.H{"error": "agent is not available"})
		case errors.Is(err, repository.ErrLanguageMismatch),
			errors.Is(err, repository.ErrDomainMismatch):
			c.JSON(http.StatusConflict, gin.H{"error": "agent does not match call requirements"})
		case isTelephonyError(err):
			// Call was claimed but conference setup failed; call is now
			// marked telephony_failed in Firestore.
			c.JSON(http.StatusBadGateway, gin.H{"error": "telephony setup failed; call marked as telephony_failed"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"call":  call,
		"agent": agent,
	})
}

// isTelephonyError reports whether the error originated from the telephony
// setup phase (i.e. after the call was successfully claimed in Firestore).
// We detect this by string prefix because the service wraps the error with
// "telephony setup failed:" to avoid a package-level import cycle.
func isTelephonyError(err error) bool {
	if err == nil {
		return false
	}
	return len(err.Error()) >= 18 && err.Error()[:18] == "telephony setup fa"
}

// CallStatus handles POST /dr/call-status.
// It marks a call as completed and decrements the assigned agent's currentCalls
// inside a single Firestore transaction. The operation is idempotent.
func (h *CallHandler) CallStatus(c *gin.Context) {
	var req struct {
		CallSid string `json:"callSid"`
		Status  string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.CallSid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "callSid is required"})
		return
	}

	if req.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported status; only \"completed\" is accepted"})
		return
	}

	result, err := h.service.CompleteCall(c.Request.Context(), req.CallSid)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrCallNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "call not found"})
		case errors.Is(err, repository.ErrAgentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		case errors.Is(err, repository.ErrCallNotAssigned):
			c.JSON(http.StatusConflict, gin.H{"error": "call is not assigned to an agent"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           result.Call.Status,
		"alreadyCompleted": result.AlreadyCompleted,
		"call":             result.Call,
		"agent":            result.Agent,
	})
}
