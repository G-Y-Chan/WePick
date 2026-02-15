import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Message } from "./types"
import { WS_BASE_URL } from "./api/client";

type ConnectOptions = {
  onMessage?: (msg: Message) => void;
  onOpen?: () => void;
  onClose?: () => void;
  onError?: (err: unknown) => void;
};

type Status = "idle" | "connecting" | "connected" | "error" | "closed";

export function useRoomWS(roomCode: string | undefined, opts: ConnectOptions = {}) {
  const [status, setStatus] = useState<Status>(roomCode ? "connecting" : "idle");
  const connRef = useRef<ReturnType<typeof connectRoomWS> | null>(null);

  // keep latest onMessage without reconnecting
  const onMessageRef = useRef(opts.onMessage);
  useEffect(() => {
    onMessageRef.current = opts.onMessage;
  }, [opts.onMessage]);

  useEffect(() => {
    if (!roomCode) return;

    setStatus("connecting");

    const conn = connectRoomWS(roomCode, {
      onOpen: () => setStatus("connected"),
      onClose: () => setStatus("closed"),
      onError: () => setStatus("error"),
      onMessage: (msg) => onMessageRef.current?.(msg),
    });

    connRef.current = conn;

    return () => {
      connRef.current?.close();
      connRef.current = null;
    };
  }, [roomCode]);

  const send = useCallback((msg: Message) => {
    return connRef.current?.send(msg) ?? false;
  }, []);

  const close = useCallback(() => {
    connRef.current?.close();
  }, []);

  return useMemo(() => ({ status, send, close }), [status, send, close]);
}

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