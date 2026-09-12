import { useEffect, useRef, useState } from 'react';

import { ChartPoint, toChartPoint } from './chartPoint';

const CHART_LIMIT = 600;
const CACHE_VERSION = 1;

type MetricsState = 'checking' | 'available' | 'unavailable';
type MetricsData =
  | { cpu: string; memory: string; timestamp: string }
  | undefined;

type CachedMetrics = {
  version: number;
  nodeCPU: number;
  chartData: ChartPoint[];
};

export function useAggregatedMetrics(
  {
    data,
    error,
  }: {
    data: MetricsData;
    error: unknown;
  },
  nodeCPU: number,
  sessionCacheKey?: string
) {
  const cacheIdentity = `${sessionCacheKey ?? ''}:${nodeCPU}`;
  const [chartData, setChartData] = useState<ChartPoint[]>(() =>
    readSessionCache(sessionCacheKey, nodeCPU)
  );
  const [metricsState, setMetricsState] = useState<MetricsState>(() =>
    chartData.length ? 'available' : 'checking'
  );
  const [activeCacheIdentity, setActiveCacheIdentity] = useState(cacheIdentity);
  const lastData = useRef<MetricsData>(undefined);
  const lastCacheIdentity = useRef(cacheIdentity);

  useEffect(() => {
    if (lastCacheIdentity.current !== cacheIdentity) {
      lastCacheIdentity.current = cacheIdentity;
      lastData.current = undefined;
      const cachedData = readSessionCache(sessionCacheKey, nodeCPU);
      setChartData(cachedData);
      setMetricsState(cachedData.length ? 'available' : 'checking');
      setActiveCacheIdentity(cacheIdentity);
    }
  }, [cacheIdentity, nodeCPU, sessionCacheKey]);

  // Latch the first result during render so a later polling error does not hide
  // charts that were already proven available.
  if (metricsState === 'checking' && error !== undefined) {
    setMetricsState(error ? 'unavailable' : 'available');
  }

  useEffect(() => {
    if (!data || data === lastData.current) {
      return;
    }
    lastData.current = data;

    const point = toChartPoint(data.cpu, data.memory, data.timestamp, nodeCPU);
    setChartData((previous) => {
      if (previous.at(-1)?.time === point.time) {
        return previous;
      }
      const next = [...previous, point];
      return next.length > CHART_LIMIT
        ? next.slice(next.length - CHART_LIMIT)
        : next;
    });
  }, [data, nodeCPU]);

  useEffect(() => {
    if (activeCacheIdentity === cacheIdentity) {
      writeSessionCache(sessionCacheKey, nodeCPU, chartData);
    }
  }, [activeCacheIdentity, cacheIdentity, chartData, nodeCPU, sessionCacheKey]);

  return { chartData, metricsState, error: error ?? null };
}

function readSessionCache(key: string | undefined, nodeCPU: number) {
  if (!key || typeof sessionStorage === 'undefined') {
    return [];
  }

  try {
    const value = JSON.parse(sessionStorage.getItem(key) ?? 'null') as unknown;
    if (!isCachedMetrics(value) || value.nodeCPU !== nodeCPU) {
      return [];
    }
    return value.chartData.slice(-CHART_LIMIT);
  } catch {
    return [];
  }
}

function writeSessionCache(
  key: string | undefined,
  nodeCPU: number,
  chartData: ChartPoint[]
) {
  if (!key || typeof sessionStorage === 'undefined') {
    return;
  }

  try {
    const value: CachedMetrics = {
      version: CACHE_VERSION,
      nodeCPU,
      chartData: chartData.slice(-CHART_LIMIT),
    };
    sessionStorage.setItem(key, JSON.stringify(value));
  } catch {
    // Storage can be disabled or full; live metrics should continue regardless.
  }
}

function isCachedMetrics(value: unknown): value is CachedMetrics {
  if (!value || typeof value !== 'object') {
    return false;
  }

  const candidate = value as Partial<CachedMetrics>;
  return (
    candidate.version === CACHE_VERSION &&
    typeof candidate.nodeCPU === 'number' &&
    Array.isArray(candidate.chartData) &&
    candidate.chartData.every(
      (point) =>
        point &&
        typeof point.time === 'string' &&
        typeof point.cpu === 'number' &&
        Number.isFinite(point.cpu) &&
        typeof point.memory === 'number' &&
        Number.isFinite(point.memory)
    )
  );
}
