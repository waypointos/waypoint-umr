import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ConnectionPanel } from './ConnectionPanel';
import { UmrStats } from './proto/umr_pb';

vi.mock('./useUmrStats', () => ({
  useUmrStats: () => ({ stats: (globalThis as any).__stats ?? null }),
}));

function setStats(s: UmrStats | null) {
  (globalThis as any).__stats = s;
}

describe('ConnectionPanel', () => {
  it('renders all N/A when stats is null', () => {
    setStats(null);
    render(<ConnectionPanel />);
    expect(screen.getAllByText('N/A').length).toBeGreaterThan(5);
  });

  it('shows wifi mode fields when active=wifiwan', () => {
    setStats(new UmrStats({
      device: { online: true, firmware: '1.17.9' },
      wan: { active: 'wifiwan', externalIp: '203.0.113.10', externalIsp: 'Orange', wifi: { ssid: 'test-ap', signalBars: 3, statusCode: 1 } },
      system: { cpuPct: 46, memoryPct: 36 },
      throughput: { downBps: 2_100_000n, upBps: 250_000n, latencyAvgMs: 4 },
    }));
    render(<ConnectionPanel />);
    expect(screen.getByText('203.0.113.10')).toBeInTheDocument();
    expect(screen.getByText('Orange')).toBeInTheDocument();
  });

  it('shows LTE fields when active=lte', () => {
    setStats(new UmrStats({
      device: { online: true },
      wan: { active: 'lte', lte: { operator: 'Orange', band: 'B7', mode: '5g-nsa', rssiDbm: -73, rsrpDbm: -98, rsrqDb: -11 } },
      system: { cpuPct: 50, memoryPct: 30 },
    }));
    render(<ConnectionPanel />);
    expect(screen.getByText(/-73/)).toBeInTheDocument();
    expect(screen.getByText(/-98/)).toBeInTheDocument();
  });

  it('shows empty-state for clients table when device online but no clients', () => {
    setStats(new UmrStats({ device: { online: true }, wan: { active: 'wifiwan' } }));
    render(<ConnectionPanel />);
    expect(screen.getByText(/no connected clients/i)).toBeInTheDocument();
  });

  it('shows "router login failing" when device.online=false', () => {
    setStats(new UmrStats({ device: { online: false }, wan: {} }));
    render(<ConnectionPanel />);
    expect(screen.getAllByText(/router login failing/i).length).toBeGreaterThan(0);
  });
});
