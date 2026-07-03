// Thin client for the two backend services described in
// incident-service-openapi.yaml and report-service-openapi.yaml.
// The frontend is read-only: it only ever issues GET requests against
// the incident-service. The report-service is not called directly yet
// (see the open question about report generation in the feature list doc).

const INCIDENT_BASE = (
  import.meta.env.VITE_INCIDENT_API_BASE || 'http://localhost:8080'
).replace(/\/$/, '');

const REPORT_BASE = (
  import.meta.env.VITE_REPORT_API_BASE || 'http://localhost:8001'
).replace(/\/$/, '');

export class ApiError extends Error {
  constructor(message, status) {
    super(message);
    this.name = 'ApiError';
    this.status = status; // null when the request never reached the server
  }
}

async function getJson(url) {
  let res;
  try {
    res = await fetch(url);
  } catch (err) {
    throw new ApiError(
      `Could not reach the server at ${url}`,
      null
    );
  }

  if (!res.ok) {
    let message = `Request to ${url} failed with status ${res.status}`;
    try {
      const body = await res.json();
      if (body && body.error) message = body.error;
    } catch {
      // response body wasn't JSON, fall back to the generic message
    }
    throw new ApiError(message, res.status);
  }

  return res.json();
}

// GET /api/v1/incidents
// When chatId is provided, only incidents belonging to that Telegram chat are returned.
export function getIncidents(chatId) {
  const query =
    chatId === undefined || chatId === null || chatId === ''
      ? ''
      : `?chatId=${encodeURIComponent(chatId)}`;
  return getJson(`${INCIDENT_BASE}/api/v1/incidents${query}`);
}

// GET /api/v1/incidents/{id}
export function getIncident(id) {
  return getJson(`${INCIDENT_BASE}/api/v1/incidents/${id}`);
}

// GET /api/v1/incidents/{id}/timeline
export function getTimeline(id) {
  return getJson(`${INCIDENT_BASE}/api/v1/incidents/${id}/timeline`);
}

// GET /api/v1/incidents/{id}/images
export function getImages(id) {
  return getJson(`${INCIDENT_BASE}/api/v1/incidents/${id}/images`);
}

export const apiConfig = {
  incidentBase: INCIDENT_BASE,
  reportBase: REPORT_BASE,
};
