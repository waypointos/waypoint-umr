// Renders UMR router telemetry in four panels: MODE, NETWORK, ROUTER, CLIENTS.
// Every nullable field renders muted "N/A" + a reason hint via Metric.naReason.
import { Panel } from './ui/Panel';
import { Metric } from './ui/Metric';
import { MetricGrid } from './ui/MetricGrid';
import { SignalBars } from './ui/SignalBars';
import { useUmrStats } from './useUmrStats';
import styles from './ConnectionPanel.module.css';

export function ConnectionPanel() {
  const { stats } = useUmrStats();
  const online = stats?.device?.online ?? false;
  const offlineReason = !online ? 'router login failing' : undefined;
  const wan = stats?.wan;
  const active = wan?.active ?? null;

  const lteRSSI = wan?.lte?.rssiDbm ?? null;
  const lteRSRP = wan?.lte?.rsrpDbm ?? null;
  const lteRSRQ = wan?.lte?.rsrqDb ?? null;
  const wifiBars = wan?.wifi?.signalBars ?? null;
  const lteBars = wan?.lte?.signalBars ?? null;
  const signalValue = active === 'lte' ? lteBars : active === 'wifiwan' ? wifiBars : null;

  return (
    <div className={styles.grid} data-testid="panel-m-umr">
      <Panel title="MODE">
        <MetricGrid>
          <Metric label="Active WAN" value={active ? active.toUpperCase() : null} naReason={offlineReason ?? 'no active WAN'} tone="key" />
          <Metric label="Carrier"    value={active === 'lte' ? (wan?.lte?.operator ?? null) : null} naReason={active === 'lte' ? 'not reported' : 'not on LTE'} />
          <Metric label="Band"       value={active === 'lte' ? (wan?.lte?.band ?? null) : null}     naReason={active === 'lte' ? 'not reported' : 'not on LTE'} />
          <Metric label="Technology" value={active === 'lte' ? (wan?.lte?.mode ?? null) : null}      naReason={active === 'lte' ? 'not reported' : 'not on LTE'} />
          <div className={styles.signalRow}>
            <span className={styles.signalLabel}>Signal</span>
            {signalValue !== null
              ? <SignalBars value={signalValue} showLabel />
              : <span className={styles.na}>N/A — {offlineReason ?? (active === 'wan' ? 'ethernet WAN — no signal value' : 'no signal reported')}</span>
            }
          </div>
          <Metric label="RSSI" value={lteRSSI !== null ? lteRSSI : null} unit="dBm" naReason={active === 'lte' ? 'not reported' : 'not on LTE'} />
          <Metric label="RSRP" value={lteRSRP !== null ? lteRSRP : null} unit="dBm" naReason={active === 'lte' ? 'not reported' : 'not on LTE'} />
          <Metric label="RSRQ" value={lteRSRQ !== null ? lteRSRQ : null} unit="dB"  naReason={active === 'lte' ? 'not reported' : 'not on LTE'} />
        </MetricGrid>
      </Panel>

      <Panel title="NETWORK">
        <MetricGrid>
          <Metric label="External IP" value={wan?.externalIp ?? null}  naReason={offlineReason ?? 'no public IP yet'} tone="key" />
          <Metric label="ISP"         value={wan?.externalIsp ?? null} naReason={offlineReason ?? 'no ISP detected'} />
          <Metric label="Up"          value={mbps(stats?.throughput?.upBps)}   unit="Mbps" naReason={offlineReason ?? 'no throughput sample'} />
          <Metric label="Down"        value={mbps(stats?.throughput?.downBps)} unit="Mbps" naReason={offlineReason ?? 'no throughput sample'} />
          <Metric label="Latency"     value={stats?.throughput?.latencyAvgMs ?? null} unit="ms" naReason={offlineReason ?? 'no latency sample'} />
        </MetricGrid>
      </Panel>

      <Panel title="ROUTER">
        <MetricGrid>
          <Metric label="Firmware"  value={stats?.device?.firmware ?? null} naReason={offlineReason ?? 'unreported'} />
          <Metric label="Uptime"    value={uptimeText(stats?.device?.uptimeS)} naReason={offlineReason ?? 'unreported'} />
          <Metric label="CPU"       value={stats?.system?.cpuPct ?? null}    unit="%" naReason={offlineReason ?? 'unreported'} />
          <Metric label="Memory"    value={stats?.system?.memoryPct ?? null} unit="%" naReason={offlineReason ?? 'unreported'} />
        </MetricGrid>
      </Panel>

      <Panel title="CLIENTS">
        {!online ? (
          <p className={styles.empty}>N/A — router login failing</p>
        ) : !stats?.clients?.length ? (
          <p className={styles.empty}>No connected clients</p>
        ) : (
          <table className={styles.clients}>
            <thead>
              <tr><th>Name</th><th>IP</th><th>Link</th><th>Speed</th></tr>
            </thead>
            <tbody>
              {[...stats.clients]
                .sort((a, b) => (b.uptimeS ?? 0) - (a.uptimeS ?? 0))
                .map((c) => (
                  <tr key={c.mac ?? c.ip ?? c.name}>
                    <td>{c.name ?? '—'}</td>
                    <td>{c.ip ?? '—'}</td>
                    <td>{c.connection ?? '—'}</td>
                    <td>{c.linkSpeedMbps ? `${c.linkSpeedMbps} Mbps` : '—'}</td>
                  </tr>
                ))}
            </tbody>
          </table>
        )}
      </Panel>
    </div>
  );
}

function mbps(bps: bigint | undefined | null): number | null {
  if (bps === undefined || bps === null) return null;
  const n = Number(bps);
  if (!Number.isFinite(n)) return null;
  return Math.round((n / 1_000_000) * 10) / 10;
}

function uptimeText(seconds: number | undefined | null): string | null {
  if (seconds === undefined || seconds === null) return null;
  const s = seconds % 60;
  const m = Math.floor(seconds / 60) % 60;
  const h = Math.floor(seconds / 3600);
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m ${s}s`;
  return `${s}s`;
}
