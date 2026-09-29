import type { Agent } from '../types';

interface Props {
  agent: Agent;
}

export function AgentSidebar({ agent }: Props) {
  const statusClass =
    agent.status === 'available'
      ? 'status-available'
      : agent.status === 'busy'
        ? 'status-busy'
        : 'status-offline';

  return (
    <section className="panel agent-sidebar">
      <div className="panel-header">
        <h2 className="panel-title">
          <span className="panel-icon">👤</span> Agent
        </h2>
      </div>

      <div className="panel-body">
        <div className="agent-avatar">
          {agent.displayName
            .split(' ')
            .map((w) => w[0])
            .join('')
            .toUpperCase()
            .slice(0, 2)}
        </div>
        <div className="agent-display-name">{agent.displayName}</div>
        <div className={`status-badge ${statusClass}`}>
          <span className="status-dot" />
          {agent.status.toUpperCase()}
        </div>

        <div className="sidebar-details">
          <div className="detail-row">
            <span className="detail-key">ID</span>
            <span className="code sidebar-id">{agent.id}</span>
          </div>
          <div className="detail-row">
            <span className="detail-key">Languages</span>
            <span className="tag-group">
              {agent.languages.map((l) => (
                <span key={l} className="tag">{l}</span>
              ))}
            </span>
          </div>
          <div className="detail-row">
            <span className="detail-key">Domains</span>
            <span className="tag-group">
              {agent.domains.map((d) => (
                <span key={d} className="tag">{d}</span>
              ))}
            </span>
          </div>
          <div className="detail-row">
            <span className="detail-key">Calls</span>
            <span>
              {agent.currentCalls} / {agent.maxConcurrentCalls}
            </span>
          </div>
        </div>
      </div>
    </section>
  );
}
