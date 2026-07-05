// Formatting and color-mapping helpers. These mirror the logic already
// approved in the Claude Design prototype, so the real app matches the
// design the team signed off on.

const SEVERITY_STYLES = {
  HIGH: { color: '#b42318', bg: '#fdeceb' },
  MEDIUM: { color: '#b45309', bg: '#f8f1e3' },
  LOW: { color: '#475569', bg: '#eef1f5' },
};

export function severityStyle(severity) {
  return SEVERITY_STYLES[severity] || SEVERITY_STYLES.LOW;
}

export function statusStyle(status) {
  return status === 'ACTIVE'
    ? { color: '#b42318', bg: '#fdeceb', dot: '#dc2626' }
    : { color: '#3f7d4f', bg: '#ecf5ee', dot: '#a9a8a0' };
}

// Short human-readable incident reference, e.g. "INC-5B6C"
export function shortRef(id) {
  return 'INC-' + String(id).slice(-4).toUpperCase();
}

const AVATAR_PALETTE = [
  '#5b5bd6', '#0e7490', '#b4257f', '#0f766e',
  '#a16207', '#475569', '#9333ea', '#c2410c',
];

// Deterministic color per username, so the same person always gets the
// same avatar color across the app.
export function colorForUsername(username) {
  let hash = 0;
  for (let i = 0; i < username.length; i++) {
    hash = (hash * 31 + username.charCodeAt(i)) >>> 0;
  }
  return AVATAR_PALETTE[hash % AVATAR_PALETTE.length];
}

export function initials(username) {
  const cleaned = String(username).replace(/^@/, '');
  return cleaned.slice(0, 2).toUpperCase();
}

export function formatDate(iso) {
  if (!iso) return null;
  const d = new Date(iso);
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
    'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  const mon = months[d.getMonth()];
  const hh = String(d.getHours()).padStart(2, '0');
  const mm = String(d.getMinutes()).padStart(2, '0');
  return {
    short: `${mon} ${d.getDate()}`,
    date: `${mon} ${d.getDate()}, ${d.getFullYear()}`,
    time: `${hh}:${mm}`,
    full: `${mon} ${d.getDate()} - ${hh}:${mm}`,
  };
}

// mediaKind classifies an attachment URL by its file extension so the UI can
// pick the right element: an <img> for images, a <video> for videos, and a
// plain download link for everything else (documents, audio, …).
const IMAGE_EXTS = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg'];
const VIDEO_EXTS = ['mp4', 'webm', 'mov', 'm4v', 'ogv'];

export function mediaKind(url) {
  const clean = String(url).split(/[?#]/)[0];
  const ext = clean.slice(clean.lastIndexOf('.') + 1).toLowerCase();
  if (IMAGE_EXTS.includes(ext)) return 'image';
  if (VIDEO_EXTS.includes(ext)) return 'video';
  return 'file';
}

// fileName returns the last path segment of a URL, used as a label for
// non-previewable attachments.
export function fileName(url) {
  const clean = String(url).split(/[?#]/)[0];
  return decodeURIComponent(clean.slice(clean.lastIndexOf('/') + 1)) || 'attachment';
}

export function formatDuration(seconds) {
  if (seconds == null) return null;
  const h = Math.floor(seconds / 3600);
  const m = Math.round((seconds % 3600) / 60);
  if (h && m) return `${h}h ${m}m`;
  if (h) return `${h}h`;
  return `${m}m`;
}
