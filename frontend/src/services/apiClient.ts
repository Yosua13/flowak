import type { Module } from '../domain/types';

export type ApiErrorCode = 'network_error' | 'invalid_response' | 'unauthorized' | 'forbidden' | 'not_found' | 'validation_error' | 'conflict' | 'server_error' | 'unknown_error';

export class ApiError extends Error {
  constructor(public readonly code: ApiErrorCode, message: string, public readonly status?: number, public readonly details?: unknown) {
    super(message);
    this.name = 'ApiError';
  }
}

export function normalizeApiError(error: unknown): ApiError {
  if (error instanceof ApiError) return error;
  if (error instanceof DOMException && error.name === 'AbortError') return new ApiError('network_error', 'Request dibatalkan.');
  if (error instanceof TypeError) return new ApiError('network_error', 'Koneksi ke server terputus.');
  return new ApiError('unknown_error', error instanceof Error ? error.message : 'Terjadi kesalahan yang tidak diketahui.');
}

type ApiOptions = Omit<RequestInit, 'body' | 'headers' | 'signal'> & {
  body?: unknown;
  signal?: AbortSignal;
  headers?: HeadersInit;
};

export class ApiClient {
  private token: string | null = null;

  constructor(private readonly baseUrl = '/api') {}

  setToken(token: string | null) { this.token = token; }

  abortable() {
    return new AbortController();
  }

  async request<T>(path: string, options: ApiOptions = {}): Promise<T> {
    const headers = new Headers(options.headers);
    headers.set('Accept', 'application/json');
    if (options.body !== undefined) headers.set('Content-Type', 'application/json');
    if (this.token) headers.set('Authorization', `Bearer ${this.token}`);
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}${path}`, { ...options, headers, body: options.body === undefined ? undefined : JSON.stringify(options.body) });
    } catch (error) {
      throw normalizeApiError(error);
    }
    const raw = await response.text();
    let payload: unknown = null;
    if (raw) {
      try { payload = JSON.parse(raw); } catch { throw new ApiError('invalid_response', 'Server mengembalikan JSON tidak valid.', response.status); }
    }
    if (!response.ok) {
      const message = typeof payload === 'object' && payload && 'error' in payload ? String(payload.error) : `Permintaan gagal (${response.status}).`;
      const code: ApiErrorCode = response.status === 409 ? 'conflict' : response.status === 401 ? 'unauthorized' : response.status === 403 ? 'forbidden' : response.status === 404 ? 'not_found' : response.status === 422 || response.status === 400 ? 'validation_error' : response.status >= 500 ? 'server_error' : 'unknown_error';
      throw new ApiError(code, message, response.status, payload);
    }
    return payload as T;
  }

  get<T>(path: string, signal?: AbortSignal) { return this.request<T>(path, { signal }); }
  post<T>(path: string, body?: unknown, signal?: AbortSignal) { return this.request<T>(path, { method: 'POST', body, signal }); }
  put<T>(path: string, body?: unknown, signal?: AbortSignal) { return this.request<T>(path, { method: 'PUT', body, signal }); }
  delete<T>(path: string, signal?: AbortSignal) { return this.request<T>(path, { method: 'DELETE', signal }); }
}

export const apiClient = new ApiClient();

export interface ProjectResponse { modules: Module[]; }
