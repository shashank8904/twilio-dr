package repository

import (
	"context"

	"twilio-go/internal/model"

	"cloud.google.com/go/firestore"
)

type AgentRepository struct {
	client *firestore.Client
}

func NewAgentRepository(client *firestore.Client) *AgentRepository {
	return &AgentRepository{
		client: client,
	}
}

func (r *AgentRepository) Create(
	ctx context.Context,
	agent *model.Agent,
) error {

	_, err := r.client.
		Collection("drAgents").
		Doc(agent.ID).
		Set(ctx, agent)

	return err
}

func (r *AgentRepository) GetAvailableAgents(
	ctx context.Context,
) ([]model.Agent, error) {

	snapshot, err := r.client.
		Collection("drAgents").
		Where("status", "==", "available").
		Documents(ctx).
		GetAll()

	if err != nil {
		return nil, err
	}

	agents := make([]model.Agent, 0, len(snapshot))

	for _, doc := range snapshot {
		var agent model.Agent

		if err := doc.DataTo(&agent); err != nil {
			return nil, err
		}

		agent.ID = doc.Ref.ID

		agents = append(agents, agent)
	}

	return agents, nil
}

func (r *AgentRepository) GetByID(
	ctx context.Context,
	id string,
) (*model.Agent, error) {

	doc, err := r.client.
		Collection("drAgents").
		Doc(id).
		Get(ctx)

	if err != nil {
		return nil, err
	}

	var agent model.Agent

	if err := doc.DataTo(&agent); err != nil {
		return nil, err
	}

	agent.ID = doc.Ref.ID

	return &agent, nil
}
