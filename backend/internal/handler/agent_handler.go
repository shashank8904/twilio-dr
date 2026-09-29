package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"twilio-go/internal/model"
	"twilio-go/internal/service"
)

type AgentHandler struct {
	service *service.AgentService
}

func NewAgentHandler(service *service.AgentService) *AgentHandler {
	return &AgentHandler{
		service: service,
	}
}

func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var agent model.Agent

	if err := c.ShouldBindJSON(&agent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if agent.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id is required",
		})
		return
	}

	if err := h.service.CreateAgent(c.Request.Context(), &agent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, agent)
}

// GetAgent handles GET /dr/agents/:id.
// Returns the current state of an agent — useful for testing after /dr/accept.
func (h *AgentHandler) GetAgent(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	agent, err := h.service.GetAgent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}
