import Avatar from '../components/Avatar.jsx';
import EmptyState from '../components/EmptyState.jsx';
import { formatDate } from '../utils.js';

const EVENT_MARKERS = {
  INCIDENT_CREATED: { label: 'Incident opened', dot: '#dc2626', bg: '#fdeceb', text: '#b42318' },
  INCIDENT_CLOSED: { label: 'Incident closed', dot: '#3f7d4f', bg: '#ecf5ee', text: '#3f7d4f' },
  INCIDENT_REOPENED: { label: 'Incident reopened', dot: '#b45309', bg: '#f8f1e3', text: '#b45309' },
  SEVERITY_CHANGED: { label: 'Severity changed', dot: '#7c3aed', bg: '#f2ecfb', text: '#6d28d9' },
};

function markerFor(type) {
  return (
    EVENT_MARKERS[type] || {
      label: 'Comment',
      dot: '#a8a69f',
      bg: '#f0efec',
      text: '#6f6e69',
    }
  );
}

export default function TimelineTab({ events }) {
  if (!events || events.length === 0) {
    return <EmptyState message="No timeline events yet." />;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {events.map((event) => {
        const marker = markerFor(event.type);
        const time = formatDate(event.createdAt);
        return (
          <div key={event.id} style={{ display: 'flex', gap: 12 }}>
            <div
              style={{
                width: 8,
                height: 8,
                borderRadius: '50%',
                background: marker.dot,
                marginTop: 6,
                flex: 'none',
              }}
            />
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                <Avatar username={event.username} size={18} />
                <span style={{ fontSize: 13, fontWeight: 500 }}>{event.username}</span>
                <span
                  className="wr-badge"
                  style={{ color: marker.text, background: marker.bg, fontSize: 10 }}
                >
                  {marker.label}
                </span>
                <span style={{ fontSize: 11, color: '#a8a69f', marginLeft: 'auto' }}>
                  {time ? time.full : ''}
                </span>
              </div>
              <div style={{ fontSize: 13.5, color: '#33332f', lineHeight: 1.5 }}>
                {event.message}
              </div>
              {event.mediaUrl && (
                <img
                  src={event.mediaUrl}
                  alt={event.message || 'attached photo'}
                  style={{
                    marginTop: 8,
                    maxWidth: 320,
                    borderRadius: 8,
                    border: '1px solid #ebe9e6',
                    display: 'block',
                  }}
                />
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
