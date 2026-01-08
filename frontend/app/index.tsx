import React, { useState } from "react";
import { SafeAreaProvider } from "react-native-safe-area-context";
import {
  ActivityIndicator,
  Button,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { router } from "expo-router";
import { getRoomCode, verifyRoomCode } from "../services/api/room";

const sleep = (ms: number) =>
  new Promise<void>(resolve => setTimeout(resolve, ms));

export default function Index() {
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleCreateRoom = async () => {
    //console.log("Create Room pressed");
    try {
      setError(null);
      setLoading(true);
      const code = await getRoomCode();
      console.log("room code:", code);

      router.push({
        pathname: "/room",
        params: { 
          roomCode: code,
          host: "true"
        },
      });
    } catch (e: unknown) {
      console.error("Error in Creating Room:", e);
      const message = "Internal Server Error"
        router.push({
          pathname: "/error",
          params: {errorMessage: message },
        })
    } finally {
      setLoading(false);
    }
  };

  const handleJoinRoom = async () => {
    router.push({
      pathname: "/join",
    });
  };

  return (
    <SafeAreaProvider>
      <View style={styles.container}>
        {loading ? (
          <>
            <Text>Verifying...{"\n"}</Text>
            <ActivityIndicator />
          </>
        ) : error ? (
          <>
            <Text>{"Message:\n"}</Text>
            <Text>Error: {error}</Text>
          </>
        ) : (
          <>
            <Button title="Create Room" onPress={handleCreateRoom} />
            <Button title="Join Room" onPress={handleJoinRoom} />
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
});
