export type ProjectEventState = 'connected' | 'disconnected';

export interface ProjectEventSubscription {
  close: () => void;
}

// Authentication for this stream is carried by the short-lived, HttpOnly SSE
// cookie. Do not put access tokens in an EventSource URL: URLs are routinely
// retained in browser and proxy logs.
export function subscribeProjectEvents(projectId: string, onInvalidate: () => void, onState?: (state: ProjectEventState) => void): ProjectEventSubscription {
  const stream = new EventSource(`/api/projects/${encodeURIComponent(projectId)}/events`, { withCredentials: true });
  const invalidate = () => onInvalidate();
  stream.addEventListener('ready', () => { onState?.('connected'); invalidate(); });
  stream.addEventListener('change', invalidate);
  stream.addEventListener('invalidate', invalidate);
  stream.onerror = () => onState?.('disconnected');
  return { close: () => stream.close() };
}
