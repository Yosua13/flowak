import { describe, expect, it, vi, afterEach } from 'vitest';
import { ApiClient, ApiError, normalizeApiError } from './apiClient';

describe('API error normalization', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('normalizes a 409 response as a stable conflict', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: 'Versi sudah berubah' }), { status: 409 })));
    await expect(new ApiClient().put('/modules/module-1', {})).rejects.toMatchObject({ code: 'conflict', status: 409 });
  });

  it('normalizes transport failures without leaking implementation errors', () => {
    expect(normalizeApiError(new TypeError('failed to fetch'))).toMatchObject({ code: 'network_error' });
  });
});
