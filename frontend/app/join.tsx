import React, { useState } from "react";
import { SafeAreaProvider } from "react-native-safe-area-context";
import {
  ActivityIndicator,
  Button,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { router } from "expo-router";
import { verifyRoomCode } from "../services/api/room";

export default function Index() {
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [text, onChangeText] = useState('');

  const handleJoinRoom = async (roomCode: number) => {
    try {
      setError(null);
      setLoading(true);
      const response = await verifyRoomCode(roomCode);
      console.log("room code:", roomCode);
      const status: boolean = response.toLowerCase() === 'true';
      setLoading(false);
      if (status) {
        router.push({
          pathname: "/room",
          params: { roomCode: roomCode },
        });
      } else {
        const message = "Invalid Room Code"
        router.push({
          pathname: "/error",
          params: {errorMessage: message },
        })
      }
    } catch (e: unknown) {
      console.error("Error in Joining Room:", e);
      const message = "Internal Server Error"
        router.push({
          pathname: "/error",
          params: {errorMessage: message },
        })
    }
  }

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
            <TextInput
                style={styles.input}
                onChangeText={onChangeText}
                value={text}
                placeholder="Enter room Code"
            />
            <Button title="Join Room" onPress={async () => {
                var code = parseInt(text)
                console.log(code)
                await handleJoinRoom(code)
            }} />
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
  input: {
    height: 40,
    width: 200,
    margin: 12,
    borderWidth: 1,
    padding: 10,
  },
});
