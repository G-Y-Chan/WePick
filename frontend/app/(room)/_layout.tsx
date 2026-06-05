import React, { useMemo } from "react";
import { Stack, useLocalSearchParams } from "expo-router";

import { RoomProvider } from "@/src/context/room-context";

export default function RoomLayout() {
  const { roomCode } = useLocalSearchParams();

  const stringCode = useMemo(() => {
    if (Array.isArray(roomCode)) {
      return roomCode.join("");
    }
    return String(roomCode ?? "");
  }, [roomCode]);

  if (!stringCode) return null;

  return (
    <RoomProvider roomCode={stringCode}>
      <Stack screenOptions={{ headerShown: false }} />
    </RoomProvider>
  );
}
