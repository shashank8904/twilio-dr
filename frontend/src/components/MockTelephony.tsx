import type { Conference } from '../types';

interface Props {
  conferences: Conference[];
  loading: boolean;
}

export function MockTelephony({ conferences, loading }: Props) {
  return (
    <section className="panel mock-panel">
      <div className="panel-header">
        <h2 className="panel-title">
          <span className="panel-icon">🔧</span> Mock Telephony
        </h2>
        <span className="dev-badge">DEV</span>
      </div>

      <div className="panel-body">
        {loading && conferences.length === 0 && (
          <div className="empty-state">
            <div className="spinner" />
          </div>
        )}

        {!loading && conferences.length === 0 && (
          <div className="empty-state">
            <span className="empty-icon">📡</span>
            <p>No active conferences.</p>
          </div>
        )}

        {conferences.map((conf) => (
          <div key={conf.name} className="conference-card">
            <div className="conference-name">
              <span className="conf-icon">🔗</span>
              <span className="code">{conf.name}</span>
              <span className={`status-pill ${conf.status === 'active' ? 'status-active' : ''}`}>
                {conf.status}
              </span>
            </div>

            <div className="participants-list">
              <div className="participants-label">Participants</div>
              {(conf.participants ?? []).map((p) => (
                <div key={p.id} className="participant-row">
                  <span className="participant-type">
                    {p.type === 'caller' ? '📱' : '🎧'}{' '}
                    {p.type}
                  </span>
                  <span className={`participant-status ${p.status === 'connected' ? 'connected' : ''}`}>
                    {p.status}
                  </span>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
