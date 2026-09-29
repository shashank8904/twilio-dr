package service

import (
	"context"
	"fmt"
	"log"

	"twilio-go/internal/model"
	"twilio-go/internal/repository"
	"twilio-go/internal/telephony"
)

type CallService struct {
	callRepository  *repository.CallRepository
	agentRepository *repository.AgentRepository
	telephony       telephony.TelephonyProvider
}

func NewCallService(
	callRepository *repository.CallRepository,
	agentRepository *repository.AgentRepository,
	tel telephony.TelephonyProvider,
) *CallService {
	return &CallService{
		callRepository:  callRepository,
		agentRepository: agentRepository,
		telephony:       tel,
	}
}

func (s *CallService) Enqueue(
	ctx context.Context,
	call *model.Call,
) error {

	call.Status = "waiting"

	return s.callRepository.Create(ctx, call)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}

func matches(agent model.Agent, call model.Call) bool {
	if !contains(agent.Languages, call.Language) {
		return false
	}

	if !contains(agent.Domains, call.Domain) {
		return false
	}

	if agent.CurrentCalls >= agent.MaxConcurrentCalls {
		return false
	}

	return true
}

func (s *CallService) GetMatchingCalls(
	ctx context.Context,
	agent *model.Agent,
) ([]model.Call, error) {

	calls, err := s.callRepository.GetWaitingCalls(ctx)
	if err != nil {
		return nil, err
	}

	var matching []model.Call

	for _, call := range calls {
		if matches(*agent, call) {
			matching = append(matching, call)
		}
	}

	return matching, nil
}

func (s *CallService) GetQueue(
	ctx context.Context,
	agentID string,
) ([]model.Call, error) {

	agent, err := s.agentRepository.GetByID(ctx, agentID)
	if err != nil {
		return nil, err
	}

	calls, err := s.callRepository.GetWaitingCalls(ctx)
	if err != nil {
		return nil, err
	}

	var matching []model.Call

	for _, call := range calls {
		if matches(*agent, call) {
			matching = append(matching, call)
		}
	}

	return matching, nil
}

// AcceptCall claims a waiting call for an agent, then creates a mock
// telephony conference and connects both the caller and the agent.
//
// The flow is deliberately split:
//
//  1. Short Firestore transaction — atomically claims the call.
//  2. Telephony operations — happen OUTSIDE the transaction so the DB lock
//     is released before any network I/O.
//  3. Second Firestore write — stores the conferenceName on the call.
//
// Failure handling:
// If the telephony step fails after the call has been claimed, we mark the
// call as "telephony_failed" in Firestore (best-effort rollback). This
// prevents the call from silently appearing as "assigned" with no live
// conference.
func (s *CallService) AcceptCall(
	ctx context.Context,
	callSid string,
	agentID string,
) (*model.Call, *model.Agent, error) {

	// ── Step 1: Atomically claim the call ─────────────────────────────────
	call, agent, err := s.callRepository.AcceptCallTx(ctx, callSid, agentID)
	if err != nil {
		return nil, nil, err
	}

	// ── Step 2: Telephony setup ────────────────────────────────────────────
	conferenceName := fmt.Sprintf("dr-%s", callSid)

	if err := s.setupConference(ctx, call, agent, conferenceName); err != nil {
		// Best-effort rollback: mark call as telephony_failed so the system
		// does not silently report a live call that has no conference.
		if rbErr := s.callRepository.MarkTelephonyFailed(ctx, callSid); rbErr != nil {
			log.Printf("ERROR: could not mark call %s as telephony_failed: %v", callSid, rbErr)
		}
		return nil, nil, fmt.Errorf("telephony setup failed: %w", err)
	}

	// ── Step 3: Persist conference name ───────────────────────────────────
	if err := s.callRepository.UpdateConferenceName(ctx, callSid, conferenceName); err != nil {
		// Non-fatal for the PoC: the conference exists; the name is just not
		// stored. Log and continue so the caller gets a useful response.
		log.Printf("WARN: could not persist conferenceName on call %s: %v", callSid, err)
	}

	call.ConferenceName = conferenceName
	return call, agent, nil
}

// setupConference performs the three telephony operations in sequence.
// If any step fails the whole function returns the error.
func (s *CallService) setupConference(
	ctx context.Context,
	call *model.Call,
	agent *model.Agent,
	conferenceName string,
) error {

	if err := s.telephony.CreateConference(ctx, conferenceName); err != nil {
		return fmt.Errorf("CreateConference: %w", err)
	}

	if err := s.telephony.JoinCaller(ctx, call.CallSid, conferenceName); err != nil {
		return fmt.Errorf("JoinCaller: %w", err)
	}

	if err := s.telephony.JoinAgent(ctx, agent.ID, conferenceName); err != nil {
		return fmt.Errorf("JoinAgent: %w", err)
	}

	return nil
}

// CompleteCall marks a call as completed and decrements the assigned agent's
// currentCalls atomically. The operation is idempotent: calling it twice for
// the same call is safe and returns AlreadyCompleted = true on the second call.
func (s *CallService) CompleteCall(
	ctx context.Context,
	callSid string,
) (*repository.CompleteCallTxResult, error) {

	return s.callRepository.CompleteCallTx(ctx, callSid)
}
