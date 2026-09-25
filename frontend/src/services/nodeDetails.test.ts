import { afterEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from './apiClient';
import { nodeDetailsApi } from './nodeDetails';

describe('node details data contract', () => {
  afterEach(() => { vi.unstubAllGlobals(); apiClient.setToken(null); });

  it('reuses detail data during an immediate drawer-to-page transition', async () => {
    const fetchMock = vi.fn().mockImplementation(() => Promise.resolve(new Response('[]', { status: 200, headers: { 'Content-Type': 'application/json' } })));
    vi.stubGlobal('fetch', fetchMock);
    apiClient.setToken('test-token');

    const first = await nodeDetailsApi.load('project-cache-test', 'node-cache-test');
    const second = await nodeDetailsApi.load('project-cache-test', 'node-cache-test');

    expect(second).toBe(first);
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });
});
