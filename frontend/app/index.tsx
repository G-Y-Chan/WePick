import React, { useEffect, useState } from "react";
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
import { getRoomCode } from "../services/api/room";

function goToRoom(code: string) {
  router.push({
    pathname: "/room",
    params: { code },
  });
}

export default function Index() {
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  /*
  if (loading) {
    return (
      <SafeAreaProvider>
        <View style={styles.container}>
          <Text>Verifying...{`\n`}</Text>
          <ActivityIndicator />
        </View>
      </SafeAreaProvider>
    );
  }
  */

  if (error) {
    return (
      <SafeAreaProvider>
        <View style={styles.container}>
          <Text>{'Message: \n'}</Text>
          <Text>Error: {error}</Text>
        </View>
      </SafeAreaProvider>
    );
  }

  return (
    <SafeAreaProvider>
      <View style={styles.container}>
        <Button
          onPress={async () => {
            const code = await getRoomCode();
            console.log("room code:", code);
            router.push({
              pathname: "/room",
              params: { 
                roomCode: code 
              },
            });
          }}
          title="Create Room"
        />
        <Button
          title="Join Room"
        />
        <Button
          title="Test Create Room"
        />
      </View>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,                 // take full screen
    justifyContent: "center", // center vertically
    alignItems: "center",     // center horizontally
    paddingHorizontal: 16,
  },
  title: {
    fontSize: 18,
    marginBottom: 8,
    textAlign: "center",
    fontWeight: "600",
  },
  text: {
    fontSize: 16,
    textAlign: "center",
  },
  input: {
    borderWidth: 1,
    borderColor: "#ccc",
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 16,
  },
});