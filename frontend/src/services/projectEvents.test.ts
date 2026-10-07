import { afterEach, describe, expect, it, vi } from 'vitest';
import { subscribeProjectEvents } from './projectEvents';

class FakeEventSource {
  static instance: FakeEventSource | undefined;
  readonly listeners = new Map<string, Array<() => void>>();
  onerror: (() => void) | null = null;
  closed = false;
  constructor(readonly url: string, readonly options: EventSourceInit) { FakeEventSource.instance = this; }
  addEventListener(type: string, listener: () => void) { this.listeners.set(type, [...(this.listeners.get(type) || []), listener]); }
  close() { this.closed = true; }
  emit(type: string) { this.listeners.get(type)?.forEach((listener) => listener()); }
}

afterEach(() => { vi.unstubAllGlobals(); FakeEventSource.instance = undefined; });

describe('project EventSource sync', () => {
  it('invalidates canonical state for ready, replay, and safe fallback events', () => {
    vi.stubGlobal('EventSource', FakeEventSource);
    const invalidate = vi.fn(); const state = vi.fn();
    const subscription = subscribeProjectEvents('project/a', invalidate, state);
    const stream = FakeEventSource.instance!;
    expect(stream.url).toBe('/api/projects/project%2Fa/events');
    expect(stream.options.withCredentials).toBe(true);
    stream.emit('ready'); stream.emit('change'); stream.emit('invalidate'); stream.onerror?.();
    expect(invalidate).toHaveBeenCalledTimes(3);
    expect(state).toHaveBeenCalledWith('connected');
    expect(state).toHaveBeenCalledWith('disconnected');
    subscription.close(); expect(stream.closed).toBe(true);
  });
});
