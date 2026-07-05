// Dashboard access-token handling.
//
// The Telegram bot's /dashboard command hands operators a link shaped like
// `<dashboard>/dashboard?token=<jwt>` (see incident-service's dashboard.Linker).
// The token authorises every incident-service API call as an
// `Authorization: Bearer <token>` header, and is scoped to a single Telegram
// chat. We stash it in localStorage so it survives navigation and reloads.

const STORAGE_KEY = 'iwr.dashboardToken';

export function getToken() {
  try {
    return localStorage.getItem(STORAGE_KEY) || null;
  } catch {
    // localStorage can throw in private-mode / sandboxed contexts.
    return null;
  }
}

export function setToken(token) {
  try {
    localStorage.setItem(STORAGE_KEY, token);
  } catch {
    // ignore — the token still lives in memory for this page load below
  }
}

export function clearToken() {
  try {
    localStorage.removeItem(STORAGE_KEY);
  } catch {
    // ignore
  }
}

// captureTokenFromURL pulls a `?token=` off the current URL (if present),
// stores it, and strips it from the address bar so the JWT isn't left sitting
// in history or copy-pasted by accident. Returns true when a token was found.
export function captureTokenFromURL() {
  const params = new URLSearchParams(window.location.search);
  const token = params.get('token');
  if (!token) return false;

  setToken(token);

  params.delete('token');
  const query = params.toString();
  const cleanUrl =
    window.location.pathname + (query ? `?${query}` : '') + window.location.hash;
  window.history.replaceState(null, '', cleanUrl);
  return true;
}
