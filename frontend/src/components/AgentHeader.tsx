import type { Agent } from '../types';

interface Props {
  agent: Agent;
}

const statusColors: Record<string, string> = {
  available: 'status-available',
  busy: 'status-busy',
  offline: 'status-offline',
};

export function AgentHeader({ agent }: Props) {
  const statusClass = statusColors[agent.status] ?? 'status-offline';

  return (
    <header className="agent-header">
      <div className="header-brand">
        <span className="brand-icon">📞</span>
        <span className="brand-name">TWILIO FLEX DR</span>
      </div>

      <div className="header-agent-info">
        <div className="agent-name">{agent.displayName}</div>
        <div className="agent-meta">
          <span className="meta-label">Languages:</span>{' '}
          <span className="meta-value">{agent.languages.join(', ')}</span>
          <span className="meta-sep">·</span>
          <span className="meta-label">Domains:</span>{' '}
          <span className="meta-value">{agent.domains.join(', ')}</span>
        </div>
      </div>

      <div className="header-right">
        <div className={`status-badge ${statusClass}`}>
          <span className="status-dot" />
          {agent.status.toUpperCase()}
        </div>
        <div className="call-count">
          {agent.currentCalls} / {agent.maxConcurrentCalls} calls
        </div>
      </div>
    </header>
  );
}
