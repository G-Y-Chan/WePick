import { Message } from "./types"
import { WS_BASE_URL } from "./api/client";

type ConnectOptions = {
  onMessage?: (msg: Message) => void;
  onOpen?: () => void;
  onClose?: () => void;
  onError?: (err: unknown) => void;
};

export function connectRoomWS(
  roomCode: string,
  opts: ConnectOptions = {},
  wsBaseUrl: string = WS_BASE_URL
) {
  const url = `${wsBaseUrl}/ws?roomCode=${encodeURIComponent(roomCode)}`;
  const ws = new WebSocket(url);

  ws.onopen = () => {
    opts.onOpen?.();
  };

  ws.onmessage = (event) => {
    try {
      const msg: Message = JSON.parse(event.data);
      opts.onMessage?.(msg);
    } catch {
      // ignore malformed messages
    }
  };

  ws.onerror = (e) => {
    opts.onError?.(e);
  };

  ws.onclose = () => {
    opts.onClose?.();
  };

  const send = (msg: Message) => {
    if (ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(msg));
      return true;
    }
    return false;
  };

  const close = () => ws.close();

  return { ws, url, send, close };
}