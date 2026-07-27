import { describe, it, expect } from 'vitest';
import { isUmrStatsFrame } from './statsFrame';
import { UmrStats } from './proto/umr_pb';

// Hand-encoded SDK ModuleStats frames (field 1 Stamp, field 2 uint64 uptime_s).
const HEARTBEAT_STAMP_ONLY = new Uint8Array([0x0a, 0x02, 0x08, 0x01]);
const HEARTBEAT_WITH_UPTIME = new Uint8Array([0x0a, 0x02, 0x08, 0x01, 0x10, 0x2a]);

describe('isUmrStatsFrame', () => {
  it('rejects a ModuleStats heartbeat carrying only a stamp', () => {
    expect(isUmrStatsFrame(HEARTBEAT_STAMP_ONLY)).toBe(false);
  });

  it('rejects a ModuleStats heartbeat carrying an uptime', () => {
    expect(isUmrStatsFrame(HEARTBEAT_WITH_UPTIME)).toBe(false);
  });

  it('accepts a minimal UmrStats', () => {
    expect(isUmrStatsFrame(new UmrStats({ device: { online: true } }).toBinary())).toBe(true);
  });

  it('accepts an offline-shaped UmrStats', () => {
    const bytes = new UmrStats({ device: { online: false }, wan: {} }).toBinary();
    expect(isUmrStatsFrame(bytes)).toBe(true);
  });

  it('rejects an empty buffer', () => {
    expect(isUmrStatsFrame(new Uint8Array())).toBe(false);
  });

  it('rejects a truncated buffer', () => {
    expect(isUmrStatsFrame(new Uint8Array([0x0a, 0x05, 0x08, 0x01]))).toBe(false);
    expect(isUmrStatsFrame(new Uint8Array([0x80]))).toBe(false);
  });
});
