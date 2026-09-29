import { useState, useId } from 'react';
import { createAgent, ApiError } from '../services/api';
import type { Agent } from '../types';

interface Props {
  onAgentCreated: (agent: Agent) => void;
}

function generateAgentId(name: string): string {
  const slug = name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
  return `${slug}-${Date.now()}`;
}

export function AgentSetup({ onAgentCreated }: Props) {
  const formId = useId();

  const [displayName, setDisplayName] = useState('');
  const [languagesInput, setLanguagesInput] = useState('EN, HI');
  const [domainsInput, setDomainsInput] = useState('Support, Sales');
  const [maxConcurrentCalls, setMaxConcurrentCalls] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);

    const languages = languagesInput
      .split(',')
      .map((s) => s.trim().toUpperCase())
      .filter(Boolean);

    const domains = domainsInput
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);

    if (!displayName.trim()) {
      setError('Display name is required.');
      return;
    }
    if (languages.length === 0) {
      setError('At least one language is required.');
      return;
    }
    if (domains.length === 0) {
      setError('At least one domain is required.');
      return;
    }

    const agentId = generateAgentId(displayName);

    setLoading(true);
    try {
      const agent = await createAgent({
        id: agentId,
        displayName: displayName.trim(),
        languages,
        domains,
        status: 'available',
        maxConcurrentCalls,
      });
      // Backend sets status = "offline"; we surface it as-is.
      localStorage.setItem('agentId', agent.id);
      onAgentCreated(agent);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(`Failed to create agent: ${err.message}`);
      } else {
        setError('Unexpected error. Is the backend running?');
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="setup-page">
      <div className="setup-card">
        <div className="setup-logo">
          <span className="brand-icon">📞</span>
          <h1 className="setup-title">TWILIO FLEX DR</h1>
          <p className="setup-subtitle">Agent Console Setup</p>
        </div>

        <form id={`${formId}-form`} onSubmit={handleSubmit} className="setup-form">
          <div className="form-group">
            <label htmlFor={`${formId}-name`} className="form-label">
              Display Name
            </label>
            <input
              id={`${formId}-name`}
              className="form-input"
              type="text"
              placeholder="e.g. Shashank"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              disabled={loading}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor={`${formId}-languages`} className="form-label">
              Languages <span className="form-hint">(comma-separated)</span>
            </label>
            <input
              id={`${formId}-languages`}
              className="form-input"
              type="text"
              placeholder="EN, HI"
              value={languagesInput}
              onChange={(e) => setLanguagesInput(e.target.value)}
              disabled={loading}
            />
          </div>

          <div className="form-group">
            <label htmlFor={`${formId}-domains`} className="form-label">
              Domains / Skills <span className="form-hint">(comma-separated)</span>
            </label>
            <input
              id={`${formId}-domains`}
              className="form-input"
              type="text"
              placeholder="Support, Sales"
              value={domainsInput}
              onChange={(e) => setDomainsInput(e.target.value)}
              disabled={loading}
            />
          </div>

          <div className="form-group">
            <label htmlFor={`${formId}-max-calls`} className="form-label">
              Max Concurrent Calls
            </label>
            <input
              id={`${formId}-max-calls`}
              className="form-input"
              type="number"
              min={1}
              max={10}
              value={maxConcurrentCalls}
              onChange={(e) => setMaxConcurrentCalls(Number(e.target.value))}
              disabled={loading}
            />
          </div>

          {error && (
            <div className="form-error" role="alert">
              {error}
            </div>
          )}

          <button
            id="btn-create-agent"
            type="submit"
            className="btn btn-primary"
            disabled={loading}
          >
            {loading ? 'Creating Agent…' : 'Create Agent & Enter Console'}
          </button>
        </form>
      </div>
    </div>
  );
}
