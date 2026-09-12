import { baseHref } from '@/portainer/helpers/pathHelper';

import { formatLogs } from './formatLogs';
import { FormattedLine } from './types';

const MAX_DOCKER_FRAME_SIZE = 16 * 1024 * 1024;
const MAX_LOG_LINES = 10_000;
const UPDATE_INTERVAL_MS = 100;

type LogResource = 'containers' | 'services' | 'tasks';

type LiveLogsOptions = {
  environmentId: number;
  resource: LogResource;
  resourceId: string;
  nodeName?: string;
  timestamps: boolean;
  since: number;
  tail: number;
  multiplexed: boolean;
  onLogs: (logs: FormattedLine[]) => void;
  onError: (error: Error) => void;
};

export class DockerLogStreamDecoder {
  private readonly decoder = new TextDecoder();

  private pending = new Uint8Array();

  constructor(private readonly multiplexed: boolean) {}

  push(chunk: ArrayBuffer | Uint8Array): string {
    const bytes = chunk instanceof Uint8Array ? chunk : new Uint8Array(chunk);
    if (!this.multiplexed) {
      return this.decoder.decode(bytes, { stream: true });
    }

    this.pending = appendBytes(this.pending, bytes);
    let output = '';

    while (this.pending.byteLength >= 8) {
      const frameSize = new DataView(
        this.pending.buffer,
        this.pending.byteOffset,
        this.pending.byteLength
      ).getUint32(4);
      if (frameSize > MAX_DOCKER_FRAME_SIZE) {
        throw new Error('Docker log frame exceeds the safety limit');
      }
      if (this.pending.byteLength < 8 + frameSize) {
        break;
      }

      output += this.decoder.decode(this.pending.subarray(8, 8 + frameSize), {
        stream: true,
      });
      this.pending = this.pending.slice(8 + frameSize);
    }

    return output;
  }

  flush(): string {
    if (this.multiplexed && this.pending.byteLength) {
      throw new Error('Docker log stream ended with an incomplete frame');
    }
    return this.decoder.decode();
  }
}

export function buildDockerLogsWebSocketUrl(
  options: Omit<LiveLogsOptions, 'onLogs' | 'onError' | 'multiplexed'>
) {
  const base = new URL(baseHref(), window.location.origin);
  const url = new URL('api/websocket/logs', base);
  url.protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  url.searchParams.set('endpointId', String(options.environmentId));
  url.searchParams.set('resource', options.resource);
  url.searchParams.set('id', options.resourceId);
  url.searchParams.set('timestamps', String(options.timestamps));
  url.searchParams.set('since', String(normalizeNonNegative(options.since, 0)));
  url.searchParams.set(
    'tail',
    String(Math.min(normalizeNonNegative(options.tail, 100), MAX_LOG_LINES))
  );
  if (options.nodeName) {
    url.searchParams.set('nodeName', options.nodeName);
  }
  return url.toString();
}

export function openDockerLogsStream(options: LiveLogsOptions) {
  let socket: WebSocket | undefined;
  let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
  let updateTimer: ReturnType<typeof setTimeout> | undefined;
  let stopped = false;
  let rawLogs = '';
  const lineLimit = Math.min(
    normalizeNonNegative(options.tail, 100),
    MAX_LOG_LINES
  );

  function publish() {
    updateTimer = undefined;
    options.onLogs(formatLogs(rawLogs, { withTimestamps: options.timestamps }));
  }

  function schedulePublish() {
    if (!updateTimer) {
      updateTimer = setTimeout(publish, UPDATE_INTERVAL_MS);
    }
  }

  function appendLogs(value: string) {
    if (!value) {
      return;
    }
    rawLogs = trimToLineLimit(rawLogs + value, lineLimit);
    schedulePublish();
  }

  function connect() {
    if (stopped) {
      return;
    }

    const decoder = new DockerLogStreamDecoder(options.multiplexed);
    socket = new WebSocket(buildDockerLogsWebSocketUrl(options));
    socket.binaryType = 'arraybuffer';

    socket.addEventListener('message', (event) => {
      try {
        if (typeof event.data === 'string') {
          throw new Error(event.data || 'Unable to stream Docker logs');
        }
        appendLogs(decoder.push(event.data as ArrayBuffer));
      } catch (error) {
        stopped = true;
        socket?.close();
        options.onError(toError(error));
      }
    });

    socket.addEventListener('close', () => {
      socket = undefined;
      try {
        appendLogs(decoder.flush());
      } catch (error) {
        stopped = true;
        options.onError(toError(error));
        return;
      }

      if (!stopped && !reconnectTimer) {
        reconnectTimer = setTimeout(() => {
          reconnectTimer = undefined;
          connect();
        }, 1000);
      }
    });

    socket.addEventListener('error', () => {
      if (!stopped) {
        stopped = true;
        socket?.close();
        options.onError(
          new Error('Unable to connect to the Docker log stream')
        );
      }
    });
  }

  connect();

  return {
    close() {
      stopped = true;
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
      }
      if (updateTimer) {
        clearTimeout(updateTimer);
        publish();
      }
      socket?.close();
    },
  };
}

function appendBytes(left: Uint8Array, right: Uint8Array) {
  if (!left.byteLength) {
    return right.slice();
  }
  const joined = new Uint8Array(left.byteLength + right.byteLength);
  joined.set(left);
  joined.set(right, left.byteLength);
  return joined;
}

function normalizeNonNegative(value: number, fallback: number) {
  return Number.isFinite(value) && value >= 0 ? Math.floor(value) : fallback;
}

function trimToLineLimit(value: string, lineLimit: number) {
  if (lineLimit === 0) {
    return '';
  }
  const lines = value.split('\n');
  return lines.length > lineLimit + 1
    ? lines.slice(-(lineLimit + 1)).join('\n')
    : value;
}

function toError(error: unknown) {
  return error instanceof Error ? error : new Error(String(error));
}
