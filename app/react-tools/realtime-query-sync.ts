import type { QueryClient } from '@tanstack/react-query';

import { baseHref } from '@/portainer/helpers/pathHelper';

let socket: WebSocket | undefined;
let reconnectTimer: ReturnType<typeof setTimeout> | undefined;

export function startRealtimeQuerySync(queryClient: QueryClient) {
  if (
    process.env.NODE_ENV === 'test' ||
    typeof window === 'undefined' ||
    typeof WebSocket === 'undefined' ||
    socket
  ) {
    return;
  }

  const base = new URL(baseHref(), window.location.origin);
  const url = new URL('api/websocket/events', base);
  url.protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  socket = new WebSocket(url);

  socket.addEventListener('message', (event) => {
    try {
      const message = JSON.parse(event.data) as { type?: string };
      if (message.type === 'invalidate') {
        queryClient.invalidateQueries({ refetchType: 'active' });
      }
    } catch {
      // Ignore malformed events and keep the connection alive.
    }
  });

  socket.addEventListener('close', () => {
    socket = undefined;
    if (!reconnectTimer) {
      reconnectTimer = setTimeout(() => {
        reconnectTimer = undefined;
        startRealtimeQuerySync(queryClient);
      }, 5000);
    }
  });
}
