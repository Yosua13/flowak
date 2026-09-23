import { ApiError, normalizeApiError } from './apiClient';

export interface AuthUser {
  id: string;
  name: string;
  email: string;
  role: 'pm' | 'uiux' | 'frontend' | 'backend';
}

export interface AuthSession {
  token: string;
  user: AuthUser;
  organization_role: 'owner' | 'member';
}

async function authRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`/api/auth/${path}`, init);
  } catch (error) {
    throw normalizeApiError(error);
  }
  const payload = await response.json().catch(() => null) as T | { error?: string } | null;
  if (!response.ok) throw new ApiError(response.status === 401 ? 'unauthorized' : response.status === 400 ? 'validation_error' : 'unknown_error', payload && typeof payload === 'object' && 'error' in payload ? String(payload.error) : 'Permintaan autentikasi gagal.', response.status, payload);
  return payload as T;
}

export const authService = {
  login(email: string, password: string) {
    return authRequest<AuthSession>('login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email, password }), credentials: 'same-origin' });
  },
  refresh() {
    return authRequest<AuthSession>('refresh', { method: 'POST', credentials: 'same-origin' });
  },
  register(name: string, email: string, password: string, role: string) {
    return authRequest<{ success: boolean }>('register', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name, email, password, role }), credentials: 'same-origin' });
  },
  async logout() {
    await fetch('/api/auth/logout', { method: 'POST', credentials: 'same-origin' });
  },
};
