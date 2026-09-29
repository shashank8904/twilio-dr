package service

import (
	"context"

	"twilio-go/internal/model"
	"twilio-go/internal/repository"
)

type AgentService struct {
	repository *repository.AgentRepository
}

func NewAgentService(
	repository *repository.AgentRepository,
) *AgentService {
	return &AgentService{
		repository: repository,
	}
}

func (s *AgentService) CreateAgent(
	ctx context.Context,
	agent *model.Agent,
) error {

	if agent.Status == "" {
		agent.Status = "offline"
	}

	if agent.MaxConcurrentCalls == 0 {
		agent.MaxConcurrentCalls = 1
	}

	return s.repository.Create(ctx, agent)
}

// GetAgent retrieves a single agent by ID.
func (s *AgentService) GetAgent(
	ctx context.Context,
	id string,
) (*model.Agent, error) {

	return s.repository.GetByID(ctx, id)
}
