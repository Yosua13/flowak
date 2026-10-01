import { apiClient } from './apiClient';

// The browser keeps only a non-sensitive hint for refresh-on-start. The access
// token remains memory-only in ApiClient and the refresh credential is the
// server-managed HttpOnly cookie.
const SESSION_HINT_KEY = 'flowak_user';

export const sessionLifecycle = {
  hasRefreshHint(): boolean {
    return localStorage.getItem(SESSION_HINT_KEY) !== null;
  },
  establish(token: string, user: unknown): void {
    apiClient.setToken(token);
    localStorage.setItem(SESSION_HINT_KEY, JSON.stringify(user));
  },
  clear(): void {
    apiClient.setToken(null);
    localStorage.removeItem(SESSION_HINT_KEY);
  },
};
