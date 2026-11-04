// WebSocket event stream with automatic reconnect.
import { applyEvent, connected } from './stores';

let socket: WebSocket | null = null;
let retryTimer: ReturnType<typeof setTimeout> | null = null;
let closed = false;

export function startWebSocket() {
  closed = false;
  connect();
}

function connect() {
  if (closed) return;
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
  const ws = new WebSocket(`${proto}//${location.host}/api/ws`);
  socket = ws;
  ws.onopen = () => connected.set(true);
  ws.onmessage = (event) => {
    try {
      const message = JSON.parse(event.data);
      if (message.type) applyEvent(message.type, message.data);
    } catch {
      // ignore malformed frames
    }
  };
  ws.onclose = () => {
    if (socket === ws) socket = null;
    connected.set(false);
    if (!closed && retryTimer == null) {
      retryTimer = setTimeout(() => {
        retryTimer = null;
        connect();
      }, 1500);
    }
  };
  ws.onerror = () => ws.close();
}

export function stopWebSocket() {
  closed = true;
  socket?.close();
  socket = null;
}
