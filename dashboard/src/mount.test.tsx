import { afterEach, describe, expect, it, vi } from 'vitest';
import { act } from 'react';
import panel from './mount';
import { UmrStats } from './proto/umr_pb';

describe('umr panel mount', () => {
  afterEach(() => {
    document.body.replaceChildren();
  });

  it('mounts ConnectionPanel and subscribes to the umr stats subject', () => {
    const container = document.createElement('div');
    document.body.appendChild(container);
    let push: ((b: Uint8Array) => void) | null = null;
    const subscribe = vi.fn((subject: string, onBytes: (b: Uint8Array) => void) => {
      expect(subject).toBe('waypoint.rover-1.module.umr.stats');
      push = onBytes;
      return () => {};
    });

    let cleanup!: () => void;
    act(() => {
      cleanup = panel.mount(container, { roverId: 'rover-1', subscribe, publish: vi.fn() });
    });

    expect(subscribe).toHaveBeenCalledOnce();
    expect(container.querySelector('[data-testid="panel-m-umr"]')).not.toBeNull();

    act(() => {
      push!(new UmrStats({ device: { online: true } }).toBinary());
    });

    act(() => {
      cleanup();
    });
    expect(container.querySelector('[data-testid="panel-m-umr"]')).toBeNull();
  });
});
