import EmptyState from '../components/EmptyState.jsx';
import { mediaKind, fileName } from '../utils.js';

// Preview renders the media thumbnail appropriate to its type.
function Preview({ url, message }) {
  const kind = mediaKind(url);
  const style = { width: '100%', height: 130, objectFit: 'cover', display: 'block' };

  if (kind === 'image') {
    return <img src={url} alt={message || 'incident photo'} style={style} />;
  }
  if (kind === 'video') {
    return <video src={url} controls style={{ ...style, objectFit: 'contain', background: '#000' }} />;
  }
  return (
    <a
      href={url}
      target="_blank"
      rel="noreferrer"
      style={{
        ...style,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 6,
        color: '#0e7490',
        textDecoration: 'none',
        background: '#f7f6f4',
        padding: '0 10px',
        boxSizing: 'border-box',
        fontSize: 12.5,
        textAlign: 'center',
        wordBreak: 'break-word',
      }}
    >
      📎 {fileName(url)}
    </a>
  );
}

export default function PhotosTab({ images }) {
  if (!images || images.length === 0) {
    return <EmptyState message="No media has been attached to this incident." />;
  }

  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(180px, 1fr))',
        gap: 14,
      }}
    >
      {images.map((img) => (
        <figure
          key={img.url}
          style={{
            margin: 0,
            border: '1px solid #ebe9e6',
            borderRadius: 10,
            overflow: 'hidden',
            background: '#fff',
          }}
        >
          <Preview url={img.url} message={img.message} />
          <figcaption style={{ padding: '8px 10px' }}>
            <div style={{ fontSize: 12.5, color: '#33332f', marginBottom: 2 }}>
              {img.message || 'No caption'}
            </div>
            <div style={{ fontSize: 11, color: '#a8a69f' }}>{img.username}</div>
          </figcaption>
        </figure>
      ))}
    </div>
  );
}
