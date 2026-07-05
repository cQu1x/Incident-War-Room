import Avatar from '../components/Avatar.jsx';
import EmptyState from '../components/EmptyState.jsx';
import { formatDate, mediaKind, fileName } from '../utils.js';

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

// MediaAttachment previews a single attachment inline: images and videos are
// shown, anything else (documents, audio, …) becomes a download link.
function MediaAttachment({ url, caption }) {
  const kind = mediaKind(url);
  const box = {
    maxWidth: 320,
    borderRadius: 8,
    border: '1px solid #ebe9e6',
    display: 'block',
  };

  if (kind === 'image') {
    return <img src={url} alt={caption || 'attachment'} style={box} />;
  }
  if (kind === 'video') {
    return <video src={url} controls style={{ ...box, maxHeight: 240 }} />;
  }
  return (
    <a
      href={url}
      target="_blank"
      rel="noreferrer"
      style={{
        ...box,
        padding: '8px 12px',
        fontSize: 12.5,
        color: '#0e7490',
        textDecoration: 'none',
        background: '#f7f6f4',
      }}
    >
      📎 {fileName(url)}
    </a>
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
              {event.mediaUrls?.length > 0 && (
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginTop: 8 }}>
                  {event.mediaUrls.map((url) => (
                    <MediaAttachment key={url} url={url} caption={event.message} />
                  ))}
                </div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
