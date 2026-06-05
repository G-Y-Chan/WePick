import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";

import { connectRoomWS } from "@/services/ws";
import { Message } from "@/services/types";

type WSStatus =
  | "idle"
  | "connecting"
  | "connected"
  | "error";

type RoomContextType = {
  roomCode: string;
  status: WSStatus;
  lastMessage: Message | null;
  send: (msg: Message) => void;
  disconnect: () => void;
};

const RoomContext = createContext<RoomContextType | null>(null);

type Props = {
  roomCode: string;
  children: React.ReactNode;
};

export function RoomProvider({ roomCode, children }: Props) {
  const [status, setStatus] = useState<WSStatus>("idle");
  const [lastMessage, setLastMessage] = useState<Message | null>(null);

  const wsRef = useRef<ReturnType<typeof connectRoomWS> | null>(null);

  useEffect(() => {
    if (!roomCode) return;

    setStatus("connecting");

    const conn = connectRoomWS(roomCode, {
      onOpen: () => {
        setStatus("connected");
      },

      onMessage: (msg) => {
        setLastMessage(msg);
      },

      onError: () => {
        setStatus("error");
      },

      onClose: () => {
        setStatus("idle");
      },
    });

    wsRef.current = conn;

    return () => {
      conn.close();
      wsRef.current = null;
    };
  }, [roomCode]);

  const send = useCallback((msg: Message) => {
    wsRef.current?.send(msg);
  }, []);

  const disconnect = useCallback(() => {
    wsRef.current?.close();
    wsRef.current = null;
  }, []);

  return (
    <RoomContext.Provider
      value={{
        roomCode,
        status,
        lastMessage,
        send,
        disconnect,
      }}
    >
      {children}
    </RoomContext.Provider>
  );
}

export function useRoom() {
  const ctx = useContext(RoomContext);

  if (!ctx) {
    throw new Error("useRoom must be used inside RoomProvider");
  }

  return ctx;
}
