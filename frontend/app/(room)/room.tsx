import React, { useEffect, useMemo, useState } from "react";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { Button, StyleSheet, Text, View } from "react-native";
import { router, useLocalSearchParams } from "expo-router";

import { startRoom } from "../../services/api/http";
import { useRoom } from "@/src/context/room-context";

export default function Room() {
  const { roomCode, host } = useLocalSearchParams();

  const isHost = host === "true";

  const stringCode = useMemo(() => {
    if (Array.isArray(roomCode)) return roomCode.join("");
    return String(roomCode ?? "");
  }, [roomCode]);

  const { status, lastMessage } = useRoom();

  const [isStarting, setIsStarting] = useState(false);

  useEffect(() => {
    if (!lastMessage) return;

    if (lastMessage.Header === "START") {
      router.push({
        pathname: "/swipe",
        params: { roomCode: stringCode },
      });

      return;
    }

    if (lastMessage.Header === "ERROR") {
      router.push({
        pathname: "/error",
        params: {
          errorMessage: lastMessage.Body || "Server error",
        },
      });
    }
  }, [lastMessage, stringCode]);

  const handleStartRoom = async () => {
    if (!stringCode) {
      router.push({
        pathname: "/error",
        params: { errorMessage: "Missing room code" },
      });

      return;
    }

    if (status !== "connected") {
      router.push({
        pathname: "/error",
        params: {
          errorMessage: "WebSocket is not connected yet",
        },
      });

      return;
    }

    try {
      setIsStarting(true);

      const response = await startRoom(stringCode);

      const ok = response === "true";

      if (!ok) {
        router.push({
          pathname: "/error",
          params: {
            errorMessage: "Unable to start room",
          },
        });
      }
    } catch (e: unknown) {
      let message = "Internal Server Error";

      if (e instanceof Error) {
        message = e.message;
      }

      router.push({
        pathname: "/error",
        params: { errorMessage: message },
      });
    } finally {
      setIsStarting(false);
    }
  };

  const wsStatusText =
    status === "connecting"
      ? "Connecting..."
      : status === "connected"
      ? "Connected"
      : status === "error"
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
              disabled={isStarting || status !== "connected"}
            />

            <Text style={styles.text}>
              Code to join: {stringCode}
            </Text>

            <Text style={styles.text}>
              {wsStatusText}
            </Text>
          </>
        ) : (
          <>
            <Text style={styles.text}>
              Waiting for host...
            </Text>

            <Text style={styles.text}>
              Code to join: {stringCode}
            </Text>

            <Text style={styles.text}>
              {wsStatusText}
            </Text>
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
