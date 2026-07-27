import { useEffect, useState } from 'react';
import { useBridge } from './bridge';
import { UmrStats } from './proto/umr_pb';
import { isUmrStatsFrame } from './statsFrame';

export function useUmrStats(): { stats: UmrStats | null } {
  const { roverId, subscribe } = useBridge();
  const [stats, setStats] = useState<UmrStats | null>(null);
  useEffect(() => {
    return subscribe(`waypoint.${roverId}.module.umr.stats`, (bytes) => {
      if (!isUmrStatsFrame(bytes)) return;
      try {
        setStats(UmrStats.fromBinary(bytes));
      } catch {
        /* ignore decode errors */
      }
    });
  }, [roverId, subscribe]);
  return { stats };
}
