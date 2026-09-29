import type {
  Agent,
  Call,
  AcceptResponse,
  CompleteResponse,
  QueueResponse,
  ConferencesResponse,
} from '../types';

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '';

async function request<T>(
  path: string,
  options?: RequestInit,
): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });

  const body = await res.json().catch(() => ({}));

  if (!res.ok) {
    const message =
      (body as { error?: string }).error ?? `HTTP ${res.status}`;
    const err = new ApiError(message, res.status);
    throw err;
  }

  return body as T;
}

export class ApiError extends Error {
  readonly status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

// ── Agent ──────────────────────────────────────────────────────────────────

export async function createAgent(agent: {
  id: string;
  displayName: string;
  languages: string[];
  domains: string[];
  status?: string;
  maxConcurrentCalls: number;
}): Promise<Agent> {
  return request<Agent>('/dr/agents', {
    method: 'POST',
    body: JSON.stringify(agent),
  });
}

export async function getAgent(id: string): Promise<Agent> {
  return request<Agent>(`/dr/agents/${encodeURIComponent(id)}`);
}

// ── Queue ──────────────────────────────────────────────────────────────────

export async function getQueue(agentId: string): Promise<Call[]> {
  const res = await request<QueueResponse>(
    `/dr/queue?agentId=${encodeURIComponent(agentId)}`,
  );
  // Backend returns null when there are no matching calls
  return res.calls ?? [];
}

// ── Calls ──────────────────────────────────────────────────────────────────

export async function acceptCall(
  callSid: string,
  agentId: string,
): Promise<AcceptResponse> {
  return request<AcceptResponse>('/dr/accept', {
    method: 'POST',
    body: JSON.stringify({ callSid, agentId }),
  });
}

export async function completeCall(
  callSid: string,
): Promise<CompleteResponse> {
  return request<CompleteResponse>('/dr/call-status', {
    method: 'POST',
    body: JSON.stringify({ callSid, status: 'completed' }),
  });
}

// ── Mock telephony ──────────────────────────────────────────────────────────

export async function getMockConferences(): Promise<ConferencesResponse> {
  return request<ConferencesResponse>('/mock/telephony/conferences');
}
