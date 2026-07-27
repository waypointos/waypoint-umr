import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import type { ReactNode } from 'react';
import { UmrStats } from './proto/umr_pb';
import { useUmrStats } from './useUmrStats';
import { BridgeProvider, type Subscribe } from './bridge';

// Hand-encoded SDK ModuleStats frames (field 1 Stamp, field 2 uint64 uptime_s).
// The stamp-only shape decodes cleanly as an empty UmrStats, so only the frame
// guard keeps it from overwriting good stats; the uptime shape throws instead.
const HEARTBEAT_STAMP_ONLY = new Uint8Array([0x0a, 0x02, 0x08, 0x01]);
const HEARTBEAT_WITH_UPTIME = new Uint8Array([0x0a, 0x02, 0x08, 0x01, 0x10, 0x2a]);

function harness() {
  let push: ((b: Uint8Array) => void) | null = null;
  const subscribe: Subscribe = (_subject, onBytes) => {
    push = onBytes;
    return () => {};
  };
  const wrapper = ({ children }: { children: ReactNode }) => (
    <BridgeProvider value={{ roverId: 'rover-1', subscribe, publish: vi.fn() }}>
      {children}
    </BridgeProvider>
  );
  const feed = (bytes: Uint8Array) => {
    act(() => {
      push!(bytes);
    });
  };
  return { wrapper, feed };
}

describe('useUmrStats', () => {
  it('decodes UmrStats from subscribed bytes', () => {
    const { wrapper, feed } = harness();
    const { result } = renderHook(() => useUmrStats(), { wrapper });
    expect(result.current.stats).toBeNull();

    feed(new UmrStats({ device: { online: true } }).toBinary());
    expect(result.current.stats?.device?.online).toBe(true);
  });

  it('keeps the last good frame when the SDK heartbeat interleaves', () => {
    const { wrapper, feed } = harness();
    const { result } = renderHook(() => useUmrStats(), { wrapper });

    feed(new UmrStats({ device: { online: true } }).toBinary());
    feed(HEARTBEAT_STAMP_ONLY);
    feed(HEARTBEAT_WITH_UPTIME);

    expect(result.current.stats?.device?.online).toBe(true);
  });
});
