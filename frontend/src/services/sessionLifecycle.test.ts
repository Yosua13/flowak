import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({ setToken: vi.fn() }));
vi.mock('./apiClient', () => ({ apiClient: { setToken: mocks.setToken } }));
import { sessionLifecycle } from './sessionLifecycle';

describe('sessionLifecycle', () => {
  beforeEach(() => {
    const values = new Map<string, string>();
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => values.get(key) ?? null,
      setItem: (key: string, value: string) => values.set(key, value),
      removeItem: (key: string) => values.delete(key),
      clear: () => values.clear(),
    });
    mocks.setToken.mockClear();
  });

  it('keeps only a refresh hint in storage and clears the in-memory transport token', () => {
    sessionLifecycle.establish('access-token', { id: 'user-1', name: 'Ada' });
    expect(sessionLifecycle.hasRefreshHint()).toBe(true);
    expect(globalThis.localStorage.getItem('flowak_user')).not.toContain('access-token');
    expect(mocks.setToken).toHaveBeenCalledWith('access-token');
    sessionLifecycle.clear();
    expect(sessionLifecycle.hasRefreshHint()).toBe(false);
    expect(mocks.setToken).toHaveBeenLastCalledWith(null);
  });
});
