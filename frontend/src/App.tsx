import { useState, useEffect, useRef, useCallback } from 'react';
import { AgentSetup } from './components/AgentSetup';
import { AgentHeader } from './components/AgentHeader';
import { AgentSidebar } from './components/AgentSidebar';
import { QueuePanel } from './components/QueuePanel';
import { ActiveCall } from './components/ActiveCall';
import { MockTelephony } from './components/MockTelephony';
import { getAgent, getQueue, getMockConferences } from './services/api';
import type { Agent, Call, Conference } from './types';

const POLL_MS = 3000;

export default function App() {
  const [agent, setAgent] = useState<Agent | null>(null);
  const [queue, setQueue] = useState<Call[]>([]);
  const [activeCall, setActiveCall] = useState<Call | null>(null);
  const [conferences, setConferences] = useState<Conference[]>([]);

  const [agentLoading, setAgentLoading] = useState(false);
  const [queueLoading, setQueueLoading] = useState(false);
  const [confLoading, setConfLoading] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);

  const agentIdRef = useRef<string | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // ── Polling ───────────────────────────────────────────────────────────────

  const pollAgent = useCallback(async () => {
    const id = agentIdRef.current;
    if (!id) return;
    try {
      const updated = await getAgent(id);
      setAgent(updated);
      setApiError(null);
    } catch {
      setApiError('Cannot reach backend — retrying…');
    }
  }, []);

  const pollQueue = useCallback(async () => {
    const id = agentIdRef.current;
    if (!id) return;
    setQueueLoading(true);
    try {
      const calls = await getQueue(id);
      setQueue(calls);
    } catch {
      // Silent: already showing apiError from agent poll
    } finally {
      setQueueLoading(false);
    }
  }, []);

  const pollConferences = useCallback(async () => {
    setConfLoading(true);
    try {
      const res = await getMockConferences();
      setConferences(res.conferences ?? []);
    } catch {
      // Non-fatal
    } finally {
      setConfLoading(false);
    }
  }, []);

  const startPolling = useCallback(() => {
    if (pollRef.current) clearInterval(pollRef.current);

    // Initial fetch
    void pollAgent();
    void pollQueue();
    void pollConferences();

    pollRef.current = setInterval(() => {
      void pollAgent();
      void pollQueue();
      void pollConferences();
    }, POLL_MS);
  }, [pollAgent, pollQueue, pollConferences]);

  const stopPolling = useCallback(() => {
    if (pollRef.current) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
  }, []);

  useEffect(() => {
    return () => stopPolling();
  }, [stopPolling]);

  // ── Agent onboarding ──────────────────────────────────────────────────────

  function handleAgentCreated(newAgent: Agent) {
    agentIdRef.current = newAgent.id;
    setAgent(newAgent);
    setAgentLoading(false);
    startPolling();
  }

  // After accepting a call the backend returns fresh call + agent objects
  function handleCallAccepted(call: Call, updatedAgent: Agent) {
    setActiveCall(call);
    setAgent(updatedAgent);
    // Remove from queue
    setQueue((prev) => prev.filter((c) => c.callSid !== call.callSid));
    // Refresh conferences so the mock viz updates immediately
    void pollConferences();
  }

  // After ending a call the backend returns the fresh agent object
  function handleCallEnded(updatedAgent: Agent) {
    setActiveCall(null);
    setAgent(updatedAgent);
    void pollQueue();
    void pollConferences();
  }

  // Manual queue refresh (called from QueuePanel)
  function handleQueueRefresh() {
    void pollAgent();
    void pollQueue();
  }

  // ── Render ────────────────────────────────────────────────────────────────

  if (!agent) {
    return <AgentSetup onAgentCreated={handleAgentCreated} />;
  }

  if (agentLoading) {
    return (
      <div className="fullpage-loading">
        <div className="spinner" />
        <p>Loading agent…</p>
      </div>
    );
  }

  return (
    <div className="console-layout">
      <AgentHeader agent={agent} />

      {apiError && (
        <div className="api-error-bar" role="alert">
          ⚠️ {apiError}
        </div>
      )}

      <div className="console-body">
        {/* ── Left: Queue + Active Call ── */}
        <div className="main-col">
          {activeCall ? (
            <ActiveCall
              call={activeCall}
              agent={agent}
              onCallEnded={handleCallEnded}
            />
          ) : (
            <QueuePanel
              calls={queue}
              agentId={agent.id}
              loading={queueLoading}
              onCallAccepted={handleCallAccepted}
              onRefresh={handleQueueRefresh}
            />
          )}
        </div>

        {/* ── Right: Agent info + Mock telephony ── */}
        <div className="side-col">
          <AgentSidebar agent={agent} />
          <MockTelephony conferences={conferences} loading={confLoading} />
        </div>
      </div>
    </div>
  );
}
