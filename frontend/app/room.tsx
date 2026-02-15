import React, { useEffect, useMemo, useRef, useState } from "react";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { Button, StyleSheet, Text, View } from "react-native";
import { router, useLocalSearchParams } from "expo-router";
import { startRoom } from "../services/api/http";
import { connectRoomWS } from "../services/ws"; // <-- your helper

export default function Room() {
  const { roomCode, host } = useLocalSearchParams();
  const isHost = host === "true";

  const stringCode = useMemo(() => {
    if (Array.isArray(roomCode)) return roomCode.join("");
    return String(roomCode ?? "");
  }, [roomCode]);

  const [wsStatus, setWsStatus] = useState<"idle" | "connecting" | "connected" | "error">("idle");
  const wsRef = useRef<ReturnType<typeof connectRoomWS> | null>(null);

  // Guest: connect websocket and wait for START
  useEffect(() => {
    if (isHost) return;
    if (!stringCode) return;

    setWsStatus("connecting");

    const conn = connectRoomWS(stringCode, {
      onOpen: () => setWsStatus("connected"),
      onMessage: (msg) => {
        if (msg.Header === "START") {
          conn.close();
          router.push({ 
            pathname: "/swipe",
            params: { roomCode: stringCode } });
          return;
        }

        if (msg.Header === "ERROR") {
          conn.close();
          router.push({
            pathname: "/error",
            params: { errorMessage: msg.Body || "Server error" },
          });
        }
      },
      onError: () => {
        setWsStatus("error");
        router.push({
          pathname: "/error",
          params: { errorMessage: "WebSocket connection failed" },
        });
      },
      onClose: () => {
        // optional: update UI
      },
    });

    wsRef.current = conn;

    return () => {
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [isHost, stringCode]);

  // Host: start room via HTTP
  const handleStartRoom = async () => {
    try {
      const response = await startRoom(stringCode);
      const status = response === "true";
      if (status) {
        router.push({ pathname: "/swipe" });
      } else {
        router.push({
          pathname: "/error",
          params: { errorMessage: "Unable to Start" },
        });
      }
    } catch (e: unknown) {
      let message = "Internal Server Error";
      if (e instanceof Error) message = e.message;
      router.push({ pathname: "/error", params: { errorMessage: message } });
    }
  };

  return (
    <SafeAreaProvider>
      <View style={styles.container}>
        {isHost ? (
          <>
            <Button title="Start" onPress={handleStartRoom} />
            <Text style={styles.text}>Code to join: {stringCode}</Text>
          </>
        ) : (
          <>
            <Text>Waiting for host...</Text>
            <Text style={styles.text}>Code to join: {stringCode}</Text>
            <Text style={styles.text}>
              {wsStatus === "connecting" ? "Connecting..." :
               wsStatus === "connected" ? "Connected" :
               wsStatus === "error" ? "Connection error" : ""}
            </Text>
          </>
        )}
      </View>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, justifyContent: "center", alignItems: "center", paddingHorizontal: 16 },
  text: { fontSize: 16, textAlign: "center" },
});
