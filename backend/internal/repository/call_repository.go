package repository

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"twilio-go/internal/model"
)

// Sentinel errors returned by AcceptCallTx so callers can map them to HTTP
// status codes without inspecting raw Firestore error strings.
var (
	ErrCallNotFound     = errors.New("call not found")
	ErrAgentNotFound    = errors.New("agent not found")
	ErrCallNotWaiting   = errors.New("call is no longer available")
	ErrCallNotAssigned  = errors.New("call is not assigned to an agent")
	ErrAgentUnavailable = errors.New("agent is not available")
	ErrAgentAtCapacity  = errors.New("agent is at max capacity")
	ErrLanguageMismatch = errors.New("agent does not handle call language")
	ErrDomainMismatch   = errors.New("agent does not handle call domain")
)

type CallRepository struct {
	client *firestore.Client
}

func NewCallRepository(client *firestore.Client) *CallRepository {
	return &CallRepository{
		client: client,
	}
}

func (r *CallRepository) Create(
	ctx context.Context,
	call *model.Call,
) error {

	_, err := r.client.
		Collection("drCalls").
		Doc(call.CallSid).
		Set(ctx, call)

	return err

}

func (r *CallRepository) GetWaitingCalls(
	ctx context.Context,
) ([]model.Call, error) {

	snapshot, err := r.client.
		Collection("drCalls").
		Where("status", "==", "waiting").
		Documents(ctx).
		GetAll()

	if err != nil {
		return nil, err
	}

	calls := make([]model.Call, 0, len(snapshot))

	for _, doc := range snapshot {
		var call model.Call

		if err := doc.DataTo(&call); err != nil {
			return nil, err
		}

		calls = append(calls, call)
	}

	return calls, nil
}

// GetByID fetches a single call document by its CallSid.
func (r *CallRepository) GetByID(
	ctx context.Context,
	callSid string,
) (*model.Call, error) {

	doc, err := r.client.
		Collection("drCalls").
		Doc(callSid).
		Get(ctx)

	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrCallNotFound
		}
		return nil, err
	}

	var call model.Call
	if err := doc.DataTo(&call); err != nil {
		return nil, err
	}

	return &call, nil
}

// UpdateConferenceName stores the conference name on a call document.
// This is called after the telephony provider successfully creates the
// conference, outside the main Firestore transaction.
func (r *CallRepository) UpdateConferenceName(
	ctx context.Context,
	callSid string,
	conferenceName string,
) error {
	_, err := r.client.
		Collection("drCalls").
		Doc(callSid).
		Update(ctx, []firestore.Update{
			{Path: "conferenceName", Value: conferenceName},
		})
	return err
}

// MarkTelephonyFailed transitions a call to "telephony_failed" status.
// This is used as a best-effort rollback when the call was claimed in
// Firestore but the telephony setup (conference/join) failed.
func (r *CallRepository) MarkTelephonyFailed(
	ctx context.Context,
	callSid string,
) error {
	_, err := r.client.
		Collection("drCalls").
		Doc(callSid).
		Update(ctx, []firestore.Update{
			{Path: "status", Value: "telephony_failed"},
		})
	return err
}

// AcceptCallTx atomically claims a waiting call for an agent.
// All reads, validations and writes happen inside a single Firestore
// transaction so two concurrent requests for the same call cannot both
// succeed — whichever commits second will see the document already
// modified and receive a conflict error from Firestore.
func (r *CallRepository) AcceptCallTx(
	ctx context.Context,
	callSid string,
	agentID string,
) (*model.Call, *model.Agent, error) {

	var updatedCall model.Call
	var updatedAgent model.Agent

	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// --- Read call ---
		callRef := r.client.Collection("drCalls").Doc(callSid)
		callSnap, err := tx.Get(callRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return ErrCallNotFound
			}
			return err
		}

		var call model.Call
		if err := callSnap.DataTo(&call); err != nil {
			return err
		}

		// --- Read agent ---
		agentRef := r.client.Collection("drAgents").Doc(agentID)
		agentSnap, err := tx.Get(agentRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return ErrAgentNotFound
			}
			return err
		}

		var agent model.Agent
		if err := agentSnap.DataTo(&agent); err != nil {
			return err
		}
		agent.ID = agentSnap.Ref.ID

		// --- Validate inside transaction ---
		if call.Status != "waiting" {
			return ErrCallNotWaiting
		}

		if agent.Status == "offline" {
			return ErrAgentUnavailable
		}

		if agent.CurrentCalls >= agent.MaxConcurrentCalls {
			return ErrAgentAtCapacity
		}

		if !containsStr(agent.Languages, call.Language) {
			return ErrLanguageMismatch
		}

		if !containsStr(agent.Domains, call.Domain) {
			return ErrDomainMismatch
		}

		// --- Apply updates ---
		call.Status = "assigned"
		call.AssignedAgentID = agent.ID

		agent.CurrentCalls++
		if agent.CurrentCalls >= agent.MaxConcurrentCalls {
			agent.Status = "busy"
		} else {
			agent.Status = "available"
		}

		if err := tx.Set(callRef, &call); err != nil {
			return err
		}

		if err := tx.Set(agentRef, &agent); err != nil {
			return err
		}

		updatedCall = call
		updatedAgent = agent

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return &updatedCall, &updatedAgent, nil
}

// CompleteCallTxResult carries the outcome of CompleteCallTx to the caller.
type CompleteCallTxResult struct {
	Call             model.Call
	Agent            model.Agent
	AlreadyCompleted bool
}

// CompleteCallTx atomically marks a call as completed and decrements the
// assigned agent's currentCalls. The operation is idempotent: if the call
// is already completed the transaction does nothing and returns
// AlreadyCompleted = true.
func (r *CallRepository) CompleteCallTx(
	ctx context.Context,
	callSid string,
) (*CompleteCallTxResult, error) {

	var result CompleteCallTxResult

	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// --- Read call ---
		callRef := r.client.Collection("drCalls").Doc(callSid)
		callSnap, err := tx.Get(callRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return ErrCallNotFound
			}
			return err
		}

		var call model.Call
		if err := callSnap.DataTo(&call); err != nil {
			return err
		}
		call.CallSid = callSnap.Ref.ID

		// --- Idempotency: already completed ---
		if call.Status == "completed" {
			result = CompleteCallTxResult{
				Call:             call,
				AlreadyCompleted: true,
			}
			return nil
		}

		// --- Validate the call has an assigned agent ---
		if call.AssignedAgentID == "" {
			return ErrCallNotAssigned
		}

		// --- Read agent ---
		agentRef := r.client.Collection("drAgents").Doc(call.AssignedAgentID)
		agentSnap, err := tx.Get(agentRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return ErrAgentNotFound
			}
			return err
		}

		var agent model.Agent
		if err := agentSnap.DataTo(&agent); err != nil {
			return err
		}
		agent.ID = agentSnap.Ref.ID

		// --- Apply updates ---
		// Decrement safely; never go below 0.
		if agent.CurrentCalls > 0 {
			agent.CurrentCalls--
		}

		if agent.CurrentCalls < agent.MaxConcurrentCalls {
			agent.Status = "available"
		} else {
			agent.Status = "busy"
		}

		call.Status = "completed"

		if err := tx.Set(callRef, &call); err != nil {
			return err
		}
		if err := tx.Set(agentRef, &agent); err != nil {
			return err
		}

		result = CompleteCallTxResult{
			Call:             call,
			Agent:            agent,
			AlreadyCompleted: false,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// containsStr reports whether slice contains target (case-sensitive).
func containsStr(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
