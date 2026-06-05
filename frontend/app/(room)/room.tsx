import React, { useEffect, useState } from "react";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { Button, StyleSheet, Text, View } from "react-native";
import { router, useLocalSearchParams } from "expo-router";

import { startRoom } from "../../services/api/http";
import { useRoom } from "@/src/context/room-context";

export default function Room() {
  // Only grab 'host' from params; the room code is managed by the Layout/Context
  const { host } = useLocalSearchParams();
  const isHost = host === "true";

  // Pull the already-parsed roomCode from Context
  const { status, lastMessage, roomCode } = useRoom();

  const [isStarting, setIsStarting] = useState(false);

  useEffect(() => {
    if (!lastMessage) return;

    if (lastMessage.Header === "START") {
      // No need to pass params; SwipeScreen is inside the RoomProvider
      router.push("/swipe");
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
  }, [lastMessage]);

  const handleStartRoom = async () => {
    if (!roomCode) {
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
      const response = await startRoom(roomCode);
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
            <Text style={styles.text}>Code to join: {roomCode}</Text>
            <Text style={styles.text}>{wsStatusText}</Text>
          </>
        ) : (
          <>
            <Text style={styles.text}>Waiting for host...</Text>
            <Text style={styles.text}>Code to join: {roomCode}</Text>
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