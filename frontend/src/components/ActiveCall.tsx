import { useState } from 'react';
import { completeCall, ApiError } from '../services/api';
import type { Call, Agent } from '../types';

interface Props {
  call: Call;
  agent: Agent;
  onCallEnded: (updatedAgent: Agent) => void;
}

export function ActiveCall({ call, agent, onCallEnded }: Props) {
  const [muted, setMuted] = useState(false);
  const [ending, setEnding] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleEndCall() {
    setEnding(true);
    setError(null);
    try {
      const res = await completeCall(call.callSid);
      onCallEnded(res.agent);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(`End call failed: ${err.message}`);
      } else {
        setError('Unexpected error ending call.');
      }
    } finally {
      setEnding(false);
    }
  }

  return (
    <section className="panel active-call-panel">
      <div className="panel-header">
        <h2 className="panel-title">
          <span className="active-call-dot" />
          ACTIVE CALL
        </h2>
      </div>

      <div className="panel-body">
        <div className="active-call-caller">
          <span className="caller-icon">📱</span>
          <span className="caller-number">{call.caller || '(unknown)'}</span>
        </div>

        <div className="connected-badge">
          <span className="pulse-ring" />
          CONNECTED
        </div>

        <div className="call-details-grid">
          <div className="detail-row">
            <span className="detail-key">Language</span>
            <span className="tag">{call.language}</span>
          </div>
          <div className="detail-row">
            <span className="detail-key">Domain</span>
            <span className="tag">{call.domain}</span>
          </div>
          <div className="detail-row">
            <span className="detail-key">Call SID</span>
            <span className="code">{call.callSid}</span>
          </div>
          {call.conferenceName && (
            <div className="detail-row">
              <span className="detail-key">Conference</span>
              <span className="code">{call.conferenceName}</span>
            </div>
          )}
          <div className="detail-row">
            <span className="detail-key">Agent</span>
            <span>{agent.displayName}</span>
          </div>
        </div>

        {error && (
          <div className="call-error" role="alert">
            {error}
          </div>
        )}

        <div className="call-actions">
          <button
            id="btn-mute"
            className={`btn ${muted ? 'btn-muted' : 'btn-ghost'}`}
            onClick={() => setMuted((m) => !m)}
          >
            {muted ? '🔇 Unmute' : '🎙 Mute'}
          </button>

          <button
            id="btn-end-call"
            className="btn btn-danger"
            onClick={handleEndCall}
            disabled={ending}
          >
            {ending ? 'Ending…' : '📵 End Call'}
          </button>
        </div>
      </div>
    </section>
  );
}
