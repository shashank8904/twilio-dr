import { useState } from 'react';
import { acceptCall, ApiError } from '../services/api';
import type { Call, Agent } from '../types';

interface Props {
  calls: Call[];
  agentId: string;
  loading: boolean;
  onCallAccepted: (call: Call, agent: Agent) => void;
  onRefresh: () => void;
}

export function QueuePanel({ calls, agentId, loading, onCallAccepted, onRefresh }: Props) {
  const [accepting, setAccepting] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

  async function handleAccept(callSid: string) {
    setAccepting(callSid);
    setErrors((prev) => ({ ...prev, [callSid]: '' }));

    try {
      const res = await acceptCall(callSid, agentId);
      onCallAccepted(res.call, res.agent);
    } catch (err) {
      let message = 'Failed to accept call.';
      if (err instanceof ApiError) {
        if (err.status === 409) {
          message = 'This call was already accepted by another agent.';
        } else if (err.status === 502) {
          message = 'Call setup failed. The call could not be connected.';
        } else {
          message = err.message;
        }
      }
      setErrors((prev) => ({ ...prev, [callSid]: message }));
      onRefresh();
    } finally {
      setAccepting(null);
    }
  }

  return (
    <section className="panel queue-panel">
      <div className="panel-header">
        <h2 className="panel-title">
          <span className="panel-icon">📋</span> Queue
          {calls.length > 0 && (
            <span className="badge badge-queue">{calls.length}</span>
          )}
        </h2>
        <button
          id="btn-refresh-queue"
          className="btn btn-ghost btn-sm"
          onClick={onRefresh}
          title="Refresh queue"
        >
          ↻
        </button>
      </div>

      <div className="panel-body">
        {loading && calls.length === 0 && (
          <div className="empty-state">
            <div className="spinner" />
            <p>Loading queue…</p>
          </div>
        )}

        {!loading && calls.length === 0 && (
          <div className="empty-state">
            <span className="empty-icon">🎧</span>
            <p>No waiting calls in your queue.</p>
          </div>
        )}

        {calls.map((call) => (
          <div key={call.callSid} className="call-card">
            <div className="call-card-header">
              <span className="call-caller">
                📱 {call.caller || '(unknown)'}
              </span>
              <span className="call-status-pill status-waiting">Waiting</span>
            </div>

            <div className="call-card-meta">
              <div className="meta-row">
                <span className="meta-key">Call SID</span>
                <span className="meta-val code">{call.callSid}</span>
              </div>
              <div className="meta-row">
                <span className="meta-key">Language</span>
                <span className="meta-val">
                  <span className="tag">{call.language}</span>
                </span>
              </div>
              <div className="meta-row">
                <span className="meta-key">Domain</span>
                <span className="meta-val">
                  <span className="tag">{call.domain}</span>
                </span>
              </div>
            </div>

            {errors[call.callSid] && (
              <div className="call-error" role="alert">
                {errors[call.callSid]}
              </div>
            )}

            <button
              id={`btn-accept-${call.callSid}`}
              className="btn btn-accept"
              disabled={accepting === call.callSid}
              onClick={() => handleAccept(call.callSid)}
            >
              {accepting === call.callSid ? 'Accepting…' : '✓ Accept'}
            </button>
          </div>
        ))}
      </div>
    </section>
  );
}
