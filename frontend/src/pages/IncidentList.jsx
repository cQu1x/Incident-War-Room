import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { getIncidents, ApiError } from '../api.js';
import { useAuth } from '../AuthContext.jsx';
import { formatDate, shortRef } from '../utils.js';
import SeverityBadge from '../components/SeverityBadge.jsx';
import StatusBadge from '../components/StatusBadge.jsx';
import Skeleton from '../components/Skeleton.jsx';
import ErrorState from '../components/ErrorState.jsx';
import EmptyState from '../components/EmptyState.jsx';

export default function IncidentList() {
  const navigate = useNavigate();
  const { chatId } = useAuth();
  const [status, setStatus] = useState('loading'); // loading | ready | error
  const [incidents, setIncidents] = useState([]);
  const [error, setError] = useState(null);

  const load = useCallback(async () => {
    setStatus('loading');
    setError(null);
    try {
      const data = await getIncidents(chatId);
      setIncidents(data || []);
      setStatus('ready');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Unexpected error');
      setStatus('error');
    }
  }, [chatId]);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <div style={{ maxWidth: 980, margin: '0 auto', padding: '28px 24px' }}>
      <h1 style={{ fontSize: 20, fontWeight: 600, margin: '0 0 18px' }}>
        Incidents
      </h1>

      {status === 'loading' && <ListSkeleton />}

      {status === 'error' && <ErrorState message={error} onRetry={load} />}

      {status === 'ready' && incidents.length === 0 && (
        <EmptyState message="No incidents yet. They'll show up here once one is opened in Telegram." />
      )}

      {status === 'ready' && incidents.length > 0 && (
        <div
          style={{
            border: '1px solid #ebe9e6',
            borderRadius: 10,
            overflow: 'hidden',
            background: '#fff',
          }}
        >
          {incidents.map((inc) => {
            const created = formatDate(inc.createdAt);
            const closed = formatDate(inc.closedAt);
            return (
              <div
                key={inc.id}
                onClick={() => navigate(`/incidents/${inc.id}`)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 14,
                  padding: '13px 16px',
                  borderBottom: '1px solid #ebe9e6',
                  cursor: 'pointer',
                }}
                onMouseEnter={(e) => (e.currentTarget.style.background = '#f6f5f3')}
                onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
              >
                <span
                  style={{
                    fontFamily: 'ui-monospace, SF Mono, Menlo, Consolas, monospace',
                    fontSize: 11,
                    color: '#8a8983',
                    flex: '0 0 80px',
                  }}
                >
                  {shortRef(inc.id)}
                </span>

                <span style={{ flex: 1, fontSize: 14, fontWeight: 500, minWidth: 0 }}>
                  {inc.title}
                </span>

                <SeverityBadge severity={inc.severity} />
                <StatusBadge status={inc.status} />

                <span style={{ fontSize: 12, color: '#6f6e69', flex: '0 0 90px', textAlign: 'right' }}>
                  {created ? created.short : '-'}
                </span>

                <span style={{ fontSize: 12, color: '#6f6e69', flex: '0 0 90px', textAlign: 'right' }}>
                  {inc.closedAt ? closed.short : 'ongoing'}
                </span>

                <span style={{ flex: '0 0 24px', textAlign: 'right' }}>
                  {inc.reportUrl && (
                    <a
                      href={inc.reportUrl}
                      target="_blank"
                      rel="noreferrer"
                      title="Download report"
                      onClick={(e) => e.stopPropagation()}
                      style={{ color: '#6f6e69' }}
                    >
                      <DownloadIcon />
                    </a>
                  )}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

function ListSkeleton() {
  return (
    <div style={{ border: '1px solid #ebe9e6', borderRadius: 10, overflow: 'hidden' }}>
      {[0, 1, 2, 3, 4].map((i) => (
        <div
          key={i}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 14,
            padding: '13px 16px',
            borderBottom: '1px solid #ebe9e6',
          }}
        >
          <Skeleton width={80} height={11} />
          <Skeleton width="40%" height={14} />
          <Skeleton width={50} height={16} />
          <Skeleton width={60} height={12} />
        </div>
      ))}
    </div>
  );
}

function DownloadIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
      <path d="M8 1.5v8.5M8 10 4.5 6.5M8 10l3.5-3.5" />
      <path d="M2.5 13.5h11" />
    </svg>
  );
}
