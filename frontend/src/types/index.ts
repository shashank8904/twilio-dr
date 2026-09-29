// Matches internal/model/agent.go
export interface Agent {
  id: string;
  displayName: string;
  languages: string[];
  domains: string[];
  status: string; // "offline" | "available" | "busy"
  currentCalls: number;
  maxConcurrentCalls: number;
}

// Matches internal/model/call.go
export interface Call {
  callSid: string;
  caller: string;
  called: string;
  language: string;
  domain: string;
  status: string; // "waiting" | "assigned" | "completed" | "telephony_failed"
  assignedAgentId: string;
  conferenceName: string;
}

// Matches mock_handler.go conferenceView
export interface Participant {
  id: string;
  type: string;
  callSid?: string;
  agentId?: string;
  status: string;
}

export interface Conference {
  name: string;
  status: string;
  participants: Participant[];
}

// Matches POST /dr/accept response
export interface AcceptResponse {
  call: Call;
  agent: Agent;
}

// Matches POST /dr/call-status response
export interface CompleteResponse {
  status: string;
  alreadyCompleted: boolean;
  call: Call;
  agent: Agent;
}

// Matches GET /dr/queue response
export interface QueueResponse {
  calls: Call[] | null;
}

// Matches GET /mock/telephony/conferences response
export interface ConferencesResponse {
  conferences: Conference[];
}
