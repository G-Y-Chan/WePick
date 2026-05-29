import React, { useEffect, useMemo, useRef, useState } from "react";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { Button, StyleSheet, Text, View } from "react-native";
import { router, useLocalSearchParams } from "expo-router";
import { startRoom } from "../services/api/http";
import { connectRoomWS } from "../services/ws";

export default function Room() {
  const { roomCode, host } = useLocalSearchParams();
  const isHost = host === "true";

  const stringCode = useMemo(() => {
    if (Array.isArray(roomCode)) return roomCode.join("");
    return String(roomCode ?? "");
  }, [roomCode]);

  const [wsStatus, setWsStatus] = useState<
    "idle" | "connecting" | "connected" | "error"
  >("idle");

  const [isStarting, setIsStarting] = useState(false);

  const wsRef = useRef<ReturnType<typeof connectRoomWS> | null>(null);

  // Host and guest both connect to websocket and wait for START
  useEffect(() => {
    if (!stringCode) return;

    setWsStatus("connecting");

    const conn = connectRoomWS(stringCode, {
      onOpen: () => {
        setWsStatus("connected");
      },

      onMessage: (msg) => {
        if (msg.Header === "START") {
          conn.close();

          router.push({
            pathname: "/swipe",
            params: { roomCode: stringCode },
          });

          return;
        }

        if (msg.Header === "ERROR") {
          conn.close();

          router.push({
            pathname: "/error",
            params: {
              errorMessage: msg.Body || "Server error",
            },
          });
        }
      },

      onError: () => {
        setWsStatus("error");

        router.push({
          pathname: "/error",
          params: {
            errorMessage: "WebSocket connection failed",
          },
        });
      },

      onClose: () => {
        // Optional: update UI if needed
      },
    });

    wsRef.current = conn;

    return () => {
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [stringCode]);

  // Host sends the "start room" command via HTTP,
  // but navigation happens only after WS receives START.
  const handleStartRoom = async () => {
    if (!stringCode) {
      router.push({
        pathname: "/error",
        params: { errorMessage: "Missing room code" },
      });
      return;
    }

    if (wsStatus !== "connected") {
      router.push({
        pathname: "/error",
        params: { errorMessage: "WebSocket is not connected yet" },
      });
      return;
    }

    try {
      setIsStarting(true);

      const response = await startRoom(stringCode);
      const status = response === "true";

      if (!status) {
        router.push({
          pathname: "/error",
          params: { errorMessage: "Unable to start room" },
        });
      }

      // Do not router.push("/swipe") here.
      // Wait for the backend to broadcast START over WebSocket.
    } catch (e: unknown) {
      let message = "Internal Server Error";
      if (e instanceof Error) message = e.message;

      router.push({
        pathname: "/error",
        params: { errorMessage: message },
      });
    } finally {
      setIsStarting(false);
    }
  };

  const wsStatusText =
    wsStatus === "connecting"
      ? "Connecting..."
      : wsStatus === "connected"
      ? "Connected"
      : wsStatus === "error"
      ? "Connection error"
      : "";

  return (
    <SafeAreaProvider>
      <View style={styles.container}>
        {isHost ? (
          <>
            <Button
              title={isStarting ? "Starting..." : "Start"}
              onPress={handleStartRoom}
              disabled={isStarting || wsStatus !== "connected"}
            />

            <Text style={styles.text}>Code to join: {stringCode}</Text>
            <Text style={styles.text}>{wsStatusText}</Text>
          </>
        ) : (
          <>
            <Text style={styles.text}>Waiting for host...</Text>
            <Text style={styles.text}>Code to join: {stringCode}</Text>
            <Text style={styles.text}>{wsStatusText}</Text>
          </>
        )}
      </View>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: 16,
  },
  text: {
    fontSize: 16,
    textAlign: "center",
    marginTop: 8,
  },
});